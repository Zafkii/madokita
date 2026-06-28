package editor

import (
	"animprite/internal/project"
	"animprite/internal/ui"
)

func (a *EditorApp) syncAnimBtns() {
	n := len(a.proj.Animations)
	for len(a.animNameInputs) < n {
		inp := ui.NewTextInput(0, 0, 100, topPanelBtnH, a.th)
		a.animNameInputs = append(a.animNameInputs, inp)
	}
	a.animNameInputs = a.animNameInputs[:n]
	for idx := 0; idx < n; idx++ {
		i := idx
		a.animNameInputs[i].Text = a.proj.Animations[i].Name
		a.animNameInputs[i].OnChange = func(text string) {
			a.proj.Animations[i].Name = text
		}
	}
	for len(a.animAddFrameBtns) < n {
		a.animAddFrameBtns = append(a.animAddFrameBtns, ui.NewButton(0, 0, 30, topPanelBtnH, "+add", a.th))
		a.animRemoveFrameBtns = append(a.animRemoveFrameBtns, ui.NewButton(0, 0, 35, topPanelBtnH, "- del", a.th))
		a.animFramePrevBtns = append(a.animFramePrevBtns, ui.NewButton(0, 0, 20, topPanelBtnH, "\U000F04AE", a.th))
		a.animFrameNextBtns = append(a.animFrameNextBtns, ui.NewButton(0, 0, 20, topPanelBtnH, "\U000F04AD", a.th))
	}
	a.animAddFrameBtns = a.animAddFrameBtns[:n]
	a.animRemoveFrameBtns = a.animRemoveFrameBtns[:n]
	a.animFramePrevBtns = a.animFramePrevBtns[:n]
	a.animFrameNextBtns = a.animFrameNextBtns[:n]
	for len(a.animFrameInputs) < n {
		inp := ui.NewTextInput(0, 0, 24, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.HasMax = true
		inp.Step = 1
		a.animFrameInputs = append(a.animFrameInputs, inp)
	}
	a.animFrameInputs = a.animFrameInputs[:n]
	if a.animTable.SelectedIdx >= n {
		a.animTable.SelectedIdx = n - 1
	}
	if n == 0 || a.animTable.SelectedIdx < 0 {
		a.animTable.SelectedIdx = -1
	}
	if a.animTable.SelectedIdx >= 0 {
		a.hurtboxAnimCtx = a.animTable.SelectedIdx
	}
	for i := range a.proj.Animations {
		if len(a.proj.Animations[i].Frames) == 0 && a.proj.Animations[i].CurrentIdx > 0 {
			count := a.proj.Animations[i].CurrentIdx + 1
			for j := 0; j < count; j++ {
				a.proj.Animations[i].Frames = append(a.proj.Animations[i].Frames, project.AnimationFrame{
					ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5,
				})
			}
		} else if len(a.proj.Animations[i].Frames) == 0 {
			a.proj.Animations[i].Frames = []project.AnimationFrame{{ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5}}
			a.proj.Animations[i].CurrentIdx = 0
		}
		if a.proj.Animations[i].CurrentIdx >= len(a.proj.Animations[i].Frames) {
			a.proj.Animations[i].CurrentIdx = len(a.proj.Animations[i].Frames) - 1
		}
		if a.proj.Animations[i].CurrentIdx < 0 {
			a.proj.Animations[i].CurrentIdx = 0
		}
	}
	for idx := 0; idx < n; idx++ {
		i := idx
		a.animAddFrameBtns[i].OnClick = func() {
			a.saveSnapshot()
			frame := project.AnimationFrame{ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5, Phase: project.PhaseWindup}
			if a.proj.Animations[i].CurrentIdx >= 0 && a.proj.Animations[i].CurrentIdx < len(a.proj.Animations[i].Frames) {
				prev := a.proj.Animations[i].Frames[a.proj.Animations[i].CurrentIdx]
				frame.SpriteIdx = prev.SpriteIdx
				frame.SpriteFrameIdx = prev.SpriteFrameIdx
				frame.Phase = prev.Phase
				frame.OriginX = prev.OriginX
				frame.OriginY = prev.OriginY
			}
			a.proj.Animations[i].Frames = append(a.proj.Animations[i].Frames, frame)
			a.proj.Animations[i].CurrentIdx = len(a.proj.Animations[i].Frames) - 1
			a.animTable.SelectedIdx = i
			a.hurtboxAnimCtx = i
			a.prevSelectedHurtboxIdx = -1
		}
		a.animRemoveFrameBtns[i].OnClick = func() {
			a.saveSnapshot()
			if len(a.proj.Animations[i].Frames) > 0 {
				a.proj.Animations[i].Frames = a.proj.Animations[i].Frames[:len(a.proj.Animations[i].Frames)-1]
				if a.proj.Animations[i].CurrentIdx >= len(a.proj.Animations[i].Frames) {
					a.proj.Animations[i].CurrentIdx = max(0, len(a.proj.Animations[i].Frames)-1)
				}
			}
			a.animTable.SelectedIdx = i
			a.hurtboxAnimCtx = i
			a.prevSelectedHurtboxIdx = -1
		}
		a.animFramePrevBtns[i].OnClick = func() {
			a.saveSnapshot()
			if a.proj.Animations[i].CurrentIdx > 0 {
				a.proj.Animations[i].CurrentIdx--
			}
			a.animTable.SelectedIdx = i
			a.hurtboxAnimCtx = i
			a.prevSelectedHurtboxIdx = -1
		}
		a.animFrameNextBtns[i].OnClick = func() {
			a.saveSnapshot()
			if a.proj.Animations[i].CurrentIdx < len(a.proj.Animations[i].Frames)-1 {
				a.proj.Animations[i].CurrentIdx++
			}
			a.animTable.SelectedIdx = i
			a.hurtboxAnimCtx = i
			a.prevSelectedHurtboxIdx = -1
		}
		a.animFrameInputs[i].Max = float64(len(a.proj.Animations[i].Frames))
		if a.proj.Animations[i].CurrentIdx >= 0 {
			a.animFrameInputs[i].SetNumeric(float64(a.proj.Animations[i].CurrentIdx + 1))
		}
		ii := i
		a.animFrameInputs[i].OnEnter = func() {
			v := a.animFrameInputs[ii].IntValue()
			total := len(a.proj.Animations[ii].Frames)
			if v < 1 {
				v = 1
			}
			if v > total {
				v = total
			}
			a.proj.Animations[ii].CurrentIdx = v - 1
			a.animFrameInputs[ii].SetNumeric(float64(v))
			a.animTable.SelectedIdx = ii
			a.hurtboxAnimCtx = ii
			a.prevSelectedHurtboxIdx = -1
		}
	}
}

func (a *EditorApp) syncSpriteBtns() {
	n := len(a.proj.Sprites)
	for len(a.spriteBrowseBtns) < n {
		a.spriteBrowseBtns = append(a.spriteBrowseBtns, ui.NewButton(0, 0, 50, topPanelBtnH, "Browse", a.th))
		a.spriteFramePrevBtns = append(a.spriteFramePrevBtns, ui.NewButton(0, 0, 20, topPanelBtnH, "\U000F04AE", a.th))
		a.spriteFrameNextBtns = append(a.spriteFrameNextBtns, ui.NewButton(0, 0, 20, topPanelBtnH, "\U000F04AD", a.th))
	}
	for len(a.spriteWidthInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.Step = 1
		a.spriteWidthInputs = append(a.spriteWidthInputs, inp)
	}
	for len(a.spriteHeightInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.Step = 1
		a.spriteHeightInputs = append(a.spriteHeightInputs, inp)
	}
	a.spriteBrowseBtns = a.spriteBrowseBtns[:n]
	a.spriteFramePrevBtns = a.spriteFramePrevBtns[:n]
	a.spriteFrameNextBtns = a.spriteFrameNextBtns[:n]
	a.spriteWidthInputs = a.spriteWidthInputs[:n]
	a.spriteHeightInputs = a.spriteHeightInputs[:n]
	if a.spriteTable.SelectedIdx >= n {
		a.spriteTable.SelectedIdx = n - 1
	}
	if n == 0 || a.spriteTable.SelectedIdx < 0 {
		a.spriteTable.SelectedIdx = -1
	}
	for idx := 0; idx < n; idx++ {
		i := idx
		a.spriteBrowseBtns[i].OnClick = func() {
			a.browseSprite(i)
		}
		a.spriteFramePrevBtns[i].OnClick = func() {
			if a.animTable.SelectedIdx >= 0 && a.animTable.SelectedIdx < len(a.proj.Animations) {
				anim := &a.proj.Animations[a.animTable.SelectedIdx]
				if anim.CurrentIdx >= 0 && anim.CurrentIdx < len(anim.Frames) {
					frame := &anim.Frames[anim.CurrentIdx]
					if frame.SpriteFrameIdx > 0 {
						frame.SpriteFrameIdx--
					}
					frame.SpriteIdx = i
					a.proj.Sprites[i].CurrentIdx = frame.SpriteFrameIdx
					a.frameSpriteDropdown.Selected = i + 1
					a.phaseDropdown.Selected = int(frame.Phase)
					return
				}
			}
			if a.proj.Sprites[i].CurrentIdx > 0 {
				a.proj.Sprites[i].CurrentIdx--
			}
		}
		a.spriteFrameNextBtns[i].OnClick = func() {
			if a.animTable.SelectedIdx >= 0 && a.animTable.SelectedIdx < len(a.proj.Animations) {
				anim := &a.proj.Animations[a.animTable.SelectedIdx]
				if anim.CurrentIdx >= 0 && anim.CurrentIdx < len(anim.Frames) {
					frame := &anim.Frames[anim.CurrentIdx]
					if frame.SpriteFrameIdx < a.proj.Sprites[i].FrameCount-1 {
						frame.SpriteFrameIdx++
					}
					frame.SpriteIdx = i
					a.proj.Sprites[i].CurrentIdx = frame.SpriteFrameIdx
					a.frameSpriteDropdown.Selected = i + 1
					a.phaseDropdown.Selected = int(frame.Phase)
					return
				}
			}
			if a.proj.Sprites[i].CurrentIdx < a.proj.Sprites[i].FrameCount-1 {
				a.proj.Sprites[i].CurrentIdx++
			}
		}
		a.spriteWidthInputs[i].SetNumeric(float64(a.proj.Sprites[i].Width))
		a.spriteWidthInputs[i].OnChange = func(_ string) {
			v := a.spriteWidthInputs[i].IntValue()
			if v < 1 {
				v = 1
			}
			a.proj.Sprites[i].Width = v
			a.recalcSpriteFrameCount(i)
		}
		a.spriteHeightInputs[i].SetNumeric(float64(a.proj.Sprites[i].Height))
		a.spriteHeightInputs[i].OnChange = func(_ string) {
			v := a.spriteHeightInputs[i].IntValue()
			if v < 1 {
				v = 1
			}
			a.proj.Sprites[i].Height = v
			a.recalcSpriteFrameCount(i)
		}
	}
	if a.frameSpriteDropdown != nil {
		opts := make([]string, 0, len(a.proj.Sprites)+1)
		opts = append(opts, "(none)")
		for _, s := range a.proj.Sprites {
			opts = append(opts, s.Name)
		}
		a.frameSpriteDropdown.Options = opts
	}
}

func (a *EditorApp) hurtboxList() *[]project.HurtboxRow {
	idx := a.animTable.SelectedIdx
	if idx < 0 {
		idx = a.hurtboxAnimCtx
	}
	if idx >= 0 && idx < len(a.proj.Animations) {
		anim := &a.proj.Animations[idx]
		if anim.CurrentIdx >= 0 && anim.CurrentIdx < len(anim.Frames) {
			return &anim.Frames[anim.CurrentIdx].Hurtboxes
		}
	}
	return nil
}

func (a *EditorApp) syncHurtboxBtns() {
	hbListPtr := a.hurtboxList()
	var hbList []project.HurtboxRow
	if hbListPtr != nil {
		hbList = *hbListPtr
	}
	n := len(hbList)
	for len(a.hurtboxXInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = -99999
		inp.Step = 0.5
		a.hurtboxXInputs = append(a.hurtboxXInputs, inp)
	}
	for len(a.hurtboxYInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = -99999
		inp.Step = 0.5
		a.hurtboxYInputs = append(a.hurtboxYInputs, inp)
	}
	for len(a.hurtboxWidthInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.Step = 1
		a.hurtboxWidthInputs = append(a.hurtboxWidthInputs, inp)
	}
	for len(a.hurtboxHeightInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.Step = 1
		a.hurtboxHeightInputs = append(a.hurtboxHeightInputs, inp)
	}
	for len(a.hurtboxDmgMultInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 0
		inp.Step = 0.1
		a.hurtboxDmgMultInputs = append(a.hurtboxDmgMultInputs, inp)
	}
	a.hurtboxXInputs = a.hurtboxXInputs[:n]
	a.hurtboxYInputs = a.hurtboxYInputs[:n]
	a.hurtboxWidthInputs = a.hurtboxWidthInputs[:n]
	a.hurtboxHeightInputs = a.hurtboxHeightInputs[:n]
	a.hurtboxDmgMultInputs = a.hurtboxDmgMultInputs[:n]
	for idx := 0; idx < n; idx++ {
		i := idx
		a.hurtboxXInputs[i].SetNumeric(hbList[i].X)
		a.hurtboxXInputs[i].OnChange = func(_ string) {
			v := a.hurtboxXInputs[i].NumericValue()
			if hbp := a.hurtboxList(); hbp != nil && i < len(*hbp) {
				(*hbp)[i].X = v
			}
		}
		a.hurtboxYInputs[i].SetNumeric(hbList[i].Y)
		a.hurtboxYInputs[i].OnChange = func(_ string) {
			v := a.hurtboxYInputs[i].NumericValue()
			if hbp := a.hurtboxList(); hbp != nil && i < len(*hbp) {
				(*hbp)[i].Y = v
			}
		}
		a.hurtboxWidthInputs[i].SetNumeric(hbList[i].Width)
		a.hurtboxWidthInputs[i].OnChange = func(_ string) {
			v := a.hurtboxWidthInputs[i].NumericValue()
			if hbp := a.hurtboxList(); hbp != nil && i < len(*hbp) {
				(*hbp)[i].Width = v
			}
		}
		a.hurtboxHeightInputs[i].SetNumeric(hbList[i].Height)
		a.hurtboxHeightInputs[i].OnChange = func(_ string) {
			v := a.hurtboxHeightInputs[i].NumericValue()
			if hbp := a.hurtboxList(); hbp != nil && i < len(*hbp) {
				(*hbp)[i].Height = v
			}
		}
		a.hurtboxDmgMultInputs[i].SetNumeric(hbList[i].DmgMult)
		a.hurtboxDmgMultInputs[i].OnChange = func(_ string) {
			v := a.hurtboxDmgMultInputs[i].NumericValue()
			if hbp := a.hurtboxList(); hbp != nil && i < len(*hbp) {
				(*hbp)[i].DmgMult = v
			}
		}
	}
	if n == 0 {
		a.hurtboxTable.SelectedIdx = -1
	} else if a.hurtboxTable.SelectedIdx >= n {
		a.hurtboxTable.SelectedIdx = n - 1
	}
}

func (a *EditorApp) syncHitboxBtns() {
	n := len(a.proj.HitDefs)
	for len(a.hitboxWidthInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.Step = 1
		a.hitboxWidthInputs = append(a.hitboxWidthInputs, inp)
	}
	for len(a.hitboxHeightInputs) < n {
		inp := ui.NewTextInput(0, 0, 50, topPanelBtnH, a.th)
		inp.Numeric = true
		inp.Min = 1
		inp.Step = 1
		a.hitboxHeightInputs = append(a.hitboxHeightInputs, inp)
	}
	a.hitboxWidthInputs = a.hitboxWidthInputs[:n]
	a.hitboxHeightInputs = a.hitboxHeightInputs[:n]
	for idx := 0; idx < n; idx++ {
		i := idx
		a.hitboxWidthInputs[i].SetNumeric(float64(a.proj.HitDefs[i].Width))
		a.hitboxWidthInputs[i].OnChange = func(_ string) {
			a.proj.HitDefs[i].Width = a.hitboxWidthInputs[i].NumericValue()
		}
		a.hitboxHeightInputs[i].SetNumeric(float64(a.proj.HitDefs[i].Height))
		a.hitboxHeightInputs[i].OnChange = func(_ string) {
			a.proj.HitDefs[i].Height = a.hitboxHeightInputs[i].NumericValue()
		}
	}
}

func (a *EditorApp) recalcSpriteFrameCount(idx int) {
	row := &a.proj.Sprites[idx]
	if row.Width <= 0 || row.Height <= 0 {
		row.FrameCount = 1
		return
	}
	img, ok := a.loadedSprites[idx]
	if !ok || img == nil {
		return
	}
	b := img.Bounds()
	cols := b.Dx() / row.Width
	rows2 := b.Dy() / row.Height
	if cols < 1 {
		cols = 1
	}
	if rows2 < 1 {
		rows2 = 1
	}
	row.FrameCount = cols * rows2
	if row.CurrentIdx >= row.FrameCount {
		row.CurrentIdx = row.FrameCount - 1
	}
}

func (a *EditorApp) computeTopPanelHeight() {
	a.animTable.RowCount = len(a.proj.Animations)
	a.spriteTable.RowCount = len(a.proj.Sprites)
	hbCount := 0
	if hbp := a.hurtboxList(); hbp != nil {
		hbCount = len(*hbp)
	}
	a.hurtboxTable.RowCount = hbCount
	a.hitboxTable.RowCount = len(a.proj.HitDefs)

	ah := a.animTable.Height()
	sh := a.spriteTable.Height()
	hbh := a.hurtboxTable.Height()
	hth := a.hitboxTable.Height()

	a.topPanelH = max(topPanelMinH, max(ah+sh, hbh+hth)+topPanelPad)
}

func (a *EditorApp) syncLayout() {
	w, h := a.win.outsideWidth, a.win.outsideHeight
	a.computeTopPanelHeight()

	canvasTop := titleBarH + modeIndicatorH + a.topPanelH
	a.canvas.X = 0
	a.canvas.Y = canvasTop
	a.canvas.Width = w - rightPanelW
	a.canvas.Height = h - canvasTop - statusbarH

	groupH := dropdownH + btnGap + rightBtnH + btnGap + rightBtnH + btnGap + rightBtnH
	groupTop := titleBarH + modeIndicatorH + (a.topPanelH-groupH)/2
	a.modeDropdown.Y = groupTop

	btnY0 := groupTop + dropdownH + btnGap
	btnY1 := btnY0 + rightBtnH + btnGap
	btnY2 := btnY1 + rightBtnH + btnGap
	a.openBtn.X = panelPad
	a.openBtn.Y = btnY0
	a.saveBtn.X = panelPad
	a.saveBtn.Y = btnY1
	a.themeBtn.X = panelPad
	a.themeBtn.Y = btnY2

	px := w - rightPanelW + rightPanelPad
	for _, inp := range a.props {
		inp.X = px
	}
	a.phaseDropdown.X = px
	a.frameSpriteDropdown.X = px
	for _, inp := range a.originInputs {
		inp.X = px
	}
	a.baseRotInput.X = px
	a.assetNameInput.X = px
	a.assetKeyInput.X = px
	for _, inp := range a.defaultOriginInput {
		inp.X = px
	}
	a.loopInput.X = px
	for _, inp := range a.atkTimingInputs {
		inp.X = px
	}
	a.fpsInput.X = px
}

func (a *EditorApp) syncMovementInputs() {
	a.assetNameInput.Text = a.proj.AssetName
	a.assetKeyInput.Text = a.proj.AssetKey
	a.defaultOriginInput[0].SetNumeric(a.proj.DefaultOriginX)
	a.defaultOriginInput[1].SetNumeric(a.proj.DefaultOriginY)
	if animIdx := a.animTable.SelectedIdx; animIdx >= 0 && animIdx < len(a.proj.Animations) {
		anim := &a.proj.Animations[animIdx]
		a.loopInput.Text = "false"
		if anim.Loop {
			a.loopInput.Text = "true"
		}
	}
}
