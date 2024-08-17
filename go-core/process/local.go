package process

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service"
	extism "github.com/extism/go-sdk"
	"github.com/vmihailenco/msgpack/v5"
)

type LocalProcess struct {
	host    node.HostID
	pid     uint32
	plugin  *extism.Plugin
	running uint32
	quit    sync.WaitGroup
	man     *Manager
}

func (man *Manager) NewLocalProcess(data []byte) (*LocalProcess, error) {
	pid, err := man.allocate()
	if err != nil {
		return nil, err
	}
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: data,
			},
		},
	}

	proc := &LocalProcess{
		host:    man.host,
		pid:     pid,
		plugin:  nil,
		running: RunStateReady,
		quit:    sync.WaitGroup{},
		man:     man,
	}

	ctx := context.Background()
	config := extism.PluginConfig{
		EnableWasi: true,
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, proc.createHostFunctions(man))
	if err != nil {
		man.release(pid)
		return nil, err
	}
	proc.plugin = plugin

	man.AliveProcess[pid] = proc
	proc.quit.Add(1)
	if man.CreateCallback != nil {
		man.CreateCallback(proc.pid)
	}
	return proc, nil
}

func (proc *LocalProcess) Host() node.HostID {
	return proc.host
}

func (proc *LocalProcess) Id() uint32 {
	return proc.pid
}

func (proc *LocalProcess) WaitQuit() {
	proc.quit.Wait()
}

func (proc *LocalProcess) Kill() {
	// TODO: maybe need a grace kill
	proc.plugin.Close()
}

func (proc *LocalProcess) createHostFunctions(man *Manager) []extism.HostFunction {
	callService := extism.NewHostFunctionWithStack(
		"_sos_call_service",
		func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
			serviceName, err := p.ReadString(stack[0])
			stack[0] = extism.EncodeI64(0)
			if err != nil {
				log.Println(err)
				return
			}
			var data []byte
			if stack[1] != 0 {
				data, err = p.ReadBytes(stack[1])
				if err != nil {
					log.Println(err)
					return
				}
			}
			ret, err := man.callFn(serviceName, data, proc)
			if err != nil {
				log.Println(err)
				return
			}
			stack[0], err = p.WriteBytes(ret)
			if err != nil {
				log.Println(err)
				stack[0] = extism.EncodeI64(0)
				return
			}
		},
		[]extism.ValueType{extism.ValueTypePTR, extism.ValueTypePTR},
		[]extism.ValueType{extism.ValueTypeI64},
	)
	return []extism.HostFunction{callService}
}

func (proc *LocalProcess) execute(fn string, data []byte) ([]byte, error) {
	exit, result, err := proc.plugin.Call(fn, data)
	if err != nil {
		return []byte{}, err
	}
	if exit != 0 {
		err = errors.New(proc.plugin.GetError())
	}
	return result, err
}

func (proc *LocalProcess) Run() error {
	proc.running = RunStateRunning
	_, err := proc.execute("_sos_entry", nil)
	proc.running = RunStateExit
	return err
}

type ProviderArg struct {
	Id   uint32 `msgpack:"id"`
	Data []byte `msgpack:"data"`
}

func (proc *LocalProcess) CallProvider(id uint32, data []byte, caller service.Provider) ([]byte, error) {
	raw, err := msgpack.Marshal(ProviderArg{Id: id, Data: data})
	if err != nil {
		return []byte{}, err
	}
	return proc.execute("_sos_call", raw)
}
