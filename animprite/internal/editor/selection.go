package editor

func (a *EditorApp) syncAnimFrameSelection() {
	animIdx := a.animTable.SelectedIdx
	var frameIdx int
	if animIdx >= 0 && animIdx < len(a.proj.Animations) {
		frameIdx = a.proj.Animations[animIdx].CurrentIdx
	}
	if animIdx == a.prevSelectedAnimIdx && frameIdx == a.prevSelectedAnimFrameIdx {
		return
	}
	if a.prevSelectedAnimIdx >= 0 && a.prevSelectedAnimIdx < len(a.proj.Animations) {
		pAnim := &a.proj.Animations[a.prevSelectedAnimIdx]
		if a.prevSelectedAnimFrameIdx >= 0 && a.prevSelectedAnimFrameIdx < len(pAnim.Frames) {
			pf := &pAnim.Frames[a.prevSelectedAnimFrameIdx]
			entry := a.frameSpriteEntry(pf, a.spriteEditIdx)
			if entry == nil && len(pf.Sprites) > 0 {
				entry = &pf.Sprites[0]
			}
			switch a.panelMode {
			case panelModeHurtbox:
				if entry != nil && a.hurtboxTable.SelectedIdx >= 0 && a.hurtboxTable.SelectedIdx < len(entry.Hurtboxes) {
					hb := &entry.Hurtboxes[a.hurtboxTable.SelectedIdx]
					hb.X = a.props[0].NumericValue()
					hb.Y = a.props[1].NumericValue()
					hb.Width = a.props[2].NumericValue()
					hb.Height = a.props[3].NumericValue()
					hb.Rotation = a.props[4].NumericValue()
				}
			case panelModeAnimFrame:
				if entry != nil {
					entry.OffsetX = a.props[0].NumericValue()
					entry.OffsetY = a.props[1].NumericValue()
					entry.Rotation = a.props[2].NumericValue()
					entry.ScaleX = a.props[3].NumericValue()
					entry.ScaleY = a.props[4].NumericValue()
					entry.OriginX = a.originInputs[0].NumericValue()
					entry.OriginY = a.originInputs[1].NumericValue()
				}
			}
		}
	}
	a.prevSelectedAnimIdx = animIdx
	a.prevSelectedAnimFrameIdx = frameIdx
	if animIdx >= 0 && frameIdx >= 0 {
		a.panelMode = panelModeAnimFrame
		frame := a.proj.Animations[animIdx].Frames[frameIdx]
		entry := a.frameSpriteEntry(&frame, a.spriteEditIdx)
		if entry == nil && len(frame.Sprites) > 0 {
			entry = &frame.Sprites[0]
		}
		if entry != nil {
			a.proj.Sprites[entry.SpriteIdx].CurrentIdx = entry.SpriteFrameIdx
			a.frameSpriteDropdown.Selected = entry.SpriteIdx + 1
			a.phaseDropdown.Selected = int(frame.Phase)
			a.props[0].SetNumeric(entry.OffsetX)
			a.props[1].SetNumeric(entry.OffsetY)
			a.props[2].SetNumeric(entry.Rotation)
			a.props[3].SetNumeric(entry.ScaleX)
			a.props[4].SetNumeric(entry.ScaleY)
			a.originInputs[0].SetNumeric(entry.OriginX)
			a.originInputs[1].SetNumeric(entry.OriginY)
		}
		a.syncHurtboxBtns()
		a.syncLayout()
		a.syncMovementInputs()
	}
}

