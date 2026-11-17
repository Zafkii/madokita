package player

import (
	"madokita-gems/internal/animation"
	"madokita-gems/internal/assets"
	"madokita-gems/internal/combat"
	"madokita-gems/internal/event"
	"madokita-gems/internal/input"
	math2 "madokita-gems/internal/math"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	FrameSize = 256

	// The solid collision body approximates the standing hurtbox union. The
	// hurtboxes themselves stay damage-only; this body is what blocks stage
	// geometry so the player cannot walk or fall through it.
	solidBodyWidth   = 95.0
	solidBodyHeight  = 190.0
	solidBodyOffsetX = 1.5
	solidBodyOffsetY = 32.0

	collisionEpsilon = 0.5
)

type State struct {
	IsControlled      bool
	IsAttacking       bool
	IsMovementLocked  bool
	IsAnimationLocked bool
	IsStaggered       bool
	IsDead            bool
	IsBursting        bool
}

type Movement struct {
	speed        float64
	jumpVelocity float64
	gravity      float64
	yVelocity    float64
	IsGrounded   bool
}

type Combat struct {
	configs     []combat.AttackConfig
	controllers map[string]*combat.Controller
	currentID   string
}

type Animations struct {
	Animator    *animation.Animator
	currentAnim string
}

type Player struct {
	X, Y       float64
	FlipX      bool
	State      State
	Movement   Movement
	Combat     Combat
	Animations Animations
	Actor      *combat.Actor
	Input      *input.Manager
	EventBus   *event.Bus
	Tracker    *combat.Tracker

	TargetX, TargetY float64
	HasTarget        bool

	AnimDef      *animation.Movement
	AttackDef    *animation.Attack
	Scale        float64
	StageGroundY float64
	WorldWidth   float64
	WorldHeight  float64

	assetMgr   *assets.AssetManager
	attackAnim *animation.AttackAnimator

	solidRects []math2.Rect

	// walkDir is the last horizontal walking direction (-1 left, 1 right, 0
	// none). Ground contact is resolved at a probe pushed a quarter of the
	// body width that way, so the base rides the surface it walks into.
	walkDir int

	defaultAttackID string
}

func New(x, y float64, actor *combat.Actor, inputMgr *input.Manager, bus *event.Bus) *Player {
	return &Player{
		X:            x,
		Y:            y,
		FlipX:        false,
		Actor:        actor,
		Input:        inputMgr,
		EventBus:     bus,
		Scale:        0.5,
		StageGroundY: GroundY,
		Tracker:      combat.NewTracker(0),
		Movement: Movement{
			speed:        300,
			jumpVelocity: -500,
			gravity:      1000,
		},
	}
}

func (p *Player) Update(dt time.Duration) {
	if p.State.IsDead || p.State.IsStaggered {
		return
	}

	dtSec := dt.Seconds()

	if !p.State.IsControlled {
		return
	}

	if p.State.IsAttacking {
		p.updateAttacking(dtSec)
	}

	if p.Input != nil && p.Input.IsJustPressed(input.ActionAttack) && p.defaultAttackID != "" && !p.State.IsAnimationLocked {
		if p.TryAttack(p.defaultAttackID) {
			return
		}
	}

	moving := false
	if !p.State.IsMovementLocked && p.Input != nil {
		dx := 0.0
		if p.Input.IsPressed(input.ActionMoveLeft) {
			dx = -p.Movement.speed * dtSec
			p.FlipX = true
			moving = true
		} else if p.Input.IsPressed(input.ActionMoveRight) {
			dx = p.Movement.speed * dtSec
			p.FlipX = false
			moving = true
		}
		if dx != 0 {
			p.moveHorizontal(dx)
		}

		wantsJump := p.Input.IsPressed(input.ActionJump) && p.Movement.IsGrounded
		if wantsJump {
			p.Movement.yVelocity = p.Movement.jumpVelocity
			p.Movement.IsGrounded = false
		}
	}

	// Walking off a ledge must release the grounded state before gravity, but a
	// small drop (the staircase of a slope) is followed instead of becoming air.
	if p.Movement.IsGrounded && !p.hasSupport() {
		if !p.snapDown() && !p.baseHasSupport() {
			p.Movement.IsGrounded = false
		}
	}

	// Gravity runs unconditionally: attacking mid-air must not freeze the fall.
	if !p.Movement.IsGrounded {
		p.applyGravity(dtSec)
		p.resolveSolids(0)
	}

	p.clampWorldBounds()

	if !p.State.IsAttacking {
		desiredAnim := "idle"
		if !p.Movement.IsGrounded {
			desiredAnim = "jump"
		} else if moving {
			desiredAnim = "walk"
		}

		if p.Animations.Animator != nil {
			if p.Animations.currentAnim != desiredAnim {
				if p.Animations.Animator.PlayAnimation(desiredAnim) {
					p.Animations.currentAnim = desiredAnim
				}
			}
			p.Animations.Animator.Update(dtSec)
		}
	}

	if p.Actor != nil {
		p.Actor.Update(dt)
	}
}

