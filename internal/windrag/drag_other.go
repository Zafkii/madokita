//go:build !windows && !linux

package windrag

func ScreenCursorPos() (x, y int) {
	return 0, 0
}
