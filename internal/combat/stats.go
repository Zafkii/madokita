package combat

type Stats struct {
	Strength                  float64
	MaxHealth                 float64
	Health                    float64
	MaxStamina                float64
	Stamina                   float64
	StaminaRegenDelay         float64
	StaminaRegenRate          float64
	Attack                    float64
	Defense                   float64
	MaxPoise                  float64
	Poise                     float64
	PoiseRegenDelay           float64
	PoiseRegenRate            float64
	StaggerImmunity           float64
	HyperArmorPoiseAbsorption float64
}

func NewStats() Stats {
	return Stats{
		Strength:  1.0,
		MaxHealth: 100,
		Health:    100,
		MaxStamina: 100,
		Stamina:   100,
		Attack:    10,
		Defense:   5,
		MaxPoise:  100,
		Poise:     100,
	}
}

func (s *Stats) TakeDamage(amount float64) {
	s.Health -= amount
	if s.Health < 0 {
		s.Health = 0
	}
}

func (s *Stats) IsDead() bool {
	return s.Health <= 0
}

func (s *Stats) ConsumeStamina(cost float64) bool {
	if s.Stamina < cost {
		return false
	}
	s.Stamina -= cost
	return true
}
