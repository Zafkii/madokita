package editor

import (
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (a *EditorApp) handleTopPanelMouse(mx, my int, justL bool) {
	ty := a.topPanelY()
	if my < ty || my >= ty+a.topPanelH {
		return
	}

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)

	var vs, ve int
	if !ctrl {
		for _, tbl := range []*ui.Table{a.animTable, a.spriteTable, a.hurtboxTable, a.hitboxTable} {
			tbl.HandleTitleRowMouse(mx, my, justL)
		}

		vs, ve = a.animTable.VisibleRange()
		for i := vs; i < ve; i++ {
			if i < len(a.animNameInputs) {
				a.animNameInputs[i].HandleMouse(mx, my, justL)
			}
			a.animAddFrameBtns[i].HandleMouse(mx, my, justL)
			a.animRemoveFrameBtns[i].HandleMouse(mx, my, justL)
			a.animFramePrevBtns[i].HandleMouse(mx, my, justL)
			a.animFrameNextBtns[i].HandleMouse(mx, my, justL)
			if i < len(a.animFrameInputs) {
				a.animFrameInputs[i].HandleMouse(mx, my, justL)
			}
		}
		vs, ve = a.spriteTable.VisibleRange()
		for i := vs; i < ve; i++ {
			a.spriteBrowseBtns[i].HandleMouse(mx, my, justL)
			a.spriteFramePrevBtns[i].HandleMouse(mx, my, justL)
			a.spriteFrameNextBtns[i].HandleMouse(mx, my, justL)
			if i < len(a.spriteWidthInputs) {
				a.spriteWidthInputs[i].HandleMouse(mx, my, justL)
			}
			if i < len(a.spriteHeightInputs) {
				a.spriteHeightInputs[i].HandleMouse(mx, my, justL)
			}
		}
		vs, ve = a.hurtboxTable.VisibleRange()
		for i := vs; i < ve; i++ {
			if i < len(a.hurtboxXInputs) {
				a.hurtboxXInputs[i].HandleMouse(mx, my, justL)
			}
			if i < len(a.hurtboxYInputs) {
				a.hurtboxYInputs[i].HandleMouse(mx, my, justL)
			}
			if i < len(a.hurtboxWidthInputs) {
				a.hurtboxWidthInputs[i].HandleMouse(mx, my, justL)
			}
			if i < len(a.hurtboxHeightInputs) {
				a.hurtboxHeightInputs[i].HandleMouse(mx, my, justL)
			}
			if i < len(a.hurtboxDmgMultInputs) {
				a.hurtboxDmgMultInputs[i].HandleMouse(mx, my, justL)
			}
		}
		vs, ve = a.hitboxTable.VisibleRange()
		for i := vs; i < ve; i++ {
			if i < len(a.hitboxWidthInputs) {
				a.hitboxWidthInputs[i].HandleMouse(mx, my, justL)
			}
			if i < len(a.hitboxHeightInputs) {
				a.hitboxHeightInputs[i].HandleMouse(mx, my, justL)
			}
		}
	}

	a.hoveredFilePath = ""
	for i := vs; i < ve; i++ {
		btn := a.spriteBrowseBtns[i]
		if mx >= btn.X && mx <= btn.X+btn.W && my >= btn.Y && my <= btn.Y+btn.H {
			if i < len(a.proj.Sprites) && a.proj.Sprites[i].File != "" {
				a.hoveredFilePath = a.proj.Sprites[i].File
			}
			break
		}
	}

	if !justL {
		return
	}

	if idx := a.animTable.HitRow(mx, my); idx >= 0 {
		a.animTable.SelectedIdx = idx
		a.hurtboxAnimCtx = idx
		a.spriteTable.SelectedIdx = -1
		a.hurtboxTable.SelectedIdx = -1
		a.hitboxTable.SelectedIdx = -1
	} else if idx := a.spriteTable.HitRow(mx, my); idx >= 0 {
		a.spriteTable.SelectedIdx = idx
		a.animTable.SelectedIdx = -1
		a.hurtboxTable.SelectedIdx = -1
		a.hitboxTable.SelectedIdx = -1
	} else if idx := a.hurtboxTable.HitRow(mx, my); idx >= 0 {
		a.hurtboxTable.SelectedIdx = idx
		a.animTable.SelectedIdx = -1
		a.spriteTable.SelectedIdx = -1
		a.hitboxTable.SelectedIdx = -1
	} else if idx := a.hitboxTable.HitRow(mx, my); idx >= 0 {
		a.hitboxTable.SelectedIdx = idx
		a.animTable.SelectedIdx = -1
		a.spriteTable.SelectedIdx = -1
		a.hurtboxTable.SelectedIdx = -1
	}
}