// SetSolidRects installs the stage collision rectangles the player must not
// pass through. The slice is copied so callers can rebuild their layout.
func (p *Player) SetSolidRects(rects []math2.Rect) {
	p.solidRects = append(p.solidRects[:0], rects...)
}

func (p *Player) collisionScale() float64 {
	if p.Scale > 0 {
		return p.Scale
	}
	return 1
}

func (p *Player) bodyRectAt(x, y float64) math2.Rect {
	scale := p.collisionScale()
	w := solidBodyWidth * scale
	h := solidBodyHeight * scale
	cx := x + solidBodyOffsetX*scale
	cy := y + solidBodyOffsetY*scale
	return math2.NewRect(cx-w/2, cy-h/2, w, h)
}

func (p *Player) bodyRect() math2.Rect {
	return p.bodyRectAt(p.X, p.Y)
}

func (p *Player) bodyBottomOffset() float64 {
	return (solidBodyOffsetY + solidBodyHeight/2) * p.collisionScale()
}

func (p *Player) bodyTopOffset() float64 {
	return (solidBodyOffsetY - solidBodyHeight/2) * p.collisionScale()
}

// maxAutoStepHeight is the tallest ledge the player walks up without jumping.
// Stage geometry approximates rotated or curved surfaces with axis-aligned
// strips, so a rounded floor is a staircase of small ledges; anything taller
// stays a wall.
const maxAutoStepHeight = 16.0

// maxStepDown is the deepest drop the player follows while walking. It keeps
// the down side of a stair-stepped slope smooth: the player stays grounded and
// keeps the walk animation instead of falling a few pixels per frame.
const maxStepDown = 12.0

// probeEdge is how far inside the body the edge floors are sampled, so a probe
// never lands exactly on a strip boundary.
const probeEdge = 0.5

// groundProbeX is the point where the body meets the floor. Descending, it is
// the body centre pushed a quarter of its width toward the walking direction, so
// three quarters of the base stand on the collision and the leading quarter may
// hang over the drop. Ascending — while grounded — the ground ahead is higher, so
// the probe moves to the trailing edge: the base rides the lower ground behind it
// and the rising ground buries the front of the base instead of leaving the whole
// base airborne. In the air the probe stays ahead, because a fall must meet the
// floor the body is moving toward and never one behind its own base.
func (p *Player) groundProbeX() float64 {
	body := p.bodyRect()
	if p.walkDir == 0 {
		return body.X + body.W/2
	}

	lead, trail := body.X+body.W-probeEdge, body.X+probeEdge
	if p.walkDir < 0 {
		lead, trail = body.X+probeEdge, body.X+body.W-probeEdge
	}
	if leadTop, leadOK := p.floorAtX(lead, body); leadOK {
		if trailTop, trailOK := p.floorAtX(trail, body); trailOK && p.Movement.IsGrounded && leadTop < trailTop-collisionEpsilon {
			return trail
		}
	}
	return lead - float64(p.walkDir)*body.W/4
}

