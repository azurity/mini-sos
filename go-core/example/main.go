package main

import (
	"context"
	"log"
	"os"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/process"
	"github.com/azurity/mini-sos/go-core/service"
	"github.com/azurity/mini-sos/go-core/workspace"
	"github.com/vmihailenco/msgpack/v5"
)

const programFile = "../../go-plugin/plugin.wasm"

func InitConsole(ctx context.Context, op *process.NativeOperator) func() error {
	return (func() error {
		stdoutId, _ := op.ProviderManager().New(func(data []byte, caller service.Provider) ([]byte, error) {
			var bytes []byte
			_ = msgpack.Unmarshal(data, &bytes)
			os.Stdout.Write(bytes)
			return []byte{}, nil
		})
		arg, _ := msgpack.Marshal(process.RegisterArgs{
			Service:  "/debug-console/stdout",
			Provider: stdoutId,
		})
		op.Call("/service/register", arg)
		stderrId, _ := op.ProviderManager().New(func(data []byte, caller service.Provider) ([]byte, error) {
			var bytes []byte
			_ = msgpack.Unmarshal(data, &bytes)
			os.Stderr.Write(bytes)
			return []byte{}, nil
		})
		arg, _ = msgpack.Marshal(process.RegisterArgs{
			Service:  "/debug-console/stderr",
			Provider: stderrId,
		})
		op.Call("/service/register", arg)
		return nil
	})
}

func main() {
	network, err := node.NewManager()
	if err != nil {
		log.Panicln(err)
	}
	log.Println("node id:", network.HostId)
	wsGroup := workspace.NewManager(network)
	baseWs, err := wsGroup.NewWorkspace()
	if err != nil {
		log.Panicln(err)
	}
	defer baseWs.Close()
	log.Println("create workspace:", baseWs.Id())

	console, _ := baseWs.Local().NewNativeProcess(InitConsole)
	go console.Run()

	log.Println("start program:", programFile)
	program, _ := os.ReadFile(programFile)
	proc, err := baseWs.Local().NewProcess(program)
	if err != nil {
		log.Panicln(err)
	}
	err = proc.Run()
	if err != nil {
		log.Panicln(err)
	}
	log.Println("process exit")
}
