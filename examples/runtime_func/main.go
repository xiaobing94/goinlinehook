package main
import (
	"fmt"
	"time"
	"unsafe"
	"github.com/xiaobing94/goinlinehook"
)

//go:linkname mynewproc runtime.newproc
func mynewproc(siz int32, fn uintptr)

var oldFunc func (siz int32, fn uintptr)

func Mynewproc(siz int32, fn uintptr) {
	funcAddr := *((*int64)(unsafe.Pointer(fn)))
	fmt.Printf("siz:%d, fn:0x%x, funcAddr:0x%x\n", siz, fn, funcAddr)
	oldFunc(siz, fn)
}

func main() {
	item, err := goinlinehook.NewAndHook(mynewproc, Mynewproc, &oldFunc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go func () {
		println("test goroutine!")
	}()
	item.UnHook()
	time.Sleep(time.Second)
}