// floorAtX returns the highest (smallest Y) solid top at x close to the body,
// i.e. within a few walkable steps of its feet. The band keeps overhead geometry
// — tree canopies, tall props — from being read as the floor.
func (p *Player) floorAtX(x float64, body math2.Rect) (float64, bool) {
	feet := body.Y + body.H
	top := math.Inf(1)
	found := false
	for _, solid := range p.solidRects {
		if x < solid.X || x >= solid.X+solid.W {
			continue
		}
		if solid.Y < feet-4*maxAutoStepHeight || solid.Y > feet+4*maxStepDown {
			continue
		}
		if solid.Y < top {
			top = solid.Y
			found = true
		}
	}
	return top, found
}

// probeFloorTop returns the highest (smallest Y) solid top under the ground
// probe whose top lies within [minTop, maxTop]. Bounding the band keeps overhead
// geometry — tree canopies, tall props — from hijacking the ground contact.
func (p *Player) probeFloorTop(minTop, maxTop float64) (float64, bool) {
	return p.floorTopAt(p.groundProbeX(), minTop, maxTop)
}

// airborneFloorTop returns the highest solid top under the base while it falls.
// The ground probe follows the travel direction, so a turn mid-fall would move
// the only landing point off a corner the body already hangs over and drop the
// feet through the block. The middle half of the base is probed instead — centre
// and both quarter points — so the landing survives the turn; with no travel
// direction it stays at the centre, where a fall onto a corner still slides off.
func (p *Player) airborneFloorTop(minTop, maxTop float64) (float64, bool) {
	body := p.bodyRect()
	points := []float64{body.X + body.W/2}
	if p.walkDir != 0 {
		points = append(points, body.X+probeEdge+body.W/4, body.X+body.W-probeEdge-body.W/4)
	}
	top, found := math.Inf(1), false
	for _, x := range points {
		if at, ok := p.floorTopAt(x, minTop, maxTop); ok && at < top {
			top, found = at, true
		}
	}
	return top, found
}

// floorTopAt returns the highest (smallest Y) solid top under x whose top lies
// within [minTop, maxTop].
func (p *Player) floorTopAt(x, minTop, maxTop float64) (float64, bool) {
	top := math.Inf(1)
	found := false
	for _, solid := range p.solidRects {
		if x < solid.X || x >= solid.X+solid.W {
			continue
		}
		if solid.Y < minTop || solid.Y > maxTop {
			continue
		}
		if solid.Y < top {
			top = solid.Y
			found = true
		}
	}
	return top, found
}

// wallAhead returns a solid the body overlaps whose top rises more than a
// walkable step above the ground at its own leading face. Measuring against the
// adjacent ground instead of the feet keeps a climb from reading as a wall: the
// base rides the low side, so the strips ahead of the feet are legitimately above
// them, while a real ledge still towers over the ground it stands on.
//
// Both walkable exemptions only apply while the player stands on the ground: the
// climb in resolveSolids lifts the base onto such a step on the same move. In the
// air nothing lifts the body, so every overlapping solid is a wall and a jump can
// never insert the body into a block.
func (p *Player) wallAhead() (math2.Rect, bool) {
	body := p.bodyRect()
	feet := body.Y + body.H
	probe := p.groundProbeX()
	for _, solid := range p.solidRects {
		if !body.Overlaps(solid) || (p.Movement.IsGrounded && p.behindProbeAt(probe, solid)) {
			continue
		}
		if p.Movement.IsGrounded {
			if feet-solid.Y <= maxAutoStepHeight {
				continue
			}
			beside := solid.X - 1
			if p.walkDir < 0 {
				beside = solid.X + solid.W + 1
			}
			if top, ok := p.floorAtX(beside, body); ok && solid.Y >= top-maxAutoStepHeight {
				continue
			}
		}
		return solid, true
	}
	return math2.Rect{}, false
}

