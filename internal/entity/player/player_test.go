package player

import (
	"math"
	"testing"
	"time"

	"madokita-gems/internal/animation"
	"madokita-gems/internal/combat"
	"madokita-gems/internal/game/stages"
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

func TestPlayerStepsUpSmallLedge(t *testing.T) {
	p := &Player{Scale: 0.5, X: 50, Y: 36.5, Movement: Movement{IsGrounded: true}}
	floor := math2.NewRect(0, 100, 400, 50)
	step := math2.NewRect(100, 90, 300, 200)
	p.SetSolidRects([]math2.Rect{floor, step})

	// The base rides the low ground, so the body is buried in the 10px step while
	// the step sits only ahead of the trailing edge, and it keeps walking.
	p.moveHorizontal(60)

	if diff := p.Y - 36.5; diff < -0.01 || diff > 0.01 {
		t.Fatalf("Y = %v, want the base still on the lower floor at %v", p.Y, 36.5)
	}
	if p.X != 110 {
		t.Fatalf("X = %v, want the player to keep walking into the step", p.X)
	}
	if !p.hasSupport() {
		t.Fatal("player walking into a small step must stay grounded")
	}

	// Once the trailing edge reaches the step the base climbs onto it.
	p.moveHorizontal(60) // X = 170: the trailing edge is on the step.

	if diff := p.Y - 26.5; diff < -0.01 || diff > 0.01 {
		t.Fatalf("Y = %v, want the player raised onto the 10px step at %v", p.Y, 26.5)
	}
	if !p.hasSupport() {
		t.Fatal("player raised onto a small step must stay grounded")
	}
}

func TestPlayerDoesNotStepUpTallWall(t *testing.T) {
	p := &Player{Scale: 0.5, X: 50, Y: 36.5, Movement: Movement{IsGrounded: true}}
	floor := math2.NewRect(0, 100, 200, 50)
	wall := math2.NewRect(100, 80, 50, 200)
	p.SetSolidRects([]math2.Rect{floor, wall})

	p.moveHorizontal(60)

	wantX := 100 - (p.bodyRect().X + p.bodyRect().W - p.X)
	if diff := p.X - wantX; diff < -0.01 || diff > 0.01 {
		t.Fatalf("X = %v, want blocked at wall left edge %v", p.X, wantX)
	}
	if diff := p.Y - 36.5; diff < -0.01 || diff > 0.01 {
		t.Fatalf("Y = %v, want the player to stay on the floor", p.Y)
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

func TestActiveHitboxesFollowAttackPhase(t *testing.T) {
	p := commitmentTestPlayer()
	p.X, p.Y = 100, 200
	p.Combat.controllers["atk"].Hitbox().Config = combat.HitboxConfig{
		W: 50, H: 40, OffsetX: 10, OffsetY: 5,
	}

	if got := p.ActiveHitboxes(); len(got) != 0 {
		t.Fatalf("idle hitboxes = %v, want none", got)
	}

	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}
	if got := p.ActiveHitboxes(); len(got) != 0 {
		t.Fatalf("windup hitboxes = %v, want none", got)
	}

	p.updateAttacking(0.05) // windup frame
	p.updateAttacking(0.05) // first active frame

	got := p.ActiveHitboxes()
	if len(got) != 1 {
		t.Fatalf("active hitboxes = %v, want exactly one", got)
	}
	want := math2.NewRect(100+10-25, 200+5-20, 50, 40)
	if got[0] != want {
		t.Fatalf("active hitbox = %v, want %v", got[0], want)
	}

	p.FlipX = true
	flipped := p.ActiveHitboxes()
	if len(flipped) != 1 {
		t.Fatalf("flipped hitboxes = %v, want exactly one", flipped)
	}
	wantFlip := math2.NewRect(100-10-25, 200+5-20, 50, 40)
	if flipped[0] != wantFlip {
		t.Fatalf("flipped hitbox = %v, want %v", flipped[0], wantFlip)
	}

	p.updateAttacking(0.1) // recover frame
	if got := p.ActiveHitboxes(); len(got) != 0 {
		t.Fatalf("recover hitboxes = %v, want none", got)
	}
}

// slopeStrips builds a staircase of axis-aligned strips like the ones the stage
// runtime produces for a rotated block: each strip is stripW wide and its top
// moves by rise pixels per strip.
func slopeStrips(startX, startTop, stripW, rise float64, count int) []math2.Rect {
	rects := make([]math2.Rect, 0, count)
	for i := 0; i < count; i++ {
		x := startX + float64(i)*stripW
		top := startTop + float64(i)*rise
		rects = append(rects, math2.NewRect(x, top, stripW, 200))
	}
	return rects
}

func TestPlayerWalksUpStripSlope(t *testing.T) {
	const (
		stripW = 6.0
		rise   = -1.6 // 15 degrees upwards to the right
	)
	p := &Player{
		State:    State{IsControlled: true},
		Movement: Movement{IsGrounded: true, gravity: 1000},
		Scale:    0.5,
		X:        80,
	}
	p.Y = 100 - p.bodyBottomOffset()
	p.SetSolidRects(append(
		[]math2.Rect{math2.NewRect(0, 100, 100, 200)},
		slopeStrips(100, 100, stripW, rise, 40)...))

	for i := 0; i < 48; i++ {
		p.moveHorizontal(5)
		p.Update(10 * time.Millisecond)
		if !p.Movement.IsGrounded {
			t.Fatalf("frame %d: player left the ground while climbing", i)
		}
	}

	feet := p.bodyRect().Y + p.bodyRect().H
	if feet > 100-30 {
		t.Fatalf("feet = %v, want the player carried up the slope", feet)
	}
	if p.X < 320 {
		t.Fatalf("X = %v, want the player to keep walking up the slope", p.X)
	}
}

func TestPlayerWalksDownStripSlopeWithoutLeavingGround(t *testing.T) {
	const (
		stripW = 6.0
		drop   = 1.6
	)
	p := &Player{
		State:    State{IsControlled: true},
		Movement: Movement{IsGrounded: true, gravity: 1000},
		Scale:    0.5,
		X:        80,
	}
	p.Y = 100 - p.bodyBottomOffset()
	p.SetSolidRects(append(
		[]math2.Rect{math2.NewRect(0, 100, 100, 200)},
		slopeStrips(100, 100, stripW, drop, 40)...))

	for i := 0; i < 40; i++ {
		p.moveHorizontal(5)
		p.Update(10 * time.Millisecond)
		if !p.Movement.IsGrounded {
			t.Fatalf("frame %d: player became airborne walking down a slope", i)
		}
	}

	if p.X < 280 {
		t.Fatalf("X = %v, want the player to keep walking down the slope", p.X)
	}
}

func TestPlayerSnapsDownSmallStep(t *testing.T) {
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 50}
	p.Y = 100 - p.bodyBottomOffset()
	p.Movement = Movement{IsGrounded: true, gravity: 1000}
	p.SetSolidRects([]math2.Rect{
		math2.NewRect(0, 100, 100, 50),
		math2.NewRect(100, 108, 200, 50),
	})

	p.X = 130
	p.Update(10 * time.Millisecond)

	if !p.Movement.IsGrounded {
		t.Fatal("an 8px drop must be followed while walking, not become airborne")
	}
	wantY := 108 - p.bodyBottomOffset()
	if diff := p.Y - wantY; diff < -0.01 || diff > 0.01 {
		t.Fatalf("Y = %v, want the player snapped to the lower surface at %v", p.Y, wantY)
	}
}

func TestPlayerFallsOffTallLedge(t *testing.T) {
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 50, StageGroundY: 1000}
	p.Y = 100 - p.bodyBottomOffset()
	p.Movement = Movement{IsGrounded: true, gravity: 1000}
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 100, 50)})

	p.X = 200
	p.Update(10 * time.Millisecond)

	if p.Movement.IsGrounded {
		t.Fatal("walking off a tall ledge must leave the grounded state")
	}
}

