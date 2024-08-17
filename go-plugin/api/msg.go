package api

//go:generate msgp
type ProviderArg struct {
	Id   uint32 `msg:"id"`
	Data []byte `msg:"data"`
}
