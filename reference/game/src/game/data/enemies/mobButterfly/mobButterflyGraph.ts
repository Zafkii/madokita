import type { AttackGraph } from "../../../types/CombatTypes"

export const mobButterflyGraph: AttackGraph = {
  start: "idle",

  nodes: {
    idle: {
      id: "idle",

      wait: 1000,

      next: ["attack"],
    },

    attack: {
      id: "attack",

      attackId: "sting",

      next: ["idle"],
    },
  },
}
