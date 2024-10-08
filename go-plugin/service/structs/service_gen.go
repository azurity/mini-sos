package structs

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *AddCapReq) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "path":
			z.Path, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Path")
				return
			}
		case "desc":
			err = z.Desc.DecodeMsg(dc)
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
func (z *AddCapReq) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "path"
	err = en.Append(0x82, 0xa4, 0x70, 0x61, 0x74, 0x68)
	if err != nil {
		return
	}
	err = en.WriteString(z.Path)
	if err != nil {
		err = msgp.WrapError(err, "Path")
		return
	}
	// write "desc"
	err = en.Append(0xa4, 0x64, 0x65, 0x73, 0x63)
	if err != nil {
		return
	}
	err = z.Desc.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Desc")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AddCapReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "path"
	o = append(o, 0x82, 0xa4, 0x70, 0x61, 0x74, 0x68)
	o = msgp.AppendString(o, z.Path)
	// string "desc"
	o = append(o, 0xa4, 0x64, 0x65, 0x73, 0x63)
	o, err = z.Desc.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Desc")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AddCapReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "path":
			z.Path, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Path")
				return
			}
		case "desc":
			bts, err = z.Desc.UnmarshalMsg(bts)
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
func (z *AddCapReq) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Path) + 5 + z.Desc.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *DelCapReq) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "path":
			z.Path, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Path")
				return
			}
		case "cap":
			err = z.Cap.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "Cap")
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
func (z *DelCapReq) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "path"
	err = en.Append(0x82, 0xa4, 0x70, 0x61, 0x74, 0x68)
	if err != nil {
		return
	}
	err = en.WriteString(z.Path)
	if err != nil {
		err = msgp.WrapError(err, "Path")
		return
	}
	// write "cap"
	err = en.Append(0xa3, 0x63, 0x61, 0x70)
	if err != nil {
		return
	}
	err = z.Cap.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Cap")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *DelCapReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "path"
	o = append(o, 0x82, 0xa4, 0x70, 0x61, 0x74, 0x68)
	o = msgp.AppendString(o, z.Path)
	// string "cap"
	o = append(o, 0xa3, 0x63, 0x61, 0x70)
	o, err = z.Cap.MarshalMsg(o)
	if err != nil {
		err = msgp.WrapError(err, "Cap")
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *DelCapReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "path":
			z.Path, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Path")
				return
			}
		case "cap":
			bts, err = z.Cap.UnmarshalMsg(bts)
			if err != nil {
				err = msgp.WrapError(err, "Cap")
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
func (z *DelCapReq) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Path) + 4 + z.Cap.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ListCapReq) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "path":
			z.Path, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Path")
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
func (z ListCapReq) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "path"
	err = en.Append(0x81, 0xa4, 0x70, 0x61, 0x74, 0x68)
	if err != nil {
		return
	}
	err = en.WriteString(z.Path)
	if err != nil {
		err = msgp.WrapError(err, "Path")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ListCapReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 1
	// string "path"
	o = append(o, 0x81, 0xa4, 0x70, 0x61, 0x74, 0x68)
	o = msgp.AppendString(o, z.Path)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ListCapReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "path":
			z.Path, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Path")
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
func (z ListCapReq) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Path)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ListCapRes) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "error":
			if dc.IsNil() {
				err = dc.ReadNil()
				if err != nil {
					err = msgp.WrapError(err, "Error")
					return
				}
				z.Error = nil
			} else {
				if z.Error == nil {
					z.Error = new(Error)
				}
				err = z.Error.DecodeMsg(dc)
				if err != nil {
					err = msgp.WrapError(err, "Error")
					return
				}
			}
		case "caps":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				err = msgp.WrapError(err, "Caps")
				return
			}
			if cap(z.Caps) >= int(zb0002) {
				z.Caps = (z.Caps)[:zb0002]
			} else {
				z.Caps = make([]Cap, zb0002)
			}
			for za0001 := range z.Caps {
				err = z.Caps[za0001].DecodeMsg(dc)
				if err != nil {
					err = msgp.WrapError(err, "Caps", za0001)
					return
				}
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
func (z *ListCapRes) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "error"
	err = en.Append(0x82, 0xa5, 0x65, 0x72, 0x72, 0x6f, 0x72)
	if err != nil {
		return
	}
	if z.Error == nil {
		err = en.WriteNil()
		if err != nil {
			return
		}
	} else {
		err = z.Error.EncodeMsg(en)
		if err != nil {
			err = msgp.WrapError(err, "Error")
			return
		}
	}
	// write "caps"
	err = en.Append(0xa4, 0x63, 0x61, 0x70, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Caps)))
	if err != nil {
		err = msgp.WrapError(err, "Caps")
		return
	}
	for za0001 := range z.Caps {
		err = z.Caps[za0001].EncodeMsg(en)
		if err != nil {
			err = msgp.WrapError(err, "Caps", za0001)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *ListCapRes) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "error"
	o = append(o, 0x82, 0xa5, 0x65, 0x72, 0x72, 0x6f, 0x72)
	if z.Error == nil {
		o = msgp.AppendNil(o)
	} else {
		o, err = z.Error.MarshalMsg(o)
		if err != nil {
			err = msgp.WrapError(err, "Error")
			return
		}
	}
	// string "caps"
	o = append(o, 0xa4, 0x63, 0x61, 0x70, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Caps)))
	for za0001 := range z.Caps {
		o, err = z.Caps[za0001].MarshalMsg(o)
		if err != nil {
			err = msgp.WrapError(err, "Caps", za0001)
			return
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ListCapRes) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "error":
			if msgp.IsNil(bts) {
				bts, err = msgp.ReadNilBytes(bts)
				if err != nil {
					return
				}
				z.Error = nil
			} else {
				if z.Error == nil {
					z.Error = new(Error)
				}
				bts, err = z.Error.UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "Error")
					return
				}
			}
		case "caps":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Caps")
				return
			}
			if cap(z.Caps) >= int(zb0002) {
				z.Caps = (z.Caps)[:zb0002]
			} else {
				z.Caps = make([]Cap, zb0002)
			}
			for za0001 := range z.Caps {
				bts, err = z.Caps[za0001].UnmarshalMsg(bts)
				if err != nil {
					err = msgp.WrapError(err, "Caps", za0001)
					return
				}
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
func (z *ListCapRes) Msgsize() (s int) {
	s = 1 + 6
	if z.Error == nil {
		s += msgp.NilSize
	} else {
		s += z.Error.Msgsize()
	}
	s += 5 + msgp.ArrayHeaderSize
	for za0001 := range z.Caps {
		s += z.Caps[za0001].Msgsize()
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ListReq) DecodeMsg(dc *msgp.Reader) (err error) {
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
func (z ListReq) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	_ = z
	err = en.Append(0x80)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ListReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	_ = z
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ListReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
func (z ListReq) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ListRes) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ListRes, zb0002)
	}
	for zb0001 := range *z {
		(*z)[zb0001], err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ListRes) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0003 := range z {
		err = en.WriteString(z[zb0003])
		if err != nil {
			err = msgp.WrapError(err, zb0003)
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ListRes) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendArrayHeader(o, uint32(len(z)))
	for zb0003 := range z {
		o = msgp.AppendString(o, z[zb0003])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ListRes) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0002 uint32
	zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ListRes, zb0002)
	}
	for zb0001 := range *z {
		(*z)[zb0001], bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err, zb0001)
			return
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ListRes) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0003 := range z {
		s += msgp.StringPrefixSize + len(z[zb0003])
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RegisterReq) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "path":
			z.Path, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Path")
				return
			}
		case "local":
			z.Local, err = dc.ReadBool()
			if err != nil {
				err = msgp.WrapError(err, "Local")
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
func (z RegisterReq) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "path"
	err = en.Append(0x82, 0xa4, 0x70, 0x61, 0x74, 0x68)
	if err != nil {
		return
	}
	err = en.WriteString(z.Path)
	if err != nil {
		err = msgp.WrapError(err, "Path")
		return
	}
	// write "local"
	err = en.Append(0xa5, 0x6c, 0x6f, 0x63, 0x61, 0x6c)
	if err != nil {
		return
	}
	err = en.WriteBool(z.Local)
	if err != nil {
		err = msgp.WrapError(err, "Local")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RegisterReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "path"
	o = append(o, 0x82, 0xa4, 0x70, 0x61, 0x74, 0x68)
	o = msgp.AppendString(o, z.Path)
	// string "local"
	o = append(o, 0xa5, 0x6c, 0x6f, 0x63, 0x61, 0x6c)
	o = msgp.AppendBool(o, z.Local)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RegisterReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "path":
			z.Path, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Path")
				return
			}
		case "local":
			z.Local, bts, err = msgp.ReadBoolBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Local")
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
func (z RegisterReq) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Path) + 6 + msgp.BoolSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *UnregisterReq) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 string
		zb0001, err = dc.ReadString()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = UnregisterReq(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z UnregisterReq) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString(string(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z UnregisterReq) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, string(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *UnregisterReq) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 string
		zb0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = UnregisterReq(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z UnregisterReq) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len(string(z))
	return
}
