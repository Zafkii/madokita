package editor

import "testing"

func TestHasUnsavedChanges(t *testing.T) {
	a := &EditorApp{}
	if a.hasUnsavedChanges() {
		t.Fatal("clean project reported unsaved changes")
	}

	a.dirty = true
	if !a.hasUnsavedChanges() {
		t.Fatal("dirty project reported clean")
	}

	a.mode = modeStage
	if a.hasUnsavedChanges() {
		t.Fatal("stage mode should be exempt from the unsaved-changes guard")
	}

	a.markClean()
	if a.dirty {
		t.Fatal("markClean did not clear the dirty flag")
	}
}

func TestSaveSnapshotTracksDirty(t *testing.T) {
	a := &EditorApp{}

	a.saveSnapshotNoDirty()
	if a.dirty {
		t.Fatal("navigation snapshot must not mark the project dirty")
	}

	a.saveSnapshot()
	if !a.dirty {
		t.Fatal("saveSnapshot must mark the project dirty")
	}

	a.markClean()
	if a.dirty {
		t.Fatal("markClean did not clear the dirty flag")
	}
}

func TestRequestOpenFolderRequiresConfirmationWhenDirty(t *testing.T) {
	a := &EditorApp{mode: modeMovement, dirty: true}

	a.requestOpenFolder(t.TempDir())

	if !a.confirm.visible {
		t.Fatal("expected unsaved-changes confirmation")
	}
	if a.confirm.onConfirm == nil {
		t.Fatal("expected a pending open action")
	}
}

func TestConfirmAcceptRunsAction(t *testing.T) {
	a := &EditorApp{}
	called := false

	a.showConfirm("Unsaved Changes", "Discard?", func() { called = true })
	a.acceptConfirm()

	if !called {
		t.Fatal("acceptConfirm did not run the pending action")
	}
	if a.confirm.visible || a.confirm.onConfirm != nil {
		t.Fatal("acceptConfirm did not clear the confirmation state")
	}
}

func TestConfirmCancelDropsAction(t *testing.T) {
	a := &EditorApp{}
	called := false

	a.showConfirm("Unsaved Changes", "Discard?", func() { called = true })
	a.cancelConfirm()

	if called {
		t.Fatal("cancelConfirm ran the pending action")
	}
	if a.confirm.visible || a.confirm.onConfirm != nil {
		t.Fatal("cancelConfirm did not clear the confirmation state")
	}
}
