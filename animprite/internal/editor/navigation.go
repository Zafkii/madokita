package editor

func (a *EditorApp) navigateToAnim(animIdx int) {
	a.animTable.SelectedIdx = animIdx
	a.spriteTable.SelectedIdx = -1
	a.hurtboxTable.SelectedIdx = -1
	a.hitboxTable.SelectedIdx = -1
	a.hurtboxAnimCtx = animIdx

	if animIdx >= 0 && animIdx < len(a.proj.Animations) {
		anim := &a.proj.Animations[animIdx]
		frameIdx := anim.CurrentIdx
		if frameIdx >= 0 && frameIdx < len(anim.Frames) {
			a.loadAnimFrameProps(animIdx, frameIdx)
		}
	}
	a.syncHurtboxBtns()
	a.syncLayout()
}

func (a *EditorApp) navigateToSprite(spriteIdx int) {
	a.spriteTable.SelectedIdx = spriteIdx
	a.animTable.SelectedIdx = -1
	a.hurtboxTable.SelectedIdx = -1
	a.hitboxTable.SelectedIdx = -1

	if spriteIdx >= 0 && spriteIdx < len(a.proj.Sprites) {
		a.panelMode = panelModeSprite
		row := a.proj.Sprites[spriteIdx]
		a.props[0].SetLabel("Offset X")
		a.props[0].SetNumeric(row.OffsetX)
		a.props[0].Min = -99999
		a.props[0].Step = 1
		a.props[1].SetLabel("Offset Y")
		a.props[1].SetNumeric(row.OffsetY)
		a.props[1].Min = -99999
		a.props[1].Step = 1
		a.props[2].SetLabel("Rotation (°)")
		a.props[2].SetNumeric(row.Rotation)
		a.props[2].Min = -360
		a.props[2].Step = 0.5
		a.props[3].SetLabel("Scale X")
		a.props[3].SetNumeric(row.ScaleX)
		a.props[3].Min = -99999
		a.props[3].Step = 0.05
		a.props[4].SetLabel("Scale Y")
		a.props[4].SetNumeric(row.ScaleY)
		a.props[4].Min = -99999
		a.props[4].Step = 0.05
		a.originInputs[0].SetNumeric(row.OriginX)
		a.originInputs[1].SetNumeric(row.OriginY)
	}

	a.prevSelectedSpriteIdx = spriteIdx
	a.syncHurtboxBtns()
	a.syncLayout()
}

func (a *EditorApp) navigateToHurtbox(hbIdx int) {
	a.hurtboxTable.SelectedIdx = hbIdx
	a.animTable.SelectedIdx = -1
	a.spriteTable.SelectedIdx = -1
	a.hitboxTable.SelectedIdx = -1
	a.syncHurtboxBtns()

	if a.hurtboxTable.SelectedIdx >= 0 {
		a.panelMode = panelModeHurtbox
		hbp := a.hurtboxList()
		if hbp != nil && a.hurtboxTable.SelectedIdx < len(*hbp) {
			hb := (*hbp)[a.hurtboxTable.SelectedIdx]
			a.props[0].SetLabel("Offset X")
			a.props[0].SetNumeric(hb.X)
			a.props[0].Min = -99999
			a.props[0].Step = 1
			a.props[1].SetLabel("Offset Y")
			a.props[1].SetNumeric(hb.Y)
			a.props[1].Min = -99999
			a.props[1].Step = 1
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

	a.prevSelectedHurtboxIdx = a.hurtboxTable.SelectedIdx
	a.syncLayout()
}

func (a *EditorApp) navigateToHitbox(hbIdx int) {
	a.hitboxTable.SelectedIdx = hbIdx
	a.animTable.SelectedIdx = -1
	a.spriteTable.SelectedIdx = -1
	a.hurtboxTable.SelectedIdx = -1

	if hbIdx >= 0 && hbIdx < len(a.proj.HitDefs) {
		a.panelMode = panelModeHitbox
		hb := a.proj.HitDefs[hbIdx]
		a.props[0].SetLabel("Width")
		a.props[0].SetNumeric(hb.Width)
		a.props[0].Min = 0
		a.props[0].Step = 1
		a.props[1].SetLabel("Height")
		a.props[1].SetNumeric(hb.Height)
		a.props[1].Min = 0
		a.props[1].Step = 1
		a.props[2].SetLabel("")
		a.props[2].SetNumeric(0)
		a.props[3].SetLabel("")
		a.props[3].SetNumeric(0)
		a.props[4].SetLabel("")
		a.props[4].SetNumeric(0)
	}

	a.syncLayout()
}