// behindProbeAt reports whether a solid sits entirely on the trailing side of a
// probe, where the base may sink into a slope without blocking anything.
func (p *Player) behindProbeAt(probe float64, solid math2.Rect) bool {
	switch {
	case p.walkDir > 0:
		return solid.X+solid.W <= probe+collisionEpsilon
	case p.walkDir < 0:
		return solid.X >= probe-collisionEpsilon
	}
	return false
}

// moveHorizontal travels sideways and then resolves whatever the move ran into.
func (p *Player) moveHorizontal(dx float64) {
	if dx == 0 {
		return
	}
	if dx > 0 {
		p.walkDir = 1
	} else {
		p.walkDir = -1
	}
	p.X += dx
	p.resolveSolids(p.walkDir)
}

// resolveSolids pushes the body out of walls and, while grounded, lifts it onto
// walkable steps. dir is the travel direction the body came from; zero means the
// overlap was not caused by a horizontal move, so each blocker is escaped through
// the side the body overlaps least. It runs after gravity as well, so a fall can
// never leave the body inside a block just because no input moved it that frame.
func (p *Player) resolveSolids(dir int) {
	if len(p.solidRects) == 0 {
		return
	}

	// Several passes handle the strips of a slope: moving sideways can enter
	// more than one step, and each pass either climbs to the probe's surface or
	// stops at a wall.
	for pass := 0; pass < 8; pass++ {
		body := p.bodyRect()
		feet := body.Y + body.H

		// A solid towering over the feet is a wall; the floor strips behind the
		// probe only bury the base and never block.
		if blocker, blocked := p.wallAhead(); blocked {
			leftOffset := body.X - p.X
			rightOffset := body.X + body.W - p.X
			leftPen := body.X + body.W - blocker.X
			rightPen := blocker.X + blocker.W - body.X
			escapeLeft := leftPen <= rightPen
			// A move this frame can only make a shallow overlap on the side it
			// came from. A deeper one was inherited from a turn, a landing or a
			// fall, so it must leave through the nearer face instead of the
			// travel direction — otherwise turning inside a face would eject the
			// body across the whole block.
			switch {
			case dir > 0 && leftPen <= maxAutoStepHeight:
				escapeLeft = true
			case dir < 0 && rightPen <= maxAutoStepHeight:
				escapeLeft = false
			}
			if escapeLeft {
				p.X = blocker.X - rightOffset
			} else {
				p.X = blocker.X + blocker.W - leftOffset
			}
			continue
		}

		// Only a grounded body climbs: in the air nothing lifts the base, so a
		// jump is never boosted onto a step's face.
		if !p.Movement.IsGrounded {
			return
		}
		top, ok := p.probeFloorTop(feet-maxAutoStepHeight, math.Inf(1))
		if !ok || top >= feet-collisionEpsilon {
			return
		}
		p.Y -= feet - top
	}
}

func (p *Player) applyGravity(dtSec float64) {
	prevBody := p.bodyRect()
	p.Movement.yVelocity += p.Movement.gravity * dtSec
	p.Y += p.Movement.yVelocity * dtSec
	body := p.bodyRect()

	if p.Movement.yVelocity > 0 {
		// Land on the highest surface under the probe that this step crossed, or
		// that the fall has already buried the feet in by less than a walkable
		// step: slope strips bury the front of the base by a few pixels, and the
		// body must ride their surface instead of dropping through it.
		minTop := math.Min(prevBody.Y+prevBody.H, body.Y+body.H-maxAutoStepHeight)
		if landingTop, ok := p.airborneFloorTop(minTop, body.Y+body.H); ok {
			p.Y = landingTop - p.bodyBottomOffset()
			p.Movement.yVelocity = 0
			p.Movement.IsGrounded = true
			return
		}

		// Compatibility fallback for stages with no authored block geometry.
		if len(p.solidRects) == 0 && p.Y >= p.StageGroundY {
			p.Y = p.StageGroundY
			p.Movement.yVelocity = 0
			p.Movement.IsGrounded = true
		}
		return
	}

	if p.Movement.yVelocity < 0 {
		ceilingBottom := math.Inf(-1)
		for _, solid := range p.solidRects {
			if !horizontalOverlap(body, solid) {
				continue
			}
			if prevBody.Y < solid.Y+solid.H-collisionEpsilon {
				continue
			}
			if body.Y > solid.Y+solid.H {
				continue
			}
			if solid.Y+solid.H > ceilingBottom {
				ceilingBottom = solid.Y + solid.H
			}
		}
		if !math.IsInf(ceilingBottom, -1) {
			p.Y = ceilingBottom - p.bodyTopOffset()
			p.Movement.yVelocity = 0
		}
	}
}

