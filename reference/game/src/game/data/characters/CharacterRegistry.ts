import { sayakaMovement } from "./sayaka/sayakaMovements"
import { madokaMovement } from "./madoka/madokaMovements"
import { kyoukoMovement } from "./kyouko/kyoukoMovements"
import { sayakaAttackConfigs, sayakaAttacks } from "./sayaka/sayakaAttacks"

import type { AttackConfig } from "../../types/CombatTypes"
import type { Attack } from "../../entities/base/combat/AttackTypes"
import type { Movement } from "../../entities/base/combat/MovementTypes"

export type CharacterData = {
  movement?: Movement
  effects: {
    burst: string
    ultimate?: string
  }
  attackConfigs?: AttackConfig[]
  attack?: Attack
  additionalTextures?: string[]
}

export const CharacterRegistry: Record<string, CharacterData> = {
  madoka: {
    movement: madokaMovement,
    effects: {
      burst: "madoka",
      ultimate: undefined,
    },
  },

  kyouko: {
    movement: kyoukoMovement,
    effects: {
      burst: "kyouko",
      ultimate: undefined,
    },
  },
  sayaka: {
    movement: sayakaMovement,
    effects: {
      burst: "sayaka",
      ultimate: undefined,
    },
    attackConfigs: sayakaAttackConfigs,
    attack: sayakaAttacks,
    additionalTextures: ["sayaka_sword"],
  },
}
