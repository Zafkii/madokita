//go:build !windows && !linux

package ui

func ClipboardGet() (string, bool) {
	return "", false
}

func ClipboardSet(_ string) bool {
	return false
}
