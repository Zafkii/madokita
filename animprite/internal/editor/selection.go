package editor

func (a *EditorApp) syncAnimFrameSelection() {
}

func (a *EditorApp) loadAnimFrameProps(animIdx, frameIdx int) {
	a.panelMode = panelModeAnimFrame
	frame := &a.proj.Animations[animIdx].Frames[frameIdx]
	if len(frame.Sprites) == 0 {
		frame.Sprites = a.defaultFrameSprites()
	}
	entry := a.frameSpriteEntry(frame, a.spriteEditIdx)
	if entry == nil {
		entry = &frame.Sprites[0]
		a.spriteEditIdx = frame.Sprites[0].SpriteIdx
		a.frameSpriteDropdown.Selected = entry.SpriteIdx + 1
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
	a.prevSelectedAnimIdx = animIdx
	a.prevSelectedAnimFrameIdx = frameIdx
	a.syncHurtboxBtns()
	a.syncLayout()
	a.syncMovementInputs()
}

func (a *EditorApp) syncSpriteSelection() {
}

func (a *EditorApp) syncHurtboxSelection() {
}
