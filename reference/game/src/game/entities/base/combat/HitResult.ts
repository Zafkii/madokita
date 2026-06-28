import type { HitboxConfig, StaggerLevel } from "../../../types/CombatTypes"
import type { CombatStats } from "./CombatStats"

export type HitResult = {
  damage: number
  poiseDamage: number
  staggerDuration: number
  staggerLevel: StaggerLevel
  knockbackX: number
  knockbackY: number
  brokePoise: boolean
  killed: boolean
}

export function calculateHit(
  attackerStats: CombatStats,
  defenderStats: CombatStats,
  hitConfig: HitboxConfig,
  damageMultiplier: number,
): HitResult {
  const baseDamage = Math.max(0, attackerStats.attack - defenderStats.defense * 0.5)
  const rawDamage = baseDamage + hitConfig.damage
  const finalDamage = Math.round(rawDamage * damageMultiplier)

  const defenderPoise = defenderStats.poise
  const poiseDamage = hitConfig.poiseDamage
  const brokePoise = poiseDamage >= defenderPoise

  return {
    damage: finalDamage,
    poiseDamage,
    staggerDuration: hitConfig.staggerDuration,
    staggerLevel: hitConfig.staggerLevel,
    knockbackX: hitConfig.knockbackX ?? 0,
    knockbackY: hitConfig.knockbackY ?? 0,
    brokePoise,
    killed: defenderStats.health - finalDamage <= 0,
  }
}
