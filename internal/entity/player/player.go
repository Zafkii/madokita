package player

import (
	"math"
	"madokita/internal/animation"
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

	Frames       []*ebiten.Image
	AnimDef      *animation.Movement
	Scale        float64
	StageGroundY float64

	sprite *ebiten.Image
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
		return
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

		footOffset := float64(FrameSize) * (1 - p.AnimDef.DefaultOriginY) * p.Scale
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

func (p *Player) SetupAnim(def *animation.Movement, frames []*ebiten.Image) {
	p.AnimDef = def
	p.Frames = frames
	p.Animations.Animator = animation.NewAnimator(*def, len(frames))
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
	originY := float64(0.5)
	if p.AnimDef != nil {
		originY = p.AnimDef.DefaultOriginY
	}
	footOffset := float64(FrameSize) * (1 - originY) * p.Scale
	p.Y = stageGroundY - footOffset
	p.Movement.yVelocity = 0
	p.Movement.IsGrounded = true
}

func (p *Player) Draw(screen *ebiten.Image, cameraX float64) {
	if p.Animations.Animator == nil {
		return
	}
	frame := p.Animations.Animator.CurrentFrame()
	if frame == nil {
		return
	}
	for i := range frame.SpriteFrames {
		sf := frame.SpriteFrames[i]
		if sf < 0 || sf >= len(p.Frames) {
			continue
		}

		ox := float64(0)
		oy := float64(0)
		rot := float64(0)
		sx := float64(1)
		sy := float64(1)

		if i < len(frame.OffsetX) {
			ox = frame.OffsetX[i]
		}
		if i < len(frame.OffsetY) {
			oy = frame.OffsetY[i]
		}
		if i < len(frame.Rotation) {
			rot = frame.Rotation[i]
		}
		if i < len(frame.ScaleX) {
			sx = frame.ScaleX[i]
		}
		if i < len(frame.ScaleY) {
			sy = frame.ScaleY[i]
		}

		originX := p.AnimDef.DefaultOriginX * FrameSize
		originY := p.AnimDef.DefaultOriginY * FrameSize

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-originX, -originY)
		op.GeoM.Translate(ox, oy)
		if rot != 0 {
			op.GeoM.Rotate(rot * math.Pi / 180)
		}
		flip := float64(1)
		if p.FlipX {
			flip = -1
		}
		op.GeoM.Scale(sx*p.Scale*flip, sy*p.Scale)
		op.GeoM.Translate(p.X, p.Y)
		op.GeoM.Translate(-cameraX, 0)
		screen.DrawImage(p.Frames[sf], op)
	}
}

const GroundY = 700
