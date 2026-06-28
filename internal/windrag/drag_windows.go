//go:build windows

package windrag

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var procGetCursorPos = windows.NewLazySystemDLL("user32.dll").NewProc("GetCursorPos")

func ScreenCursorPos() (x, y int) {
	var pt struct {
		X, Y int32
	}
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}
