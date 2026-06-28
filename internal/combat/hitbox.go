package combat

import (
	math2 "madokita/internal/math"
)

type HitboxConfig struct {
	W, H         float64
	OffsetX      float64
	OffsetY      float64
	Damage       float64
	PoiseDamage  float64
	StaggerLevel StaggerLevel
}

type Hitbox struct {
	Config  HitboxConfig
	OwnerID string
	Active  bool
}

func NewHitbox(config HitboxConfig, ownerID string) *Hitbox {
	return &Hitbox{
		Config:  config,
		OwnerID: ownerID,
	}
}

func (h *Hitbox) Rect(pos math2.Vec2, flipX bool) math2.Rect {
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

func (h *Hitbox) Activate()   { h.Active = true }
func (h *Hitbox) Deactivate() { h.Active = false }
