import type { AttackConfig } from "../../../types/CombatTypes"

export const mobButterflyAttacks: AttackConfig[] = [
  {
    id: "sting",

    animation: "mobbutterfly_attack",

    type: "lunge",

    damage: 10,

    cooldown: 2000,

    windup: 650,

    activeTime: 120,

    lungeSpeed: 520,

    poiseDamage: 20,

    hitbox: {
      subHitboxes: [
        {
          width: [90],
          height: [120],
          offsetX: [40],
          offsetY: [0],
        },
      ],

      damage: 10,

      poiseDamage: 20,
      staggerDuration: 400,
      staggerLevel: "stagger",

      knockbackX: 100,
      knockbackY: 0,
    },
  },
]
