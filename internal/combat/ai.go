package combat

import "time"

type AIState int

const (
	AIStateIdle AIState = iota
	AIStateApproach
	AIStateAttack
	AIStateRetreat
	AIStateStaggered
	AIStateDead
)

type AIController struct {
	Actor      *Actor
	State      AIState
	timer      time.Duration
	nextAttack time.Time
}

func NewAIController(actor *Actor) *AIController {
	return &AIController{
		Actor: actor,
		State: AIStateIdle,
	}
}

func (ai *AIController) Update(dt time.Duration) {
	if !ai.Actor.Alive {
		ai.State = AIStateDead
		return
	}
	if ai.Actor.IsStaggered {
		ai.State = AIStateStaggered
		return
	}
}
