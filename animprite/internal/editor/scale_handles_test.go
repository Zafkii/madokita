package editor

import (
	"testing"

	"animprite/internal/camera"
	"animprite/internal/canvas"
	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"
)

func newScaleTestApp() *EditorApp {
	th := theme.NewManager()
	cam := camera.New()
	cam.SetViewport(1000, 700)
	c := &canvas.Canvas{X: 0, Y: 0, Width: 1000, Height: 700, Cam: cam}

	a := &EditorApp{th: th, canvas: c}
	a.win.outsideWidth = 1280
	a.win.outsideHeight = 720
	a.proj = project.ProjectData{
		Sprites: []project.SpriteRow{
			{Name: "base", Width: 256, Height: 256, FrameCount: 1, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
		},
		Animations: []project.AnimationRow{
			{Name: "idle", FPS: 3, Loop: true, CurrentIdx: 0, Frames: []project.AnimationFrame{
				{
					Sprites:   []project.FrameSpriteEntry{{SpriteIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5}},
					Hurtboxes: []project.HurtboxRow{{X: 1.5, Y: -32.5, Width: 95, Height: 61, DmgMult: 1}},
				},
			}},
		},
	}
	a.animTable = &ui.Table{SelectedIdx: 0}
	a.spriteTable = &ui.Table{SelectedIdx: -1}
	a.hurtboxTable = &ui.Table{SelectedIdx: 0}
	a.hitboxTable = &ui.Table{SelectedIdx: -1}
	a.modeDropdown = ui.NewDropdown(0, 0, 0, 0, th)
	a.phaseDropdown = ui.NewDropdown(0, 0, 0, 0, th)
	a.frameSpriteDropdown = ui.NewDropdown(0, 0, 0, 0, th)
	a.openBtn = ui.NewButton(0, 0, 0, 0, "Open", th)
	a.openDirBtn = ui.NewButton(0, 0, 0, 0, "Open Dir", th)
	a.saveBtn = ui.NewButton(0, 0, 0, 0, "Save", th)
	a.themeBtn = ui.NewButton(0, 0, 0, 0, "Dark", th)
	a.baseRotInput = ui.NewTextInput(0, 0, 0, 0, th)
	a.fpsInput = ui.NewTextInput(0, 0, 0, 0, th)
	a.loopInput = ui.NewTextInput(0, 0, 0, 0, th)
	a.loopInput.Text = "true"
	for i := range a.props {
		a.props[i] = ui.NewTextInput(0, 0, 0, 0, th)
		a.props[i].Numeric = true
	}
	for i := range a.originInputs {
		a.originInputs[i] = ui.NewTextInput(0, 0, 0, 0, th)
	}
	for i := range a.atkTimingInputs {
		a.atkTimingInputs[i] = ui.NewTextInput(0, 0, 0, 0, th)
	}
	a.syncLayout()
	cam.SetViewport(float64(c.Width), float64(c.Height))
	return a
}

// topRightClick converts the selection's top-right corner to screen space.
func topRightClick(a *EditorApp) (mx, my int) {
	sel := a.canvas.Selection
	cx, cy := a.canvas.Cam.WorldToCanvas(sel.X+sel.W, sel.Y)
	return int(cx) + a.canvas.X, int(cy) + a.canvas.Y
}

func spriteEntry(a *EditorApp) project.FrameSpriteEntry {
	return a.proj.Animations[0].Frames[0].Sprites[0]
}

func hurtbox(a *EditorApp) project.HurtboxRow {
	return a.proj.Animations[0].Frames[0].Hurtboxes[0]
}

func assertEntryUntouched(t *testing.T, e project.FrameSpriteEntry) {
	t.Helper()
	if e.ScaleX != 1 || e.ScaleY != 1 || e.OffsetX != 0 || e.OffsetY != 0 {
		t.Fatalf("sprite entry was scaled instead of the hurtbox: %+v", e)
	}
}

// A stale panelMode (after preview playback or a frame change) must not make a
// Ctrl+drag on the hurtbox handles scale the frame sprite instead.
func TestHandleScaleFollowsCanvasSelection(t *testing.T) {
	a := newScaleTestApp()
	a.panelMode = panelModeAnimFrame // stale, hurtbox still selected on canvas
	a.buildSpriteRenders()

	if !a.canvas.Selection.Visible {
		t.Fatal("expected a visible selection for the selected hurtbox")
	}

	mx, my := topRightClick(a)
	a.saveSnapshot()
	a.syncPanelToSelection()
	if a.panelMode != panelModeHurtbox {
		t.Fatalf("panelMode = %v, want panelModeHurtbox", a.panelMode)
	}

	a.scaleHandle = 1
	a.initScaleOrig(1)
	a.applyHandleScale(mx+10, my-3)
	a.flushInputsToData()

	assertEntryUntouched(t, spriteEntry(a))
	hb := hurtbox(a)
	if hb.Width != 105 || hb.Height != 64 {
		t.Fatalf("hurtbox size = %vx%v, want 105x64", hb.Width, hb.Height)
	}
	if hb.X != 6.5 || hb.Y != -34 {
		t.Fatalf("hurtbox center = (%v, %v), want (6.5, -34)", hb.X, hb.Y)
	}
}

// Fractional hurtbox offsets must survive the per-frame input flush.
func TestHurtboxOffsetsKeepDecimals(t *testing.T) {
	a := newScaleTestApp()
	a.panelMode = panelModeAnimFrame
	a.buildSpriteRenders()
	a.syncPanelToSelection()
	for i := 0; i < 3; i++ {
		a.flushInputsToData()
	}
	hb := hurtbox(a)
	if hb.X != 1.5 || hb.Y != -32.5 {
		t.Fatalf("fractional offsets were rounded: (%v, %v)", hb.X, hb.Y)
	}
}

// The drag target is captured on press, so a mode change mid-drag (preview
// advancing) cannot hijack the edit.
func TestHandleScaleKeepsTargetAcrossModeChange(t *testing.T) {
	a := newScaleTestApp()
	a.panelMode = panelModeAnimFrame
	a.buildSpriteRenders()
	a.syncPanelToSelection()

	mx, my := topRightClick(a)
	a.scaleHandle = 1
	a.initScaleOrig(1)
	a.panelMode = panelModeAnimFrame // preview advanced after the drag started

	a.applyHandleScale(mx+10, my-3)
	a.flushInputsToData()

	assertEntryUntouched(t, spriteEntry(a))
	hb := hurtbox(a)
	if hb.Width != 105 || hb.Height != 64 {
		t.Fatalf("hurtbox size = %vx%v, want 105x64", hb.Width, hb.Height)
	}
}
