export const StaggerLevel = {
  NONE: "none",
  FLINCH: "flinch",
  STAGGER: "stagger",
  KNOCKDOWN: "knockdown",
  LAUNCH: "launch",
} as const

export type StaggerLevel = (typeof StaggerLevel)[keyof typeof StaggerLevel]

export type SubHitboxConfig = {
  width: number[]
  height: number[]
  offsetX: number[]
  offsetY: number[]
}

export type HitboxConfig = {
  id?: string

  subHitboxes: SubHitboxConfig[]

  damage: number

  poiseDamage: number
  staggerDuration: number
  staggerLevel: StaggerLevel

  knockbackX?: number
  knockbackY?: number

  active?: boolean
}

export type AttackType = "static" | "lunge" | "dash" | "body"

export type AttackConfig = {
  id: string
  animation: string
  texture?: string
  windupAnim?: string
  recoverAnim?: string
  type: AttackType
  damage: number
  hitbox: HitboxConfig
  cooldown: number
  activeTime?: number
  windup?: number
  spacingDistance?: number
  spacingSpeed?: number

  // MOVEMENT

  lungeSpeed?: number
  lungeDuration?: number

  // TRACKING

  trackingStrength?: number

  // VERTICALITY

  allowVerticalTracking?: boolean

  // COMBAT

  hyperArmor?: boolean
  hyperArmorStart?: number
  hyperArmorEnd?: number
  staminaCost?: number
  poiseDamage?: number

  // TIMING

  recover?: number
  guard?: number
}

export type AttackNode = {
  id: string
  attackId?: string
  wait?: number
  next: string[]
}

export type AttackGraph = {
  start: string
  nodes: Record<string, AttackNode>
}
