package data

import (
	"madokita/internal/animation"
	"madokita/internal/combat"
)

type CharacterData struct {
	Animations         []animation.Movement
	Hurtboxes          []combat.HurtboxConfig
	AttackConfigs      []combat.AttackConfig
	Attack             *animation.Attack
	AdditionalTextures []string
	Effects            struct {
		Burst    string
		Ultimate string
	}
}

type StageData struct {
	ID                int
	Name              string
	Background        string
	PlayableCharacter string
	Allies            []string
	Enemies           []string
	BaseSpeed         float64
	NextStage         *int
}

var Registry = map[string]CharacterData{}

func Register(key string, data CharacterData) {
	Registry[key] = data
}

func Get(key string) (CharacterData, bool) {
	d, ok := Registry[key]
	return d, ok
}
