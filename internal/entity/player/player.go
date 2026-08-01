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

	if p.State.IsAnimationLocked {
		p.updateAnimationLocked(dtSec)
		return
	}

	if p.State.IsAttacking {
		p.updateAttacking(dtSec)
		return
	}

	if p.Input != nil && p.Input.IsJustPressed(input.ActionAttack) && p.defaultAttackID != "" {
		if p.TryAttack(p.defaultAttackID) {
			return
		}
	}

	moving := false
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

	if p.Actor != nil {
		p.Actor.Update(dt)
	}
}

func (p *Player) updateAnimationLocked(dtSec float64) {
	if p.HasTarget {
		p.Tracker.Update(
			math2.Vec2{X: p.X, Y: p.Y},
			math2.Vec2{X: p.TargetX, Y: p.TargetY},
			&p.FlipX,
			float64(dtSec),
		)
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
// allows it (stamina, cooldown). The AttackAnimator drives the phases; the
// controller follows via SyncPhase so hitbox activation matches the frames.
func (p *Player) TryAttack(id string) bool {
	if p.State.IsAttacking || p.State.IsAnimationLocked || p.State.IsStaggered || p.State.IsDead {
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
	return true
}

func (p *Player) updateAttacking(dtSec float64) {
	if p.attackAnim == nil {
		p.State.IsAttacking = false
		return
	}
	p.attackAnim.Update(dtSec)

	if controller, ok := p.Combat.controllers[p.Combat.currentID]; ok {
		controller.SyncPhase(mapAnimToCombatPhase(p.attackAnim.Phase()))
	}

	if !p.attackAnim.Done() {
		return
	}
	if controller, ok := p.Combat.controllers[p.Combat.currentID]; ok {
		controller.Interrupt()
	}
	p.Combat.currentID = ""
	p.State.IsAttacking = false
	p.PlayAnim("idle")
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

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-originX, -originY)
		op.GeoM.Translate(s.OffsetX, s.OffsetY)
		if s.Rotation != 0 {
			op.GeoM.Rotate(s.Rotation * math.Pi / 180)
		}
		flip := float64(1)
		if p.FlipX {
			flip = -1
		}
		op.GeoM.Scale(s.ScaleX*p.Scale*flip, s.ScaleY*p.Scale)
		op.GeoM.Translate(p.X, p.Y)
		op.GeoM.Translate(-cameraX, 0)
		screen.DrawImage(img, op)
	}
}

const GroundY = 700
