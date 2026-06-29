package editor

import (
	"strings"

	"animprite/internal/project"
	"animprite/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (a *EditorApp) Update() error {
	mx, my := ebiten.CursorPosition()
	justL := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	if ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) {
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				a.redo()
			} else {
				a.undo()
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyY) {
			a.redo()
		}
	}

	a.hoveredFilePath = ""

	a.syncTitleBarScale()
	leftDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	justPressed := leftDown && !a.win.prevLeftBtn
	isMaxed := ebiten.IsWindowMaximized()

	skipMouse, err := a.handleWindowChrome(mx, my, leftDown, justPressed, isMaxed)
	if err != nil {
		return err
	}
	a.win.prevLeftBtn = leftDown

	_, wy := ebiten.Wheel()
	if wy != 0 {
		a.dispatchWheel(mx, my, wy)
	}

	a.handleCanvasMouse(mx, my, leftDown, justPressed)
	a.updateHandleHighlight(mx, my)

	if !skipMouse {
		popupHit := a.modeDropdown.PopupHit(mx, my)
		a.themeBtn.HandleMouse(mx, my, justL && !popupHit)
		a.openBtn.HandleMouse(mx, my, justL && !popupHit)
		a.saveBtn.HandleMouse(mx, my, justL && !popupHit)
		a.modeDropdown.HandleMouse(mx, my, justL)

		a.handleTopPanelMouse(mx, my, justL)
		a.handleRightPanelMouse(mx, my, justL)
		_, yoff := ebiten.Wheel()
		if yoff != 0 {
			a.handleScrollWheel(mx, my, yoff)
		}
	}

	a.syncAnimFrameSelection()
	a.syncSpriteSelection()
	a.syncHurtboxSelection()

	if !skipMouse {
		a.handleDragSelect()
	}

	a.handleInputUpdate()
	a.handleRightPanelKeys()

	a.updateWindowCursor(isMaxed)
	a.handleWindowRestoreState()
	a.handleResetView(mx, my, justPressed)

	if a.statusTime > 0 {
		a.statusTime--
		if a.statusTime == 0 {
			a.statusMsg = ""
		}
	}

	a.flushInputsToData()
	a.advancePreview()

	a.prevMouseX = mx
	a.prevMouseY = my
	return nil
}

func (a *EditorApp) flushInputsToData() {
	sel := a.spriteTable.SelectedIdx

	if animIdx := a.animTable.SelectedIdx; animIdx >= 0 && animIdx < len(a.proj.Animations) {
		anim := &a.proj.Animations[animIdx]
		anim.Windup = a.atkTimingInputs[0].NumericValue()
		anim.Active = a.atkTimingInputs[1].NumericValue()
		anim.Recover = a.atkTimingInputs[2].NumericValue()
		anim.Armed = a.atkTimingInputs[3].NumericValue()
		anim.ArmedFPS = a.atkTimingInputs[4].NumericValue()
		anim.FPS = a.fpsInput.NumericValue()
		anim.Loop = strings.EqualFold(a.loopInput.Text, "true")
	}

	if a.panelMode == panelModeHurtbox && a.hurtboxTable.SelectedIdx >= 0 {
		hbp := a.hurtboxList()
		if hbp != nil && a.hurtboxTable.SelectedIdx < len(*hbp) {
			hb := &(*hbp)[a.hurtboxTable.SelectedIdx]
			hb.X = a.props[0].NumericValue()
			hb.Y = a.props[1].NumericValue()
			hb.Width = a.props[2].NumericValue()
			hb.Height = a.props[3].NumericValue()
			hb.Rotation = a.props[4].NumericValue()
		}
	} else if a.panelMode == panelModeAnimFrame {
		if !a.prev.previewPlaying {
			if entry := a.currentFrameSpriteEntry(); entry != nil {
				entry.OffsetX = a.props[0].NumericValue()
				entry.OffsetY = a.props[1].NumericValue()
				entry.Rotation = a.props[2].NumericValue()
				entry.ScaleX = a.props[3].NumericValue()
				entry.ScaleY = a.props[4].NumericValue()
				entry.OriginX = a.originInputs[0].NumericValue()
				entry.OriginY = a.originInputs[1].NumericValue()
			}
		}
	} else if sel >= 0 && sel < len(a.proj.Sprites) {
		row := &a.proj.Sprites[sel]
		row.OffsetX = a.props[0].NumericValue()
		row.OffsetY = a.props[1].NumericValue()
		row.Rotation = a.props[2].NumericValue()
		row.ScaleX = a.props[3].NumericValue()
		row.ScaleY = a.props[4].NumericValue()
		row.OriginX = a.originInputs[0].NumericValue()
		row.OriginY = a.originInputs[1].NumericValue()
	}
}

func (a *EditorApp) handleInputUpdate() {
	runes := ebiten.AppendInputChars(nil)
	focused := a.focusedInput()
	if focused == nil {
		return
	}
	focused.HandleKeys()
	if len(runes) > 0 {
		focused.HandleRunes(runes)
	}
}

