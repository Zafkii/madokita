export const EnemyState = {
  IDLE: "IDLE",
  CHASE: "CHASE",
  ATTACK: "ATTACK",
  RETREAT: "RETREAT",
  STAGGERED: "STAGGERED",
  FLINCH: "FLINCH",
  DEAD: "DEAD",
} as const

export type EnemyState = (typeof EnemyState)[keyof typeof EnemyState]