// snapDown lowers the player onto the closest surface under its body when the
// drop is small enough to walk down. It returns false when there is nothing to
// stand on within maxStepDown, so real ledges still make the player airborne.
func (p *Player) snapDown() bool {
	if len(p.solidRects) == 0 {
		return false
	}
	feet := p.bodyRect().Y + p.bodyRect().H
	top, ok := p.probeFloorTop(feet-maxAutoStepHeight, math.Inf(1))
	if !ok {
		return false
	}
	drop := top - feet
	if drop <= collisionEpsilon || drop > maxStepDown {
		return false
	}

	p.Y += drop
	if _, walled := p.wallAhead(); walled {
		p.Y -= drop
		return false
	}
	return true
}

func (p *Player) hasSupport() bool {
	feet := p.bodyRect().Y + p.bodyRect().H
	if top, ok := p.probeFloorTop(feet-maxAutoStepHeight, math.Inf(1)); ok && math.Abs(feet-top) <= collisionEpsilon {
		return true
	}
	// The logical ground line only backs stages with no authored geometry.
	return len(p.solidRects) == 0 && p.StageGroundY > 0 && math.Abs(p.Y-p.StageGroundY) <= collisionEpsilon
}

// ledgeSupport is how much of the base must still overlap a floor for the player
// to stay grounded. A couple of pixels are enough: the base reaches the very edge
// of the collision before a drop releases, which keeps the jump available at the
// ledge and stops the fall from starting while most of the base still stands on
// it. Raise it to release earlier with more of the body hanging over the drop, or
// set it to 0 to release only once the base has cleared the collision.
const ledgeSupport = 2.0

// baseHasSupport reports whether any part of the base still rests on a floor at
// feet level. The ground probe is a single point a quarter of the width ahead of
// the centre, so it leaves a ledge while most of the base is still standing; the
// release follows the whole base instead, down to the last couple of pixels.
func (p *Player) baseHasSupport() bool {
	body := p.bodyRect()
	feet := body.Y + body.H
	for _, solid := range p.solidRects {
		if math.Abs(solid.Y-feet) > collisionEpsilon {
			continue
		}
		overlap := math.Min(solid.X+solid.W, body.X+body.W) - math.Max(solid.X, body.X)
		if overlap >= ledgeSupport {
			return true
		}
	}
	// The logical ground line only backs stages with no authored geometry.
	return len(p.solidRects) == 0 && p.StageGroundY > 0 && math.Abs(feet-p.StageGroundY) <= collisionEpsilon
}

func horizontalOverlap(a, b math2.Rect) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X
}

func (p *Player) SetWorldBounds(width, height float64) {
	p.WorldWidth = width
	p.WorldHeight = height
}

func (p *Player) clampWorldBounds() {
	if p.WorldWidth > 0 {
		if p.X < 0 {
			p.X = 0
		}
		if p.X > p.WorldWidth {
			p.X = p.WorldWidth
		}
	}
	if p.WorldHeight <= 0 {
		return
	}
	if p.Y < 0 {
		p.Y = 0
		if p.Movement.yVelocity < 0 {
			p.Movement.yVelocity = 0
		}
	}
	if p.Y > p.WorldHeight && p.WorldHeight >= p.StageGroundY {
		p.Y = p.WorldHeight
		if p.Movement.yVelocity > 0 {
			p.Movement.yVelocity = 0
		}
	}
}

