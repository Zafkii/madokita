package animation

import "testing"

func verticalTestAttack() Attack {
	return Attack{
		Animations: map[string]AttackAnimDef{
			"atk": AttackAnim(10, false, 0, 0, 0, 1000, 4, 2, 2, 1,
				AttackF(PhaseWindup, S(0, 0)),
				AttackF(PhaseWindup, S(0, 1)),
				AttackF(PhaseActive, S(0, 2)),
				AttackF(PhaseActive, S(0, 3)),
				AttackF(PhaseRecover, S(0, 4)),
				AttackF(PhaseArmed, S(0, 5)),
				AttackF(PhaseArmed, S(0, 6)),
			),
		},
	}
}

func TestAttackAnimatorPlayUnknownAnimation(t *testing.T) {
	a := NewAttackAnimator(verticalTestAttack())
	if a.PlayAnimation("nope") {
		t.Fatal("PlayAnimation should fail for unknown animation")
	}
	if a.IsPlaying() {
		t.Fatal("animator should not be playing after failed start")
	}
}

func TestAttackAnimatorPhaseTransitions(t *testing.T) {
	a := NewAttackAnimator(verticalTestAttack())
	if !a.PlayAnimation("atk") {
		t.Fatal("PlayAnimation failed")
	}

	cases := []struct {
		name       string
		advance    float64
		wantPhase  AttackPhase
		wantSprite int
	}{
		{"windup starts", 0, PhaseWindup, 0},
		{"windup second frame", 0.1, PhaseWindup, 1},
		{"active begins", 0.1, PhaseActive, 2},
		{"active second frame", 0.1, PhaseActive, 3},
		{"recover begins", 0.1, PhaseRecover, 4},
		{"armed begins", 0.1, PhaseArmed, 5},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			a.Update(tt.advance)
			if got := a.Phase(); got != tt.wantPhase {
				t.Errorf("Phase() = %q, want %q", got, tt.wantPhase)
			}
			f := a.CurrentFrame()
			if f == nil {
				t.Fatal("CurrentFrame() = nil")
			}
			if len(f.Sprites) == 0 {
				t.Fatal("frame has no sprites")
			}
			if got := f.Sprites[0].SpriteFrameIdx; got != tt.wantSprite {
				t.Errorf("sprite frame = %d, want %d", got, tt.wantSprite)
			}
			if a.Done() {
				t.Error("attack ended during playback")
			}
		})
	}
}

func TestAttackAnimatorArmedLoopsAtArmedFPS(t *testing.T) {
	a := NewAttackAnimator(verticalTestAttack())
	if !a.PlayAnimation("atk") {
		t.Fatal("PlayAnimation failed")
	}
	a.Update(0.5) // 5 non-armed frames at 10fps

	if a.Phase() != PhaseArmed {
		t.Fatalf("expected armed phase, got %q", a.Phase())
	}

	// ArmedFPS=4 → 0.25s per frame over 2 armed frames.
	a.Update(0.25)
	if f := a.CurrentFrame(); f == nil || f.Sprites[0].SpriteFrameIdx != 6 {
		t.Errorf("armed frame after 0.25s: got %v, want sprite 6", spriteOf(a))
	}
	a.Update(0.25) // elapsed 0.5s → wraps back to the first armed frame
	if f := a.CurrentFrame(); f == nil || f.Sprites[0].SpriteFrameIdx != 5 {
		t.Errorf("armed frame after 0.5s: got %v, want wrapped sprite 5", spriteOf(a))
	}
	if a.Done() {
		t.Error("attack should still be in armed loop (1s armed duration)")
	}
}

func TestAttackAnimatorDoneAfterArmedDuration(t *testing.T) {
	a := NewAttackAnimator(verticalTestAttack())
	if !a.PlayAnimation("atk") {
		t.Fatal("PlayAnimation failed")
	}
	a.Update(0.5) // reach armed
	a.Update(0.5) // armed elapsed 0.5s of 1s
	if a.Done() {
		t.Fatal("attack ended before the armed duration elapsed")
	}
	a.Update(0.5) // armed elapsed 1.0s
	if !a.Done() {
		t.Fatal("attack should be done after the armed duration")
	}
	if f := a.CurrentFrame(); f != nil {
		t.Error("CurrentFrame() should be nil once done")
	}
}

func TestAttackAnimatorDoneWithoutArmedPhase(t *testing.T) {
	def := Attack{Animations: map[string]AttackAnimDef{
		"atk": AttackAnim(10, false, 0, 0, 0, 0, 0, 2, 0, 0,
			AttackF(PhaseWindup, S(0, 0)),
			AttackF(PhaseWindup, S(0, 1)),
		),
	}}
	a := NewAttackAnimator(def)
	if !a.PlayAnimation("atk") {
		t.Fatal("PlayAnimation failed")
	}
	a.Update(0.1)
	if a.Done() {
		t.Fatal("attack ended before the last frame")
	}
	a.Update(0.1)
	if !a.Done() {
		t.Fatal("attack should be done after the last frame")
	}
}

func TestAttackAnimatorNilPhaseDefaultsToWindup(t *testing.T) {
	def := Attack{Animations: map[string]AttackAnimDef{
		"atk": {FPS: 10, Frames: []AttackFrame{{Sprites: []FrameSprite{S(0, 0)}}}},
	}}
	a := NewAttackAnimator(def)
	if !a.PlayAnimation("atk") {
		t.Fatal("PlayAnimation failed")
	}
	if got := a.Phase(); got != PhaseWindup {
		t.Errorf("Phase() = %q, want windup", got)
	}
	if f := a.CurrentFrame(); f == nil {
		t.Fatal("CurrentFrame() = nil")
	}
}

func spriteOf(a *AttackAnimator) any {
	f := a.CurrentFrame()
	if f == nil || len(f.Sprites) == 0 {
		return nil
	}
	return f.Sprites[0].SpriteFrameIdx
}
