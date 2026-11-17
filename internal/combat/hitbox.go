package combat

import (
	math2 "madokita-gems/internal/math"
)

type HitboxConfig struct {
	W, H         float64
	OffsetX      float64
	OffsetY      float64
	Damage       float64
	PoiseDamage  float64
	StaggerLevel StaggerLevel
}

// HitboxGeometry is the geometric part of a hitbox without the damage values:
// size and centre offset from the owner's pivot. The editor authors one of
// these per attack frame, so a frame can carry more than one box.
type HitboxGeometry struct {
	W, H, OffsetX, OffsetY float64
}

type Hitbox struct {
	Config  HitboxConfig
	OwnerID string
	Active  bool
	boxes   []HitboxGeometry
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

// SetGeometry replaces the authored geometry. An empty list clears it, which
// makes Rects fall back to the config box (the pre-editor behaviour).
func (h *Hitbox) SetGeometry(boxes []HitboxGeometry) {
	if len(boxes) == 0 {
		h.boxes = nil
		return
	}
	h.boxes = append(h.boxes[:0], boxes...)
}

// Rects returns one world-space rectangle per authored box, or the config box
// when nothing was authored.
func (h *Hitbox) Rects(pos math2.Vec2, flipX bool) []math2.Rect {
	if len(h.boxes) == 0 {
		return []math2.Rect{h.Rect(pos, flipX)}
	}
	rects := make([]math2.Rect, 0, len(h.boxes))
	for _, b := range h.boxes {
		ox := b.OffsetX
		if flipX {
			ox = -ox
		}
		rects = append(rects, math2.NewRect(
			pos.X+ox-b.W/2,
			pos.Y+b.OffsetY-b.H/2,
			b.W,
			b.H,
		))
	}
	return rects
}

func (h *Hitbox) Activate()   { h.Active = true }
func (h *Hitbox) Deactivate() { h.Active = false }
