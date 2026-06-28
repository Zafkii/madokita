package editor

import (
	"image"
	"math"

	"animprite/internal/canvas"
	"animprite/internal/project"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) buildSpriteRenders() {
	var activeFrame *project.AnimationFrame
	if a.animTable.SelectedIdx >= 0 && a.animTable.SelectedIdx < len(a.proj.Animations) {
		anim := &a.proj.Animations[a.animTable.SelectedIdx]
		if anim.CurrentIdx >= 0 && anim.CurrentIdx < len(anim.Frames) {
			activeFrame = &anim.Frames[anim.CurrentIdx]
		}
	}

	var renders []canvas.SpriteRender
	for idx := range a.proj.Sprites {
		img, ok := a.loadedSprites[idx]
		if !ok || img == nil {
			continue
		}

		row := a.proj.Sprites[idx]

		fw := row.Width
		fh := row.Height
		if fw <= 0 {
			fw = img.Bounds().Dx()
		}
		if fh <= 0 {
			fh = img.Bounds().Dy()
		}
		cols := img.Bounds().Dx() / fw
		if cols < 1 {
			cols = 1
		}
		frameIdx := row.CurrentIdx
		if activeFrame != nil && idx == activeFrame.SpriteIdx {
			frameIdx = activeFrame.SpriteFrameIdx
		}
		if frameIdx >= row.FrameCount {
			frameIdx = row.FrameCount - 1
		}
		if frameIdx < 0 {
			frameIdx = 0
		}

		ox, oy, rot, sx, sy, orx, ory := row.OffsetX, row.OffsetY, row.Rotation, row.ScaleX, row.ScaleY, row.OriginX, row.OriginY
		if activeFrame != nil && idx == activeFrame.SpriteIdx {
			ox = activeFrame.OffsetX
			oy = activeFrame.OffsetY
			rot = activeFrame.Rotation
			sx = activeFrame.ScaleX
			sy = activeFrame.ScaleY
			orx = activeFrame.OriginX
			ory = activeFrame.OriginY
		}

		fx := (frameIdx % cols) * fw
		fy := (frameIdx / cols) * fh
		frameImg := img.SubImage(image.Rect(fx, fy, fx+fw, fy+fh)).(*ebiten.Image)

		renders = append(renders, canvas.SpriteRender{
			Image:    frameImg,
			OffsetX:  ox,
			OffsetY:  oy,
			Rotation: rot * math.Pi / 180,
			ScaleX:   sx,
			ScaleY:   sy,
			OriginX:  orx,
			OriginY:  ory,
		})
	}

	if len(renders) == 0 {
		renders = nil
	}
	a.canvas.SetSpriteRenders(renders)

	if hbIdx := a.hurtboxTable.SelectedIdx; hbIdx >= 0 {
		hbp := a.hurtboxList()
		if hbp != nil && hbIdx < len(*hbp) {
			hb := (*hbp)[hbIdx]
			sx, sy := 1.0, 1.0
			spriteRot := 0.0
			ox, oy := 0.0, 0.0
			if activeFrame != nil {
				sx = activeFrame.ScaleX
				sy = activeFrame.ScaleY
				spriteRot = activeFrame.Rotation
				ox = activeFrame.OffsetX
				oy = activeFrame.OffsetY
			} else if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
				sx = a.proj.Sprites[s].ScaleX
				sy = a.proj.Sprites[s].ScaleY
				spriteRot = a.proj.Sprites[s].Rotation
				ox = a.proj.Sprites[s].OffsetX
				oy = a.proj.Sprites[s].OffsetY
			}
			rotRad := spriteRot * math.Pi / 180
			cos := math.Cos(rotRad)
			sin := math.Sin(rotRad)
			wcx := (hb.X*sx)*cos - (hb.Y*sy)*sin + ox
			wcy := (hb.X*sx)*sin + (hb.Y*sy)*cos + oy
			ww := hb.Width * sx
			wh := hb.Height * sy
			hbRotRad := (spriteRot + hb.Rotation) * math.Pi / 180
			a.canvas.SetSelectionRect(canvas.SelectionRect{X: wcx - ww/2, Y: wcy - wh/2, W: ww, H: wh, Rotation: hbRotRad, Visible: true})
		} else {
			a.canvas.SetSelectionRect(canvas.SelectionRect{Visible: false})
		}
	} else if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
		row := a.proj.Sprites[s]
		aw := float64(row.Width) * row.ScaleX
		ah := float64(row.Height) * row.ScaleY
		wx := row.OffsetX - row.OriginX*aw
		wy := row.OffsetY - row.OriginY*ah
		a.canvas.SetSelectionRect(canvas.SelectionRect{X: wx, Y: wy, W: aw, H: ah, Rotation: row.Rotation * math.Pi / 180, Visible: true})
	} else {
		a.canvas.SetSelectionRect(canvas.SelectionRect{Visible: false})
	}

	a.buildHurtboxRenders()
}

func (a *EditorApp) buildHurtboxRenders() {
	hbp := a.hurtboxList()
	if hbp == nil {
		a.canvas.SetHurtboxRenders(nil)
		return
	}
	hbList := *hbp

	sx, sy, ox, oy, spriteRot := 1.0, 1.0, 0.0, 0.0, 0.0

	animIdx := a.animTable.SelectedIdx
	if animIdx < 0 {
		animIdx = a.hurtboxAnimCtx
	}
	if animIdx >= 0 && animIdx < len(a.proj.Animations) {
		anim := &a.proj.Animations[animIdx]
		if anim.CurrentIdx >= 0 && anim.CurrentIdx < len(anim.Frames) {
			frame := anim.Frames[anim.CurrentIdx]
			sx = frame.ScaleX
			sy = frame.ScaleY
			ox = frame.OffsetX
			oy = frame.OffsetY
			spriteRot = frame.Rotation
		}
	} else if len(a.proj.Sprites) > 0 {
		row := a.proj.Sprites[0]
		sx = row.ScaleX
		sy = row.ScaleY
		ox = row.OffsetX
		oy = row.OffsetY
		spriteRot = row.Rotation
	}

	hbRenders := make([]canvas.HurtboxRender, 0, len(hbList))
	rotRad := spriteRot * math.Pi / 180
	cos := math.Cos(rotRad)
	sin := math.Sin(rotRad)
	for i, hb := range hbList {
		if hb.Width <= 0 || hb.Height <= 0 {
			continue
		}
		hbRot := (spriteRot + hb.Rotation) * math.Pi / 180
		wx := (hb.X*sx)*cos - (hb.Y*sy)*sin + ox
		wy := (hb.X*sx)*sin + (hb.Y*sy)*cos + oy
		hbRenders = append(hbRenders, canvas.HurtboxRender{
			OffsetX:     wx,
			OffsetY:     wy,
			WorldWidth:  hb.Width * sx,
			WorldHeight: hb.Height * sy,
			Rotation:    hbRot,
			Selected:    i == a.hurtboxTable.SelectedIdx,
		})
	}
	a.canvas.SetHurtboxRenders(hbRenders)
}
