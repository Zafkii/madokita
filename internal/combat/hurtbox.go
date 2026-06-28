package combat

import math2 "madokita/internal/math"

type HurtboxType int

const (
	HurtboxHead HurtboxType = iota
	HurtboxBody
	HurtboxLegs
)

type HurtboxConfig struct {
	W, H, OffsetX, OffsetY float64
	Type                   HurtboxType
	DamageMultiplier       float64
}

type Hurtbox struct {
	Config HurtboxConfig
	rect   math2.Rect
	world  *math2.Rect
}

func NewHurtbox(config HurtboxConfig) *Hurtbox {
	return &Hurtbox{
		Config: config,
	}
}

func (h *Hurtbox) Rect(pos math2.Vec2, flipX bool) math2.Rect {
	ox := h.Config.OffsetX
	if flipX {
		ox = -ox
	}
	return math2.NewRect(
		pos.X+ox-h.Config.W/2,
		pos.Y+h.Config.OffsetY-h.Config.H/2,
		h.Config.W,
		h.Config.H,
	)
}

type HurtboxSystem struct {
	hurtboxes []*Hurtbox
}

func NewHurtboxSystem(configs []HurtboxConfig) *HurtboxSystem {
	hs := &HurtboxSystem{}
	for _, c := range configs {
		hs.hurtboxes = append(hs.hurtboxes, NewHurtbox(c))
	}
	return hs
}

func (hs *HurtboxSystem) Hurtboxes() []*Hurtbox {
	return hs.hurtboxes
}
