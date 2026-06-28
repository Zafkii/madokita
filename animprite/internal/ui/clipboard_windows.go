//go:build windows

package ui

import (
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard   = user32.NewProc("OpenClipboard")
	procCloseClipboard  = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procGlobalAlloc     = kernel32.NewProc("GlobalAlloc")
	procGlobalLock      = kernel32.NewProc("GlobalLock")
	procGlobalUnlock    = kernel32.NewProc("GlobalUnlock")
	procGlobalFree      = kernel32.NewProc("GlobalFree")
	procLstrlenW        = kernel32.NewProc("lstrlenW")
	procRtlMoveMemory   = kernel32.NewProc("RtlMoveMemory")
)

const (
	cFText       = 1
	gMemMoveable = 0x0002
)

func ClipboardGet() (string, bool) {
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return "", false
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(cFText)
	if h == 0 {
		return "", false
	}

	p, _, _ := syscall.SyscallN(procGlobalLock.Addr(), h)
	if p == 0 {
		return "", false
	}
	defer procGlobalUnlock.Call(h)

	clen, _, _ := procLstrlenW.Call(p)
	if clen == 0 {
		return "", true
	}

	buf := make([]uint16, int(clen))
	procRtlMoveMemory.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		p,
		clen*2,
	)

	text := syscall.UTF16ToString(buf)
	return text, true
}

func ClipboardSet(text string) bool {
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return false
	}
	defer procCloseClipboard.Call()

	procEmptyClipboard.Call()

	utf16 := syscall.StringToUTF16(text)
	size := uintptr(len(utf16) * 2)

	h, _, _ := procGlobalAlloc.Call(gMemMoveable, size)
	if h == 0 {
		return false
	}

	p, _, _ := syscall.SyscallN(procGlobalLock.Addr(), h)
	if p == 0 {
		procGlobalFree.Call(h)
		return false
	}

	procRtlMoveMemory.Call(
		p,
		uintptr(unsafe.Pointer(&utf16[0])),
		size,
	)

	procGlobalUnlock.Call(h)

	r, _, _ = procSetClipboardData.Call(cFText, h)
	return r != 0
}
