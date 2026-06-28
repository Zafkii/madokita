//go:build windows

package windrag

import (
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getCursorPosProc = user32.NewProc("GetCursorPos")
)

type point struct {
	X, Y int32
}

func ScreenCursorPos() (x, y int) {
	var pt point
	ret, _, _ := getCursorPosProc.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return -1, -1
	}
	return int(pt.X), int(pt.Y)
}