// TestPlayerWalksRotatedStageHill walks the real test_stage hill, which is built
// from rotated grass blocks. Rotated collision boxes are decomposed into strips,
// so the walk must stay grounded and never bump more than a few pixels at once.
// TestPlayerKeepsClimbingBaseOnLowGround pins the ascending side of the ground
// probe: the ground ahead is higher, so the base rides the trailing edge and the
// rise buries the front of the base instead of leaving the body airborne. The
// body climbs once its trailing edge reaches the higher ground.
func TestPlayerKeepsClimbingBaseOnLowGround(t *testing.T) {
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 160}
	p.Y = 100 - p.bodyBottomOffset()
	p.Movement = Movement{IsGrounded: true, gravity: 1000}
	p.SetSolidRects([]math2.Rect{
		math2.NewRect(0, 100, 200, 50),
		math2.NewRect(200, 92, 200, 50),
	})

	for p.X < 185 {
		p.moveHorizontal(5)
		p.Update(10 * time.Millisecond)
	}

	body := p.bodyRect()
	feet := body.Y + body.H
	if math.Abs(feet-100) > 0.01 {
		t.Fatalf("feet = %v, want the base held on the lower ground at 100", feet)
	}
	if lead, ok := p.floorAtX(body.X+body.W-probeEdge, body); !ok || lead >= feet {
		t.Fatalf("leading floor = %v (found %v), want it above the base so nothing floats", lead, ok)
	}
	if !p.Movement.IsGrounded {
		t.Fatal("player must stay grounded while climbing")
	}

	for p.X < 225 {
		p.moveHorizontal(5)
		p.Update(10 * time.Millisecond)
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-92) > 0.01 {
		t.Fatalf("feet = %v, want the base raised once the trailing edge reaches the rise", feet)
	}
}

