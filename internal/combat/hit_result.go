package combat

import "math"

type StaggerLevel int

const (
	StaggerNone StaggerLevel = iota
	StaggerFlinch
	StaggerStagger
	StaggerKnockdown
	StaggerLaunch
)

type HitResult struct {
	Damage          float64
	PoiseDamage     float64
	StaggerDuration float64
	StaggerLevel    StaggerLevel
	KnockbackX      float64
	KnockbackY      float64
	BrokePoise      bool
	Killed          bool
	HitLocation     HurtboxType
}

type HitConfig struct {
	Damage       float64
	PoiseDamage  float64
	StaggerLevel StaggerLevel
	KnockbackX   float64
	KnockbackY   float64
}

func CalculateHit(attack, defense float64, config HitConfig, dmgMult, strength float64) HitResult {
	raw := math.Round((attack*strength - defense*0.5) + config.Damage)
	damage := raw * dmgMult
	if damage < 0 {
		damage = 0
	}
	return HitResult{
		Damage:          damage,
		PoiseDamage:     config.PoiseDamage,
		StaggerDuration: staggerDurationFromLevel(config.StaggerLevel),
		StaggerLevel:    config.StaggerLevel,
		KnockbackX:      config.KnockbackX,
		KnockbackY:      config.KnockbackY,
	}
}

func staggerDurationFromLevel(level StaggerLevel) float64 {
	switch level {
	case StaggerNone:
		return 0
	case StaggerFlinch:
		return 0.2
	case StaggerStagger:
		return 0.5
	case StaggerKnockdown:
		return 1.0
	case StaggerLaunch:
		return 1.5
	default:
		return 0
	}
}
