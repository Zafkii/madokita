package assets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

func SliceFrames(src *ebiten.Image, frameW, frameH, count int) []*ebiten.Image {
	if count <= 1 || frameW <= 0 || frameH <= 0 {
		return []*ebiten.Image{src}
	}

	frames := make([]*ebiten.Image, 0, count)
	sheetW := src.Bounds().Dx()
	if sheetW == 0 {
		return []*ebiten.Image{src}
	}
	cols := sheetW / frameW
	if cols < 1 {
		cols = 1
	}

	for i := 0; i < count; i++ {
		x := (i % cols) * frameW
		y := (i / cols) * frameH
		rect := image.Rect(x, y, x+frameW, y+frameH)
		sub := src.SubImage(rect)
		frames = append(frames, sub.(*ebiten.Image))
	}
	return frames
}
