package combat

import (
	"time"
)

type Actor struct {
	ActorID            string
	Team               Team
	Stats              *Stats
	HurtboxSystem      *HurtboxSystem
	IsInvincible       bool
	IsStaggered        bool
	IsFlinching        bool
	IsHyperArmoring    bool
	PoiseDepleted      bool
	HasStaggerImmunity bool
	Alive              bool

	invincibleTimer      time.Duration
	staggerTimer         time.Duration
	flinchTimer          time.Duration
	hyperArmorAbsorption float64
}

func NewActor(id string, team Team, stats *Stats, hurtboxes *HurtboxSystem) *Actor {
	return &Actor{
		ActorID:       id,
		Team:          team,
		Stats:         stats,
		HurtboxSystem: hurtboxes,
		Alive:         true,
	}
}

func (a *Actor) Update(dt time.Duration) {
	if !a.Alive {
		return
	}
	if a.IsInvincible {
		a.invincibleTimer -= dt
		if a.invincibleTimer <= 0 {
			a.IsInvincible = false
		}
	}
	if a.IsStaggered {
		a.staggerTimer -= dt
		if a.staggerTimer <= 0 {
			a.exitStagger()
		}
	}
	if a.IsFlinching {
		a.flinchTimer -= dt
		if a.flinchTimer <= 0 {
			a.exitFlinch()
		}
	}
}

func (a *Actor) SetInvincible(duration time.Duration) {
	a.IsInvincible = true
	a.invincibleTimer = duration
}

func (a *Actor) SetHyperArmor(active bool, absorption float64) {
	a.IsHyperArmoring = active
	if active {
		a.hyperArmorAbsorption = absorption
	}
}

func (a *Actor) enterStagger(duration time.Duration) {
	a.IsStaggered = true
	a.staggerTimer = duration
}

func (a *Actor) exitStagger() {
	a.IsStaggered = false
	a.HasStaggerImmunity = true
}

func (a *Actor) enterFlinch(duration time.Duration) {
	a.IsFlinching = true
	a.flinchTimer = duration
}

func (a *Actor) exitFlinch() {
	a.IsFlinching = false
}

func (a *Actor) ApplyPoiseDamage(dmg float64) bool {
	if a.PoiseDepleted || !a.Alive {
		return false
	}
	if a.IsHyperArmoring && a.hyperArmorAbsorption > 0 {
		dmg -= a.hyperArmorAbsorption
		if dmg < 0 {
			dmg = 0
		}
	}
	a.Stats.Poise -= dmg
	if a.Stats.Poise <= 0 {
		a.Stats.Poise = 0
		a.PoiseDepleted = true
		return true
	}
	return false
}

func (a *Actor) ReceiveHit(result HitResult) {
	if a.IsInvincible || !a.Alive {
		return
	}
	a.Stats.TakeDamage(result.Damage)
	if a.Stats.IsDead() {
		a.Alive = false
	}
}

func (a *Actor) Die() {
	a.Alive = false
}