func (p *Player) SetTarget(x, y float64) {
	p.TargetX = x
	p.TargetY = y
	p.HasTarget = true
}

func (p *Player) SetupAnim(def *animation.Movement, am *assets.AssetManager) {
	p.AnimDef = def
	p.assetMgr = am
	p.Animations.Animator = animation.NewAnimator(*def)
}

// SetupCombat wires the attack data (editor-driven animation) with the
// combat configs (code-driven damage). Each config gets its own controller
// and hitbox, registered against the player's actor.
func (p *Player) SetupCombat(def *animation.Attack, configs []combat.AttackConfig, am *assets.AssetManager) {
	p.AttackDef = def
	p.assetMgr = am
	p.Combat.configs = configs
	p.Combat.controllers = make(map[string]*combat.Controller, len(configs))
	for i := range configs {
		cfg := configs[i]
		if p.Actor == nil {
			continue
		}
		hb := combat.NewHitbox(cfg.Hitbox, p.Actor.ActorID)
		p.Combat.controllers[cfg.ID] = combat.NewController(cfg, hb, p.Actor)
	}
	if len(configs) > 0 {
		p.defaultAttackID = configs[0].ID
	}
	if def != nil {
		p.attackAnim = animation.NewAttackAnimator(*def)
	}
}

// ActiveHitboxes returns the world-space hitboxes of the current attack that
// are active during this frame: the boxes authored on the playing attack
// frame, or the config box when the frame carries none.
func (p *Player) ActiveHitboxes() []math2.Rect {
	if p.Combat.controllers == nil || p.Combat.currentID == "" {
		return nil
	}
	controller, ok := p.Combat.controllers[p.Combat.currentID]
	if !ok {
		return nil
	}
	hitbox := controller.Hitbox()
	if hitbox == nil || !hitbox.Active {
		return nil
	}
	return hitbox.Rects(math2.Vec2{X: p.X, Y: p.Y}, p.FlipX)
}

// CurrentAttackHurtboxes returns the hurtboxes authored on the playing attack
// frame, so the debug overlay can draw the geometry the editor animated
// instead of the frozen movement frame. Combat does not resolve damage against
// them yet.
func (p *Player) CurrentAttackHurtboxes() []animation.FrameHurtbox {
	if p.attackAnim == nil {
		return nil
	}
	frame := p.attackAnim.CurrentFrame()
	if frame == nil {
		return nil
	}
	return frame.Hurtboxes
}

// CurrentAttackHitboxes returns the hitboxes authored on the playing attack
// frame, in world space. Damage is still owned by the combat controllers; this
// exposes the editor-authored geometry (X, Y, W, H per phase) so gameplay and
// the debug overlay can read the same numbers the editor exported.
func (p *Player) CurrentAttackHitboxes() []math2.Rect {
	if p.attackAnim == nil {
		return nil
	}
	frame := p.attackAnim.CurrentFrame()
	if frame == nil || len(frame.Hitboxes) == 0 {
		return nil
	}
	rects := make([]math2.Rect, 0, len(frame.Hitboxes))
	for _, hb := range frame.Hitboxes {
		ox := hb.OffsetX
		if p.FlipX {
			ox = -ox
		}
		rects = append(rects, math2.NewRect(
			p.X+ox-hb.W/2,
			p.Y+hb.OffsetY-hb.H/2,
			hb.W,
			hb.H,
		))
	}
	return rects
}

