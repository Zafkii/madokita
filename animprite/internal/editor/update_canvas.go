package editor

import (
	"math"

	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

func (a *EditorApp) handleCanvasWheel(mx, my int, wy float64, ctrl, shift bool) {
	cx, cy2, cw, ch := a.canvas.X, a.canvas.Y, a.canvas.Width, a.canvas.Height
	if mx >= cx && mx <= cx+cw && my >= cy2 && my <= cy2+ch {
		switch {
		case ctrl:
			if hbIdx := a.hurtboxTable.SelectedIdx; hbIdx >= 0 {
				hbp := a.hurtboxList()
				if hbp != nil && hbIdx < len(*hbp) {
					hb := &(*hbp)[hbIdx]
					hb.Width = math.Round((hb.Width+wy*0.5)*100) / 100
					if hb.Width < 1 {
						hb.Width = 1
					}
					hb.Height = math.Round((hb.Height+wy*0.5)*100) / 100
					if hb.Height < 1 {
						hb.Height = 1
					}
					a.props[2].SetNumeric(hb.Width)
					a.props[3].SetNumeric(hb.Height)
				}
			} else if sel := a.spriteTable.SelectedIdx; sel >= 0 && sel < len(a.proj.Sprites) {
				row := &a.proj.Sprites[sel]
				row.ScaleX = math.Round((row.ScaleX+wy*0.05)*100) / 100
				row.ScaleY = math.Round((row.ScaleY+wy*0.05)*100) / 100
				a.props[3].SetNumeric(row.ScaleX)
				a.props[4].SetNumeric(row.ScaleY)
			}
		case shift:
			if hbIdx := a.hurtboxTable.SelectedIdx; hbIdx >= 0 {
				hbp := a.hurtboxList()
				if hbp != nil && hbIdx < len(*hbp) {
					hb := &(*hbp)[hbIdx]
					hb.Rotation = math.Round((hb.Rotation+wy)*100) / 100
					a.props[4].SetNumeric(hb.Rotation)
				}
			} else if sel := a.spriteTable.SelectedIdx; sel >= 0 && sel < len(a.proj.Sprites) {
				row := &a.proj.Sprites[sel]
				row.Rotation += wy
				a.props[2].SetNumeric(row.Rotation)
			}
		default:
			amt := 1.0 + wy*0.1
			a.canvas.Cam.ZoomAt(amt, float64(mx-cx), float64(my-cy2))
		}
	}
}

func (a *EditorApp) handleCanvasMouse(mx, my int, leftDown, justPressed bool) {
	if !leftDown {
		a.scaleHandle = -1
	}
	if !leftDown {
		return
	}

	cx, cy2, cw, ch := a.canvas.X, a.canvas.Y, a.canvas.Width, a.canvas.Height
	if mx < cx || mx > cx+cw || my < cy2 || my > cy2+ch {
		return
	}

	dx := float64(mx - a.prevMouseX)
	dy := float64(my - a.prevMouseY)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)

	if ctrl && a.canvas.Selection.Visible {
		if justPressed {
			if h := a.canvas.HandleHitTest(mx, my); h >= 0 {
				a.saveSnapshot()
				a.scaleHandle = h
				a.initScaleOrig(h)
			}
		}
		if a.scaleHandle >= 0 {
			a.applyHandleScale(mx, my)
		} else {
			a.canvas.Cam.Pan(dx, dy)
		}
	} else if alt {
		if justPressed {
			a.saveSnapshot()
		}
		if hbIdx := a.hurtboxTable.SelectedIdx; hbIdx >= 0 {
			hbp := a.hurtboxList()
			if hbp != nil && hbIdx < len(*hbp) {
				sel := a.spriteTable.SelectedIdx
				sx, sy := 1.0, 1.0
				if sel >= 0 && sel < len(a.proj.Sprites) {
					sx = a.proj.Sprites[sel].ScaleX
					sy = a.proj.Sprites[sel].ScaleY
				}
				hb := &(*hbp)[hbIdx]
				rot := 0.0
				if sel >= 0 && sel < len(a.proj.Sprites) {
					rot = a.proj.Sprites[sel].Rotation * math.Pi / 180
				}
				cos := math.Cos(rot)
				sin := math.Sin(rot)
				wdx := dx / a.canvas.Cam.Zoom
				wdy := dy / a.canvas.Cam.Zoom
				ldx := (wdx*cos + wdy*sin) / sx
				ldy := (-wdx*sin + wdy*cos) / sy
				hb.X = math.Round((hb.X+ldx)*100) / 100
				hb.Y = math.Round((hb.Y+ldy)*100) / 100
				a.props[0].SetNumeric(hb.X)
				a.props[1].SetNumeric(hb.Y)
				a.syncHurtboxBtns()
			}
		} else if sel := a.spriteTable.SelectedIdx; sel >= 0 && sel < len(a.proj.Sprites) {
			row := &a.proj.Sprites[sel]
			row.OffsetX = math.Round(row.OffsetX + dx/a.canvas.Cam.Zoom)
			row.OffsetY = math.Round(row.OffsetY + dy/a.canvas.Cam.Zoom)
			a.props[0].SetNumeric(row.OffsetX)
			a.props[1].SetNumeric(row.OffsetY)
		}
	} else {
		a.canvas.Cam.Pan(dx, dy)
	}
}

func (a *EditorApp) updateHandleHighlight(mx, my int) {
	ctrlHeld := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrlHeld && a.canvas.Selection.Visible {
		cx, cy2, cw, ch := a.canvas.X, a.canvas.Y, a.canvas.Width, a.canvas.Height
		if mx >= cx && mx <= cx+cw && my >= cy2 && my <= cy2+ch {
			a.canvas.HighlightedHandle = a.canvas.HandleHitTest(mx, my)
		} else {
			a.canvas.HighlightedHandle = -1
		}
	} else {
		a.canvas.HighlightedHandle = -1
	}
}

func (a *EditorApp) handleResetView(mx, my int, justPressed bool) {
	if justPressed &&
		mx >= a.win.resetViewBtnX && mx <= a.win.resetViewBtnX+a.win.resetViewBtnW &&
		my >= a.win.resetViewBtnY && my <= a.win.resetViewBtnY+a.win.resetViewBtnH {
		a.canvas.Cam.Reset()
	}
}

func (a *EditorApp) dispatchWheel(mx, my int, wy float64) {
	ty := a.topPanelY()
	amt := int(wy * scrollStep)
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	if (ctrl || shift) && !a.wheelChanged {
		a.saveSnapshot()
		a.wheelChanged = true
	}
	if !ctrl && !shift {
		a.wheelChanged = false
	}
	if my >= ty && my < ty+a.topPanelH {
		tables := []*ui.Table{a.animTable, a.spriteTable, a.hurtboxTable, a.hitboxTable}
		for _, tbl := range tables {
			if tbl.HitTest(mx, my) {
				tbl.ScrollBy(amt)
				break
			}
		}
	} else if mx >= a.win.outsideWidth-rightPanelW && my >= a.rightPanelY() && my < a.rightPanelY()+a.rightPanelH() {
		a.rp.scroll -= int(wy * scrollStep)
		maxScroll := a.rp.contentH - a.rightPanelH() + scrollMargin
		if maxScroll < 0 {
			maxScroll = 0
		}
		if a.rp.scroll < 0 {
			a.rp.scroll = 0
		}
		if a.rp.scroll > maxScroll {
			a.rp.scroll = maxScroll
		}
	} else {
		a.handleCanvasWheel(mx, my, wy, ctrl, shift)
	}
}