// TestPlayerGroundsThreeQuartersOfTheBase pins the ground probe: the body rests
// where the centre pushed a quarter of its width toward the walk direction meets
// the floor, so the leading quarter may hang past an edge while three quarters
// still stand on the ground.
func TestPlayerGroundsThreeQuartersOfTheBase(t *testing.T) {
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 160}
	p.Y = 100 - p.bodyBottomOffset()
	p.Movement = Movement{IsGrounded: true, gravity: 1000}
	p.SetSolidRects([]math2.Rect{
		math2.NewRect(0, 100, 200, 50),
		math2.NewRect(200, 108, 200, 50),
	})

	const stepW = 5.0
	settle := func() {
		p.moveHorizontal(stepW)
		p.Update(10 * time.Millisecond)
	}

	// probe = X + 0.75 + bodyW/4 crosses the step's edge at X = 187.4, which is
	// exactly where three quarters of the 47.5px base still stand on the upper
	// floor.
	for p.X < 185 {
		settle()
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-100) > 0.01 {
		t.Fatalf("feet = %v, want the body held on the upper floor while three quarters stand on it", feet)
	}
	if !p.Movement.IsGrounded {
		t.Fatal("player must stay grounded while three quarters of the base are supported")
	}

	settle() // X = 190: only the leading quarter hangs over the drop now.
	if p.X != 190 {
		t.Fatalf("X = %v, want 190", p.X)
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-108) > 0.01 {
		t.Fatalf("feet = %v, want the body followed the drop once the probe left the upper floor", feet)
	}
	if !p.Movement.IsGrounded {
		t.Fatal("an 8px drop at the probe must be followed while walking")
	}

	// Mirror: climbing back left the base stays on the lower ground, so the rise
	// buries the front of the base instead of floating the body. It climbs once
	// its trailing edge reaches the higher floor at 200.
	p.X = 210
	p.Y = 108 - p.bodyBottomOffset()
	p.Movement.IsGrounded = true
	p.moveHorizontal(-stepW) // X = 205: the trailing edge is still on the drop.
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-108) > 0.01 {
		t.Fatalf("feet = %v, want the base held on the lower ground at 108", feet)
	}

	for p.X > 175 {
		p.moveHorizontal(-stepW)
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-100) > 0.01 {
		t.Fatalf("feet = %v, want the body raised onto the upper floor", feet)
	}
}

