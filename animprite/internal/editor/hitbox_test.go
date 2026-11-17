package editor

import (
	"testing"

	"animprite/internal/canvas"
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

func TestHitboxesRenderOnTheCanvas(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hurtboxes = []project.HurtboxRow{
		{X: 2, Y: 3, Width: 10, Height: 12},
	}
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 10, Y: -4, Width: 40, Height: 20},
	}
	app.hitboxTable.SelectedIdx = 0

	renders := app.boxRenders()
	if len(renders) != 2 {
		t.Fatalf("box renders = %d, want 2 (one hurtbox, one hitbox)", len(renders))
	}
	if renders[0].Kind != canvas.RenderHurtbox || renders[0].Selected {
		t.Fatalf("hurtbox render = %+v, want unselected hurtbox", renders[0])
	}
	ht := renders[1]
	if ht.Kind != canvas.RenderHitbox {
		t.Fatalf("hitbox render kind = %v, want RenderHitbox", ht.Kind)
	}
	if !ht.Selected {
		t.Fatal("the selected hitbox row must render as selected")
	}
	if ht.OffsetX != 10 || ht.OffsetY != -4 || ht.WorldWidth != 40 || ht.WorldHeight != 20 {
		t.Fatalf("hitbox render = %+v, want world box at (10,-4) 40x20", ht)
	}
	if ht.Rotation != 0 {
		t.Fatalf("hitbox rotation = %v, want 0", ht.Rotation)
	}

	// An empty frame draws nothing from the attack box list.
	app.proj.Animations[0].Frames[0].Hitboxes = nil
	app.proj.Animations[0].Frames[0].Hurtboxes = nil
	if got := app.boxRenders(); len(got) != 0 {
		t.Fatalf("box renders without boxes = %d, want 0", len(got))
	}
}

func TestHitboxSelectionSurvivesAPanelSync(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 0, Y: 0, Width: 40, Height: 20},
	}
	// A sprite row stays selected after opening a project, which is what the
	// canvas used to re-sync the panel to.
	app.spriteTable.SelectedIdx = 0
	app.navigateToHitbox(0)

	if mode, ok := app.canvasSelectionTarget(); !ok || mode != panelModeHitbox {
		t.Fatalf("canvas selection target = %v/%v, want hitbox", mode, ok)
	}
	app.syncPanelToSelection()
	if app.panelMode != panelModeHitbox {
		t.Fatalf("panel mode after a canvas press = %v, want hitbox", app.panelMode)
	}
	app.flushInputsToData()

	got := app.proj.Animations[0].Frames[0].Hitboxes[0]
	if got.Width != 40 || got.Height != 20 {
		t.Fatalf("hitbox after a canvas press = %+v, want Width 40 Height 20", got)
	}
}

func TestHitboxDragKeepsWidthAndHeight(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 0, Y: 0, Width: 40, Height: 20},
	}
	app.spriteTable.SelectedIdx = 0
	app.navigateToHitbox(0)

	// The panel was showing the sprite entry (rotation 0, scale 1) right
	// before the drag, which is how Width/Height used to become 0 and 1.
	app.loadAnimFrameProps(0, 0)
	app.dragHitbox(0, 12, -6)
	app.flushInputsToData()

	got := app.proj.Animations[0].Frames[0].Hitboxes[0]
	want := project.HitboxRow{X: 12, Y: -6, Width: 40, Height: 20}
	if got != want {
		t.Fatalf("hitbox after dragging = %+v, want %+v", got, want)
	}
}

func TestSelectingAHitboxTakesTheSelectionFromTheHurtbox(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hurtboxes = []project.HurtboxRow{
		{X: 0, Y: 0, Width: 10, Height: 10},
	}
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 0, Y: 0, Width: 40, Height: 20},
	}
	app.syncHurtboxBtns()
	app.navigateToHurtbox(0)
	app.navigateToHitbox(0)

	if got := app.hurtboxTable.SelectedIdx; got != -1 {
		t.Fatalf("hurtbox row still selected = %d, want -1", got)
	}
	if got := app.hitboxTable.SelectedIdx; got != 0 {
		t.Fatalf("hitbox row = %d, want 0", got)
	}
	if mode, ok := app.canvasSelectionTarget(); !ok || mode != panelModeHitbox {
		t.Fatalf("canvas selection target = %v/%v, want hitbox", mode, ok)
	}

	// The next flush must edit the hitbox, never the hurtbox it left behind.
	app.flushInputsToData()
	hb := app.proj.Animations[0].Frames[0].Hurtboxes[0]
	if hb != (project.HurtboxRow{X: 0, Y: 0, Width: 10, Height: 10, DmgMult: 0}) && hb.Width != 10 {
		t.Fatalf("hurtbox was rewritten = %+v", hb)
	}

	// And the reverse: picking the hurtbox drops the hitbox selection.
	app.navigateToHurtbox(0)
	if got := app.hitboxTable.SelectedIdx; got != -1 {
		t.Fatalf("hitbox row after picking the hurtbox = %d, want -1", got)
	}
}

func TestHitboxScaleHandlesResizeTheBox(t *testing.T) {
	app := newHitboxTestApp(t)
	app.proj.Animations[0].Frames[0].Hitboxes = []project.HitboxRow{
		{X: 0, Y: 0, Width: 40, Height: 20},
	}
	app.navigateToHitbox(0)
	// Same rect the renderer publishes for a hitbox centred on the origin.
	app.canvas.Selection = canvas.SelectionRect{X: -20, Y: -10, W: 40, H: 20, Visible: true}

	// The press handler records the grabbed handle before snapshotting the
	// original box.
	app.scaleHandle = 2 // bottom-right
	app.initScaleOrig(2)
	// Drag the bottom-right corner out to world (40, 30): the opposite
	// (top-left) corner stays at (-20, -10), so the box becomes 60x40 centred
	// on (10, 10). X/Y hold the centre, like hurtboxes.
	mx, my := hitboxScreenPoint(app, 40, 30)
	app.applyHandleScale(mx, my)

	got := app.proj.Animations[0].Frames[0].Hitboxes[0]
	want := project.HitboxRow{X: 10, Y: 10, Width: 60, Height: 40}
	if got != want {
		t.Fatalf("scaled hitbox = %+v, want %+v", got, want)
	}
	app.flushInputsToData()
	if got := app.proj.Animations[0].Frames[0].Hitboxes[0]; got != want {
		t.Fatalf("hitbox after flush = %+v, want %+v", got, want)
	}
}

// hitboxScreenPoint mirrors the screen<-world conversion applyHandleScale
// undoes, so tests can express drags in world coordinates.
func hitboxScreenPoint(app *EditorApp, wx, wy float64) (int, int) {
	zoom := app.canvas.Cam.Zoom
	mx := wx*zoom + float64(app.canvas.Width)/2 + app.canvas.Cam.X + float64(app.canvas.X)
	my := wy*zoom + float64(app.canvas.Height)/2 + app.canvas.Cam.Y + float64(app.canvas.Y)
	return int(mx), int(my)
}