func (a *EditorApp) syncSpriteSelection() {
	if a.spriteTable.SelectedIdx == a.prevSelectedSpriteIdx {
		return
	}
	if a.panelMode == panelModeAnimFrame {
		if ai := a.prevSelectedAnimIdx; ai >= 0 && ai < len(a.proj.Animations) {
			if afi := a.prevSelectedAnimFrameIdx; afi >= 0 && afi < len(a.proj.Animations[ai].Frames) {
				pf := &a.proj.Animations[ai].Frames[afi]
				entry := a.frameSpriteEntry(pf, a.spriteEditIdx)
				if entry == nil && len(pf.Sprites) > 0 {
					entry = &pf.Sprites[0]
				}
				if entry != nil {
					entry.OffsetX = a.props[0].NumericValue()
					entry.OffsetY = a.props[1].NumericValue()
					entry.Rotation = a.props[2].NumericValue()
					entry.ScaleX = a.props[3].NumericValue()
					entry.ScaleY = a.props[4].NumericValue()
					entry.OriginX = a.originInputs[0].NumericValue()
					entry.OriginY = a.originInputs[1].NumericValue()
				}
			}
		}
	} else if a.panelMode == panelModeSprite {
		if p := a.prevSelectedSpriteIdx; p >= 0 && p < len(a.proj.Sprites) {
			row := &a.proj.Sprites[p]
			row.OffsetX = a.props[0].NumericValue()
			row.OffsetY = a.props[1].NumericValue()
			row.Rotation = a.props[2].NumericValue()
			row.ScaleX = a.props[3].NumericValue()
			row.ScaleY = a.props[4].NumericValue()
			row.OriginX = a.originInputs[0].NumericValue()
			row.OriginY = a.originInputs[1].NumericValue()
		}
	}
	a.prevSelectedSpriteIdx = a.spriteTable.SelectedIdx
	if idx := a.spriteTable.SelectedIdx; idx >= 0 && idx < len(a.proj.Sprites) {
		a.panelMode = panelModeSprite
		row := a.proj.Sprites[idx]
		a.props[0].SetNumeric(row.OffsetX)
		a.props[1].SetNumeric(row.OffsetY)
		a.props[2].SetNumeric(row.Rotation)
		a.props[3].SetNumeric(row.ScaleX)
		a.props[4].SetNumeric(row.ScaleY)
		a.originInputs[0].SetNumeric(row.OriginX)
		a.originInputs[1].SetNumeric(row.OriginY)
	}
	a.syncHurtboxBtns()
	a.syncLayout()
}

func (a *EditorApp) syncHurtboxSelection() {
	if a.hurtboxTable.SelectedIdx == a.prevSelectedHurtboxIdx {
		return
	}
	a.prevSelectedHurtboxIdx = a.hurtboxTable.SelectedIdx

	entry := a.currentFrameSpriteEntry()

	if a.hurtboxTable.SelectedIdx >= 0 {
		if a.panelMode == panelModeAnimFrame {
			if entry != nil {
				entry.OffsetX = a.props[0].NumericValue()
				entry.OffsetY = a.props[1].NumericValue()
				entry.Rotation = a.props[2].NumericValue()
				entry.ScaleX = a.props[3].NumericValue()
				entry.ScaleY = a.props[4].NumericValue()
				entry.OriginX = a.originInputs[0].NumericValue()
				entry.OriginY = a.originInputs[1].NumericValue()
			}
		}
		a.panelMode = panelModeHurtbox
		hbp := a.hurtboxList()
		if hbp != nil && a.hurtboxTable.SelectedIdx < len(*hbp) {
			hb := (*hbp)[a.hurtboxTable.SelectedIdx]
			a.props[0].SetLabel("Offset X")
			a.props[0].SetNumeric(hb.X)
			a.props[1].SetLabel("Offset Y")
			a.props[1].SetNumeric(hb.Y)
			a.props[2].SetLabel("Width")
			a.props[2].SetNumeric(hb.Width)
			a.props[2].Min = 0
			a.props[2].Step = 1
			a.props[3].SetLabel("Height")
			a.props[3].SetNumeric(hb.Height)
			a.props[3].Min = 0
			a.props[3].Step = 1
			a.props[4].SetLabel("Rotation (°)")
			a.props[4].SetNumeric(hb.Rotation)
			a.props[4].Min = -360
			a.props[4].Step = 0.5
		}
	} else {
		if a.animTable.SelectedIdx >= 0 && a.animTable.SelectedIdx < len(a.proj.Animations) {
			return
		}
		a.panelMode = panelModeSprite
		a.props[0].SetLabel("Offset X")
		a.props[0].Min = -99999
		a.props[0].Step = 1
		a.props[1].SetLabel("Offset Y")
		a.props[1].Min = -99999
		a.props[1].Step = 1
		a.props[2].SetLabel("Rotation (°)")
		a.props[2].Min = -360
		a.props[2].Step = 0.5
		a.props[3].SetLabel("Scale X")
		a.props[3].Min = -99999
		a.props[3].Step = 0.05
		a.props[4].SetLabel("Scale Y")
		a.props[4].Min = -99999
		a.props[4].Step = 0.05
		if idx := a.spriteTable.SelectedIdx; idx >= 0 && idx < len(a.proj.Sprites) {
			row := a.proj.Sprites[idx]
			a.props[0].SetNumeric(row.OffsetX)
			a.props[1].SetNumeric(row.OffsetY)
			a.props[2].SetNumeric(row.Rotation)
			a.props[3].SetNumeric(row.ScaleX)
			a.props[4].SetNumeric(row.ScaleY)
			a.originInputs[0].SetNumeric(row.OriginX)
			a.originInputs[1].SetNumeric(row.OriginY)
		}
	}
}
