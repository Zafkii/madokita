package editor

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) handleDragSelect() {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return
	}
	f := a.focusedInput()
	if f == nil {
		return
	}
	mx, my := ebiten.CursorPosition()
	f.HandleDrag(mx, my)
}

func (a *EditorApp) initScaleOrig(h int) {
	sel := a.canvas.Selection
	if !sel.Visible {
		a.scaleHandle = -1
		return
	}

	hw := sel.W / 2
	hh := sel.H / 2
	cx := sel.X + hw
	cy := sel.Y + hh
	cos := math.Cos(sel.Rotation)
	sin := math.Sin(sel.Rotation)

	aIdx := oppositeHandles[h]
	var alx, aly float64
	switch aIdx {
	case 0:
		alx, aly = -hw, -hh
	case 1:
		alx, aly = hw, -hh
	case 2:
		alx, aly = hw, hh
	case 3:
		alx, aly = -hw, hh
	case 4:
		alx, aly = 0, -hh
	case 5:
		alx, aly = 0, hh
	case 6:
		alx, aly = -hw, 0
	case 7:
		alx, aly = hw, 0
	}

	o := &a.scaleOrig
	o.handleIdx = h
	o.anchorWx = cx + alx*cos - aly*sin
	o.anchorWy = cy + alx*sin + aly*cos
	o.origW = sel.W
	o.origH = sel.H
	o.origRot = sel.Rotation

	switch a.panelMode {
	case panelModeHurtbox:
		sx, sy, ox, oy, rotDeg := a.getSpriteTransformForHurtboxSel()
		o.hbSx, o.hbSy = sx, sy
		o.hbOx, o.hbOy = ox, oy
		rotRad := rotDeg * math.Pi / 180
		o.hbCos = math.Cos(rotRad)
		o.hbSin = math.Sin(rotRad)
	case panelModeSprite:
		if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
			row := a.proj.Sprites[s]
			o.originX = row.OriginX
			o.originY = row.OriginY
			o.origOffsetX = row.OffsetX
			o.origOffsetY = row.OffsetY
			o.origScaleX = row.ScaleX
			o.origScaleY = row.ScaleY
			o.pxW = float64(row.Width)
			o.pxH = float64(row.Height)
		}
	case panelModeAnimFrame:
		if entry := a.currentFrameSpriteEntry(); entry != nil &&
			entry.SpriteIdx >= 0 && entry.SpriteIdx < len(a.proj.Sprites) {
			row := a.proj.Sprites[entry.SpriteIdx]
			o.originX = entry.OriginX
			o.originY = entry.OriginY
			o.origOffsetX = entry.OffsetX
			o.origOffsetY = entry.OffsetY
			o.origScaleX = entry.ScaleX
			o.origScaleY = entry.ScaleY
			o.pxW = float64(row.Width)
			o.pxH = float64(row.Height)
		}
	}
}

