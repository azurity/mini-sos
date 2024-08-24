package structs

import "errors"

//go:generate msgp

type Error struct {
	Desc string `msg:"desc"`
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
