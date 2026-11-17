package editor

import (
	"testing"

	"animprite/internal/project"
)

// newHitboxTestApp builds an attack-mode editor with one animation holding two
// frames, which is enough to prove hitboxes belong to the frame being edited.
func newHitboxTestApp(t *testing.T) *EditorApp {
	t.Helper()
	app := NewEditorApp()
	app.mode = modeAttack
	sprite := []project.FrameSpriteEntry{
		{SpriteIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
	}
	app.proj = project.ProjectData{
		AssetName: "TestAttack",
		AssetKey:  "test_attack",
		Sprites: []project.SpriteRow{
			{Name: "base", Width: 64, Height: 64, FrameCount: 1, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
		},
		Animations: []project.AnimationRow{
			{
				Name: "slash", FPS: 12,
				Frames: []project.AnimationFrame{
					{Phase: project.PhaseWindup, Sprites: append([]project.FrameSpriteEntry(nil), sprite...)},
					{Phase: project.PhaseActive, Sprites: append([]project.FrameSpriteEntry(nil), sprite...)},
				},
			},
		},
	}
	app.animTable.SelectedIdx = 0
	app.syncHitboxBtns()
	return app
}

func TestHitboxRowsBelongToTheCurrentFrame(t *testing.T) {
	app := newHitboxTestApp(t)

	app.hitboxTable.AddBtn.OnClick()
	if got := len(app.proj.Animations[0].Frames[0].Hitboxes); got != 1 {
		t.Fatalf("frame 0 hitboxes after add = %d, want 1", got)
	}
	if got := len(app.proj.Animations[0].Frames[1].Hitboxes); got != 0 {
		t.Fatalf("frame 1 gained %d hitboxes from an edit on frame 0", got)
	}

	app.proj.Animations[0].CurrentIdx = 1
	app.syncHitboxBtns()
	if got := app.hitboxTable.RowCount; got != 0 {
		t.Fatalf("row count on frame 1 = %d, want 0", got)
	}

	app.hitboxTable.AddBtn.OnClick()
	if got := len(app.proj.Animations[0].Frames[1].Hitboxes); got != 1 {
		t.Fatalf("frame 1 hitboxes after add = %d, want 1", got)
	}
	if got := len(app.proj.Animations[0].Frames[0].Hitboxes); got != 1 {
		t.Fatalf("frame 0 hitboxes changed to %d, want 1", got)
	}
}

func TestHitboxRowInputsWriteAllFourValues(t *testing.T) {
	app := newHitboxTestApp(t)
	app.hitboxTable.AddBtn.OnClick()

	app.hitboxXInputs[0].SetNumeric(12)
	app.hitboxXInputs[0].OnChange("12")
	app.hitboxYInputs[0].SetNumeric(-7.5)
	app.hitboxYInputs[0].OnChange("-7.5")
	app.hitboxWidthInputs[0].SetNumeric(40)
	app.hitboxWidthInputs[0].OnChange("40")
	app.hitboxHeightInputs[0].SetNumeric(30)
	app.hitboxHeightInputs[0].OnChange("30")

	want := project.HitboxRow{X: 12, Y: -7.5, Width: 40, Height: 30}
	if got := app.proj.Animations[0].Frames[0].Hitboxes[0]; got != want {
		t.Fatalf("hitbox = %+v, want %+v", got, want)
	}
}

func TestHitboxPanelWritesXAndY(t *testing.T) {
	app := newHitboxTestApp(t)
	app.hitboxTable.AddBtn.OnClick()

	app.props[0].SetNumeric(24)
	app.props[1].SetNumeric(-16)
	app.props[2].SetNumeric(50)
	app.props[3].SetNumeric(20)
	app.flushInputsToData()

	want := project.HitboxRow{X: 24, Y: -16, Width: 50, Height: 20}
	if got := app.proj.Animations[0].Frames[0].Hitboxes[0]; got != want {
		t.Fatalf("hitbox = %+v, want %+v", got, want)
	}
}

func TestHitboxCopyPreviousFrame(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 10, Y: 4, Width: 30, Height: 12},
	}
	app.proj.Animations[0].CurrentIdx = 1
	app.syncHitboxBtns()

	copyBtn := app.hitboxTable.ExtraBtns[0]
	copyBtn.OnClick()

	got := app.proj.Animations[0].Frames[1].Hitboxes
	if len(got) != 1 {
		t.Fatalf("copied hitboxes = %d, want 1", len(got))
	}
	if want := app.proj.Animations[0].Frames[0].Hitboxes[0]; got[0] != want {
		t.Fatalf("copied hitbox = %+v, want %+v", got[0], want)
	}
	// The copy is a value clone: editing it must not touch the source frame.
	got[0].X = 999
	if app.proj.Animations[0].Frames[0].Hitboxes[0].X != 10 {
		t.Fatal("copied hitbox shares storage with the previous frame")
	}
}

