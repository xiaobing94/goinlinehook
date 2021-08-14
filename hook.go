package goinlinehook

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	"golang.org/x/arch/x86/x86asm"
)

var (
	archMode = 64
	maxInstSize = 16
	g_all = make(map[uintptr]*HookItem)
)

// StatusType 状态类型
type StatusType int32

const (
	StatusHooked StatusType = iota
	StatusUnHook
)

func init() {
	sz := unsafe.Sizeof(uintptr(0))
	if sz == 4 {
		archMode = 32
	}
}

// GetArchMode 获取系统位宽
func GetArchMode() int {
	return archMode
}

type valueStruct struct {
	typ uintptr
	ptr uintptr
}

func getDataPtrFromValue(v reflect.Value) uintptr {
	return (uintptr)((*valueStruct)(unsafe.Pointer(&v)).ptr)
}

// HookItem hook项
type HookItem struct {
	TargetAddr reflect.Value
	NewAddr reflect.Value
	TrampolineAddr uintptr
	OriginalInst []byte
	Status StatusType
}

// NewHookItem 创建并hook
func NewHookItem(target, newFunc interface{}) *HookItem{
	t := reflect.ValueOf(target)
	n := reflect.ValueOf(newFunc)
	item := &HookItem{
		TargetAddr: t,
		NewAddr: n,
	}
	return item
}

// UnHook 取消hook
func (item *HookItem) UnHook() {
	target := item.TargetAddr.Pointer()
	CopyInstruction(target, item.OriginalInst)
	item.Status = StatusUnHook
	return
}

// GetOldFunc 获取旧函数地址
func (item *HookItem) GetOldFunc(oldFunc interface{}) error {
	if item.Status != StatusHooked {
		return fmt.Errorf("HookItem is not Hooked")
	}
	oldValue := reflect.ValueOf(oldFunc)
	oldFuncAddr := oldValue.Pointer()
	*(*uintptr)(unsafe.Pointer(oldFuncAddr)) = (uintptr)(unsafe.Pointer(&item.TrampolineAddr))
	return nil
}

// Hook hook目标函数
func (item *HookItem) Hook() error {
	if item.TargetAddr.Kind() != reflect.Func {
		return fmt.Errorf("target must be a Func")
	}
	if item.NewAddr.Kind() != reflect.Func {
		return fmt.Errorf("NewAddr must be a Func")
	}
	if item.TrampolineAddr == 0 {
		pageSize := syscall.Getpagesize()
		trampoline := make([]byte, pageSize)
		item.TrampolineAddr = (uintptr)(unsafe.Pointer(&trampoline[0]))
	}
	from := item.TargetAddr.Pointer()
	to := item.NewAddr.Pointer()
	mode := GetArchMode()
	code := genJumpCode(mode, to, from)
	lenght := len(code)
	codeLen := 0
	for lenght > codeLen {
		fs := makeSliceFromPointer(from + uintptr(codeLen), maxInstSize)
		inst, err := x86asm.Decode(fs, mode)
		if err != nil {
			return err
		}
		codeLen += inst.Len
	}
	l := 0
	fromCode := makeSliceFromPointer(from, codeLen)
	item.OriginalInst = append(item.OriginalInst, fromCode...)
	var fixedOriginalCode []byte
	for l < codeLen {
		curFromCode := fromCode[l:]
		sz, ft, nc := FixInstruction(mode, from, from+uintptr(l), curFromCode, item.TrampolineAddr, codeLen)
		if ft == FT_OVERFLOW {
			return fmt.Errorf("Can not fix instruction, the new offset is overflow")
		}
		fixedOriginalCode = append(fixedOriginalCode, nc...)
		l += sz
	}
	// 拷贝跳转代码到目标函数
	CopyInstruction(from, code[:])
	// 复制修正过得代码到跳板，且在跳板末尾增加跳转到原函数的代码
	relocateInstruction(fixedOriginalCode, item.TrampolineAddr, from+(uintptr)(codeLen))
	item.Status = StatusHooked
	g_all[from] = item
	return nil
}

// NewAndHook 实例化并hook
func NewAndHook(target, newFunc, oldFunc interface{}) (*HookItem, error){
	item := NewHookItem(target, newFunc)
	err := item.Hook()
	if err != nil {
		return nil, err
	}
	err = item.GetOldFunc(oldFunc)
	return item, err
}

// GetOldFunc 获取旧函数指针
func GetOldFunc(target, oldFunc interface{}) error {
	t := reflect.ValueOf(target)
	targetAddr := t.Pointer()
	hookitem, ok := g_all[targetAddr]
	if !ok {
		return fmt.Errorf("target Func not Found")
	}
	err := hookitem.GetOldFunc(oldFunc)
	return err
}