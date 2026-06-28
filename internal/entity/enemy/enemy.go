package enemy

import (
	"madokita/internal/combat"
	"madokita/internal/entity/player"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Type int

const (
	TypeMobButterfly Type = iota
	TypeCharlottePhase1
	TypeTestExtension
)

type Enemy struct {
	X, Y   float64
	Type   Type
	Actor  *combat.Actor
	AI     *combat.AIController
	Target *player.Player
	FlipX  bool
	sprite *ebiten.Image
}

func New(x, y float64, eType Type, actor *combat.Actor) *Enemy {
	return &Enemy{
		X:     x,
		Y:     y,
		Type:  eType,
		Actor: actor,
		AI:    combat.NewAIController(actor),
	}
}

func (e *Enemy) Update(dt time.Duration) {
	if e.Actor != nil {
		e.Actor.Update(dt)
	}
	if e.AI != nil {
		e.AI.Update(dt)
	}
}
