package service

import (
	"github.com/azurity/mini-sos/go-plugin/api"
	"github.com/azurity/mini-sos/go-plugin/service/capability"
	"github.com/azurity/mini-sos/go-plugin/service/structs"
	"github.com/google/uuid"
)

type Void struct{}

func Register(path string, local bool) error {
	res, err := api.CallService[structs.RegisterRes]("/service/register", capability.CapCall, &structs.RegisterReq{
		Path:  path,
		Local: local,
	})
	if err != nil {
		return err
	}
	return res.Err()
}

func Unregister(path string) error {
	arg := structs.UnregisterReq(path)
	res, err := api.CallService[structs.UnregisterRes]("/service/unregister", capability.CapCall, &arg)
	if err != nil {
		return err
	}
	return res.Err()
}

func AddCap(path string, cap uuid.UUID, provider uint32) error {
	res, err := api.CallService[structs.AddCapRes]("/service/cap/add", capability.CapCall, &structs.AddCapReq{
		Path: path,
		Desc: structs.Cap{
			Name:     api.UUID(cap),
			Host:     api.UUID(uuid.Nil),
			Process:  0,
			Provider: provider,
		},
	})
	if err != nil {
		return err
	}
	return res.Err()
}

func DelCap(path string, cap uuid.UUID) error {
	res, err := api.CallService[structs.DelCapRes]("/service/cap/del", capability.CapCall, &structs.DelCapReq{
		Path: path,
		Cap:  api.UUID(cap),
	})
	if err != nil {
		return err
	}
	return res.Err()
}

func ListCap(path string, cap uuid.UUID) ([]uuid.UUID, error) {
	res, err := api.CallService[structs.ListCapRes]("/service/cap/list", capability.CapCall, &structs.ListCapReq{
		Path: path,
	})
	if err != nil {
		return nil, err
	}
	if res.Error.Err() != nil {
		return nil, res.Error.Err()
	}
	ret := []uuid.UUID{}
	for _, cap := range res.Caps {
		ret = append(ret, uuid.UUID(cap.Name))
	}
	return ret, nil
}
