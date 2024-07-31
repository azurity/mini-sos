package main

import (
	"log"
	"os"

	"github.com/azurity/mini-sos/go-core/node"
	"github.com/azurity/mini-sos/go-core/workspace"
	"github.com/google/uuid"
)

const programFile = "../../go-plugin/plugin.wasm"

func main() {
	network, err := node.NewManager()
	if err != nil {
		log.Panicln(err)
	}
	log.Println("node id:", network.HostId)
	wsGroup := workspace.NewManager(network)
	baseWs, err := wsGroup.NewWorkspace(uuid.Nil)
	if err != nil {
		log.Panicln(err)
	}
	defer baseWs.Close()
	log.Println("create workspace:", baseWs.Id())

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
	log.Println("process exit:", proc.ExitCode())
}
