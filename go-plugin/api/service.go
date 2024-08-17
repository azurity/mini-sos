package api

import (
	"errors"

	"github.com/extism/go-pdk"
	"github.com/tinylib/msgp/msgp"
)

//go:wasmimport extism:host/user _sos_call_service
func _sos_call_service(uint64, uint64) uint64

var ErrorCallFailed = errors.New("call failed")

type typeHolder[T any] interface {
	*T
	msgp.Unmarshaler
}

func CallService[T any, PT typeHolder[T]](service string, arg msgp.Marshaler) (PT, error) {
	servicePtr := pdk.AllocateString(service)
	defer servicePtr.Free()
	var dataPtr uint64 = 0
	if arg != nil {
		data, err := arg.MarshalMsg(nil)
		if err != nil {
			return nil, err
		}
		argPtr := pdk.AllocateBytes(data)
		defer argPtr.Free()
		dataPtr = argPtr.Offset()
	}
	retAddr := _sos_call_service(servicePtr.Offset(), dataPtr)
	if retAddr == 0 {
		return nil, ErrorCallFailed
	}
	retPtr := pdk.FindMemory(retAddr)
	defer retPtr.Free()
	bytes := retPtr.ReadBytes()
	var ret PT = new(T)
	_, err := ret.UnmarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	return ret, err
}
