package process

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/azurity/mini-sos/go-core/message"
	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/service"
	extism "github.com/extism/go-sdk"
)

type Process interface {
	Host() node.HostID
	Id() uint32
	WaitQuit()
	Kill()
	Run() error
	ExitCode() uint32
}

type LocalProcess struct {
	host       node.HostID
	pid        uint32
	plugin     *extism.Plugin
	running    uint32
	exitCode   atomic.Uint32
	Messages   chan message.Message
	channelMap map[string]string
	quit       sync.WaitGroup
	man        *Manager
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
		host:       man.host,
		pid:        pid,
		plugin:     nil,
		running:    RunStateReady,
		exitCode:   atomic.Uint32{},
		Messages:   make(chan message.Message),
		channelMap: map[string]string{},
		quit:       sync.WaitGroup{},
		man:        man,
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
	proc.plugin.Close()
}

func (proc *LocalProcess) ExitCode() uint32 {
	return proc.exitCode.Load()
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
		[]extism.ValueType{extism.ValueTypePTR},
	)
	return []extism.HostFunction{callService}
}

func (proc *LocalProcess) execute(fn string, data []byte) ([]byte, error) {
	exit, result, err := proc.plugin.Call(fn, data)
	if exit != 0 {
		proc.exitCode.CompareAndSwap(0, exit)
	}
	return result, err
}

func (proc *LocalProcess) messageChannel() {
	for {
		msg := <-proc.Messages
		if msg.Slot == fmt.Sprintf("process/%d/signal", proc.pid) {
			if len(msg.Data) == 4 && binary.LittleEndian.Uint32(msg.Data) == SignalKill {
				proc.plugin.Close()
				proc.running = RunStateExit
				break
			}
		}
		if fn, ok := proc.channelMap[msg.Slot]; ok {
			ret, err := proc.execute(fn, msg.Data)
			msg.Callback(ret, err)
		}
		if msg.Slot == fmt.Sprintf("process/%d/signal", proc.pid) {
			if len(msg.Data) == 4 && binary.LittleEndian.Uint32(msg.Data) == SignalExit {
				break
			}
		}
	}
	proc.quit.Done()
	proc.man.release(proc.pid)
}

func (proc *LocalProcess) Run() error {
	go proc.messageChannel()
	proc.running = RunStateRunning
	_, err := proc.execute("_sos_entry", nil)
	proc.running = RunStateExit
	return err
}

func (proc *LocalProcess) CallService(service string, data []byte, caller service.Provider) ([]byte, error) {
	fn, ok := proc.channelMap[service]
	if !ok {
		return nil, ErrNotFound
	}
	return proc.execute(fn, data)
}
