package attacks

import (
	"time"

	. "madokita/internal/animation"
	"madokita/internal/combat"
	"madokita/internal/data"
	"madokita/internal/data/characters/movements"
)

// SayakaVerticalAttack is generated from the animprite editor export
// (animprite/tmp/wea/atk2.go), with sprite paths rewritten to game-relative
// paths and the animation renamed for gameplay use.
var SayakaVerticalAttack = Attack{
	AssetKey:       "sayaka_vertical_attack",
	DefaultOriginX: 0.5,
	DefaultOriginY: 0.5,
	Sprites: []SpriteSheetDef{
		{File: "sprites/players/sayaka_miki/vertical_attack.png", FrameW: 256, FrameH: 256, FrameCount: 20},
		{File: "sprites/players/sayaka_miki/sayaka_weapon.png", FrameW: 145, FrameH: 145, FrameCount: 6},
	},
	Animations: map[string]AttackAnimDef{
		"vertical_attack": AttackAnim(14, true, 60, 80, 150, 3000, 4, 6, 5, 2,
			AttackF(PhaseWindup,
				S(0, 0, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 0, -18, 17, 0, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseWindup,
				S(0, 1, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 1, 12, 17, 0, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseWindup,
				S(0, 2, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 2, 71, 20, 8, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseWindup,
				S(0, 3, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 59, -19, -8, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseWindup,
				S(0, 4, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, -8, -28, -52, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseWindup,
				S(0, 5, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, -60, 84, -154, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseActive,
				S(0, 6, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, -8, -31, -48, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseActive,
				S(0, 7, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 67, -19, -5, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseActive,
				S(0, 8, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 105, 25, 39, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseActive,
				S(0, 9, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 119, 84, 87, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseActive,
				S(0, 10, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 83, 137, 131, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseRecover,
				S(0, 11, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 121, 91, 93, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseRecover,
				S(0, 12, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 110, 36, 45, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseArmed,
				S(0, 13, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 96, 20, 28, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseArmed,
				S(0, 14, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 96, 20, 28, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseArmed,
				S(0, 15, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 96, 20, 28, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseArmed,
				S(0, 16, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 96, 20, 28, 1, 1, 0.5, 0.5),
			),
			AttackF(PhaseArmed,
				S(0, 17, 0, 0, 0, 1, 1, 0.5, 0.5),
				S(1, 5, 96, 20, 28, 1, 1, 0.5, 0.5),
			),
		),
	},
}

// SayakaAttackConfigs holds the combat values (damage, hitbox, stamina) that
// the editor does not define. Timing fields mirror the attack def but are
// overridden at runtime by the AttackAnimator's phase transitions.
var SayakaAttackConfigs = []combat.AttackConfig{
	{
		ID:        "vertical_attack",
		Animation: "vertical_attack",
		Type:      combat.AttackStatic,
		Damage:    10,
		Hitbox: combat.HitboxConfig{
			W: 150, H: 260,
			OffsetX: 60, OffsetY: 0,
			Damage:       10,
			PoiseDamage:  15,
			StaggerLevel: combat.StaggerFlinch,
		},
		Cooldown:    800 * time.Millisecond,
		ActiveTime:  80 * time.Millisecond,
		Windup:      60 * time.Millisecond,
		Recover:     150 * time.Millisecond,
		ArmedTime:   3000 * time.Millisecond,
		StaminaCost: 10,
		PoiseDamage: 15,
	},
}

func RegisterSayaka() {
	data.Register("sayaka", data.CharacterData{
		Animations:    []Movement{movements.SayakaMovement},
		Attack:        &SayakaVerticalAttack,
		AttackConfigs: SayakaAttackConfigs,
		Effects: struct {
			Burst    string
			Ultimate string
		}{Burst: "sayaka"},
	})
}
