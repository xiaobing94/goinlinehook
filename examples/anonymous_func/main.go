package main

import (
	"fmt"
	"github.com/xiaobing94/goinlinehook"
)

var oldFunc func (int, int) int

//go:noinline
func MyAdd(a, b int) int {
	return oldFunc(a, b) + 10
}

func main() {
	f := func (a, b int) int {
		return a+b
	}
	res := f(2, 3)
	// result: 5
	fmt.Println("result:", res)
	hookItem, err := goinlinehook.NewAndHook(f, MyAdd, &oldFunc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// hooked result: 15
	fmt.Println("hooked result:", f(2,3))
	// old result: 5
	fmt.Println("old result:", oldFunc(2,3))
	hookItem.UnHook()
	// UnHooked result: 5
	fmt.Println("UnHooked result:", f(2,3))
}