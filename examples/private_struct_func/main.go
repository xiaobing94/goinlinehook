package main

import (
	"fmt"
	_ "unsafe"
	"github.com/xiaobing94/goinlinehook/examples/private_struct_func/private_struct"
	"github.com/xiaobing94/goinlinehook"
)

//go:linkname Add github.com/xiaobing94/goinlinehook/examples/private_struct_func/private_struct.(*AddStruct).add
func Add(a uintptr) int

//go:noinline
func MyAdd(ptr uintptr) int {
	var oldFunc func (ptr uintptr) int
	goinlinehook.GetOldFunc(Add, &oldFunc)
	return oldFunc(ptr) + 10
}

func main() {
	var oldFunc func (ptr uintptr) int
	as := &private_struct.AddStruct{
		A: 4,
		B: 5,
	}
	res := as.Add()
	// result: 5
	fmt.Println("result:", res)
	hookItem, err := goinlinehook.NewAndHook(Add, MyAdd, &oldFunc)
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