package main

import (
	"context"
	"log"
	"os"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/process"
	"github.com/azurity/mini-sos/go-core/service/capability"
	"github.com/azurity/mini-sos/go-core/service/structs"
	"github.com/azurity/mini-sos/go-core/utils/tree"
	"github.com/azurity/mini-sos/go-core/workspace"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

const programFile = "../../go-plugin/plugin.wasm"

func Console(initChan chan bool) func(ctx context.Context, op *process.NativeOperator) func() error {
	return func(ctx context.Context, op *process.NativeOperator) func() error {
		return (func() error {
			transfer, _ := tree.NewBasicTree(tree.NewPath("/")).Transfer()
			arg, _ := msgpack.Marshal(structs.SetDirReq{
				Abs:    tree.NewPath("/"),
				Path:   tree.NewPath("debug-console"),
				Option: false,
				Item:   transfer,
			})
			op.Call("/", capability.CapSetDir, arg)
			stdoutId, _ := op.ProviderManager().New(func(data []byte, caller process.Process) ([]byte, error) {
				var bytes []byte
				_ = msgpack.Unmarshal(data, &bytes)
				os.Stdout.Write(bytes)
				return []byte{}, nil
			})
			arg, _ = msgpack.Marshal(structs.RegisterReq{
				Path:  "/debug-console/stdout",
				Local: false,
			})
			op.Call("/service/register", capability.CapCall, arg)
			arg, _ = msgpack.Marshal(structs.AddCapReq{
				Path: "/debug-console/stdout",
				Desc: structs.Cap{
					Name:     capability.CapCall,
					Host:     uuid.Nil,
					Process:  op.Pid(),
					Provider: stdoutId,
				},
			})
			op.Call("/service/cap/add", capability.CapCall, arg)
			stderrId, _ := op.ProviderManager().New(func(data []byte, caller process.Process) ([]byte, error) {
				var bytes []byte
				_ = msgpack.Unmarshal(data, &bytes)
				os.Stderr.Write(bytes)
				return []byte{}, nil
			})
			arg, _ = msgpack.Marshal(structs.RegisterReq{
				Path:  "/debug-console/stderr",
				Local: false,
			})
			op.Call("/service/register", capability.CapCall, arg)
			arg, _ = msgpack.Marshal(structs.AddCapReq{
				Path: "/debug-console/stderr",
				Desc: structs.Cap{
					Name:     capability.CapCall,
					Host:     uuid.Nil,
					Process:  op.Pid(),
					Provider: stderrId,
				},
			})
			op.Call("/service/cap/add", capability.CapCall, arg)
			initChan <- true
			<-ctx.Done()
			return nil
		})
	}
}

func main() {
	network, err := node.NewManager()
	if err != nil {
		log.Panicln(err)
	}
	defer network.Close()
	log.Println("node id:", network.HostId)
	wsGroup := workspace.NewManager(network)
	baseWs, err := wsGroup.NewWorkspace()
	if err != nil {
		log.Panicln(err)
	}
	defer baseWs.Close()
	log.Println("create workspace:", baseWs.Id())

	initChan := make(chan bool)
	console, _ := baseWs.ProcMan().NewNativeProcess(Console(initChan))
	go console.Run()
	<-initChan

	log.Println("start program:", programFile)
	program, _ := os.ReadFile(programFile)
	proc, err := baseWs.ProcMan().NewLocalProcess(program)
	if err != nil {
		log.Panicln(err)
	}
	err = proc.Run()
	if err != nil {
		log.Panicln(err)
	}
	log.Println("process exit")
}