// TestPlayerHoldsTheLedgeDownToTwoPixels pins the ledge tolerance: the release
// follows the whole base, so the drop only starts once the base has left the
// collision instead of stopping while most of the body still stands on it.
func TestPlayerHoldsTheLedgeDownToTwoPixels(t *testing.T) {
	newPlayer := func() *Player {
		p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 160, StageGroundY: 1000}
		p.Y = 100 - p.bodyBottomOffset()
		p.Movement = Movement{IsGrounded: true, gravity: 1000, jumpVelocity: -500}
		p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 200, 50)})
		return p
	}
	// The body spans X-23 .. X+24.5, so the overlap with a ledge ending at 200
	// is 223 - X.
	overlapAt := func(x float64) float64 { return 223 - x }

	// Three pixels of the base still on the ledge: the drop has not released.
	p := newPlayer()
	p.X = 220
	p.Update(10 * time.Millisecond)
	if !p.Movement.IsGrounded {
		t.Fatalf("X = %v: player fell with %vpx of the base on the ledge", p.X, overlapAt(p.X))
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-100) > 0.01 {
		t.Fatalf("feet = %v, want the body held on the ledge top at 100", feet)
	}

	// Exactly the tolerance still supports the body, so walking to the very
	// edge of the collision keeps the jump available.
	p = newPlayer()
	p.X = 200 + (23 - ledgeSupport)
	p.Update(10 * time.Millisecond)
	if !p.Movement.IsGrounded {
		t.Fatalf("X = %v: player fell with %vpx of the base on the ledge, want the tolerance to hold", p.X, ledgeSupport)
	}

	// One pixel less and the drop releases.
	p = newPlayer()
	p.X = 200 + (23 - ledgeSupport) + 1
	p.Update(10 * time.Millisecond)
	if p.Movement.IsGrounded {
		t.Fatalf("X = %v: player held the ledge with %vpx of the base on it", p.X, overlapAt(p.X))
	}
}

// TestPlayerJumpingIntoAStepNeverEntersTheBlock pins the airborne side of the
// step tolerance: the 16px "walkable step" exemption only exists so the grounded
// climb can lift the base onto the step on the same move. In the air nothing
// lifts the body, so an overlapping solid is a wall until the feet clear its top
// and the jump must never insert the body into the block.
func TestPlayerJumpingIntoAStepNeverEntersTheBlock(t *testing.T) {
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 272, StageGroundY: 0}
	p.Movement = Movement{gravity: 1000, jumpVelocity: -500}
	p.Y = 100 - p.bodyBottomOffset()
	p.Movement.IsGrounded = true

	stepTop := 60.0
	step := math2.NewRect(300, stepTop, 700, 140)
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 300, 100), step})

	p.Movement.yVelocity = p.Movement.jumpVelocity
	p.Movement.IsGrounded = false

	for i := 0; i < 120; i++ {
		p.moveHorizontal(5)
		p.Update(16 * time.Millisecond)

		body := p.bodyRect()
		feet := body.Y + body.H
		if body.Overlaps(step) && feet > stepTop+collisionEpsilon {
			t.Fatalf("frame %d: body entered the step (X = %v, feet = %v, step top = %v)", i, p.X, feet, stepTop)
		}
		if p.Movement.IsGrounded {
			break
		}
	}

	// The arc must still carry the jump onto the step instead of stalling at the
	// face forever.
	if !p.Movement.IsGrounded {
		t.Fatalf("player must land on the step, IsGrounded = false (X = %v)", p.X)
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-stepTop) > 0.01 {
		t.Fatalf("feet = %v, want the player standing on the step top at %v", feet, stepTop)
	}
}

