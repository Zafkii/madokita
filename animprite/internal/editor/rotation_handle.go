package editor

import "math"

func (a *EditorApp) beginRotationDrag(mx, my int) bool {
	if !a.canvas.Selection.Visible {
		return false
	}
	target, ok := a.canvasSelectionTarget()
	if !ok {
		return false
	}
	orig, ok := a.rotationForTarget(target)
	if !ok {
		return false
	}
	sel := a.canvas.Selection
	cx := sel.X + sel.W/2
	cy := sel.Y + sel.H/2
	a.rotationDrag = rotationDragData{
		active:       true,
		target:       target,
		centerWX:     cx,
		centerWY:     cy,
		startAngle:   a.canvasMouseAngle(mx, my, cx, cy),
		origRotation: orig,
	}
	return true
}

func (a *EditorApp) applyRotationDrag(mx, my int) {
	d := &a.rotationDrag
	if !d.active {
		return
	}
	angle := a.canvasMouseAngle(mx, my, d.centerWX, d.centerWY)
	delta := angle - d.startAngle
	for delta > math.Pi {
		delta -= 2 * math.Pi
	}
	for delta < -math.Pi {
		delta += 2 * math.Pi
	}
	deg := math.Round((d.origRotation+delta*180/math.Pi)*10) / 10
	a.setRotationForTarget(d.target, deg)
}

func (a *EditorApp) canvasMouseAngle(mx, my int, cx, cy float64) float64 {
	wx, wy := a.canvas.Cam.CanvasToWorld(float64(mx-a.canvas.X), float64(my-a.canvas.Y))
	return math.Atan2(wy-cy, wx-cx)
}

func (a *EditorApp) rotationForTarget(target rightPanelMode) (float64, bool) {
	switch target {
	case panelModeHurtbox:
		hbp := a.hurtboxList()
		if hbp != nil && a.hurtboxTable.SelectedIdx >= 0 && a.hurtboxTable.SelectedIdx < len(*hbp) {
			return (*hbp)[a.hurtboxTable.SelectedIdx].Rotation, true
		}
	case panelModeSprite:
		if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
			return a.proj.Sprites[s].Rotation, true
		}
	case panelModeAnimFrame:
		if entry := a.currentFrameSpriteEntry(); entry != nil {
			return entry.Rotation, true
		}
	case panelModeStage:
		if obj := a.selectedStageObject(); obj != nil {
			return obj.Rotation, true
		}
	}
	return 0, false
}

func (a *EditorApp) setRotationForTarget(target rightPanelMode, deg float64) bool {
	switch target {
	case panelModeHurtbox:
		hbp := a.hurtboxList()
		if hbp != nil && a.hurtboxTable.SelectedIdx >= 0 && a.hurtboxTable.SelectedIdx < len(*hbp) {
			hb := &(*hbp)[a.hurtboxTable.SelectedIdx]
			hb.Rotation = deg
			a.props[4].SetNumeric(deg)
			return true
		}
	case panelModeSprite:
		if s := a.spriteTable.SelectedIdx; s >= 0 && s < len(a.proj.Sprites) {
			row := &a.proj.Sprites[s]
			row.Rotation = deg
			a.props[2].SetNumeric(deg)
			return true
		}
	case panelModeAnimFrame:
		if entry := a.currentFrameSpriteEntry(); entry != nil {
			entry.Rotation = deg
			a.props[2].SetNumeric(deg)
			return true
		}
	case panelModeStage:
		if obj := a.selectedStageObject(); obj != nil {
			obj.Rotation = deg
			a.props[4].SetNumeric(deg)
			a.markStageObjectLiveEdit()
			return true
		}
	}
	return false
}
