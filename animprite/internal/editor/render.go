package editor

import (
	"image"
	"math"

	"animprite/internal/canvas"
	"animprite/internal/project"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) buildSpriteRenders() {
	if a.isStageMode() {
		a.buildStageRenders()
		return
	}

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
					Opacity:  1,
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
					Opacity:      1,
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
					Opacity:  1,
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
					Opacity:      1,
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
			// Hurtboxes render against the frame's first sprite entry
			// (see boxRenders), so the selection must use the same
			// transform instead of the sprite-selected entry.
			sx, sy, ox, oy, spriteRot := a.spriteTransformForFrame(activeFrame)
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
	} else if htIdx := a.hitboxTable.SelectedIdx; htIdx >= 0 {
		htp := a.hitboxList()
		if htp != nil && htIdx < len(*htp) {
			ht := (*htp)[htIdx]
			sx, sy, ox, oy, spriteRot := a.spriteTransformForFrame(activeFrame)
			rotRad := spriteRot * math.Pi / 180
			cos := math.Cos(rotRad)
			sin := math.Sin(rotRad)
			wcx := (ht.X*sx)*cos - (ht.Y*sy)*sin + ox
			wcy := (ht.X*sx)*sin + (ht.Y*sy)*cos + oy
			ww := ht.Width * sx
			wh := ht.Height * sy
			a.canvas.SetSelectionRect(canvas.SelectionRect{X: wcx - ww/2, Y: wcy - wh/2, W: ww, H: wh, Rotation: rotRad, Visible: true})
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

	a.buildBoxRenders()
}

// boxRenders projects the current frame's hurtboxes and attack hitboxes into
// world space, using the same transform for both so the overlay lines up with
// the artwork. The canvas picks the palette from Kind.
func (a *EditorApp) boxRenders() []canvas.BoxRender {
	frame := a.currentFrame()
	if frame == nil {
		return nil
	}
	sx, sy, ox, oy, spriteRot := a.spriteTransformForFrame(frame)
	rotRad := spriteRot * math.Pi / 180
	cos := math.Cos(rotRad)
	sin := math.Sin(rotRad)

	renders := make([]canvas.BoxRender, 0, len(frame.Hurtboxes)+len(frame.Hitboxes))
	selectedHbIdx := a.hurtboxTable.SelectedIdx
	for i, hb := range frame.Hurtboxes {
		if hb.Width <= 0 || hb.Height <= 0 {
			continue
		}
		renders = append(renders, canvas.BoxRender{
			OffsetX:     (hb.X*sx)*cos - (hb.Y*sy)*sin + ox,
			OffsetY:     (hb.X*sx)*sin + (hb.Y*sy)*cos + oy,
			WorldWidth:  hb.Width * sx,
			WorldHeight: hb.Height * sy,
			Rotation:    (spriteRot + hb.Rotation) * math.Pi / 180,
			Selected:    i == selectedHbIdx,
			Kind:        canvas.RenderHurtbox,
		})
	}

	selectedHtIdx := a.hitboxTable.SelectedIdx
	for i, ht := range frame.Hitboxes {
		if ht.Width <= 0 || ht.Height <= 0 {
			continue
		}
		renders = append(renders, canvas.BoxRender{
			OffsetX:     (ht.X*sx)*cos - (ht.Y*sy)*sin + ox,
			OffsetY:     (ht.X*sx)*sin + (ht.Y*sy)*cos + oy,
			WorldWidth:  ht.Width * sx,
			WorldHeight: ht.Height * sy,
			Rotation:    spriteRot * math.Pi / 180,
			Selected:    i == selectedHtIdx,
			Kind:        canvas.RenderHitbox,
		})
	}
	return renders
}

func (a *EditorApp) buildBoxRenders() {
	a.canvas.SetBoxRenders(a.boxRenders())
}

func (a *EditorApp) spriteTransformForFrame(frame *project.AnimationFrame) (sx, sy, ox, oy, rotDeg float64) {
	sx, sy = 1.0, 1.0
	if len(frame.Sprites) > 0 {
		e := &frame.Sprites[0]
		sx, sy = e.ScaleX, e.ScaleY
		ox, oy = e.OffsetX, e.OffsetY
		rotDeg = e.Rotation
	} else if len(a.proj.Sprites) > 0 {
		row := a.proj.Sprites[0]
		sx, sy = row.ScaleX, row.ScaleY
		ox, oy = row.OffsetX, row.OffsetY
		rotDeg = row.Rotation
	}
	return
}
