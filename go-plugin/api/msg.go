package api

//go:generate msgp

type UUID [16]byte

type ProviderArg struct {
	Id         uint32 `msg:"id"`
	CallerHost UUID   `msg:"caller_host"`
	CallerId   uint32 `msg:"caller_id"`
	Data       []byte `msg:"data"`
}
