package combat

import "time"

type AttackType int

const (
	AttackStatic AttackType = iota
	AttackLunge
	AttackDash
	AttackBody
)

type AttackConfig struct {
	ID          string
	Animation   string
	Type        AttackType
	Damage      float64
	Hitbox      HitboxConfig
	Cooldown    time.Duration
	ActiveTime  time.Duration
	Windup      time.Duration
	Recover     time.Duration
	ArmedTime   time.Duration
	LungeSpeed  float64
	StaminaCost float64
	PoiseDamage float64
}

type AttackNode struct {
	ID       string
	AttackID string
	Wait     time.Duration
	Next     []string
}

type AttackGraph struct {
	Start string
	Nodes map[string]AttackNode
}

type Controller struct {
	config   AttackConfig
	hitbox   *Hitbox
	owner    *Actor
	timer    time.Duration
	phase    AttackPhase
	lastTime time.Time
	cooldown time.Time
}

type AttackPhase int

const (
	PhaseIdle AttackPhase = iota
	PhaseWindup
	PhaseActive
	PhaseRecover
	PhaseArmed
)

func NewController(config AttackConfig, hitbox *Hitbox, owner *Actor) *Controller {
	return &Controller{
		config: config,
		hitbox: hitbox,
		owner:  owner,
		phase:  PhaseIdle,
	}
}

func (c *Controller) Start() bool {
	if time.Now().Before(c.cooldown) {
		return false
	}
	if !c.owner.Stats.ConsumeStamina(c.config.StaminaCost) {
		return false
	}
	c.phase = PhaseWindup
	c.timer = c.config.Windup
	return true
}

func (c *Controller) Update(dt time.Duration) {
	switch c.phase {
	case PhaseWindup:
		c.timer -= dt
		if c.timer <= 0 {
			c.phase = PhaseActive
			c.timer = c.config.ActiveTime
			c.hitbox.Activate()
		}
	case PhaseActive:
		c.timer -= dt
		if c.timer <= 0 {
			c.phase = PhaseRecover
			c.timer = c.config.Recover
			c.hitbox.Deactivate()
		}
	case PhaseRecover:
		c.timer -= dt
		if c.timer <= 0 {
			if c.config.ArmedTime > 0 {
				c.phase = PhaseArmed
				c.timer = c.config.ArmedTime
			} else {
				c.phase = PhaseIdle
				c.cooldown = time.Now().Add(c.config.Cooldown)
			}
		}
	case PhaseArmed:
		c.timer -= dt
		if c.timer <= 0 {
			c.phase = PhaseIdle
			c.cooldown = time.Now().Add(c.config.Cooldown)
		}
	}
}

func (c *Controller) Interrupt() {
	c.phase = PhaseIdle
	c.hitbox.Deactivate()
	c.cooldown = time.Now().Add(c.config.Cooldown)
}

func (c *Controller) Phase() AttackPhase   { return c.phase }
func (c *Controller) Config() AttackConfig { return c.config }
func (c *Controller) Hitbox() *Hitbox      { return c.hitbox }
