package combat

import math2 "madokita/internal/math"

type Tracker struct {
	Speed float64
}

func NewTracker(speed float64) *Tracker {
	return &Tracker{Speed: speed}
}

func (t *Tracker) Update(attackerPos, targetPos math2.Vec2, flipX *bool, dt float64) {
	if t.Speed <= 0 {
		return
	}
	dir := targetPos.X - attackerPos.X
	if dir < -1 && !*flipX {
		*flipX = true
	} else if dir > 1 && *flipX {
		*flipX = false
	}
}