// TestPlayerFallingBesideAStepNeverSinksIntoIt pins the corner case: falling
// while the base straddles a step's face. The ground probe must not move to the
// trailing edge just because the ground ahead is higher, and the body must not
// sink into the block while the centre of the base is still outside it.
func TestPlayerFallingBesideAStepNeverSinksIntoIt(t *testing.T) {
	const stepTop = 50.0
	step := math2.NewRect(300, stepTop, 700, 50)
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 265, StageGroundY: 0}
	p.Movement = Movement{gravity: 1000, jumpVelocity: -500}
	p.Y = 100 - p.bodyBottomOffset()
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 300, 100), step})
	p.Movement.IsGrounded = true
	p.Movement.yVelocity = p.Movement.jumpVelocity
	p.Movement.IsGrounded = false

	landed := false
	for i := 0; i < 120; i++ {
		// Steer right only during the fall, which is what makes the front of the
		// base reach the step's face while the feet are already dropping past it.
		if p.Movement.yVelocity > 0 {
			p.moveHorizontal(5)
		}
		p.Update(16 * time.Millisecond)

		body := p.bodyRect()
		feet := body.Y + body.H
		if body.Overlaps(step) && feet > stepTop+collisionEpsilon {
			t.Fatalf("frame %d: body sank into the step (X = %v, feet = %v, top = %v)", i, p.X, feet, stepTop)
		}
		if p.Movement.IsGrounded {
			landed = true
			break
		}
	}
	if !landed {
		t.Fatal("player never landed")
	}
}

// TestPlayerLandsOnlyWhenTheBaseCentreIsOverTheStep pins the landing rule: a fall
// that puts just the leading quarter over a step's corner must not land on the
// step and must not sink into the block either. The base is pushed out of the
// face and slides down to the low ground, and the step is only caught once the
// probe — the centre, with no travel direction — is over it.
func TestPlayerLandsOnlyWhenTheBaseCentreIsOverTheStep(t *testing.T) {
	const stepTop = 50.0
	step := math2.NewRect(300, stepTop, 700, 50)
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 290, StageGroundY: 0}
	p.Movement = Movement{gravity: 1000}
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 300, 100), step})

	// Drop from just above the step with the centre 9px short of its face, so the
	// leading quarter hangs over the corner but the base cannot rest on it.
	p.X = 290
	p.Y = stepTop - 5 - p.bodyBottomOffset()
	p.Movement.yVelocity = 300

	for i := 0; i < 60; i++ {
		p.Update(16 * time.Millisecond)
		body := p.bodyRect()
		feet := body.Y + body.H
		if body.Overlaps(step) && feet > stepTop+collisionEpsilon {
			t.Fatalf("frame %d: body sank into the step (X = %v, feet = %v)", i, p.X, feet)
		}
		if p.Movement.IsGrounded {
			// The base centre never crossed the face, so the only floor it may
			// stand on is the low ground beside the step.
			if math.Abs(feet-100) > 0.01 {
				t.Fatalf("frame %d: stood on feet = %v, want the low ground at 100 (X = %v, centre = %v)", i, feet, p.X, body.X+body.W/2)
			}
			break
		}
	}
	body := p.bodyRect()
	if feet := body.Y + body.H; math.Abs(feet-100) > 0.01 {
		t.Fatalf("feet = %v, want the fall to reach the ground at 100 after sliding off the face", feet)
	}
	if body.X+body.W > step.X+collisionEpsilon {
		t.Fatalf("body front = %v, want the fall pushed out of the step face at %v", body.X+body.W, step.X)
	}
}

