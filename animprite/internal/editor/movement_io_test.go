package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"animprite/internal/project"
)

func TestMovementRejectsBareFrame(t *testing.T) {
	src := `package movements

import . "madokita/internal/animation"

var Legacy = Movement{
	AssetKey: "legacy",
	Animations: map[string]MovementAnimDef{
		"idle": Anim(3, true,
			F(0),
		),
	},
}
`
	path := filepath.Join(t.TempDir(), "legacy.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportMovement(path); err == nil {
		t.Fatal("expected error for sprite-less F(...) frame, got nil")
	}
}

func TestMovementRoundTripPerFrameSprites(t *testing.T) {
	orig := &project.ProjectData{
		AssetName:       "TestMulti",
		AssetKey:        "test_multi",
		DefaultOriginX:  0.506,
		DefaultOriginY:  0.586,
		Sprites: []project.SpriteRow{
			{Name: "base", File: "sprites/players/sayaka_miki/sayaka_miki.png", Width: 256, Height: 256, FrameCount: 25},
			{Name: "sprite 2", File: "sprites/players/sayaka_miki/weapon.png", Width: 128, Height: 128, FrameCount: 4},
		},
		Animations: []project.AnimationRow{
			{
				Name: "idle", FPS: 3, Loop: true,
				Frames: []project.AnimationFrame{
					{
						Phase: project.PhaseWindup,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 0, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
							{SpriteIdx: 1, SpriteFrameIdx: 0, OffsetX: 5, OffsetY: -3, Rotation: 10, ScaleX: 1.5, ScaleY: 0.75, OriginX: 0.25, OriginY: 0.75},
						},
					},
					{
						Phase: project.PhaseActive,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 1, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
							{SpriteIdx: 1, SpriteFrameIdx: 2, OffsetX: -8, OffsetY: 12, Rotation: -45, ScaleX: 2, ScaleY: 1.25, OriginX: 0.5, OriginY: 0.5},
						},
					},
					{
						Phase: project.PhaseRecover,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 2, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
							{SpriteIdx: 1, SpriteFrameIdx: 3, OffsetX: 0, OffsetY: 0, Rotation: 90, ScaleX: 1, ScaleY: 1, OriginX: 0.75, OriginY: 0.25},
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test_multi.go")

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

	want := orig.Animations[0]
	if len(got.Animations) != 1 {
		t.Fatalf("anim count: got %d, want 1", len(got.Animations))
	}
	gotA := got.Animations[0]
	if len(gotA.Frames) != len(want.Frames) {
		t.Fatalf("frame count: got %d, want %d", len(gotA.Frames), len(want.Frames))
	}
	for j, wantF := range want.Frames {
		gotF := gotA.Frames[j]
		if len(gotF.Sprites) != len(wantF.Sprites) {
			t.Fatalf("frame[%d] sprite count: got %d, want %d", j, len(gotF.Sprites), len(wantF.Sprites))
		}
		for si, wantS := range wantF.Sprites {
			gotS := gotF.Sprites[si]
			if gotS.SpriteIdx != wantS.SpriteIdx {
				t.Errorf("frame[%d].sprite[%d].SpriteIdx: got %d, want %d", j, si, gotS.SpriteIdx, wantS.SpriteIdx)
			}
			if gotS.SpriteFrameIdx != wantS.SpriteFrameIdx {
				t.Errorf("frame[%d].sprite[%d].SpriteFrameIdx: got %d, want %d", j, si, gotS.SpriteFrameIdx, wantS.SpriteFrameIdx)
			}
			if gotS.OffsetX != wantS.OffsetX {
				t.Errorf("frame[%d].sprite[%d].OffsetX: got %v, want %v", j, si, gotS.OffsetX, wantS.OffsetX)
			}
			if gotS.OffsetY != wantS.OffsetY {
				t.Errorf("frame[%d].sprite[%d].OffsetY: got %v, want %v", j, si, gotS.OffsetY, wantS.OffsetY)
			}
			if gotS.Rotation != wantS.Rotation {
				t.Errorf("frame[%d].sprite[%d].Rotation: got %v, want %v", j, si, gotS.Rotation, wantS.Rotation)
			}
			if gotS.ScaleX != wantS.ScaleX {
				t.Errorf("frame[%d].sprite[%d].ScaleX: got %v, want %v", j, si, gotS.ScaleX, wantS.ScaleX)
			}
			if gotS.ScaleY != wantS.ScaleY {
				t.Errorf("frame[%d].sprite[%d].ScaleY: got %v, want %v", j, si, gotS.ScaleY, wantS.ScaleY)
			}
			if gotS.OriginX != wantS.OriginX {
				t.Errorf("frame[%d].sprite[%d].OriginX: got %v, want %v", j, si, gotS.OriginX, wantS.OriginX)
			}
			if gotS.OriginY != wantS.OriginY {
				t.Errorf("frame[%d].sprite[%d].OriginY: got %v, want %v", j, si, gotS.OriginY, wantS.OriginY)
			}
		}
	}
}

func TestMovementRoundTrip(t *testing.T) {
	orig := &project.ProjectData{
		AssetName:       "TestMovement",
		AssetKey:        "test_movement",
		DefaultOriginX:  0.506,
		DefaultOriginY:  0.586,
		Sprites: []project.SpriteRow{
			{Name: "Sprite 0", File: "sprites/players/sayaka_miki/sayaka_miki.png", Width: 256, Height: 256, FrameCount: 25},
			{Name: "Sprite 1", File: "sprites/players/sayaka_miki/weapon.png", Width: 128, Height: 128, FrameCount: 4},
		},
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
	if len(got.Sprites) != len(orig.Sprites) {
		t.Fatalf("sprite count: got %d, want %d", len(got.Sprites), len(orig.Sprites))
	}
	for i, wantS := range orig.Sprites {
		gotS := got.Sprites[i]
		wantName := "base"
		if i > 0 {
			wantName = fmt.Sprintf("sprite %d", i+1)
		}
		if gotS.Name != wantName {
			t.Errorf("sprites[%d].Name: got %q, want %q", i, gotS.Name, wantName)
		}
		if gotS.File != wantS.File {
			t.Errorf("sprites[%d].File: got %q, want %q", i, gotS.File, wantS.File)
		}
		if gotS.Width != wantS.Width {
			t.Errorf("sprites[%d].Width: got %d, want %d", i, gotS.Width, wantS.Width)
		}
		if gotS.Height != wantS.Height {
			t.Errorf("sprites[%d].Height: got %d, want %d", i, gotS.Height, wantS.Height)
		}
		if gotS.FrameCount != wantS.FrameCount {
			t.Errorf("sprites[%d].FrameCount: got %d, want %d", i, gotS.FrameCount, wantS.FrameCount)
		}
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
