package main

import (
	"errors"
	"image"
	"log"

	"animprite/internal/editor"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	app := editor.NewEditorApp()
	ebiten.SetWindowSize(editor.DefaultWinW, editor.DefaultWinH)
	ebiten.SetWindowTitle("Animprite")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowDecorated(false)

	if img, err := editor.LoadICO("assets/logo.ico", 0); err == nil && img != nil {
		ebiten.SetWindowIcon([]image.Image{img})
	}

	if p := editor.LoadWindowPrefs(); p.WindowW > 0 && p.WindowH > 0 {
		ebiten.SetWindowPosition(p.WindowX, p.WindowY)
		ebiten.SetWindowSize(p.WindowW, p.WindowH)
	}

	if err := ebiten.RunGame(app); err != nil && !errors.Is(err, editor.ErrWindowClose) {
		log.Fatal(err)
	}
}
