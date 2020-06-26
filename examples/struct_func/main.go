package main

import (
	"fmt"
	_ "unsafe"
	"github.com/xiaobing94/goinlinehook"
)

type AddStruct struct {
	A int
	B int
}

func (as *AddStruct) Add() int {
	return as.A + as.B
}

//go:noinline
func MyAdd(ptr uintptr) int {
	var oldFunc func (ptr uintptr) int
	goinlinehook.GetOldFunc((*AddStruct).Add, &oldFunc)
	return oldFunc(ptr) + 10
}

func main() {
	var oldFunc func (ptr uintptr) int
	as := &AddStruct{
		A: 2,
		B: 3,
	}
	res := as.Add()
	// result: 5
	fmt.Println("result:", res)
	hookItem, err := goinlinehook.NewAndHook((*AddStruct).Add, MyAdd, &oldFunc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// hooked result: 15
	fmt.Println("hooked result:", as.Add())
	// old result: 5
	hookItem.UnHook()
	// UnHooked result: 5
	fmt.Println("UnHooked result:", as.Add())
}