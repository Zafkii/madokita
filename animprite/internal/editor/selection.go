package editor

import "animprite/internal/project"

func (a *EditorApp) ensureFrameSprites() {
	spriteCount := len(a.proj.Sprites)
	for ai := range a.proj.Animations {
		for fi := range a.proj.Animations[ai].Frames {
			frame := &a.proj.Animations[ai].Frames[fi]

			seen := make(map[int]bool)
			valid := frame.Sprites[:0]
			for _, e := range frame.Sprites {
				if e.SpriteIdx >= 0 && e.SpriteIdx < spriteCount && !seen[e.SpriteIdx] {
					seen[e.SpriteIdx] = true
					valid = append(valid, e)
				}
			}
			frame.Sprites = valid

			for si := 0; si < spriteCount; si++ {
				if !seen[si] {
					s := a.proj.Sprites[si]
					frame.Sprites = append(frame.Sprites, project.FrameSpriteEntry{
						SpriteIdx:      si,
						SpriteFrameIdx: 0,
						OffsetX:        s.OffsetX,
						OffsetY:        s.OffsetY,
						Rotation:       s.Rotation,
						OriginX:        s.OriginX,
						OriginY:        s.OriginY,
						ScaleX:         s.ScaleX,
						ScaleY:         s.ScaleY,
					})
				}
			}
		}
	}
}

func (a *EditorApp) resetAnimFramePropLabels() {
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
}

func (a *EditorApp) loadAnimFrameProps(animIdx, frameIdx int) {
	if animIdx < 0 || animIdx >= len(a.proj.Animations) {
		return
	}
	if frameIdx < 0 || frameIdx >= len(a.proj.Animations[animIdx].Frames) {
		return
	}
	a.resetAnimFramePropLabels()
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
	if entry != nil && entry.SpriteIdx >= 0 && entry.SpriteIdx < len(a.proj.Sprites) {
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
