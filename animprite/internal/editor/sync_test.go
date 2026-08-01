package editor

import (
	"testing"

	"animprite/internal/project"
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
