package private_struct

type AddStruct struct {
	A int
	B int
}

//go:noinline
func (as *AddStruct) add() int {
	return as.A + as.B
}

func (as *AddStruct) Add() int {
	return as.add()
}