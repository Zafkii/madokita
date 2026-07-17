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
	if activeFrame != nil {
		for _, entry := range activeFrame.Sprites {
			idx := entry.SpriteIdx
			img, ok := a.loadedSprites[idx]
			row := &a.proj.Sprites[idx]

			ox := entry.OffsetX
			oy := entry.OffsetY
			rot := entry.Rotation
			sx := entry.ScaleX
			sy := entry.ScaleY
			orx := entry.OriginX
			ory := entry.OriginY
			if a.panelMode == panelModeSprite && a.spriteTable.SelectedIdx == idx {
				ox = row.OffsetX
				oy = row.OffsetY
				rot = row.Rotation
				sx = row.ScaleX
				sy = row.ScaleY
				orx = row.OriginX
				ory = row.OriginY
			}

			pw, ph := float64(row.Width), float64(row.Height)
			if ok && img != nil {
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
				frameIdx := entry.SpriteFrameIdx
				if frameIdx >= row.FrameCount {
					frameIdx = row.FrameCount - 1
				}
				if frameIdx < 0 {
					frameIdx = 0
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
			} else {
				renders = append(renders, canvas.SpriteRender{
					OffsetX:      ox,
					OffsetY:      oy,
					Rotation:     rot * math.Pi / 180,
					ScaleX:       sx,
					ScaleY:       sy,
					OriginX:      orx,
					OriginY:      ory,
					PlaceholderW: pw,
					PlaceholderH: ph,
				})
			}
		}
	} else {
		for idx := range a.proj.Sprites {
			img, ok := a.loadedSprites[idx]
			row := a.proj.Sprites[idx]
			pw, ph := float64(row.Width), float64(row.Height)
			if ok && img != nil {
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
				if frameIdx >= row.FrameCount {
					frameIdx = row.FrameCount - 1
				}
				if frameIdx < 0 {
					frameIdx = 0
				}
				fx := (frameIdx % cols) * fw
				fy := (frameIdx / cols) * fh
				frameImg := img.SubImage(image.Rect(fx, fy, fx+fw, fy+fh)).(*ebiten.Image)
				renders = append(renders, canvas.SpriteRender{
					Image:    frameImg,
					OffsetX:  row.OffsetX,
					OffsetY:  row.OffsetY,
					Rotation: row.Rotation * math.Pi / 180,
					ScaleX:   row.ScaleX,
					ScaleY:   row.ScaleY,
					OriginX:  row.OriginX,
					OriginY:  row.OriginY,
				})
			} else {
				renders = append(renders, canvas.SpriteRender{
					OffsetX:      row.OffsetX,
					OffsetY:      row.OffsetY,
					Rotation:     row.Rotation * math.Pi / 180,
					ScaleX:       row.ScaleX,
					ScaleY:       row.ScaleY,
					OriginX:      row.OriginX,
					OriginY:      row.OriginY,
					PlaceholderW: pw,
					PlaceholderH: ph,
				})
			}
		}
	}

	if len(renders) == 0 {
		renders = nil
	}
	a.canvas.SetSpriteRenders(renders)

	entry := a.currentFrameSpriteEntry()
	if hbIdx := a.hurtboxTable.SelectedIdx; hbIdx >= 0 && entry != nil {
		hbp := a.hurtboxList()
		if hbp != nil && hbIdx < len(*hbp) {
			hb := (*hbp)[hbIdx]
			sx, sy := entry.ScaleX, entry.ScaleY
			spriteRot := entry.Rotation
			ox, oy := entry.OffsetX, entry.OffsetY
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
	} else if entry := a.currentFrameSpriteEntry(); entry != nil && a.animTable.SelectedIdx >= 0 {
		if entry.SpriteIdx >= 0 && entry.SpriteIdx < len(a.proj.Sprites) {
			row := a.proj.Sprites[entry.SpriteIdx]
			aw := float64(row.Width) * entry.ScaleX
			ah := float64(row.Height) * entry.ScaleY
			wx := entry.OffsetX - entry.OriginX*aw
			wy := entry.OffsetY - entry.OriginY*ah
			a.canvas.SetSelectionRect(canvas.SelectionRect{X: wx, Y: wy, W: aw, H: ah, Rotation: entry.Rotation * math.Pi / 180, Visible: true})
		} else {
			a.canvas.SetSelectionRect(canvas.SelectionRect{Visible: false})
		}
	} else {
		a.canvas.SetSelectionRect(canvas.SelectionRect{Visible: false})
	}

	a.buildHurtboxRenders()
}

func (a *EditorApp) buildHurtboxRenders() {
	var activeFrame *project.AnimationFrame
	if a.animTable.SelectedIdx >= 0 && a.animTable.SelectedIdx < len(a.proj.Animations) {
		anim := &a.proj.Animations[a.animTable.SelectedIdx]
		if anim.CurrentIdx >= 0 && anim.CurrentIdx < len(anim.Frames) {
			activeFrame = &anim.Frames[anim.CurrentIdx]
		}
	}

	currentEntry := a.currentFrameSpriteEntry()
	selectedHbIdx := a.hurtboxTable.SelectedIdx

	if activeFrame != nil {
		hbRenders := make([]canvas.HurtboxRender, 0)
		for _, entry := range activeFrame.Sprites {
			sx, sy := entry.ScaleX, entry.ScaleY
			ox, oy := entry.OffsetX, entry.OffsetY
			spriteRot := entry.Rotation
			rotRad := spriteRot * math.Pi / 180
			cos := math.Cos(rotRad)
			sin := math.Sin(rotRad)
			isCurrent := currentEntry != nil && entry.SpriteIdx == currentEntry.SpriteIdx

			for i, hb := range entry.Hurtboxes {
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
					Selected:    isCurrent && i == selectedHbIdx,
				})
			}
		}
		a.canvas.SetHurtboxRenders(hbRenders)
		return
	}

	// Fallback: no active animation frame
	hbp := a.hurtboxList()
	if hbp == nil {
		a.canvas.SetHurtboxRenders(nil)
		return
	}
	hbList := *hbp

	entry := a.currentFrameSpriteEntry()
	sx, sy, ox, oy, spriteRot := 1.0, 1.0, 0.0, 0.0, 0.0
	if entry != nil {
		sx = entry.ScaleX
		sy = entry.ScaleY
		ox = entry.OffsetX
		oy = entry.OffsetY
		spriteRot = entry.Rotation
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
			Selected:    i == selectedHbIdx,
		})
	}
	a.canvas.SetHurtboxRenders(hbRenders)
}
