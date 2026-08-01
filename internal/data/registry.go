package data

import (
	"fmt"

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

// Validate performs a boot-time sanity pass over the character registry.
// Mismatched asset keys or out-of-range sprite references surface as startup
// errors instead of silently rendering invisible sprites mid-match.
func Validate() error {
	seenKeys := make(map[string]string)
	for charKey, cd := range Registry {
		for i := range cd.Animations {
			if err := validateMovement(&cd.Animations[i], charKey, seenKeys); err != nil {
				return err
			}
		}
		if cd.Attack != nil {
			if err := validateAttack(cd.Attack, charKey, seenKeys); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateMovement(m *animation.Movement, charKey string, seenKeys map[string]string) error {
	if m.AssetKey == "" {
		return fmt.Errorf("character %q: movement has empty AssetKey", charKey)
	}
	if prev, dup := seenKeys[m.AssetKey]; dup {
		return fmt.Errorf("character %q: duplicate AssetKey %q (also used by %q)", charKey, m.AssetKey, prev)
	}
	seenKeys[m.AssetKey] = charKey
	for _, def := range m.Sprites {
		if err := validateSheet(m.AssetKey, def, charKey); err != nil {
			return err
		}
	}
	for name, anim := range m.Animations {
		for _, f := range anim.Frames {
			for _, s := range f.Sprites {
				if err := validateSpriteRef(s, m.Sprites, charKey, m.AssetKey, name); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateAttack(a *animation.Attack, charKey string, seenKeys map[string]string) error {
	if a.AssetKey == "" {
		return fmt.Errorf("character %q: attack has empty AssetKey", charKey)
	}
	if prev, dup := seenKeys[a.AssetKey]; dup {
		return fmt.Errorf("character %q: duplicate AssetKey %q (also used by %q)", charKey, a.AssetKey, prev)
	}
	seenKeys[a.AssetKey] = charKey
	for _, def := range a.Sprites {
		if err := validateSheet(a.AssetKey, def, charKey); err != nil {
			return err
		}
	}
	for name, anim := range a.Animations {
		for _, f := range anim.Frames {
			for _, s := range f.Sprites {
				if err := validateSpriteRef(s, a.Sprites, charKey, a.AssetKey, name); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func validateSheet(assetKey string, def animation.SpriteSheetDef, charKey string) error {
	if def.File == "" {
		return fmt.Errorf("character %q: %q: sprite sheet has empty File", charKey, assetKey)
	}
	if def.FrameW <= 0 || def.FrameH <= 0 || def.FrameCount <= 0 {
		return fmt.Errorf("character %q: %q: sprite sheet %q has invalid frame dims (FrameW=%d FrameH=%d FrameCount=%d)", charKey, assetKey, def.File, def.FrameW, def.FrameH, def.FrameCount)
	}
	return nil
}

func validateSpriteRef(s animation.FrameSprite, sprites []animation.SpriteSheetDef, charKey, assetKey, animName string) error {
	if s.SpriteIdx < 0 || s.SpriteIdx >= len(sprites) {
		return fmt.Errorf("character %q: %q animation %q: SpriteIdx %d out of range (have %d sheets)", charKey, assetKey, animName, s.SpriteIdx, len(sprites))
	}
	if s.SpriteFrameIdx < 0 || s.SpriteFrameIdx >= sprites[s.SpriteIdx].FrameCount {
		return fmt.Errorf("character %q: %q animation %q: SpriteFrameIdx %d out of range (sheet has %d frames)", charKey, assetKey, animName, s.SpriteFrameIdx, sprites[s.SpriteIdx].FrameCount)
	}
	return nil
}