// TestPlayerTurningMidFallStillLandsOnTheBlock pins the air-side twin of the
// corner fix: a jump that overshoots a block and turns back on the way down must
// still land on the block. The landing used to follow the travel direction only,
// so the turn moved the single landing point off the corner and the body dropped
// through the block onto the floor.
func TestPlayerTurningMidFallStillLandsOnTheBlock(t *testing.T) {
	const blockTop = 50.0
	block := math2.NewRect(400, blockTop, 50, 100)
	floor := math2.NewRect(0, 100, 1000, 100)
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 350, StageGroundY: 0}
	p.Movement = Movement{gravity: 1000, jumpVelocity: -500}
	p.Y = 100 - p.bodyBottomOffset()
	p.SetSolidRects([]math2.Rect{floor, block})
	p.Movement.IsGrounded = true
	p.Movement.yVelocity = p.Movement.jumpVelocity
	p.Movement.IsGrounded = false

	landed := false
	for i := 0; i < 140; i++ {
		// Jump past the block, then turn back once the fall is under way with the
		// base still above the block: that is the input that used to drop the
		// body through it.
		dir := 5.0
		if i >= 34 {
			dir = -5
		}
		p.moveHorizontal(dir)
		p.Update(16 * time.Millisecond)

		body := p.bodyRect()
		feet := body.Y + body.H
		for _, solid := range []math2.Rect{floor, block} {
			if body.Overlaps(solid) && feet > solid.Y+collisionEpsilon && body.Y < solid.Y+solid.H {
				t.Fatalf("frame %d: body entered the solid with top %v (X = %v, feet = %v)", i, solid.Y, p.X, feet)
			}
		}
		if p.Movement.IsGrounded {
			landed = true
			break
		}
	}
	if !landed {
		t.Fatal("player never landed")
	}
	if feet := p.bodyRect().Y + p.bodyRect().H; math.Abs(feet-blockTop) > 0.01 {
		t.Fatalf("feet = %v, want the player standing on the block top at %v (X = %v)", feet, blockTop, p.X)
	}
}

// TestPlayerTurningInsideAStepFaceIsPushedOutToTheNearSide pins the other half
// of the turn fix: with the base already overlapping a step's face in the air,
// turning back must not sink further into the block, and it must not eject the
// body across the whole block either. That overlap is inherited from the fall,
// so the travel direction cannot pick the way out; the nearer face must.
func TestPlayerTurningInsideAStepFaceIsPushedOutToTheNearSide(t *testing.T) {
	const stepTop = 60.0
	step := math2.NewRect(300, stepTop, 700, 140)
	p := &Player{State: State{IsControlled: true}, Scale: 0.5, X: 310, StageGroundY: 0}
	p.Movement = Movement{gravity: 1000}
	p.Y = 100 - p.bodyBottomOffset()
	p.SetSolidRects([]math2.Rect{math2.NewRect(0, 100, 300, 100), step})

	p.moveHorizontal(-5) // turn back while the base still overlaps the face

	body := p.bodyRect()
	if body.Overlaps(step) && body.Y+body.H > stepTop+collisionEpsilon && body.Y < stepTop+step.H {
		t.Fatalf("body stayed inside the step after turning (X = %v, feet = %v)", p.X, body.Y+body.H)
	}
	if p.X > step.X {
		t.Fatalf("X = %v, want the body pushed out through the near face at %v", p.X, step.X)
	}
}

