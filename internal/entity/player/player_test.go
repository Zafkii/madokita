package player

import (
	"testing"
	"time"

	"madokita/internal/animation"
	"madokita/internal/combat"
)

// commitmentTestDef is a synthetic attack with one frame per phase at 10fps
// (0.1s per frame) and a 1s armed pose.
func commitmentTestDef() animation.Attack {
	return animation.Attack{
		Animations: map[string]animation.AttackAnimDef{
			"atk": animation.AttackAnim(10, false, 0, 0, 0, 1000, 4, 1, 1, 1,
				animation.AttackF(animation.PhaseWindup, animation.S(0, 0)),
				animation.AttackF(animation.PhaseActive, animation.S(0, 1)),
				animation.AttackF(animation.PhaseRecover, animation.S(0, 2)),
				animation.AttackF(animation.PhaseArmed, animation.S(0, 3)),
			),
		},
	}
}

func commitmentTestPlayer() *Player {
	stats := combat.NewStats()
	stats.MaxStamina = 100
	stats.Stamina = 100
	actor := combat.NewActor("p1", combat.TeamPlayer, &stats, nil)
	cfg := combat.AttackConfig{
		ID:          "atk",
		Animation:   "atk",
		Windup:      100 * time.Millisecond,
		ActiveTime:  100 * time.Millisecond,
		Recover:     100 * time.Millisecond,
		StaminaCost: 0,
	}
	p := &Player{
		Scale:   1,
		Tracker: combat.NewTracker(0),
	}
	p.Combat.configs = []combat.AttackConfig{cfg}
	p.Combat.controllers = map[string]*combat.Controller{
		"atk": combat.NewController(cfg, combat.NewHitbox(combat.HitboxConfig{Damage: 10}, "p1"), actor),
	}
	p.attackAnim = animation.NewAttackAnimator(commitmentTestDef())
	p.defaultAttackID = "atk"
	return p
}

func TestCommitmentLocksFollowPhases(t *testing.T) {
	p := commitmentTestPlayer()
	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}

	cases := []struct {
		name           string
		advance        float64
		wantMoveLocked bool
		wantAnimLocked bool
	}{
		{"windup", 0.05, true, true},
		{"active", 0.05, true, true},
		{"recover", 0.1, false, true},
		{"armed", 0.1, false, false},
	}
	for _, tt := range cases {
		p.updateAttacking(tt.advance)
		if p.State.IsMovementLocked != tt.wantMoveLocked {
			t.Errorf("%s: IsMovementLocked = %v, want %v", tt.name, p.State.IsMovementLocked, tt.wantMoveLocked)
		}
		if p.State.IsAnimationLocked != tt.wantAnimLocked {
			t.Errorf("%s: IsAnimationLocked = %v, want %v", tt.name, p.State.IsAnimationLocked, tt.wantAnimLocked)
		}
	}
}

func TestTryAttackGateByCommitment(t *testing.T) {
	p := commitmentTestPlayer()
	if !p.TryAttack("atk") {
		t.Fatal("first attack should start")
	}

	p.updateAttacking(0.05) // windup
	if p.TryAttack("atk") {
		t.Error("attack must be rejected during windup")
	}

	p.updateAttacking(0.1) // active
	p.updateAttacking(0.1) // recover
	if p.TryAttack("atk") {
		t.Error("attack must be rejected during recover")
	}

	p.updateAttacking(0.1) // armed begins
	if !p.TryAttack("atk") {
		t.Fatal("re-attack must be allowed once the armed pose begins")
	}
	if p.attackAnim.Phase() != animation.PhaseWindup {
		t.Errorf("re-attack should restart at windup, got %q", p.attackAnim.Phase())
	}
	if !p.State.IsMovementLocked || !p.State.IsAnimationLocked {
		t.Error("re-attack must re-lock movement and animation")
	}
}

func TestTryAttackRejectedWhenStaggeredOrDead(t *testing.T) {
	p := commitmentTestPlayer()
	p.State.IsStaggered = true
	if p.TryAttack("atk") {
		t.Error("attack must be rejected while staggered")
	}
	p.State.IsStaggered = false
	p.State.IsDead = true
	if p.TryAttack("atk") {
		t.Error("attack must be rejected while dead")
	}
}

func TestAttackEndsAfterArmedDuration(t *testing.T) {
	p := commitmentTestPlayer()
	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}

	// Advance one frame at a time (0.1s at 10fps); single large dt values can
	// leave float-point residue in the animator's frame timer.
	p.updateAttacking(0.1) // windup
	p.updateAttacking(0.1) // active
	p.updateAttacking(0.1) // recover
	p.updateAttacking(0.1) // armed begins
	if !p.State.IsAttacking {
		t.Fatal("attack should still be active in the armed pose")
	}
	if p.State.IsAnimationLocked {
		t.Error("armed pose must not lock new attacks")
	}

	p.updateAttacking(1.0) // armed duration (1s) elapsed
	if p.State.IsAttacking {
		t.Error("attack must end after the armed duration")
	}
	if p.Combat.currentID != "" {
		t.Errorf("currentID = %q, want empty", p.Combat.currentID)
	}
}
