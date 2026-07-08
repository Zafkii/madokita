package editor

import "animprite/internal/project"

const maxUndo = 50

func (a *EditorApp) saveSnapshot() {
	a.undoStack = append(a.undoStack, project.DeepCopy(&a.proj))
	if len(a.undoStack) > maxUndo {
		a.undoStack = a.undoStack[1:]
	}
	a.redoStack = nil
}

func (a *EditorApp) undo() {
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