func (a *EditorApp) focusedInput() *ui.TextInput {
	avs, ave := a.animTable.VisibleRange()
	for i := avs; i < ave; i++ {
		if i < len(a.animNameInputs) && a.animNameInputs[i].Focused {
			return a.animNameInputs[i]
		}
	}
	for i := avs; i < ave; i++ {
		if i < len(a.animFrameInputs) && a.animFrameInputs[i].Focused {
			return a.animFrameInputs[i]
		}
	}
	vs, ve := a.spriteTable.VisibleRange()
	for i := vs; i < ve; i++ {
		if i < len(a.spriteWidthInputs) && a.spriteWidthInputs[i].Focused {
			return a.spriteWidthInputs[i]
		}
		if i < len(a.spriteHeightInputs) && a.spriteHeightInputs[i].Focused {
			return a.spriteHeightInputs[i]
		}
	}
	hvs, hve := a.hurtboxTable.VisibleRange()
	for i := hvs; i < hve; i++ {
		if i < len(a.hurtboxXInputs) && a.hurtboxXInputs[i].Focused {
			return a.hurtboxXInputs[i]
		}
		if i < len(a.hurtboxYInputs) && a.hurtboxYInputs[i].Focused {
			return a.hurtboxYInputs[i]
		}
		if i < len(a.hurtboxWidthInputs) && a.hurtboxWidthInputs[i].Focused {
			return a.hurtboxWidthInputs[i]
		}
		if i < len(a.hurtboxHeightInputs) && a.hurtboxHeightInputs[i].Focused {
			return a.hurtboxHeightInputs[i]
		}
		if i < len(a.hurtboxDmgMultInputs) && a.hurtboxDmgMultInputs[i].Focused {
			return a.hurtboxDmgMultInputs[i]
		}
	}
	htvs, htve := a.hitboxTable.VisibleRange()
	for i := htvs; i < htve; i++ {
		if i < len(a.hitboxWidthInputs) && a.hitboxWidthInputs[i].Focused {
			return a.hitboxWidthInputs[i]
		}
		if i < len(a.hitboxHeightInputs) && a.hitboxHeightInputs[i].Focused {
			return a.hitboxHeightInputs[i]
		}
	}
	for i := range a.props {
		if a.props[i].Focused {
			return a.props[i]
		}
	}
	for i := range a.originInputs {
		if a.originInputs[i].Focused {
			return a.originInputs[i]
		}
	}
	if a.baseRotInput.Focused {
		return a.baseRotInput
	}
	if a.loopInput.Focused {
		return a.loopInput
	}
	for i := range a.atkTimingInputs {
		if a.atkTimingInputs[i].Focused {
			return a.atkTimingInputs[i]
		}
	}
	if a.fpsInput.Focused {
		return a.fpsInput
	}
	return nil
}

func (a *EditorApp) allInputs() []*ui.TextInput {
	all := make([]*ui.TextInput, 0, 25)
	all = append(all, a.props[:]...)
	all = append(all, a.originInputs[:]...)
	all = append(all, a.baseRotInput)
	all = append(all, a.loopInput)
	all = append(all, a.atkTimingInputs[:]...)
	all = append(all, a.fpsInput)
	return all
}

func (a *EditorApp) previewY() int {
	rpY := a.rightPanelY()
	iy := rpY + 4 - a.rp.scroll

	iy += 24 + 5*(rightPanelInputH+2) + 4
	if a.isAttackMode() {
		iy += rightPanelInputH + 2
	}
	if a.hurtboxTable.SelectedIdx < 0 {
		iy += 24 + 3*(rightPanelInputH+2) + 4
	}
	iy += 24 + 18
	if a.isAttackMode() {
		iy += 16 + 5*(rightPanelInputH+2)
	} else {
		iy += rightPanelInputH + 2
	}
	iy += rightPanelInputH + 2
	iy += 24 + 4*(rightPanelInputH+2) + 4

	return iy
}

func (a *EditorApp) advancePreview() {
	if !a.prev.previewPlaying {
		return
	}
	animIdx := a.animTable.SelectedIdx
	if animIdx < 0 || animIdx >= len(a.proj.Animations) {
		a.prev.previewPlaying = false
		return
	}
	anim := &a.proj.Animations[animIdx]
	if len(anim.Frames) == 0 {
		a.prev.previewPlaying = false
		return
	}

	if animIdx != a.prev.previewAnimIdx {
		a.prev.previewAnimIdx = animIdx
		a.prev.previewAccumulator = 0
	}

	tps := ebiten.ActualTPS()
	if tps <= 0 {
		tps = 60
	}
	dt := 1.0 / tps
	a.prev.previewAccumulator += dt

	for {
		var frameDur float64
		if a.isAttackMode() {
			frame := anim.Frames[anim.CurrentIdx]
			switch frame.Phase {
			case project.PhaseWindup:
				frameDur = anim.Windup
			case project.PhaseActive:
				frameDur = anim.Active
			case project.PhaseRecover:
				frameDur = anim.Recover
			case project.PhaseArmed:
				frameDur = anim.Armed
			default:
				frameDur = anim.Windup
			}
			frameDur /= 1000.0
		} else {
			if anim.FPS <= 0 {
				return
			}
			frameDur = 1.0 / anim.FPS
		}

		frameDur /= a.prev.previewSpeed
		if frameDur <= 0 {
			return
		}

		if a.prev.previewAccumulator < frameDur {
			break
		}
		a.prev.previewAccumulator -= frameDur

		if anim.CurrentIdx+1 >= len(anim.Frames) {
			if a.prev.loopChecked {
				anim.CurrentIdx = 0
			} else {
				anim.CurrentIdx = len(anim.Frames) - 1
				a.prev.previewPlaying = false
				a.prev.previewAccumulator = 0
				return
			}
		} else {
			anim.CurrentIdx++
		}
	}
}
