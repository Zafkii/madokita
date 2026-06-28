export const CombatTeam = {
  PLAYER: "player",
  ALLY: "ally",
  ENEMY: "enemy",
  NEUTRAL: "neutral",
} as const

export type CombatTeam = (typeof CombatTeam)[keyof typeof CombatTeam]
