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
			switch a.panelMode {
			case panelModeHurtbox:
				if a.hurtboxTable.SelectedIdx >= 0 && a.hurtboxTable.SelectedIdx < len(pf.Hurtboxes) {
					hb := &pf.Hurtboxes[a.hurtboxTable.SelectedIdx]
					hb.X = a.props[0].NumericValue()
					hb.Y = a.props[1].NumericValue()
					hb.Width = a.props[2].NumericValue()
					hb.Height = a.props[3].NumericValue()
					hb.Rotation = a.props[4].NumericValue()
				}
			case panelModeAnimFrame:
				pf.OffsetX = a.props[0].NumericValue()
				pf.OffsetY = a.props[1].NumericValue()
				pf.Rotation = a.props[2].NumericValue()
				pf.ScaleX = a.props[3].NumericValue()
				pf.ScaleY = a.props[4].NumericValue()
				pf.OriginX = a.originInputs[0].NumericValue()
				pf.OriginY = a.originInputs[1].NumericValue()
			}
		}
	}
	a.prevSelectedAnimIdx = animIdx
	a.prevSelectedAnimFrameIdx = frameIdx
	if animIdx >= 0 && frameIdx >= 0 {
		a.panelMode = panelModeAnimFrame
		frame := a.proj.Animations[animIdx].Frames[frameIdx]
		if frame.SpriteIdx >= 0 && frame.SpriteIdx < len(a.proj.Sprites) {
			a.proj.Sprites[frame.SpriteIdx].CurrentIdx = frame.SpriteFrameIdx
		}
		a.frameSpriteDropdown.Selected = frame.SpriteIdx + 1
		a.phaseDropdown.Selected = int(frame.Phase)
		a.props[0].SetNumeric(frame.OffsetX)
		a.props[1].SetNumeric(frame.OffsetY)
		a.props[2].SetNumeric(frame.Rotation)
		a.props[3].SetNumeric(frame.ScaleX)
		a.props[4].SetNumeric(frame.ScaleY)
		a.originInputs[0].SetNumeric(frame.OriginX)
		a.originInputs[1].SetNumeric(frame.OriginY)
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
				pf.OffsetX = a.props[0].NumericValue()
				pf.OffsetY = a.props[1].NumericValue()
				pf.Rotation = a.props[2].NumericValue()
				pf.ScaleX = a.props[3].NumericValue()
				pf.ScaleY = a.props[4].NumericValue()
				pf.OriginX = a.originInputs[0].NumericValue()
				pf.OriginY = a.originInputs[1].NumericValue()
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

	if a.hurtboxTable.SelectedIdx >= 0 {
		if a.panelMode == panelModeAnimFrame {
			animIdx := a.animTable.SelectedIdx
			if animIdx >= 0 && animIdx < len(a.proj.Animations) {
				frameIdx := a.proj.Animations[animIdx].CurrentIdx
				if frameIdx >= 0 && frameIdx < len(a.proj.Animations[animIdx].Frames) {
					frame := &a.proj.Animations[animIdx].Frames[frameIdx]
					frame.OffsetX = a.props[0].NumericValue()
					frame.OffsetY = a.props[1].NumericValue()
					frame.Rotation = a.props[2].NumericValue()
					frame.ScaleX = a.props[3].NumericValue()
					frame.ScaleY = a.props[4].NumericValue()
					frame.OriginX = a.originInputs[0].NumericValue()
					frame.OriginY = a.originInputs[1].NumericValue()
				}
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
