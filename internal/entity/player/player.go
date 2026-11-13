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

	// Walking off a ledge must release the grounded state before gravity.
	if p.Movement.IsGrounded && !p.hasSupport() {
		p.Movement.IsGrounded = false
	}

	// Gravity runs unconditionally: attacking mid-air must not freeze the fall.
	if !p.Movement.IsGrounded {
		p.applyGravity(dtSec)
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

func (p *Player) moveHorizontal(dx float64) {
	p.X += dx
	if len(p.solidRects) == 0 {
		return
	}

	// A couple of passes handles corners shared by adjacent blocks.
	for pass := 0; pass < 2; pass++ {
		body := p.bodyRect()
		hit := false
		for _, solid := range p.solidRects {
			if !body.Overlaps(solid) {
				continue
			}
			leftOffset := body.X - p.X
			rightOffset := body.X + body.W - p.X
			if dx > 0 {
				p.X = solid.X - rightOffset
			} else {
				p.X = solid.X + solid.W - leftOffset
			}
			body = p.bodyRect()
			hit = true
		}
		if !hit {
			return
		}
	}
}

func (p *Player) applyGravity(dtSec float64) {
	prevBody := p.bodyRect()
	p.Movement.yVelocity += p.Movement.gravity * dtSec
	p.Y += p.Movement.yVelocity * dtSec
	body := p.bodyRect()

	if p.Movement.yVelocity > 0 {
		landingTop := math.Inf(1)
		for _, solid := range p.solidRects {
			if !horizontalOverlap(body, solid) {
				continue
			}
			// Only land when the body crossed the top during this step.
			if prevBody.Y+prevBody.H > solid.Y+collisionEpsilon {
				continue
			}
			if body.Y+body.H < solid.Y {
				continue
			}
			if solid.Y < landingTop {
				landingTop = solid.Y
			}
		}
		if !math.IsInf(landingTop, 1) {
			p.Y = landingTop - p.bodyBottomOffset()
			p.Movement.yVelocity = 0
			p.Movement.IsGrounded = true
			return
		}

		// Compatibility fallback for stages with no authored block geometry.
		if p.Y >= p.StageGroundY {
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

func (p *Player) hasSupport() bool {
	body := p.bodyRect()
	bottom := body.Y + body.H
	for _, solid := range p.solidRects {
		if !horizontalOverlap(body, solid) {
			continue
		}
		if math.Abs(bottom-solid.Y) <= collisionEpsilon {
			return true
		}
	}
	return p.StageGroundY > 0 && math.Abs(p.Y-p.StageGroundY) <= collisionEpsilon
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
