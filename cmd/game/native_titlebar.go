package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// NativeTitleBar is a development toggle to compare the two title bar
// implementations:
//
//   - true: the operating system draws its native title bar (window
//     decorations, borders, drag), so minimize/maximize/close and their
//     hover states are handled by the OS. The window becomes freely
//     resizable without the 16:9 snap of the custom implementation.
//
//   - false: the window is undecorated (ebiten.SetWindowDecorated(false))
//     and the game draws its own title bar over the canvas (titlebar.go),
//     with manual drag, resize edges and button hit-testing (window.go).
//
// The fake title bar code is kept intact behind the !NativeTitleBar guards.
const NativeTitleBar = true

// applyWindowDecorations configures the OS window for the active title bar
// mode. Replaces the previous unconditional SetWindowDecorated(false).
func applyWindowDecorations() {
	if !NativeTitleBar {
		ebiten.SetWindowDecorated(false)
		return
	}
	// With a native title bar the OS close button must not kill the window
	// immediately: we intercept IsWindowBeingClosed in Update to persist the
	// window position before returning ErrWindowClose.
	ebiten.SetWindowClosingHandled(true)
}