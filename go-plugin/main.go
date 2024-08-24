package main

import (
	"github.com/azurity/mini-sos/go-plugin/api"
	"github.com/azurity/mini-sos/go-plugin/service/capability"
)

//go:generate msgp
type ConsoleData []byte

//go:export _sos_entry
func entry() int32 {
	api.CallService[ConsoleData]("/debug-console/stdout", capability.CapCall, ConsoleData("Hello,world!"))
	return 0
}

func main() {}
