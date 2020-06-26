package main

import (
	"fmt"
	"time"
	_ "unsafe"
	"github.com/xiaobing94/goinlinehook"
)

//go:linkname mynow time.now
func mynow() (sec int64, nsec int32, mono int64)

//go:nosplit
func Mynow() (sec int64, nsec int32, mono int64) {
	var oldFunc func () (sec int64, nsec int32, mono int64)
	goinlinehook.GetOldFunc(mynow, &oldFunc)
	sec, nsec, mono = oldFunc()
	fmt.Printf("hook now, sec:%d, nsec:%d, mono:%d\n", sec, nsec, mono)
	return
}

func HookNow() error {
	var oldFunc func (uintptr)
	_, err := goinlinehook.NewAndHook(mynow, Mynow, &oldFunc)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := HookNow(); err != nil {
		fmt.Println(err.Error())
		return 
	}
	// 会先输hook now, sec:xxx, nsec:xxx, mono:xxx
	fmt.Println(time.Now().Unix())
}