// TryAttack starts the given attack if the player is free and the controller
// allows it (stamina). The AttackAnimator drives the phases; the controller
// follows via SyncPhase so hitbox activation matches the frames. Movement is
// locked during windup+active, and new attacks unlock once recover ends (the
// armed pose is free and cancelable).
func (p *Player) TryAttack(id string) bool {
	if p.State.IsAnimationLocked || p.State.IsStaggered || p.State.IsDead {
		return false
	}
	if p.attackAnim == nil {
		return false
	}
	controller, ok := p.Combat.controllers[id]
	if !ok {
		return false
	}
	animName := controller.Config().Animation
	if animName == "" {
		animName = id
	}
	if !controller.Start() {
		return false
	}
	if !p.attackAnim.PlayAnimation(animName) {
		controller.Interrupt()
		return false
	}
	p.Combat.currentID = id
	p.State.IsAttacking = true
	p.State.IsMovementLocked = true
	p.State.IsAnimationLocked = true
	return true
}

func (p *Player) updateAttacking(dtSec float64) {
	if p.attackAnim == nil {
		p.State.IsAttacking = false
		return
	}

	if p.HasTarget {
		p.Tracker.Update(
			math2.Vec2{X: p.X, Y: p.Y},
			math2.Vec2{X: p.TargetX, Y: p.TargetY},
			&p.FlipX,
			dtSec,
		)
	}

	p.attackAnim.Update(dtSec)

	// Finish first: when the animator is done, Phase() reports windup and the
	// phase switch below would re-lock the player forever.
	if p.attackAnim.Done() {
		if controller, ok := p.Combat.controllers[p.Combat.currentID]; ok {
			controller.Interrupt()
		}
		p.Combat.currentID = ""
		p.State.IsAttacking = false
		p.State.IsMovementLocked = false
		p.State.IsAnimationLocked = false
		p.PlayAnim("idle")
		return
	}

	if controller, ok := p.Combat.controllers[p.Combat.currentID]; ok {
		controller.SyncPhase(mapAnimToCombatPhase(p.attackAnim.Phase()))
		controller.SetGeometry(authoredGeometry(p.attackAnim.CurrentFrame()))
	}

	// Commitment by phase: immobile while windup+active, free to move during
	// recover, free to act again once the armed pose begins.
	switch p.attackAnim.Phase() {
	case animation.PhaseRecover:
		p.State.IsMovementLocked = false
		p.State.IsAnimationLocked = true
	case animation.PhaseArmed:
		p.State.IsMovementLocked = false
		p.State.IsAnimationLocked = false
	default:
		p.State.IsMovementLocked = true
		p.State.IsAnimationLocked = true
	}

	// Moving cancels the attack from recover on (like the reference TS
	// cancelGuard): walking away must not drag attack frames behind.
	phase := p.attackAnim.Phase()
	if (phase == animation.PhaseRecover || phase == animation.PhaseArmed) &&
		p.Input != nil &&
		(p.Input.IsPressed(input.ActionMoveLeft) || p.Input.IsPressed(input.ActionMoveRight)) {
		if controller, ok := p.Combat.controllers[p.Combat.currentID]; ok {
			controller.Interrupt()
		}
		p.Combat.currentID = ""
		p.State.IsAttacking = false
		p.State.IsAnimationLocked = false
		p.PlayAnim("walk")
		return
	}
}

// authoredGeometry converts the editor-authored hitboxes of one attack frame
// into the geometry the combat hitbox resolves. Frames without boxes clear
// it, so the config fallback applies.
func authoredGeometry(frame *animation.AttackFrame) []combat.HitboxGeometry {
	if frame == nil || len(frame.Hitboxes) == 0 {
		return nil
	}
	boxes := make([]combat.HitboxGeometry, 0, len(frame.Hitboxes))
	for _, hb := range frame.Hitboxes {
		boxes = append(boxes, combat.HitboxGeometry{
			W: hb.W, H: hb.H, OffsetX: hb.OffsetX, OffsetY: hb.OffsetY,
		})
	}
	return boxes
}

func mapAnimToCombatPhase(p animation.AttackPhase) combat.AttackPhase {
	switch p {
	case animation.PhaseActive:
		return combat.PhaseActive
	case animation.PhaseRecover:
		return combat.PhaseRecover
	case animation.PhaseArmed:
		return combat.PhaseArmed
	default:
		return combat.PhaseWindup
	}
}