func TestNewFrameInheritsHitboxes(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 10, Y: 4, Width: 30, Height: 12},
		{X: -6, Y: 8, Width: 16, Height: 16},
	}
	app.syncAnimBtns()

	app.animAddFrameBtns[0].OnClick()

	frames := app.proj.Animations[0].Frames
	if len(frames) != 3 {
		t.Fatalf("frames = %d, want 3", len(frames))
	}
	fresh := frames[len(frames)-1]
	if len(fresh.Hitboxes) != 2 {
		t.Fatalf("inherited hitboxes = %d, want 2", len(fresh.Hitboxes))
	}
	if fresh.Hitboxes[0] != frames[0].Hitboxes[0] {
		t.Fatalf("inherited hitbox = %+v, want %+v", fresh.Hitboxes[0], frames[0].Hitboxes[0])
	}
	// Frames must not share the backing array.
	fresh.Hitboxes[0].X = 999
	if frames[0].Hitboxes[0].X != 10 {
		t.Fatal("new frame shares hitbox storage with the previous frame")
	}
}

func TestFrameSwitchRefreshesHitboxRows(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 10, Y: 4, Width: 30, Height: 12},
	}
	app.proj.Animations[0].Frames[1].Hitboxes = []project.HitboxRow{
		{X: -4, Y: 9, Width: 20, Height: 6},
	}
	app.loadAnimFrameProps(0, 0)
	if got := app.hitboxTable.RowCount; got != 1 {
		t.Fatalf("row count on frame 0 = %d, want 1", got)
	}
	if len(app.hitboxXInputs) != 1 || len(app.hitboxHeightInputs) != 1 {
		t.Fatalf("frame 0 row inputs = %d/%d, want 1/1", len(app.hitboxXInputs), len(app.hitboxHeightInputs))
	}
	if got := app.hitboxXInputs[0].NumericValue(); got != 10 {
		t.Fatalf("frame 0 row X = %v, want 10", got)
	}

	// Frame buttons set the animation's current frame before refreshing props.
	app.proj.Animations[0].CurrentIdx = 1
	app.loadAnimFrameProps(0, 1)
	if got := app.hitboxTable.RowCount; got != 1 {
		t.Fatalf("row count on frame 1 = %d, want 1", got)
	}
	if len(app.hitboxXInputs) != 1 || len(app.hitboxHeightInputs) != 1 {
		t.Fatalf("frame 1 row inputs = %d/%d, want 1/1", len(app.hitboxXInputs), len(app.hitboxHeightInputs))
	}
	if got := app.hitboxXInputs[0].NumericValue(); got != -4 {
		t.Fatalf("frame 1 row X = %v, want -4: the table still shows the previous frame", got)
	}
	if got := app.hitboxHeightInputs[0].NumericValue(); got != 6 {
		t.Fatalf("frame 1 row Height = %v, want 6", got)
	}

	// An empty frame clears both the rows and the selection.
	app.proj.Animations[0].Frames[1].Hitboxes = nil
	app.loadAnimFrameProps(0, 1)
	if got := app.hitboxTable.RowCount; got != 0 {
		t.Fatalf("row count on an empty frame = %d, want 0", got)
	}
	if got := app.hitboxTable.SelectedIdx; got != -1 {
		t.Fatalf("selection on an empty frame = %d, want -1", got)
	}
}
