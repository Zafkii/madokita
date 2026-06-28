package combat

import (
	math2 "madokita/internal/math"
)

type CollisionSystem struct {
	hitboxes  []*Hitbox
	actors    []*Actor
	processed map[string]map[string]bool
}

func NewCollisionSystem() *CollisionSystem {
	return &CollisionSystem{
		processed: make(map[string]map[string]bool),
	}
}

func (cs *CollisionSystem) RegisterHitbox(h *Hitbox) {
	cs.hitboxes = append(cs.hitboxes, h)
}

func (cs *CollisionSystem) RegisterActor(a *Actor) {
	cs.actors = append(cs.actors, a)
}

func (cs *CollisionSystem) Update() []HitResult {
	var results []HitResult
	cs.processed = make(map[string]map[string]bool)

	for _, h := range cs.hitboxes {
		if !h.Active {
			continue
		}
		for _, a := range cs.actors {
			if a.ActorID == h.OwnerID || !a.Alive || a.IsInvincible {
				continue
			}
			owner := cs.findOwner(h.OwnerID)
			if owner == nil {
				continue
			}
			if !owner.Team.IsHostile(a.Team) {
				continue
			}
			if cs.alreadyProcessed(h.OwnerID, a.ActorID) {
				continue
			}
			for _, hurtbox := range a.HurtboxSystem.Hurtboxes() {
				hbRect := h.Rect(math2.Vec2{}, false)
				hbRect2 := hurtbox.Rect(math2.Vec2{}, false)
				if hbRect.Overlaps(hbRect2) {
					result := CalculateHit(
						owner.Stats.Attack,
						a.Stats.Defense,
						HitConfig{
							Damage:       h.Config.Damage,
							PoiseDamage:  h.Config.PoiseDamage,
							StaggerLevel: h.Config.StaggerLevel,
						},
						hurtbox.Config.DamageMultiplier,
						owner.Stats.Strength,
					)
					result.HitLocation = hurtbox.Config.Type
					result.BrokePoise = a.ApplyPoiseDamage(result.PoiseDamage)
					a.ReceiveHit(result)
					cs.markProcessed(h.OwnerID, a.ActorID)
					results = append(results, result)
					break
				}
			}
		}
	}
	return results
}

func (cs *CollisionSystem) findOwner(id string) *Actor {
	for _, a := range cs.actors {
		if a.ActorID == id {
			return a
		}
	}
	return nil
}

func (cs *CollisionSystem) alreadyProcessed(attacker, defender string) bool {
	if m, ok := cs.processed[attacker]; ok {
		return m[defender]
	}
	return false
}

func (cs *CollisionSystem) markProcessed(attacker, defender string) {
	if _, ok := cs.processed[attacker]; !ok {
		cs.processed[attacker] = make(map[string]bool)
	}
	cs.processed[attacker][defender] = true
}
