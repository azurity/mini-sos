package structs

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Error) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "desc":
			z.Desc, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Desc")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Error) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "desc"
	err = en.Append(0x81, 0xa4, 0x64, 0x65, 0x73, 0x63)
	if err != nil {
		return
	}
	err = en.WriteString(z.Desc)
	if err != nil {
		err = msgp.WrapError(err, "Desc")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Error) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "desc"
	o = append(o, 0x81, 0xa4, 0x64, 0x65, 0x73, 0x63)
	o = msgp.AppendString(o, z.Desc)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Error) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "desc":
			z.Desc, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Desc")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Error) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Desc)
	return
}
