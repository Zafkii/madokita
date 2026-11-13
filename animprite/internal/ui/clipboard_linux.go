//go:build linux

package ui

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

const clipboardCommandTimeout = 1500 * time.Millisecond

func clipboardCommands(wayland bool) (get, set [][]string) {
	getX11 := [][]string{
		{"xclip", "-selection", "clipboard", "-o"},
		{"xsel", "--clipboard", "--output"},
	}
	setX11 := [][]string{
		{"xclip", "-selection", "clipboard", "-i"},
		{"xsel", "--clipboard", "--input"},
	}
	getWayland := [][]string{{"wl-paste", "--no-newline"}}
	setWayland := [][]string{{"wl-copy"}}
	if wayland {
		return append(getWayland, getX11...), append(setWayland, setX11...)
	}
	return append(getX11, getWayland...), append(setX11, setWayland...)
}

func ClipboardGet() (string, bool) {
	get, _ := clipboardCommands(os.Getenv("WAYLAND_DISPLAY") != "")
	for _, args := range get {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), clipboardCommandTimeout)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		cancel()
		if err == nil {
			return strings.TrimRight(out.String(), "\r\n"), true
		}
	}
	return "", false
}

func ClipboardSet(text string) bool {
	_, set := clipboardCommands(os.Getenv("WAYLAND_DISPLAY") != "")
	for _, args := range set {
		if _, err := exec.LookPath(args[0]); err != nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), clipboardCommandTimeout)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		err := cmd.Run()
		cancel()
		if err == nil {
			return true
		}
	}
	return false
}
