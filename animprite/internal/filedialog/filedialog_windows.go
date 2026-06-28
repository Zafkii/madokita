//go:build windows

package filedialog

import (
	"syscall"
	"unsafe"
)

var (
	comdlg32          = syscall.NewLazyDLL("comdlg32.dll")
	getOpenFileNameW  = comdlg32.NewProc("GetOpenFileNameW")
	getSaveFileNameW  = comdlg32.NewProc("GetSaveFileNameW")
)

type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        unsafe.Pointer
	dwReserved        uint32
	flagsEx           uint32
}

func openFile(title, filter string) (string, error) {
	buf := make([]uint16, 1024)
	buf[0] = 0

	titleW, _ := syscall.UTF16PtrFromString(title)
	filterW, _ := syscall.UTF16PtrFromString(filter)

	ofn := &openFileNameW{
		lStructSize: uint32(unsafe.Sizeof(openFileNameW{})),
		lpstrFilter: filterW,
		lpstrFile:   &buf[0],
		nMaxFile:    uint32(len(buf)),
		lpstrTitle:  titleW,
		flags:       0x00000800 | 0x00000004, // OFN_FILEMUSTEXIST | OFN_HIDEREADONLY
	}

	ret, _, _ := getOpenFileNameW.Call(uintptr(unsafe.Pointer(ofn)))
	if ret == 0 {
		return "", errNoSelection
	}

	return syscall.UTF16ToString(buf), nil
}

func saveFile(title, filter string) (string, error) {
	buf := make([]uint16, 1024)
	buf[0] = 0

	titleW, _ := syscall.UTF16PtrFromString(title)
	filterW, _ := syscall.UTF16PtrFromString(filter)

	ofn := &openFileNameW{
		lStructSize: uint32(unsafe.Sizeof(openFileNameW{})),
		lpstrFilter: filterW,
		lpstrFile:   &buf[0],
		nMaxFile:    uint32(len(buf)),
		lpstrTitle:  titleW,
		flags:       0x00000002, // OFN_OVERWRITEPROMPT
	}

	ret, _, _ := getSaveFileNameW.Call(uintptr(unsafe.Pointer(ofn)))
	if ret == 0 {
		return "", errNoSelection
	}

	return syscall.UTF16ToString(buf), nil
}
