package data

import (
	"strings"
	"testing"

	. "madokita/internal/animation"
)

func resetRegistry() {
	Registry = map[string]CharacterData{}
}

func validCharacter() CharacterData {
	return CharacterData{
		Animations: []Movement{
			{
				AssetKey: "test_movement",
				Sprites:  []SpriteSheetDef{{File: "sprites/test.png", FrameW: 256, FrameH: 256, FrameCount: 4}},
				Animations: map[string]MovementAnimDef{
					"idle": Anim(3, true, F(S(0, 0))),
				},
			},
		},
		Attack: &Attack{
			AssetKey: "test_attack",
			Sprites:  []SpriteSheetDef{{File: "sprites/test_atk.png", FrameW: 256, FrameH: 256, FrameCount: 2}},
			Animations: map[string]AttackAnimDef{
				"strike": AttackAnim(14, false, 60, 80, 150, 0, 4, 1, 1, 1,
					AttackF(PhaseWindup, S(0, 0)),
				),
			},
		},
	}
}

func TestValidateOK(t *testing.T) {
	resetRegistry()
	Register("hero", validCharacter())
	if err := Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil", err)
	}
}

func TestValidateDuplicateAssetKey(t *testing.T) {
	resetRegistry()
	Register("hero", validCharacter())
	other := validCharacter()
	other.Animations[0].AssetKey = "test_attack"
	Register("villain", other)
	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "duplicate AssetKey") {
		t.Fatalf("Validate() = %v, want duplicate AssetKey error", err)
	}
}

func TestValidateEmptyAssetKey(t *testing.T) {
	resetRegistry()
	cd := validCharacter()
	cd.Animations[0].AssetKey = ""
	Register("hero", cd)
	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "empty AssetKey") {
		t.Fatalf("Validate() = %v, want empty AssetKey error", err)
	}
}

func TestValidateSpriteIdxOutOfRange(t *testing.T) {
	resetRegistry()
	cd := validCharacter()
	cd.Animations[0].Animations["idle"] = Anim(3, true, F(S(3, 0)))
	Register("hero", cd)
	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "SpriteIdx 3 out of range") {
		t.Fatalf("Validate() = %v, want SpriteIdx out of range error", err)
	}
}

func TestValidateSpriteFrameIdxOutOfRange(t *testing.T) {
	resetRegistry()
	cd := validCharacter()
	cd.Animations[0].Animations["idle"] = Anim(3, true, F(S(0, 9)))
	Register("hero", cd)
	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "SpriteFrameIdx 9 out of range") {
		t.Fatalf("Validate() = %v, want SpriteFrameIdx out of range error", err)
	}
}

func TestValidateBadSheetDims(t *testing.T) {
	resetRegistry()
	cd := validCharacter()
	cd.Animations[0].Sprites[0].FrameW = 0
	Register("hero", cd)
	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid frame dims") {
		t.Fatalf("Validate() = %v, want invalid frame dims error", err)
	}
}