func (p *Player) PlayAnim(name string) bool {
	if p.Animations.Animator == nil {
		return false
	}
	if p.Animations.Animator.PlayAnimation(name) {
		p.Animations.currentAnim = name
		return true
	}
	return false
}

// Spawn places the player's origin on the logical ground line. Visual
// grounding is intentionally not derived from the animation frame: the game
// should resolve ground contact with collision boxes (or place the visual
// floor independently).
func (p *Player) Spawn(x, stageGroundY float64) {
	p.SpawnAt(x, stageGroundY, stageGroundY)
}

// SpawnAt places the player at an explicit X/Y. A spawn below the stage ground
// line is clamped to it; a spawn above stays airborne until gravity lands it.
func (p *Player) SpawnAt(x, y, stageGroundY float64) {
	if stageGroundY > 0 && y > stageGroundY {
		y = stageGroundY
	}
	p.X = x
	p.StageGroundY = stageGroundY
	p.Y = y
	p.Movement.yVelocity = 0
	p.Movement.IsGrounded = y >= stageGroundY
}

func (p *Player) Draw(screen *ebiten.Image, cameraX float64) {
	if p.State.IsAttacking && p.attackAnim != nil && p.AttackDef != nil {
		frame := p.attackAnim.CurrentFrame()
		if frame == nil {
			return
		}
		def := p.AttackDef
		p.drawSprites(screen, cameraX, frame.Sprites, def.AssetKey, def.Sprites, def.DefaultOriginX, def.DefaultOriginY)
		return
	}

	if p.Animations.Animator == nil || p.AnimDef == nil {
		return
	}
	frame := p.Animations.Animator.CurrentFrame()
	if frame == nil {
		return
	}
	def := p.AnimDef
	p.drawSprites(screen, cameraX, frame.Sprites, def.AssetKey, def.Sprites, def.DefaultOriginX, def.DefaultOriginY)
}

func (p *Player) drawSprites(screen *ebiten.Image, cameraX float64, sprites []animation.FrameSprite, assetKey string, defSprites []animation.SpriteSheetDef, defaultOriginX, defaultOriginY float64) {
	for i := range sprites {
		s := &sprites[i]
		if p.assetMgr == nil {
			continue
		}
		img := p.assetMgr.GetFrame(assetKey, s.SpriteIdx, s.SpriteFrameIdx)
		if img == nil {
			continue
		}

		originX := defaultOriginX * FrameSize
		originY := defaultOriginY * FrameSize
		if s.SpriteIdx >= 0 && s.SpriteIdx < len(defSprites) {
			sheet := defSprites[s.SpriteIdx]
			originX = s.OriginX * float64(sheet.FrameW)
			originY = s.OriginY * float64(sheet.FrameH)
		}

		flip := float64(1)
		if p.FlipX {
			flip = -1
		}

		// Offset scaled by character scale and mirrored, but NOT rotated:
		// GeoM.Rotate rotates the accumulated translation too, so the offset
		// must be applied after the rotation (same as the editor).
		ox := s.OffsetX * p.Scale
		oy := s.OffsetY * p.Scale
		if p.FlipX {
			ox = -ox
		}

		// Rotation must be negated when flipped: R(θ)·S(-1) = M·R(-θ), but the
		// mirror of the right-facing render is M·R(θ). Without the negation the
		// weapon points the wrong way on rotated frames.
		rot := s.Rotation
		if p.FlipX {
			rot = -rot
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-originX, -originY)
		op.GeoM.Scale(s.ScaleX*p.Scale*flip, s.ScaleY*p.Scale)
		if rot != 0 {
			op.GeoM.Rotate(rot * math.Pi / 180)
		}
		op.GeoM.Translate(ox, oy)
		op.GeoM.Translate(p.X, p.Y)
		op.GeoM.Translate(-cameraX, 0)
		screen.DrawImage(img, op)
	}
}

const GroundY = 700