func (a *EditorApp) handleRightPanelMouse(mx, my int, justL bool) {
	for i := range a.props {
		a.props[i].HandleMouse(mx, my, justL)
	}
	a.phaseDropdown.HandleMouse(mx, my, justL)

	for i := range a.originInputs {
		a.originInputs[i].HandleMouse(mx, my, justL)
	}
	a.baseRotInput.HandleMouse(mx, my, justL)

	a.loopInput.HandleMouse(mx, my, justL)

	for i := range a.atkTimingInputs {
		a.atkTimingInputs[i].HandleMouse(mx, my, justL)
	}
	a.fpsInput.HandleMouse(mx, my, justL)

	a.prev.speedSlider.HandleMouse(mx, my, justL)
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		a.prev.speedSlider.HandleRelease()
	}

	btnH := 22

	a.prev.previewPlayHovered = mx >= a.prev.previewPlayX && mx <= a.prev.previewPlayX+28 &&
		my >= a.prev.previewPlayY && my <= a.prev.previewPlayY+btnH

	chk := a.prev.previewChkW > 0
	a.prev.previewLoopHovered = chk && mx >= a.prev.previewChkX && mx <= a.prev.previewChkX+a.prev.previewChkW &&
		my >= a.prev.previewChkY && my <= a.prev.previewChkY+a.prev.previewChkH

	btnSize := 20
	a.prev.previewMinusHovered = mx >= a.prev.previewMinusX && mx <= a.prev.previewMinusX+btnSize &&
		my >= a.prev.previewMinusY && my <= a.prev.previewMinusY+btnSize

	a.prev.previewPlusHovered = mx >= a.prev.previewPlusX && mx <= a.prev.previewPlusX+btnSize &&
		my >= a.prev.previewPlusY && my <= a.prev.previewPlusY+btnSize

	if justL {
		if mx >= a.prev.previewPlayX && mx <= a.prev.previewPlayX+28 &&
			my >= a.prev.previewPlayY && my <= a.prev.previewPlayY+btnH {
			a.prev.previewPlaying = !a.prev.previewPlaying
		}

		if chk && mx >= a.prev.previewChkX && mx <= a.prev.previewChkX+a.prev.previewChkW &&
			my >= a.prev.previewChkY && my <= a.prev.previewChkY+a.prev.previewChkH {
			a.prev.loopChecked = !a.prev.loopChecked
		}

		if mx >= a.prev.previewMinusX && mx <= a.prev.previewMinusX+btnSize &&
			my >= a.prev.previewMinusY && my <= a.prev.previewMinusY+btnSize {
			a.prev.previewSpeed -= 0.1
			if a.prev.previewSpeed < 0.1 {
				a.prev.previewSpeed = 0.1
			}
		}

		if mx >= a.prev.previewPlusX && mx <= a.prev.previewPlusX+btnSize &&
			my >= a.prev.previewPlusY && my <= a.prev.previewPlusY+btnSize {
			a.prev.previewSpeed += 0.1
			if a.prev.previewSpeed > 5.0 {
				a.prev.previewSpeed = 5.0
			}
		}
	}
}

func (a *EditorApp) handleRightPanelKeys() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		return
	}
	inputs := a.allInputs()
	current := -1
	for i, inp := range inputs {
		if inp.Focused {
			current = i
			inp.Focused = false
			break
		}
	}
	if current == -1 {
		return
	}
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	next := current + 1
	if shift {
		next = current - 1
	}
	if next < 0 {
		next = len(inputs) - 1
	}
	if next >= len(inputs) {
		next = 0
	}
	inputs[next].Focused = true
}

func (a *EditorApp) handleScrollWheel(mx, my int, yoff float64) {
	vs, ve := a.animTable.VisibleRange()
	for i := vs; i < ve; i++ {
		x0 := a.animFramePrevBtns[i].X
		x1 := a.animFrameNextBtns[i].X + a.animFrameNextBtns[i].W
		y0 := a.animFramePrevBtns[i].Y
		y1 := y0 + a.animFramePrevBtns[i].H
		if mx >= x0 && mx < x1 && my >= y0 && my < y1 {
			if yoff > 0 {
				a.animFrameNextBtns[i].OnClick()
			} else {
				a.animFramePrevBtns[i].OnClick()
			}
			return
		}
	}
	vs, ve = a.spriteTable.VisibleRange()
	for i := vs; i < ve; i++ {
		x0 := a.spriteFramePrevBtns[i].X
		x1 := a.spriteFrameNextBtns[i].X + a.spriteFrameNextBtns[i].W
		y0 := a.spriteFramePrevBtns[i].Y
		y1 := y0 + a.spriteFramePrevBtns[i].H
		if mx >= x0 && mx < x1 && my >= y0 && my < y1 {
			if yoff > 0 {
				a.spriteFrameNextBtns[i].OnClick()
			} else {
				a.spriteFramePrevBtns[i].OnClick()
			}
			return
		}
	}
}
