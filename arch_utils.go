package goinlinehook

import (
	"math"
	"unsafe"

	"golang.org/x/arch/x86/x86asm"
)
var (
	minJmpCodeSize = 0
)

const (
	FT_CondJmp  = 1
	FT_JMP      = 2
	FT_CALL     = 3
	FT_RET      = 4
	FT_OTHER    = 5
	FT_INVALID  = 6
	FT_SKIP     = 7
	FT_OVERFLOW = 8
)

func SetMinJmpCodeSize(sz int) {
	minJmpCodeSize = sz
}

func isByteOverflow(v int32) bool {
	if v > 0 {
		if v > math.MaxInt8 {
			return true
		}
	} else {
		if v < math.MinInt8 {
			return true
		}
	}

	return false
}

func isIntOverflow(v int64) bool {
	if v > 0 {
		if v > math.MaxInt32 {
			return true
		}
	} else {
		if v < math.MinInt32 {
			return true
		}
	}

	return false
}

func calcOffset(insSz int, startAddr, curAddr, to uintptr, to_sz int, offset int32) int64 {
	newAddr := curAddr
	absAddr := curAddr + uintptr(insSz) + uintptr(offset)

	if curAddr < startAddr+uintptr(to_sz) {
		newAddr = to + (curAddr - startAddr)
	}

	if absAddr >= startAddr && absAddr < startAddr+uintptr(to_sz) {
		absAddr = to + (absAddr - startAddr)
	}
	return int64(uint64(absAddr) - uint64(newAddr) - uint64(insSz))
}

// 指令修复
func FixInstruction(mode int, startAddr, curAddr uintptr, code []byte, to uintptr, to_sz int) (int, int, []byte) {
	nc := make([]byte, len(code))
	copy(nc, code)
	// 2 bytes
	if code[0] == 0xe3 || code[0] == 0xeb || (code[0] >= 0x70 && code[0] <= 0x7f) {
		nc = nc[:2]
		off := calcOffset(2, startAddr, curAddr, to, to_sz, int32(int8(code[1])))
		if off != int64(int8(nc[1])) {
			if isByteOverflow(int32(off)) {
				return 2, FT_OVERFLOW, nc
			}
			nc[1] = byte(off)
			return 2, FT_CondJmp, nc
		}
		return 2, FT_SKIP, nc
	}

	// 6 bytes 
	if code[0] == 0x0f && (code[1] >= 0x80 && code[1] <= 0x8f) {
		nc = nc[:6]
		off1 := (uint32(code[2]) | (uint32(code[3]) << 8) | (uint32(code[4]) << 16) | (uint32(code[5]) << 24))
		off2 := uint64(calcOffset(6, startAddr, curAddr, to, to_sz, int32(off1)))
		if uint64(int32(off1)) != off2 {
			if isIntOverflow(int64(off2)) {
				return 6, FT_OVERFLOW, nc
			}
			nc[2] = byte(off2)
			nc[3] = byte(off2 >> 8)
			nc[4] = byte(off2 >> 16)
			nc[5] = byte(off2 >> 24)
			return 6, FT_CondJmp, nc
		}
		return 6, FT_SKIP, nc
	}

	// 5 bytes
	if code[0] == 0xe9 || code[0] == 0xe8 {
		nc = nc[:5]
		off1 := (uint32(code[1]) | (uint32(code[2]) << 8) | (uint32(code[3]) << 16) | (uint32(code[4]) << 24))

		off2 := uint64(0)
		if code[0] == 0xe8 && startAddr == (curAddr+uintptr(5)+uintptr(int32(off1))) {
			off2 = uint64(int32(off1))
		} else {
			off2 = uint64(calcOffset(5, startAddr, curAddr, to, to_sz, int32(off1)))
		}

		if uint64(int32(off1)) != off2 {
			if isIntOverflow(int64(off2)) {
				return 5, FT_OVERFLOW, nc
			}
			nc[1] = byte(off2)
			nc[2] = byte(off2 >> 8)
			nc[3] = byte(off2 >> 16)
			nc[4] = byte(off2 >> 24)
			return 5, FT_JMP, nc
		}
		return 5, FT_SKIP, nc
	}

	// ret指令不需要修复
	if code[0] == 0xc3 || code[0] == 0xcb {
		nc = nc[:1]
		return 1, FT_RET, nc
	}

	if code[0] == 0xc2 || code[0] == 0xca {
		nc = nc[:3]
		return 3, FT_RET, nc
	}

	inst, err := x86asm.Decode(code, mode)
	if err != nil || (inst.Opcode == 0 && inst.Len == 1 && inst.Prefix[0] == x86asm.Prefix(code[0])) {
		return 0, FT_INVALID, nc
	}

	if inst.Len == 1 && code[0] == 0xcc {
		return 0, FT_INVALID, nc
	}

	sz := inst.Len
	nc = nc[:sz]
	return sz, FT_OTHER, nc
}


func genJumpCode(mode int, to, from uintptr) []byte {
	// 1. |from-to| < 2G 使用相对跳转 
	// 2. 否则， 使用 push target, 然后ret

	var code []byte
	delta := int64(from - to)
	if unsafe.Sizeof(uintptr(0)) == unsafe.Sizeof(int32(0)) {
		delta = int64(int32(from - to))
	}

	relative := (delta <= 0x7fffffff)

	if delta < 0 {
		delta = -delta
		relative = (delta <= 0x80000000)
	}

	if relative {
		var dis uint32
		if to > from {
			dis = uint32(int32(to-from) - 5)
		} else {
			dis = uint32(-int32(from-to) - 5)
		}
		code = []byte{
			0xe9,
			byte(dis),
			byte(dis >> 8),
			byte(dis >> 16),
			byte(dis >> 24),
		}
	} else if mode == 32 {
		code = []byte{
			0x68, // push
			byte(to),
			byte(to >> 8),
			byte(to >> 16),
			byte(to >> 24),
			0xc3, // retn
		}
	} else if mode == 64 {
		code = []byte{
			0x68, //push
			byte(to), byte(to >> 8), byte(to >> 16), byte(to >> 24),
			0xc7, 0x44, 0x24, // mov $value, 4%rsp
			0x04, // rsp + 4
			byte(to >> 32), byte(to >> 40), byte(to >> 48), byte(to >> 56),
			0xc3, // retn
		}
	} else {
		panic("invalid mode")
	}

	sz := len(code)
	if minJmpCodeSize > 0 && sz < minJmpCodeSize {
		nop := make([]byte, 0, minJmpCodeSize-sz)
		for {
			if len(nop) >= minJmpCodeSize-sz {
				break
			}
			nop = append(nop, 0x90)
		}

		code = append(code, nop...)
	}

	return code
}