func (a *EditorApp) applyHandleScale(mx, my int) {
	o := &a.scaleOrig
	if o.pxW <= 0 || o.pxH <= 0 {
		if a.panelMode == panelModeSprite || a.panelMode == panelModeAnimFrame || a.panelMode == panelModeHitbox {
			return
		}
	}

	cw := float64(a.canvas.Width)
	ch := float64(a.canvas.Height)
	cam := a.canvas.Cam
	mouseWx := (float64(mx-a.canvas.X) - cw/2 - cam.X) / cam.Zoom
	mouseWy := (float64(my-a.canvas.Y) - ch/2 - cam.Y) / cam.Zoom

	cos := math.Cos(o.origRot)
	sin := math.Sin(o.origRot)

	var newW, newH, newCx, newCy float64

	switch o.handleIdx {
	case 0, 1, 2, 3:
		dx := mouseWx - o.anchorWx
		dy := mouseWy - o.anchorWy
		newCx = (o.anchorWx + mouseWx) / 2
		newCy = (o.anchorWy + mouseWy) / 2
		ldx := dx*cos + dy*sin
		ldy := -dx*sin + dy*cos
		newW = math.Abs(ldx)
		newH = math.Abs(ldy)

	case 4:
		dx := mouseWx - o.anchorWx
		dy := mouseWy - o.anchorWy
		t := dx*sin - dy*cos
		if t < 1 {
			t = 1
		}
		projWx := o.anchorWx + t*sin
		projWy := o.anchorWy - t*cos
		newCx = (o.anchorWx + projWx) / 2
		newCy = (o.anchorWy + projWy) / 2
		newW = o.origW
		newH = t

	case 5:
		dx := mouseWx - o.anchorWx
		dy := mouseWy - o.anchorWy
		t := -dx*sin + dy*cos
		if t < 1 {
			t = 1
		}
		projWx := o.anchorWx - t*sin
		projWy := o.anchorWy + t*cos
		newCx = (o.anchorWx + projWx) / 2
		newCy = (o.anchorWy + projWy) / 2
		newW = o.origW
		newH = t

	case 6:
		dx := mouseWx - o.anchorWx
		dy := mouseWy - o.anchorWy
		t := -dx*cos - dy*sin
		if t < 1 {
			t = 1
		}
		projWx := o.anchorWx - t*cos
		projWy := o.anchorWy - t*sin
		newCx = (o.anchorWx + projWx) / 2
		newCy = (o.anchorWy + projWy) / 2
		newW = t
		newH = o.origH

	case 7:
		dx := mouseWx - o.anchorWx
		dy := mouseWy - o.anchorWy
		t := dx*cos + dy*sin
		if t < 1 {
			t = 1
		}
		projWx := o.anchorWx + t*cos
		projWy := o.anchorWy + t*sin
		newCx = (o.anchorWx + projWx) / 2
		newCy = (o.anchorWy + projWy) / 2
		newW = t
		newH = o.origH
	}

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	newX := newCx - newW/2
	newY := newCy - newH/2

	a.canvas.Selection.X = newX
	a.canvas.Selection.Y = newY
	a.canvas.Selection.W = newW
	a.canvas.Selection.H = newH
	a.canvas.Selection.Rotation = o.origRot
	a.canvas.Selection.Visible = true

	switch a.panelMode {
	case panelModeHurtbox:
		hbp := a.hurtboxList()
		if hbp != nil && a.hurtboxTable.SelectedIdx >= 0 && a.hurtboxTable.SelectedIdx < len(*hbp) {
			hb := &(*hbp)[a.hurtboxTable.SelectedIdx]
			hb.Width = math.Round(newW/o.hbSx*100) / 100
			hb.Height = math.Round(newH/o.hbSy*100) / 100
			if hb.Width < 1 {
				hb.Width = 1
			}
			if hb.Height < 1 {
				hb.Height = 1
			}
			hb.X = math.Round(((newCx-o.hbOx)*o.hbCos+(newCy-o.hbOy)*o.hbSin)/o.hbSx*100) / 100
			hb.Y = math.Round((-(newCx-o.hbOx)*o.hbSin+(newCy-o.hbOy)*o.hbCos)/o.hbSy*100) / 100
			a.props[0].SetNumeric(hb.X)
			a.props[1].SetNumeric(hb.Y)
			a.props[2].SetNumeric(hb.Width)
			a.props[3].SetNumeric(hb.Height)
		}
	case panelModeSprite:
		if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
			row := &a.proj.Sprites[s]
			nsx := math.Round(newW/o.pxW*100) / 100
			nsy := math.Round(newH/o.pxH*100) / 100
			if nsx < 0.01 {
				nsx = 0.01
			}
			if nsy < 0.01 {
				nsy = 0.01
			}
			row.ScaleX = nsx
			row.ScaleY = nsy
			row.OffsetX = math.Round(newX + o.originX*newW)
			row.OffsetY = math.Round(newY + o.originY*newH)
			a.props[0].SetNumeric(row.OffsetX)
			a.props[1].SetNumeric(row.OffsetY)
			a.props[3].SetNumeric(row.ScaleX)
			a.props[4].SetNumeric(row.ScaleY)
		}
	case panelModeAnimFrame:
		if entry := a.currentFrameSpriteEntry(); entry != nil {
			nsx := math.Round(newW/o.pxW*100) / 100
			nsy := math.Round(newH/o.pxH*100) / 100
			if nsx < 0.01 {
				nsx = 0.01
			}
			if nsy < 0.01 {
				nsy = 0.01
			}
			entry.ScaleX = nsx
			entry.ScaleY = nsy
			entry.OffsetX = math.Round(newX + o.originX*newW)
			entry.OffsetY = math.Round(newY + o.originY*newH)
			a.props[0].SetNumeric(entry.OffsetX)
			a.props[1].SetNumeric(entry.OffsetY)
			a.props[3].SetNumeric(entry.ScaleX)
			a.props[4].SetNumeric(entry.ScaleY)
		}
	}
}

func (a *EditorApp) getSpriteTransformForHurtboxSel() (sx, sy, ox, oy, rotDeg float64) {
	sx, sy = 1.0, 1.0
	ox, oy = 0.0, 0.0
	rotDeg = 0.0

	entry := a.currentFrameSpriteEntry()
	if entry != nil {
		sx = entry.ScaleX
		sy = entry.ScaleY
		ox = entry.OffsetX
		oy = entry.OffsetY
		rotDeg = entry.Rotation
	} else if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
		sx = a.proj.Sprites[s].ScaleX
		sy = a.proj.Sprites[s].ScaleY
		ox = a.proj.Sprites[s].OffsetX
		oy = a.proj.Sprites[s].OffsetY
		rotDeg = a.proj.Sprites[s].Rotation
	}
	return
}
