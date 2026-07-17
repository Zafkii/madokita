package editor

import (
	"os"
	"path/filepath"
	"testing"

	"animprite/internal/project"
)

func TestMovementRoundTrip(t *testing.T) {
	orig := &project.ProjectData{
		AssetName:       "TestMovement",
		AssetKey:        "test_movement",
		DefaultOriginX:  0.506,
		DefaultOriginY:  0.586,
		Animations: []project.AnimationRow{
			{
				Name: "idle", FPS: 3, Loop: true,
				Frames: []project.AnimationFrame{
					{
						Phase: project.PhaseWindup,
						Sprites: []project.FrameSpriteEntry{
							{
								SpriteIdx: 0, SpriteFrameIdx: 0,
								ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5,
							},
						},
						Hurtboxes: []project.HurtboxRow{
							{Width: 95, Height: 61, X: 1.5, Y: -32.5},
							{Width: 54, Height: 130, X: 4, Y: 62},
						},
					},
					{
						Phase: project.PhaseActive,
						Sprites: []project.FrameSpriteEntry{
							{
								SpriteIdx: 0, SpriteFrameIdx: 1,
								ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5,
							},
						},
						Hurtboxes: []project.HurtboxRow{
							{Width: 100, Height: 57, X: 1, Y: -32.5, Rotation: 45},
						},
					},
				},
			},
			{
				Name: "walk", FPS: 10, Loop: false,
				Frames: []project.AnimationFrame{
					{
						Phase: project.PhaseWindup,
						Sprites: []project.FrameSpriteEntry{
							{
								SpriteIdx: 0, SpriteFrameIdx: 4,
								ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5,
							},
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test_movement.go")

	if err := ExportMovement(path, orig); err != nil {
		t.Fatalf("ExportMovement: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	t.Logf("Generated file:\n%s", string(data))

	got, err := ImportMovement(path)
	if err != nil {
		t.Fatalf("ImportMovement: %v", err)
	}

	if got.AssetName != orig.AssetName {
		t.Errorf("AssetName: got %q, want %q", got.AssetName, orig.AssetName)
	}
	if got.AssetKey != orig.AssetKey {
		t.Errorf("AssetKey: got %q, want %q", got.AssetKey, orig.AssetKey)
	}
	if got.DefaultOriginX != orig.DefaultOriginX {
		t.Errorf("DefaultOriginX: got %v, want %v", got.DefaultOriginX, orig.DefaultOriginX)
	}
	if got.DefaultOriginY != orig.DefaultOriginY {
		t.Errorf("DefaultOriginY: got %v, want %v", got.DefaultOriginY, orig.DefaultOriginY)
	}
	if len(got.Animations) != len(orig.Animations) {
		t.Fatalf("anim count: got %d, want %d", len(got.Animations), len(orig.Animations))
	}

	for i, want := range orig.Animations {
		gotA := got.Animations[i]
		if gotA.Name != want.Name {
			t.Errorf("anim[%d].Name: got %q, want %q", i, gotA.Name, want.Name)
		}
		if gotA.FPS != want.FPS {
			t.Errorf("anim[%d].FPS: got %v, want %v", i, gotA.FPS, want.FPS)
		}
		if gotA.Loop != want.Loop {
			t.Errorf("anim[%d].Loop: got %v, want %v", i, gotA.Loop, want.Loop)
		}
		if len(gotA.Frames) != len(want.Frames) {
			t.Fatalf("anim[%d] frame count: got %d, want %d", i, len(gotA.Frames), len(want.Frames))
		}
		for j, wantF := range want.Frames {
			gotF := gotA.Frames[j]
			if len(gotF.Sprites) != len(wantF.Sprites) {
				t.Fatalf("anim[%d].frame[%d] sprite count: got %d, want %d", i, j, len(gotF.Sprites), len(wantF.Sprites))
			}
			for si, wantS := range wantF.Sprites {
				gotS := gotF.Sprites[si]
				if gotS.SpriteFrameIdx != wantS.SpriteFrameIdx {
					t.Errorf("anim[%d].frame[%d].sprite[%d].SpriteFrameIdx: got %d, want %d", i, j, si, gotS.SpriteFrameIdx, wantS.SpriteFrameIdx)
				}
			}
			if len(gotF.Hurtboxes) != len(wantF.Hurtboxes) {
				t.Fatalf("anim[%d].frame[%d] hb count: got %d, want %d", i, j, len(gotF.Hurtboxes), len(wantF.Hurtboxes))
			}
			for k, wantHB := range wantF.Hurtboxes {
				gotHB := gotF.Hurtboxes[k]
				if gotHB.Width != wantHB.Width {
					t.Errorf("anim[%d].frame[%d].hb[%d].Width: got %v, want %v", i, j, k, gotHB.Width, wantHB.Width)
				}
				if gotHB.Height != wantHB.Height {
					t.Errorf("anim[%d].frame[%d].hb[%d].Height: got %v, want %v", i, j, k, gotHB.Height, wantHB.Height)
				}
				if gotHB.X != wantHB.X {
					t.Errorf("anim[%d].frame[%d].hb[%d].X: got %v, want %v", i, j, k, gotHB.X, wantHB.X)
				}
				if gotHB.Y != wantHB.Y {
					t.Errorf("anim[%d].frame[%d].hb[%d].Y: got %v, want %v", i, j, k, gotHB.Y, wantHB.Y)
				}
				if gotHB.Rotation != wantHB.Rotation {
					t.Errorf("anim[%d].frame[%d].hb[%d].Rotation: got %v, want %v", i, j, k, gotHB.Rotation, wantHB.Rotation)
				}
			}
		}
	}
}
