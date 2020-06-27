# goinlinehook
 goinlinehook 是go语言环境下的inline hook库。

## 特征
* 可以调用原函数
* 可以hook私有函数
* 可以hook成员函数
* 可以hook runtime函数

## 使用方法
- 原函数
```go
//go:noinline
func Add(a, b int) int {
	return a+b
}
```
1. 编写代理函数

```go
//go:noinline
func MyAdd(a, b int) int {
	var oldFunc func (int, int) int
	goinlinehook.GetOldFunc(Add, &oldFunc)
	return oldFunc(a, b) + 10
}
```

2. hook目标函数
```go
fmt.Println("result:", res)
hookItem, err := goinlinehook.NewAndHook(Add, MyAdd, &oldFunc)
if err != nil {
    fmt.Println(err.Error())
    return
}
```
这种方式默认是使用 `push xx, ret`的形式进行跳转的
也可以使用如下的形式:
```go
trampoline[8191]
func Hook() {
    item := goinlinehook.NewHookItem(Add, MyAdd)
    ptr := (uintptr)(unsafe.Pointer(&trampoline[0]))
    item.TrampolineAddr = ptr
    err := item.Hook()
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    hook.GetOldFunc(mynewproc, &oldFunc)
}
```
这种方式由于在4字节跳转范围内，使用`jmp xxx`进行跳转，不占用栈空间。
