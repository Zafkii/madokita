package editor

import (
	"testing"

	"animprite/internal/project"
	"animprite/internal/theme"
	"animprite/internal/ui"
)

func TestEnsureFrameEntry(t *testing.T) {
	proj := &project.ProjectData{
		Sprites: []project.SpriteRow{
			{Name: "base", FrameCount: 25, CurrentIdx: 7},
			{Name: "sprite 2", FrameCount: 4, CurrentIdx: 2, OffsetX: 5, OffsetY: -3, Rotation: 10, ScaleX: 1.5, ScaleY: 0.75, OriginX: 0.25, OriginY: 0.75},
		},
	}
	a := &EditorApp{proj: *proj}

	frame := &project.AnimationFrame{}

	e := a.ensureFrameEntry(frame, 1)
	if e == nil {
		t.Fatal("expected entry to be created")
	}
	if e.SpriteIdx != 1 {
		t.Errorf("SpriteIdx: got %d, want 1", e.SpriteIdx)
	}
	if e.SpriteFrameIdx != 2 {
		t.Errorf("SpriteFrameIdx: got %d, want 2 (inherited from row CurrentIdx)", e.SpriteFrameIdx)
	}
	if e.OffsetX != 5 || e.OffsetY != -3 || e.Rotation != 10 {
		t.Errorf("offset/rotation not inherited: %+v", e)
	}
	if e.ScaleX != 1.5 || e.ScaleY != 0.75 || e.OriginX != 0.25 || e.OriginY != 0.75 {
		t.Errorf("scale/origin not inherited: %+v", e)
	}
	if len(frame.Sprites) != 1 {
		t.Fatalf("frame sprite count: got %d, want 1", len(frame.Sprites))
	}

	e2 := a.ensureFrameEntry(frame, 1)
	if e2 != e {
		t.Error("expected existing entry to be returned")
	}
	if len(frame.Sprites) != 1 {
		t.Errorf("must not duplicate entry: got %d sprites", len(frame.Sprites))
	}

	if got := a.ensureFrameEntry(frame, 99); got != nil {
		t.Error("invalid sprite idx should return nil")
	}
}

func TestFlushInputsSpriteModeKeepsFrameEntry(t *testing.T) {
	th := theme.NewManager()
	proj := &project.ProjectData{
		Sprites: []project.SpriteRow{
			{Name: "base", Width: 256, Height: 256, FrameCount: 20, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
			{Name: "sprite 2", Width: 145, Height: 145, FrameCount: 6, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
		},
		Animations: []project.AnimationRow{
			{
				Name: "idle", FPS: 14, Loop: true, CurrentIdx: 0,
				Frames: []project.AnimationFrame{
					{
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
							{SpriteIdx: 1, SpriteFrameIdx: 0, OffsetX: -18, OffsetY: 16, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
						},
					},
				},
			},
		},
	}

	a := &EditorApp{proj: *proj, panelMode: panelModeSprite, th: th}
	for i := range a.atkTimingInputs {
		a.atkTimingInputs[i] = ui.NewTextInput(0, 0, 1, 1, th)
	}
	a.fpsInput = ui.NewTextInput(0, 0, 1, 1, th)
	a.loopInput = ui.NewTextInput(0, 0, 1, 1, th)
	a.loopInput.Text = "true"
	for i := range a.props {
		a.props[i] = ui.NewTextInput(0, 0, 1, 1, th)
	}
	for i := range a.originInputs {
		a.originInputs[i] = ui.NewTextInput(0, 0, 1, 1, th)
	}
	// Panel en modo sprite mostrando el row global del sprite 2 (valores iniciales).
	a.props[0].SetNumeric(0)
	a.props[1].SetNumeric(0)
	a.props[2].SetNumeric(0)
	a.props[3].SetNumeric(1)
	a.props[4].SetNumeric(1)
	a.originInputs[0].SetNumeric(0.5)
	a.originInputs[1].SetNumeric(0.5)

	a.animTable = ui.NewTable("Animations", nil, 1, th)
	a.spriteTable = ui.NewTable("Sprites", nil, 1, th)
	a.animTable.SelectedIdx = 0
	a.spriteTable.SelectedIdx = 1

	a.flushInputsToData()

	entry := a.proj.Animations[0].Frames[0].Sprites[1]
	if entry.OffsetX != -18 || entry.OffsetY != 16 {
		t.Fatalf("frame entry was overwritten by sprite-mode flush: got offset (%v, %v), want (-18, 16)", entry.OffsetX, entry.OffsetY)
	}
	if got := a.proj.Sprites[1].OffsetX; got != 0 {
		t.Errorf("sprite row should still be written in sprite mode: got OffsetX %v, want 0", got)
	}
}
