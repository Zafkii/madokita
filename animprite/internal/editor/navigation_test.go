package editor

import "testing"

func TestNavigateToHurtboxEmptyRefreshesTable(t *testing.T) {
	a := newScaleTestApp()
	a.proj.Animations[0].Frames[0].Hurtboxes = nil
	a.syncHurtboxBtns()

	// Simulate a stale table layout from before the last hurtbox was removed.
	a.hurtboxTable.RowCount = 2
	a.prevSelectedHurtboxIdx = 0

	a.navigateToHurtbox(-1)

	if a.hurtboxTable.SelectedIdx != -1 {
		t.Fatalf("selected hurtbox = %d, want -1", a.hurtboxTable.SelectedIdx)
	}
	if a.hurtboxTable.RowCount != 0 {
		t.Fatalf("row count = %d, want 0", a.hurtboxTable.RowCount)
	}
	if len(a.hurtboxXInputs) != 0 {
		t.Fatalf("hurtbox inputs = %d, want 0", len(a.hurtboxXInputs))
	}
	if a.prevSelectedHurtboxIdx != -1 {
		t.Fatalf("prevSelectedHurtboxIdx = %d, want -1", a.prevSelectedHurtboxIdx)
	}
}
