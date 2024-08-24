package node

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrInvalidMetadata = errors.New("invalid metadata")
var ErrInvalidTarget = errors.New("invalid target")
var ErrInvalidAction = errors.New("invalid action")

type HostID = uuid.UUID

type Node interface {
	Host() HostID
	WaitClose(ctx context.Context)
	Call(action string, arg []byte) ([]byte, error)
}

type nodeImpl struct {
	man       *Manager
	host      HostID
	routePath []HostID
	close     sync.WaitGroup
}

func (node *nodeImpl) Host() HostID {
	return node.host
}

func (node *nodeImpl) WaitClose(ctx context.Context) {
	c := make(chan bool)
	go func() {
		node.close.Wait()
		c <- true
	}()
	select {
	case <-ctx.Done():
		break
	case <-c:
		break
	}
}

func (node *nodeImpl) Call(action string, arg []byte) ([]byte, error) {
	metadata, _ := msgpack.Marshal(Entry{
		Index:  1,
		Route:  append([]HostID{node.man.HostId}, node.routePath...),
		Action: action,
	})
	data, err := node.man.Connections[node.routePath[0]].socket.RequestResponse(payload.New([]byte{}, metadata)).Block(context.Background())
	if err != nil {
		return nil, err
	}
	return data.Data(), nil
}

type Connection struct {
	man    *Manager
	host   HostID
	socket rsocket.CloseableRSocket
	routes []HostID
}

func (conn *Connection) Close() {
	conn.socket.Close()
	delete(conn.man.Connections, conn.host)
	conn.man.updateNodes()
}

type Manager struct {
	HostId      HostID
	serveCloser func()
	Connections map[HostID]*Connection
	Nodes       map[HostID]*nodeImpl
	Callback    map[string]func([]byte, HostID) ([]byte, error)
}

type Entry struct {
	Index  int      `msgpack:"index"`
	Route  []HostID `msgpack:"route"`
	Action string   `msgpack:"action"`
}

type RouteInfo struct {
	Host        HostID   `msgpack:"host"`
	Connections []HostID `msgpack:"connections"`
}

func NewManager() (*Manager, error) {
	host, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	man := &Manager{
		HostId:      host,
		serveCloser: nil,
		Connections: map[HostID]*Connection{},
		Nodes:       map[HostID]*nodeImpl{},
		Callback:    map[string]func([]byte, HostID) ([]byte, error){},
	}
	man.Callback["info"] = func(arg []byte, caller HostID) ([]byte, error) {
		return msgpack.Marshal(man.makeRouteInfo())
	}
	man.Callback["update"] = func(arg []byte, caller HostID) ([]byte, error) {
		man.updateNodes()
		return []byte{}, nil
	}
	return man, nil
}

func (man *Manager) Close() {
	for _, conn := range man.Connections {
		conn.socket.Close()
	}
	man.Connections = map[HostID]*Connection{}
	if man.serveCloser != nil {
		man.serveCloser()
	}
	man.serveCloser = nil
}

func (man *Manager) makeRouteInfo() RouteInfo {
	connections := []HostID{}
	for key := range man.Connections {
		connections = append(connections, key)
	}
	return RouteInfo{
		Host:        man.HostId,
		Connections: connections,
	}
}

func (man *Manager) createAbstractSocket() rsocket.RSocket {
	return rsocket.NewAbstractSocket(
		rsocket.RequestResponse(
			func(request payload.Payload) (response mono.Mono) {
				entryRaw, ok := request.Metadata()
				if !ok {
					return mono.Error(ErrInvalidMetadata)
				}
				entry := Entry{}
				err := msgpack.Unmarshal(entryRaw, &entryRaw)
				if err != nil {
					return mono.Error(err)
				}
				if entry.Route[entry.Index] != man.HostId && entry.Route[entry.Index] != uuid.Nil {
					return mono.Error(ErrInvalidTarget)
				}
				data := request.Data()
				if entry.Index+1 < len(entry.Route) {
					// redirect call
					next := entry.Route[entry.Index+1]
					conn, ok := man.Connections[next]
					if !ok {
						return mono.Error(ErrInvalidTarget)
					}
					entry.Index += 1
					metadata, err := msgpack.Marshal(entry)
					if err != nil {
						return mono.Error(err)
					}
					return conn.socket.RequestResponse(payload.New(data, metadata))
				}
				fn, ok := man.Callback[entry.Action]
				if !ok {
					return mono.Error(ErrInvalidAction)
				}
				ret, err := fn(data, entry.Route[0])
				if err != nil {
					return mono.Error(err)
				}
				return mono.Just(payload.New(ret, nil))
			},
		),
	)
}

