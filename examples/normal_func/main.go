package main

import (
	"fmt"
	"github.com/xiaobing94/goinlinehook"
)

//go:noinline
func MyAdd(a, b int) int {
	var oldFunc func (int, int) int
	goinlinehook.GetOldFunc(Add, &oldFunc)
	return oldFunc(a, b) + 10
}

//go:noinline
func Add(a, b int) int {
	return a+b
}

func main() {
	var oldFunc func (int, int) int
	res := Add(2, 3)
	// result: 5
	fmt.Println("result:", res)
	hookItem, err := goinlinehook.NewAndHook(Add, MyAdd, &oldFunc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// hooked result: 15
	fmt.Println("hooked result:", Add(2,3))
	// old result: 5
	fmt.Println("old result:", oldFunc(2,3))
	hookItem.UnHook()
	// UnHooked result: 5
	fmt.Println("UnHooked result:", Add(2,3))
}