package service

import "github.com/azurity/mini-sos/go-plugin/api"

//go:generate msgp

type RegisterArgs struct {
	Service  string `msg:"service"`
	Function string `msg:"function"`
}

type UnregisterArgs struct {
	Service string `msg:"service"`
}

type Void struct{}

type ServiceListRet []string

func Register(entry string, fn string) bool {
	_, err := api.CallService[Void]("/service/register", &RegisterArgs{
		Service:  entry,
		Function: fn,
	})
	return err != nil
}

func Unregister(entry string) bool {
	_, err := api.CallService[Void]("/service/unregister", &UnregisterArgs{
		Service: entry,
	})
	return err != nil
}

func List() []string {
	ret, err := api.CallService[ServiceListRet]("/service/list", nil)
	if err != nil {
		return nil
	}
	return *ret
}