func (man *Manager) StartServe(transporter transport.ServerTransporter) error {
	ctx, closer := context.WithCancel(context.Background())
	man.serveCloser = closer
	return rsocket.Receive().Acceptor(
		func(ctx context.Context, setup payload.SetupPayload, socket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
			initInfo := RouteInfo{}
			err := msgpack.Unmarshal(setup.Data(), &initInfo)
			if err != nil {
				return nil, err
			}
			connections := []HostID{}
			for _, item := range initInfo.Connections {
				if item != man.HostId {
					connections = append(connections, item)
				}
			}
			man.Connections[initInfo.Host] = &Connection{
				man:    man,
				host:   initInfo.Host,
				socket: socket,
				routes: connections,
			}
			return man.createAbstractSocket(), nil
		},
	).Transport(transporter).Serve(ctx)
}

func (man *Manager) StartClient(transport transport.ClientTransporter) error {
	setupData, err := msgpack.Marshal(man.makeRouteInfo())
	if err != nil {
		return err
	}
	initChan := make(chan *RouteInfo, 1)
	client, err := rsocket.Connect().SetupPayload(payload.New(setupData, nil)).Acceptor(
		func(ctx context.Context, socket rsocket.RSocket) rsocket.RSocket {
			infoMetadata, _ := msgpack.Marshal(Entry{
				Index:  1,
				Route:  []HostID{man.HostId, uuid.Nil},
				Action: "info",
			})
			socket.RequestResponse(payload.New([]byte{}, infoMetadata)).DoOnSuccess(func(res payload.Payload) error {
				initInfo := RouteInfo{}
				err := msgpack.Unmarshal(res.Data(), &initInfo)
				if err == nil {
					initChan <- &initInfo
				}
				return err
			}).DoOnError(func(e error) {
				log.Println(err)
				initChan <- nil
			})
			return man.createAbstractSocket()
		},
	).Transport(transport).Start(context.Background())
	if err != nil {
		return err
	}
	initInfo := <-initChan
	if initChan == nil {
		client.Close()
	} else {
		connections := []HostID{}
		for _, item := range initInfo.Connections {
			if item != man.HostId {
				connections = append(connections, item)
			}
		}
		man.Connections[initInfo.Host] = &Connection{
			man:    man,
			host:   initInfo.Host,
			socket: client,
			routes: connections,
		}
	}
	man.updateNodes()
	// update total network
	for _, node := range man.Nodes {
		node.Call("update", []byte{})
	}
	return nil
}

func (man *Manager) updateNodes() {
	paths := map[HostID][]HostID{}
	targets := map[HostID][]HostID{}
	for key, conn := range man.Connections {
		paths[key] = []HostID{key}
		targets[key] = conn.routes
	}
	newer := true
	for newer {
		newList := map[HostID][]HostID{}
		for it, path := range paths {
			if _, ok := targets[it]; !ok {
				infoMetadata, _ := msgpack.Marshal(Entry{
					Index:  1,
					Route:  append([]HostID{man.HostId}, path...),
					Action: "info",
				})
				data, err := man.Connections[path[0]].socket.RequestResponse(payload.New([]byte{}, infoMetadata)).Block(context.Background())
				if err != nil {
					continue
				}
				info := RouteInfo{}
				err = msgpack.Unmarshal(data.Data(), &info)
				if err != nil {
					continue
				}
				targets[it] = info.Connections
			}
			extra := targets[it]
			for _, item := range extra {
				if _, ok := paths[item]; !ok {
					newList[item] = append(append([]HostID{}, path...), item)
				}
			}
		}
		for k, v := range newList {
			paths[k] = v
		}
		newer = len(newList) > 0
	}
	for it, path := range paths {
		if node, ok := man.Nodes[it]; ok {
			node.routePath = path
			continue
		}
		man.Nodes[it] = &nodeImpl{
			man:       man,
			host:      it,
			routePath: path,
			close:     sync.WaitGroup{},
		}
		man.Nodes[it].close.Add(1)
	}
	removes := []HostID{}
	for it, node := range man.Nodes {
		if _, ok := paths[it]; !ok {
			removes = append(removes, it)
			node.close.Done()
		}
	}
	for _, it := range removes {
		delete(man.Nodes, it)
	}
}
