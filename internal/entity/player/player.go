package player

import (
	"math"
	"madokita/internal/animation"
	"madokita/internal/assets"
	"madokita/internal/combat"
	math2 "madokita/internal/math"
	"madokita/internal/event"
	"madokita/internal/input"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const FrameSize = 256

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

	assetMgr   *assets.AssetManager
	attackAnim *animation.AttackAnimator

	defaultAttackID string
}

func New(x, y float64, actor *combat.Actor, inputMgr *input.Manager, bus *event.Bus) *Player {
	return &Player{
		X:        x,
		Y:        y,
		FlipX:    false,
		Actor:    actor,
		Input:    inputMgr,
		EventBus: bus,
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
	if !p.State.IsMovementLocked {
		if p.Input.IsPressed(input.ActionMoveLeft) {
			p.X -= p.Movement.speed * dtSec
			p.FlipX = true
			moving = true
		} else if p.Input.IsPressed(input.ActionMoveRight) {
			p.X += p.Movement.speed * dtSec
			p.FlipX = false
			moving = true
		}

		if p.X < 0 {
			p.X = 0
		}

		wantsJump := p.Input.IsPressed(input.ActionJump) && p.Movement.IsGrounded
		if wantsJump {
			p.Movement.yVelocity = p.Movement.jumpVelocity
			p.Movement.IsGrounded = false
		}
	}

	// Gravity runs unconditionally: attacking mid-air must not freeze the fall.
	if !p.Movement.IsGrounded {
		p.Movement.yVelocity += p.Movement.gravity * dtSec
		p.Y += p.Movement.yVelocity * dtSec

		footOffset := p.footHeight() * p.Scale
		if p.Y+footOffset >= p.StageGroundY {
			p.Y = p.StageGroundY - footOffset
			p.Movement.yVelocity = 0
			p.Movement.IsGrounded = true
		}
	}

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

func (p *Player) footHeight() float64 {
	h := float64(FrameSize)
	if p.AnimDef == nil {
		return h * 0.5
	}
	if len(p.AnimDef.Sprites) > 0 && p.AnimDef.Sprites[0].FrameH > 0 {
		h = float64(p.AnimDef.Sprites[0].FrameH)
	}
	return h * (1 - p.AnimDef.DefaultOriginY)
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

func (p *Player) Spawn(x, stageGroundY float64) {
	p.X = x
	p.StageGroundY = stageGroundY
	footOffset := p.footHeight() * p.Scale
	p.Y = stageGroundY - footOffset
	p.Movement.yVelocity = 0
	p.Movement.IsGrounded = true
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
