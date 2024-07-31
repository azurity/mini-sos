package message

type Message struct {
	Slot     string
	Data     []byte
	Callback func(data []byte, err error)
}
