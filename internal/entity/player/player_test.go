package player

import (
	"testing"
	"time"

	"madokita-gems/internal/animation"
	"madokita-gems/internal/combat"
	math2 "madokita-gems/internal/math"
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

func TestGravityContinuesDuringMidAirAttack(t *testing.T) {
	p := commitmentTestPlayer()
	p.State.IsControlled = true
	p.StageGroundY = 1e6 // never lands
	p.Movement.gravity = 1000
	p.Movement.jumpVelocity = -500
	p.Movement.yVelocity = -50
	p.Movement.IsGrounded = false
	p.Y = 100

	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}
	p.Update(time.Second / 10) // windup frame, movement locked
	if !p.State.IsMovementLocked {
		t.Fatal("windup should lock movement")
	}
	// Gravity must apply even while attacking mid-air.
	if p.Movement.yVelocity <= -50 {
		t.Errorf("yVelocity = %v, want it to increase (fall) during windup", p.Movement.yVelocity)
	}
	if p.Y <= 100 {
		t.Errorf("Y = %v, want it to increase during windup", p.Y)
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
	// The phase switch must not re-lock the player when the animator reports
	// done (Phase() falls back to windup): all locks must be released.
	if p.State.IsMovementLocked {
		t.Error("IsMovementLocked must be released after the attack ends")
	}
	if p.State.IsAnimationLocked {
		t.Error("IsAnimationLocked must be released after the attack ends")
	}
	// And the player must be able to attack again.
	if !p.TryAttack("atk") {
		t.Error("re-attack must be allowed after the attack ended naturally")
	}
}

func TestSpawnAtAboveGroundStartsAirborne(t *testing.T) {
	p := &Player{Scale: 1}
	p.SpawnAt(320, 480, 700)

	if p.X != 320 || p.Y != 480 {
		t.Fatalf("spawn = (%v, %v), want (320, 480)", p.X, p.Y)
	}
	if p.Movement.IsGrounded {
		t.Error("spawn above the ground line must start airborne")
	}
}

func TestSpawnAtBelowGroundClampsToGround(t *testing.T) {
	p := &Player{Scale: 1}
	p.SpawnAt(320, 900, 700)

	if p.Y != 700 {
		t.Fatalf("Y = %v, want clamp to 700", p.Y)
	}
	if !p.Movement.IsGrounded {
		t.Error("clamped spawn must be grounded")
	}
}

func TestSpawnUsesLogicalGroundLine(t *testing.T) {
	p := &Player{Scale: 1}
	p.Spawn(123, 700)

	if p.X != 123 {
		t.Errorf("X = %v, want 123", p.X)
	}
	if p.Y != 700 {
		t.Errorf("Y = %v, want the logical ground line 700", p.Y)
	}
	if p.StageGroundY != 700 {
		t.Errorf("StageGroundY = %v, want 700", p.StageGroundY)
	}
	if !p.Movement.IsGrounded {
		t.Error("spawned player must start grounded")
	}
}

func TestGroundClampUsesPlayerY(t *testing.T) {
	p := &Player{
		State:        State{IsControlled: true, IsMovementLocked: true},
		Movement:     Movement{gravity: 1000, yVelocity: 1000},
		StageGroundY: 700,
		Y:            650,
		Scale:        1,
	}
	p.Update(100 * time.Millisecond)

	if p.Y != 700 {
		t.Errorf("Y = %v, want clamped to 700", p.Y)
	}
	if !p.Movement.IsGrounded {
		t.Error("player must be grounded after reaching the ground line")
	}
}

func TestSetWorldBoundsClampsPosition(t *testing.T) {
	p := &Player{
		StageGroundY: 700,
		X:            150,
		Y:            -10,
		Movement:     Movement{yVelocity: -100},
	}
	p.SetWorldBounds(100, 720)
	p.clampWorldBounds()

	if p.X != 100 {
		t.Errorf("X = %v, want clamped to 100", p.X)
	}
	if p.Y != 0 {
		t.Errorf("Y = %v, want clamped to 0", p.Y)
	}
	if p.Movement.yVelocity != 0 {
		t.Errorf("yVelocity = %v, want 0", p.Movement.yVelocity)
	}
}

func TestWorldBoundsClampBottom(t *testing.T) {
	p := &Player{
		StageGroundY: 700,
		WorldWidth:   100,
		WorldHeight:  720,
		X:            -5,
		Y:            800,
		Movement:     Movement{yVelocity: 100},
	}
	p.clampWorldBounds()

	if p.X != 0 {
		t.Errorf("X = %v, want clamped to 0", p.X)
	}
	if p.Y != 720 {
		t.Errorf("Y = %v, want clamped to 720", p.Y)
	}
	if p.Movement.yVelocity != 0 {
		t.Errorf("yVelocity = %v, want 0", p.Movement.yVelocity)
	}
}

func TestPlayerLandsOnSolidStageBox(t *testing.T) {
	p := &Player{
		State:        State{IsControlled: true, IsMovementLocked: true},
		Movement:     Movement{gravity: 1000},
		Scale:        0.5,
		X:            50,
		Y:            0,
		StageGroundY: 1000,
	}
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 200, 50)})

	for i := 0; i < 300; i++ {
		p.Update(10 * time.Millisecond)
	}

	wantY := 100 - p.bodyBottomOffset()
	if diff := p.Y - wantY; diff < -0.01 || diff > 0.01 {
		t.Fatalf("Y = %v, want landing on box top at %v", p.Y, wantY)
	}
	if !p.Movement.IsGrounded {
		t.Fatal("player must be grounded after landing on a solid box")
	}
	if p.Movement.yVelocity != 0 {
		t.Fatalf("yVelocity = %v, want 0", p.Movement.yVelocity)
	}
}

func TestPlayerStopsAtSolidWall(t *testing.T) {
	p := &Player{Scale: 0.5, X: 20, Y: 36.5}
	p.SetSolidRects([]math2.Rect{math2.NewRect(100, 0, 50, 200)})
	p.moveHorizontal(100)

	wantX := 100 - (p.bodyRect().X + p.bodyRect().W - p.X)
	if diff := p.X - wantX; diff < -0.01 || diff > 0.01 {
		t.Fatalf("X = %v, want blocked at wall left edge %v", p.X, wantX)
	}
}

func TestPlayerSupportRequiresCurrentSurface(t *testing.T) {
	p := &Player{Scale: 0.5, X: 50, Y: 36.5, StageGroundY: 1000}
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 100, 50)})
	if !p.hasSupport() {
		t.Fatal("player standing on the box must have support")
	}

	p.X = 200
	if p.hasSupport() {
		t.Fatal("player past the box edge must lose support")
	}
}