func TestPlayerWalksRotatedStageHill(t *testing.T) {
	layout := &stages.TestStage
	var rects []math2.Rect
	for i := range layout.Objects {
		rects = append(rects, layout.Objects[i].SolidRects()...)
	}
	if len(rects) < 200 {
		t.Fatalf("stage produced %d solid rects, want rotated blocks decomposed", len(rects))
	}

	const halfBody = 47.5 / 2
	surfaceAt := func(x float64) float64 {
		top := math.Inf(1)
		for _, r := range rects {
			if x+halfBody <= r.X || x-halfBody >= r.X+r.W {
				continue
			}
			top = math.Min(top, r.Y)
		}
		return top
	}

	p := &Player{
		State:    State{IsControlled: true},
		Movement: Movement{IsGrounded: true, gravity: 1000},
		Scale:    0.5,
		X:        20,
	}
	p.Y = surfaceAt(p.X) - p.bodyBottomOffset()
	p.SetSolidRects(rects)

	walk := func(dir float64, frames int) (maxRise, maxDrop float64) {
		prev := p.bodyRect().Y + p.bodyRect().H
		for i := 0; i < frames; i++ {
			p.moveHorizontal(5 * dir)
			p.Update(10 * time.Millisecond)
			if !p.Movement.IsGrounded {
				t.Fatalf("x=%.1f frame %d: player became airborne on the hill", p.X, i)
			}
			feet := p.bodyRect().Y + p.bodyRect().H
			maxRise = math.Max(maxRise, prev-feet)
			maxDrop = math.Max(maxDrop, feet-prev)
			prev = feet
		}
		return maxRise, maxDrop
	}

	// The exported stage now closes the hill with a step taller than
	// maxAutoStepHeight, so the walk must carry the base all the way to that
	// face and stop there instead of crossing it.
	startFeet := p.bodyRect().Y + p.bodyRect().H
	face := math.Inf(1)
	for _, r := range rects {
		if r.Y < startFeet-maxAutoStepHeight && r.X > p.X && r.X < face {
			face = r.X
		}
	}
	if math.IsInf(face, 1) {
		t.Fatal("want a step taller than maxAutoStepHeight ahead of the hill")
	}

	if rise, drop := walk(1, 96); rise > 6 || drop > 6 {
		t.Fatalf("climbing: feet changed %.2f/%.2f px in one frame, want a walkable surface", rise, drop)
	}
	body := p.bodyRect()
	if front := body.X + body.W; math.Abs(front-(face-1)) > 2 {
		t.Fatalf("X = %v front = %v, want the base to reach the step face at %v", p.X, front, face-1)
	}

	if rise, drop := walk(-1, 88); rise > 6 || drop > 6 {
		t.Fatalf("descending: feet changed %.2f/%.2f px in one frame, want a walkable surface", rise, drop)
	}
	if p.X > 64 {
		t.Fatalf("X = %v, want the player to walk back down the hill", p.X)
	}
}

// authoredHitboxDef puts one authored hitbox on the active frame so the
// world-space rects can be checked against the numbers the editor exports.
func authoredHitboxDef() animation.Attack {
	return animation.Attack{
		Animations: map[string]animation.AttackAnimDef{
			"atk": animation.AttackAnim(10, false, 0, 0, 0, 1000, 4, 1, 1, 1,
				animation.AttackF(animation.PhaseWindup, animation.S(0, 0)),
				animation.AttackF(animation.PhaseActive,
					animation.S(0, 1),
					animation.HIT(40, 30, 60, -20),
				),
				animation.AttackF(animation.PhaseRecover, animation.S(0, 2)),
			),
		},
	}
}

func TestCurrentAttackHitboxesFollowsTheActiveFrame(t *testing.T) {
	p := commitmentTestPlayer()
	p.attackAnim = animation.NewAttackAnimator(authoredHitboxDef())
	p.X, p.Y = 100, 200
	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}

	// Windup frame carries no authored hitbox.
	p.updateAttacking(0.05)
	if got := p.CurrentAttackHitboxes(); len(got) != 0 {
		t.Fatalf("windup hitboxes = %+v, want none", got)
	}

	p.updateAttacking(0.05)
	got := p.CurrentAttackHitboxes()
	if len(got) != 1 {
		t.Fatalf("active hitboxes = %d, want 1", len(got))
	}
	// Centre at (160, 180), 40x30 around it.
	want := math2.Rect{X: 140, Y: 165, W: 40, H: 30}
	if got[0] != want {
		t.Fatalf("hitbox = %+v, want %+v", got[0], want)
	}

	// Facing left mirrors the offset around the player.
	p.FlipX = true
	got = p.CurrentAttackHitboxes()
	// Mirroring the +60 offset puts the centre at 40, so the 40-wide box starts
	// at 20.
	if len(got) != 1 || got[0].X != 20 {
		t.Fatalf("flipped hitbox = %+v, want X 20", got)
	}
}

