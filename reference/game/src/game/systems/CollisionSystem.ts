import Phaser from "phaser"
import { CombatActor } from "../entities/base/combat/CombatActor"
import { Hitbox } from "../entities/base/combat/Hitbox"
import { HurtboxSystem } from "../entities/base/combat/HurtboxSystem"
import { CombatTeam } from "../entities/base/combat/CombatTeam"
import { calculateHit } from "../entities/base/combat/HitResult"
import { EventBus } from "../core/EventBus"

type HitboxEntry = {
  hitbox: Hitbox
  owner: CombatActor
}

type HurtboxEntry = {
  hurtboxSystem: HurtboxSystem
  owner: CombatActor
}

export class CollisionSystem {
  private hitboxEntries: HitboxEntry[] = []
  private hurtboxEntries: HurtboxEntry[] = []
  private processedPairs = new Set<string>()

  private static pairKey(attackerId: string, defenderId: string): string {
    return `${attackerId}>${defenderId}`
  }

  registerHitbox(hitbox: Hitbox, owner: CombatActor): void {
    this.hitboxEntries.push({ hitbox, owner })
  }

  unregisterHitbox(hitbox: Hitbox): void {
    this.hitboxEntries = this.hitboxEntries.filter((e) => e.hitbox !== hitbox)
  }

  registerHurtbox(hurtboxSystem: HurtboxSystem, owner: CombatActor): void {
    this.hurtboxEntries.push({ hurtboxSystem, owner })
  }

  unregisterHurtbox(hurtboxSystem: HurtboxSystem): void {
    this.hurtboxEntries = this.hurtboxEntries.filter(
      (e) => e.hurtboxSystem !== hurtboxSystem,
    )
  }

  clearActor(actor: CombatActor): void {
    this.hitboxEntries = this.hitboxEntries.filter((e) => e.owner !== actor)
    this.hurtboxEntries = this.hurtboxEntries.filter((e) => e.owner !== actor)
  }

  update(): void {
    this.processedPairs.clear()

    for (const hEntry of this.hitboxEntries) {
      if (!hEntry.hitbox.isActive()) continue
      if (!hEntry.owner.isAlive()) continue

      const hRect = hEntry.hitbox.getRect()

      for (const tEntry of this.hurtboxEntries) {
        if (!this.isHostile(hEntry.owner, tEntry.owner)) continue
        if (tEntry.owner.isInvincible) continue
        if (!tEntry.owner.isAlive()) continue

        const pairKey = CollisionSystem.pairKey(
          hEntry.owner.actorId,
          tEntry.owner.actorId,
        )
        if (this.processedPairs.has(pairKey)) continue

        const bodyRects = tEntry.hurtboxSystem.getRects()
        for (const br of bodyRects) {
          if (Phaser.Geom.Rectangle.Overlaps(hRect, br.rect)) {
            this.processedPairs.add(pairKey)
            this.processHit(hEntry, tEntry, br.damageMultiplier)
            break
          }
        }
      }
    }
  }

  private isHostile(a: CombatActor, b: CombatActor): boolean {
    if (a.team === CombatTeam.NEUTRAL || b.team === CombatTeam.NEUTRAL) {
      return false
    }
    if (a.team === CombatTeam.PLAYER || a.team === CombatTeam.ALLY) {
      return b.team === CombatTeam.ENEMY
    }
    if (a.team === CombatTeam.ENEMY) {
      return b.team === CombatTeam.PLAYER || b.team === CombatTeam.ALLY
    }
    return false
  }

  private processHit(
    hEntry: HitboxEntry,
    tEntry: HurtboxEntry,
    damageMultiplier: number,
  ): void {
    const attacker = hEntry.owner
    const defender = tEntry.owner
    const hitConfig = hEntry.hitbox.getConfig()

    const result = calculateHit(
      attacker.stats,
      defender.stats,
      hitConfig,
      damageMultiplier,
    )

    defender.receiveHit(result)

    EventBus.emit("hit-landed", {
      attackerId: attacker.actorId,
      defenderId: defender.actorId,
      damage: result.damage,
      poiseDamage: result.poiseDamage,
      brokePoise: result.brokePoise,
      killed: result.killed,
    })
  }
}
