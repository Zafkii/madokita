package editor

import "animprite/internal/project"

const maxUndo = 50

func (a *EditorApp) saveSnapshot() {
	a.pushSnapshot()
	a.dirty = true
}

func (a *EditorApp) saveSnapshotNoDirty() {
	a.pushSnapshot()
}

func (a *EditorApp) pushSnapshot() {
	a.endInfoUndo()
	a.pushSnapshotOf(project.DeepCopy(&a.proj))
}

// pushSnapshotOf adds an already captured project state as the next undo step.
func (a *EditorApp) pushSnapshotOf(data project.ProjectData) {
	a.undoStack = append(a.undoStack, data)
	if len(a.undoStack) > maxUndo {
		a.undoStack = a.undoStack[1:]
	}
	a.redoStack = nil
}

// beginInfoUndo captures the project state before an Info panel editing burst
// (typing in a property field or dragging the opacity slider). Every change of
// that burst collapses into one undo step instead of one per keystroke.
func (a *EditorApp) beginInfoUndo() {
	if a.infoUndoPrev != nil {
		return
	}
	prev := project.DeepCopy(&a.proj)
	a.infoUndoPrev = &prev
	a.infoUndoPushed = false
}

// endInfoUndo closes the burst. A burst that changed nothing leaves no step.
func (a *EditorApp) endInfoUndo() {
	a.infoUndoPrev = nil
	a.infoUndoPushed = false
}

// pushInfoUndoIfNeeded records the captured pre-burst state once the burst
// actually changed something.
func (a *EditorApp) pushInfoUndoIfNeeded() {
	if a.infoUndoPrev == nil || a.infoUndoPushed {
		return
	}
	a.pushSnapshotOf(*a.infoUndoPrev)
	a.infoUndoPushed = true
	a.dirty = true
}

func (a *EditorApp) undo() {
	a.endInfoUndo()
	if len(a.undoStack) == 0 {
		return
	}
	a.redoStack = append(a.redoStack, project.DeepCopy(&a.proj))

	entry := a.undoStack[len(a.undoStack)-1]
	a.undoStack = a.undoStack[:len(a.undoStack)-1]
	a.proj = entry
	a.rebuildFromProj()
}

func (a *EditorApp) redo() {
	a.endInfoUndo()
	if len(a.redoStack) == 0 {
		return
	}
	a.undoStack = append(a.undoStack, project.DeepCopy(&a.proj))

	entry := a.redoStack[len(a.redoStack)-1]
	a.redoStack = a.redoStack[:len(a.redoStack)-1]
	a.proj = entry
	a.rebuildFromProj()
}

func (a *EditorApp) rebuildFromProj() {
	if a.isStageMode() {
		if a.proj.Stage.Version == 0 {
			a.proj.Stage.Version = project.StageVersion
		}
		idx := a.st.selectedObj
		if idx >= len(a.proj.Stage.Objects) {
			idx = len(a.proj.Stage.Objects) - 1
		}
		if idx < 0 {
			idx = -1
		}
		a.setStageCanvasBounds()
		a.selectStageObject(idx)
		a.syncLayout()
		return
	}

	a.ensureFrameSprites()
	a.syncAnimBtns()
	a.syncSpriteBtns()
	a.syncHurtboxBtns()
	a.syncHitboxBtns()
	a.syncLayout()

	if a.spriteTable.SelectedIdx >= len(a.proj.Sprites) {
		a.spriteTable.SelectedIdx = -1
	}
	if a.animTable.SelectedIdx >= len(a.proj.Animations) {
		a.animTable.SelectedIdx = -1
	}
	if a.animTable.SelectedIdx < 0 && len(a.proj.Animations) > 0 {
		a.animTable.SelectedIdx = 0
	}
	a.prevSelectedSpriteIdx = -1
	a.prevSelectedAnimIdx = -1
	a.prevSelectedAnimFrameIdx = -1
	a.spriteEditIdx = 0

	if a.animTable.SelectedIdx >= 0 {
		a.navigateToAnim(a.animTable.SelectedIdx)
	} else if a.spriteTable.SelectedIdx >= 0 {
		a.navigateToSprite(a.spriteTable.SelectedIdx)
	}
	a.syncMovementInputs()
}

func (a *EditorApp) setStatus(msg string) {
	a.statusMsg = msg
	a.statusTime = 180
}

func (a *EditorApp) isAttackMode() bool {
	return a.mode == modeAttack
}

func (a *EditorApp) isStageMode() bool {
	return a.mode == modeStage
}
