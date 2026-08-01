package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"animprite/internal/project"
)

func TestAttackRoundTrip(t *testing.T) {
	orig := &project.ProjectData{
		AssetName:       "TestAttack",
		AssetKey:        "test_attack",
		DefaultOriginX:  0.506,
		DefaultOriginY:  0.586,
		Sprites: []project.SpriteRow{
			{Name: "base", File: "sprites/players/sayaka_miki/sayaka_miki.png", Width: 256, Height: 256, FrameCount: 25},
			{Name: "sprite 2", File: "sprites/players/sayaka_miki/weapon.png", Width: 128, Height: 128, FrameCount: 4},
		},
		Animations: []project.AnimationRow{
			{
				Name: "combo", FPS: 12, Loop: false,
				Windup: 150, Active: 80, Recover: 200, Armed: 3000, ArmedFPS: 14,
				Frames: []project.AnimationFrame{
					{
						Phase: project.PhaseWindup,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 8, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
						},
					},
					{
						Phase: project.PhaseActive,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 9, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
						},
					},
					{
						Phase: project.PhaseRecover,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 10, ScaleX: 1, ScaleY: 1, OriginX: 0.5, OriginY: 0.5},
						},
					},
				},
			},
			{
				Name: "hold", FPS: 12, Loop: true,
				Windup: 100, Active: 60, Recover: 100, Armed: 5000, ArmedFPS: 20,
				Frames: []project.AnimationFrame{
					{
						Phase: project.PhaseArmed,
						Sprites: []project.FrameSpriteEntry{
							{SpriteIdx: 0, SpriteFrameIdx: 11, OffsetX: 5, OffsetY: -3, Rotation: 10, ScaleX: 1.5, ScaleY: 0.75, OriginX: 0.25, OriginY: 0.75},
							{SpriteIdx: 1, SpriteFrameIdx: 2, OffsetX: -8, OffsetY: 12, Rotation: -45, ScaleX: 2, ScaleY: 1.25, OriginX: 0.5, OriginY: 0.5},
						},
					},
				},
			},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test_attack.go")

	if err := ExportAttack(path, orig); err != nil {
		t.Fatalf("ExportAttack: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := string(data)
	t.Logf("Generated file:\n%s", text)

	for _, banned := range []string{"AttackFrame{", "PhasePtr", "OffsetX:", "OffsetY:", "SpriteFrames"} {
		if strings.Contains(text, banned) {
			t.Errorf("generated file must not contain %q (constructor-only format)", banned)
		}
	}
	for _, required := range []string{"AttackF(PhaseWindup,", "AttackF(PhaseActive,", "AttackF(PhaseRecover,", "AttackF(PhaseArmed,", "Armed"} {
		if !strings.Contains(text, required) {
			t.Errorf("generated file must contain %q", required)
		}
	}

	got, err := ImportAttack(path)
	if err != nil {
		t.Fatalf("ImportAttack: %v", err)
	}

	if got.AssetName != orig.AssetName {
		t.Errorf("AssetName: got %q, want %q", got.AssetName, orig.AssetName)
	}
	if got.AssetKey != orig.AssetKey {
		t.Errorf("AssetKey: got %q, want %q", got.AssetKey, orig.AssetKey)
	}
	if got.DefaultOriginX != orig.DefaultOriginX || got.DefaultOriginY != orig.DefaultOriginY {
		t.Errorf("origins: got (%v, %v), want (%v, %v)", got.DefaultOriginX, got.DefaultOriginY, orig.DefaultOriginX, orig.DefaultOriginY)
	}
	if len(got.Sprites) != len(orig.Sprites) {
		t.Fatalf("sprite count: got %d, want %d", len(got.Sprites), len(orig.Sprites))
	}

	wantByName := map[string]*project.AnimationRow{}
	for i := range orig.Animations {
		wantByName[orig.Animations[i].Name] = &orig.Animations[i]
	}
	if len(got.Animations) != len(orig.Animations) {
		t.Fatalf("anim count: got %d, want %d", len(got.Animations), len(orig.Animations))
	}
	for i := range got.Animations {
		gotA := &got.Animations[i]
		wantA, ok := wantByName[gotA.Name]
		if !ok {
			t.Errorf("unexpected animation %q", gotA.Name)
			continue
		}
		if gotA.FPS != wantA.FPS {
			t.Errorf("anim[%s].FPS: got %v, want %v", gotA.Name, gotA.FPS, wantA.FPS)
		}
		if gotA.Loop != wantA.Loop {
			t.Errorf("anim[%s].Loop: got %v, want %v", gotA.Name, gotA.Loop, wantA.Loop)
		}
		if gotA.Windup != wantA.Windup || gotA.Active != wantA.Active || gotA.Recover != wantA.Recover {
			t.Errorf("anim[%s] timings: got (wu=%v, atk=%v, rc=%v), want (wu=%v, atk=%v, rc=%v)",
				gotA.Name, gotA.Windup, gotA.Active, gotA.Recover, wantA.Windup, wantA.Active, wantA.Recover)
		}
		if gotA.Armed != wantA.Armed {
			t.Errorf("anim[%s].Armed: got %v, want %v", gotA.Name, gotA.Armed, wantA.Armed)
		}
		if gotA.ArmedFPS != wantA.ArmedFPS {
			t.Errorf("anim[%s].ArmedFPS: got %v, want %v", gotA.Name, gotA.ArmedFPS, wantA.ArmedFPS)
		}
		if len(gotA.Frames) != len(wantA.Frames) {
			t.Fatalf("anim[%s] frame count: got %d, want %d", gotA.Name, len(gotA.Frames), len(wantA.Frames))
		}
		for j := range wantA.Frames {
			wantF := &wantA.Frames[j]
			gotF := &gotA.Frames[j]
			if gotF.Phase != wantF.Phase {
				t.Errorf("anim[%s].frame[%d].Phase: got %v, want %v", gotA.Name, j, gotF.Phase, wantF.Phase)
			}
			if len(gotF.Sprites) != len(wantF.Sprites) {
				t.Fatalf("anim[%s].frame[%d] sprite count: got %d, want %d", gotA.Name, j, len(gotF.Sprites), len(wantF.Sprites))
			}
			for si, wantS := range wantF.Sprites {
				gotS := gotF.Sprites[si]
				if gotS.SpriteIdx != wantS.SpriteIdx || gotS.SpriteFrameIdx != wantS.SpriteFrameIdx {
					t.Errorf("anim[%s].frame[%d].sprite[%d] idx: got (%d,%d), want (%d,%d)",
						gotA.Name, j, si, gotS.SpriteIdx, gotS.SpriteFrameIdx, wantS.SpriteIdx, wantS.SpriteFrameIdx)
				}
				if gotS.OffsetX != wantS.OffsetX || gotS.OffsetY != wantS.OffsetY {
					t.Errorf("anim[%s].frame[%d].sprite[%d] offset: got (%v,%v), want (%v,%v)",
						gotA.Name, j, si, gotS.OffsetX, gotS.OffsetY, wantS.OffsetX, wantS.OffsetY)
				}
				if gotS.Rotation != wantS.Rotation {
					t.Errorf("anim[%s].frame[%d].sprite[%d].Rotation: got %v, want %v", gotA.Name, j, si, gotS.Rotation, wantS.Rotation)
				}
				if gotS.ScaleX != wantS.ScaleX || gotS.ScaleY != wantS.ScaleY {
					t.Errorf("anim[%s].frame[%d].sprite[%d] scale: got (%v,%v), want (%v,%v)",
						gotA.Name, j, si, gotS.ScaleX, gotS.ScaleY, wantS.ScaleX, wantS.ScaleY)
				}
				if gotS.OriginX != wantS.OriginX || gotS.OriginY != wantS.OriginY {
					t.Errorf("anim[%s].frame[%d].sprite[%d] origin: got (%v,%v), want (%v,%v)",
						gotA.Name, j, si, gotS.OriginX, gotS.OriginY, wantS.OriginX, wantS.OriginY)
				}
			}
		}
	}
}

func writeTempGo(t *testing.T, name, src string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAttackRejectsLegacyStructFrame(t *testing.T) {
	src := `package attacks

import . "madokita/internal/animation"

var Legacy = Attack{
	AssetKey: "legacy",
	Animations: map[string]AttackAnimDef{
		"combo": AttackAnim(12, false, 150, 80, 200, 1, 1, 1,
			AttackFrame{
				Sprites: []FrameSprite{S(0, 8)},
				OffsetX: []float64{1},
				OffsetY: []float64{-12},
				Phase:   PhasePtr(PhaseWindup),
			},
		),
	},
}
`
	if _, err := ImportAttack(writeTempGo(t, "legacy.go", src)); err == nil {
		t.Fatal("expected error for legacy AttackFrame{...} struct literal, got nil")
	}
}

func TestAttackRejectsOldArgumentOrder(t *testing.T) {
	src := `package attacks

import . "madokita/internal/animation"

var Legacy = Attack{
	AssetKey: "legacy",
	Animations: map[string]AttackAnimDef{
		"combo": AttackAnim(12, false, 150, 80, 200, 3000, 14, 1, 1, 1,
			AttackF(S(0, 2), PhaseWindup),
		),
	},
}
`
	if _, err := ImportAttack(writeTempGo(t, "old_order.go", src)); err == nil {
		t.Fatal("expected error for AttackF(S(...), Phase) order, got nil")
	}
}

func TestAttackRejectsSpriteLessFrame(t *testing.T) {
	src := `package attacks

import . "madokita/internal/animation"

var Legacy = Attack{
	AssetKey: "legacy",
	Animations: map[string]AttackAnimDef{
		"combo": AttackAnim(12, false, 150, 80, 200, 3000, 14, 1, 1, 1,
			AttackF(PhaseRecover),
		),
	},
}
`
	if _, err := ImportAttack(writeTempGo(t, "no_sprite.go", src)); err == nil {
		t.Fatal("expected error for sprite-less AttackF(...), got nil")
	}
}
