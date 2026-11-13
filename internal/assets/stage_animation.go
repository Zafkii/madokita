package assets

import (
	"image"
	"math"

	"madokita-gems/internal/stage"

	"github.com/hajimehoshi/ebiten/v2"
)

// StageAnimation is a pre-composed animation for one animated stage asset.
// Frames are independent images so the scene can draw the current frame with
// the same transform used by static stage objects.
type StageAnimation struct {
	Frames []*ebiten.Image
	FPS    float64
	Loop   bool
}

// Frame returns the image for the given elapsed time in seconds.
func (a *StageAnimation) Frame(clock float64) *ebiten.Image {
	if a == nil || len(a.Frames) == 0 {
		return nil
	}
	if clock < 0 {
		clock = 0
	}
	fps := a.FPS
	if fps <= 0 {
		fps = 1
	}
	idx := int(clock * fps)
	if a.Loop {
		idx %= len(a.Frames)
	} else if idx >= len(a.Frames) {
		idx = len(a.Frames) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return a.Frames[idx]
}

func composeStageAnimation(asset *stage.Asset, load func(string, *ebiten.Image) *ebiten.Image) *StageAnimation {
	if asset == nil || asset.Animation == nil ||
		len(asset.Animation.Frames) == 0 || len(asset.Sprites) == 0 {
		return nil
	}

	sheets := make([]*ebiten.Image, len(asset.Sprites))
	for i, def := range asset.Sprites {
		if def.File == "" {
			return nil
		}
		img := load(def.File, nil)
		if img == nil {
			return nil
		}
		sheets[i] = img
	}

	frameW, frameH := asset.Animation.FrameW, asset.Animation.FrameH
	if frameW <= 0 || frameH <= 0 {
		frameW, frameH = asset.Sprites[0].FrameW, asset.Sprites[0].FrameH
	}
	if frameW <= 0 || frameH <= 0 {
		bounds := sheets[0].Bounds()
		frameW, frameH = bounds.Dx(), bounds.Dy()
	}
	if frameW <= 0 || frameH <= 0 {
		return nil
	}

	anim := &StageAnimation{FPS: asset.Animation.FPS, Loop: asset.Animation.Loop}
	anim.Frames = make([]*ebiten.Image, 0, len(asset.Animation.Frames))
	for _, frame := range asset.Animation.Frames {
		img := ebiten.NewImage(frameW, frameH)
		drawStageAnimationFrame(img, sheets, asset.Sprites, frame)
		anim.Frames = append(anim.Frames, img)
	}
	return anim
}

func drawStageAnimationFrame(dst *ebiten.Image, sheets []*ebiten.Image, defs []stage.SpriteSheet, frame stage.AnimFrame) {
	cx := float64(dst.Bounds().Dx()) / 2
	cy := float64(dst.Bounds().Dy()) / 2
	for _, entry := range frame.Sprites {
		if entry.SpriteIdx < 0 || entry.SpriteIdx >= len(defs) || entry.SpriteIdx >= len(sheets) {
			continue
		}
		sheet := sheets[entry.SpriteIdx]
		def := defs[entry.SpriteIdx]
		if sheet == nil || def.FrameW <= 0 || def.FrameH <= 0 {
			continue
		}
		fw, fh := def.FrameW, def.FrameH
		cols := sheet.Bounds().Dx() / fw
		if cols < 1 {
			cols = 1
		}
		idx := entry.SpriteFrameIdx
		if idx < 0 {
			idx = 0
		}
		x := (idx % cols) * fw
		y := (idx / cols) * fh
		if x+fw > sheet.Bounds().Dx() || y+fh > sheet.Bounds().Dy() {
			continue
		}
		sub, ok := sheet.SubImage(image.Rect(x, y, x+fw, y+fh)).(*ebiten.Image)
		if !ok {
			continue
		}

		originX := entry.OriginX * float64(fw)
		originY := entry.OriginY * float64(fh)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-originX, -originY)
		op.GeoM.Scale(entry.ScaleX, entry.ScaleY)
		op.GeoM.Rotate(entry.Rotation * math.Pi / 180)
		op.GeoM.Translate(cx+entry.OffsetX, cy+entry.OffsetY)
		dst.DrawImage(sub, op)
	}
}
