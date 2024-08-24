package structs

import "errors"

type Error struct {
	Desc string `msgpack:"desc"`
}

func NewErr(err error) *Error {
	return &Error{
		Desc: err.Error(),
	}
}

func NewSuccess() *Error {
	return &Error{Desc: ""}
}

func (err *Error) Err() error {
	if err.Desc == "" {
		return nil
	}
	return errors.New(err.Desc)
}
