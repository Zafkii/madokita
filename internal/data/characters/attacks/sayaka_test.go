package attacks

import "testing"

// The attack definition is a mechanical copy of the animprite export, so this
// test pins the parts the editor owns: every frame carries its sprite and its
// authored hurtboxes, and only the active frames carry attack hitboxes.
func TestSayakaVerticalAttackCarriesTheExportedBoxes(t *testing.T) {
	anim, ok := SayakaVerticalAttack.Animations["vertical_attack"]
	if !ok {
		t.Fatalf("vertical_attack missing from %v", SayakaVerticalAttack.Animations)
	}
	if len(anim.Frames) != 18 {
		t.Fatalf("frames = %d, want 18", len(anim.Frames))
	}
	if anim.Windup != 20 || anim.ActiveTime != 80 || anim.Recover != 150 || anim.Armed != 3000 {
		t.Fatalf("timings = %+v, want windup 20 active 80 recover 150 armed 3000", anim)
	}

	activeFrames := 0
	for i, f := range anim.Frames {
		if len(f.Sprites) == 0 {
			t.Errorf("frame %d has no sprite", i)
		}
		if len(f.Hurtboxes) == 0 {
			t.Errorf("frame %d has no authored hurtbox", i)
		}
		if f.Phase == nil {
			t.Fatalf("frame %d has no phase", i)
		}
		if *f.Phase == "atk" {
			activeFrames++
			if len(f.Hitboxes) == 0 {
				t.Errorf("active frame %d has no authored hitbox", i)
			}
			continue
		}
		if len(f.Hitboxes) != 0 {
			t.Errorf("frame %d (%s) should not carry hitboxes", i, *f.Phase)
		}
	}
	if activeFrames != 5 {
		t.Fatalf("active frames = %d, want 5", activeFrames)
	}

	// The last active frame keeps the width the editor exported last.
	last := anim.Frames[10].Hitboxes
	if len(last) != 1 || last[0].W != 193 || last[0].H != 222 || last[0].OffsetX != 126 || last[0].OffsetY != 18 {
		t.Fatalf("frame 10 hitbox = %+v, want 193x222 at (126,18)", last)
	}
}

// Geometry must come only from the editor frames: a box in the config would
// silently draw a second, duplicate hitbox next to the authored one.
func TestSayakaAttackConfigCarriesNoGeometry(t *testing.T) {
	for _, cfg := range SayakaAttackConfigs {
		if cfg.Hitbox.W != 0 || cfg.Hitbox.H != 0 || cfg.Hitbox.OffsetX != 0 || cfg.Hitbox.OffsetY != 0 {
			t.Fatalf("config %q carries geometry %+v, want zero (frames own it)", cfg.ID, cfg.Hitbox)
		}
		if cfg.Hitbox.Damage == 0 {
			t.Fatalf("config %q lost its damage values", cfg.ID)
		}
	}
}
