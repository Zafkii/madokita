//void currently
export const BossState = {
  IDLE: 0,
  PHASE_INTRO: 1,
  CHASE: 2,
  ATTACK: 3,
  SPECIAL: 4,
  ENRAGED: 5,
  STAGGERED: 6,
  DEAD: 7,
} as const

export type BossState = (typeof BossState)[keyof typeof BossState]
