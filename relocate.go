package goinlinehook

func relocateInstruction(origInstructions []byte, trampoline uintptr, origAddr uintptr) {
	CopyInstruction(trampoline, origInstructions[:])
	sz := len(origInstructions)
	tmpTpl := trampoline + uintptr(sz)
	code := genJumpCode(GetArchMode(), origAddr, tmpTpl)
	CopyInstruction(tmpTpl, code[:])
}