func TestCurrentAttackHitboxesIsEmptyWithoutAuthoredBoxes(t *testing.T) {
	p := commitmentTestPlayer()
	p.X, p.Y = 0, 0
	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}
	p.updateAttacking(0.05)
	p.updateAttacking(0.05)
	if got := p.CurrentAttackHitboxes(); len(got) != 0 {
		t.Fatalf("hitboxes = %+v, want none for an attack without authored boxes", got)
	}
}

// The frame's authored hitbox is what the game resolves: the config geometry
// is only a fallback for defs that predate the editor.
func TestActiveHitboxesUseTheAuthoredFrameGeometry(t *testing.T) {
	p := commitmentTestPlayer()
	p.attackAnim = animation.NewAttackAnimator(authoredHitboxDef())
	p.X, p.Y = 100, 200
	p.Combat.controllers["atk"].Hitbox().Config = combat.HitboxConfig{
		W: 999, H: 999, OffsetX: 500, OffsetY: 500,
	}

	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}
	p.updateAttacking(0.05) // windup frame
	p.updateAttacking(0.05) // first active frame

	got := p.ActiveHitboxes()
	want := math2.Rect{X: 140, Y: 165, W: 40, H: 30}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("active hitboxes = %v, want the authored %v", got, want)
	}

	p.FlipX = true
	flipped := p.ActiveHitboxes()
	wantFlip := math2.Rect{X: 20, Y: 165, W: 40, H: 30}
	if len(flipped) != 1 || flipped[0] != wantFlip {
		t.Fatalf("flipped hitboxes = %v, want %v", flipped, wantFlip)
	}
}

// attackHurtboxDef gives every phase a different authored hurtbox so the
// overlay can only show the right one by following the playing frame.
func attackHurtboxDef() animation.Attack {
	return animation.Attack{
		Animations: map[string]animation.AttackAnimDef{
			"atk": animation.AttackAnim(10, false, 0, 0, 0, 1000, 4, 1, 1, 1,
				animation.AttackF(animation.PhaseWindup,
					animation.S(0, 0), animation.HB(10, 12, 1, 2)),
				animation.AttackF(animation.PhaseActive,
					animation.S(0, 1), animation.HB(40, 50, 30, -20)),
				animation.AttackF(animation.PhaseRecover,
					animation.S(0, 2), animation.HB(70, 80, 5, 6)),
			),
		},
	}
}

func TestCurrentAttackHurtboxesFollowThePlayingFrame(t *testing.T) {
	p := commitmentTestPlayer()
	p.attackAnim = animation.NewAttackAnimator(attackHurtboxDef())
	if got := p.CurrentAttackHurtboxes(); len(got) != 0 {
		t.Fatalf("idle attack hurtboxes = %v, want none", got)
	}

	if !p.TryAttack("atk") {
		t.Fatal("attack should start")
	}
	if got := p.CurrentAttackHurtboxes(); len(got) != 1 || got[0].W != 10 {
		t.Fatalf("windup hurtboxes = %+v, want the windup box", got)
	}

	p.updateAttacking(0.05)
	p.updateAttacking(0.05) // first active frame
	got := p.CurrentAttackHurtboxes()
	if len(got) != 1 || got[0].W != 40 || got[0].H != 50 || got[0].OffsetX != 30 || got[0].OffsetY != -20 {
		t.Fatalf("active hurtboxes = %+v, want the active box", got)
	}

	p.updateAttacking(0.05)
	p.updateAttacking(0.05) // recover frame
	if got := p.CurrentAttackHurtboxes(); len(got) != 1 || got[0].W != 70 {
		t.Fatalf("recover hurtboxes = %+v, want the recover box", got)
	}
}
