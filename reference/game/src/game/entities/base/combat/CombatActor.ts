import Phaser from "phaser"
import { HurtboxSystem } from "./HurtboxSystem"
import type { CombatStats } from "./CombatStats"
import { CombatTeam } from "./CombatTeam"
import { EventBus } from "../../../core/EventBus"
import type { HitResult } from "./HitResult"
import { StaggerLevel } from "../../../types/CombatTypes"

export type CombatActorConfig = {
  actorId: string
  team: CombatTeam
  stats: CombatStats
  hurtboxes: HurtboxSystem
  sprite: Phaser.Physics.Arcade.Sprite
}

export class CombatActor {
  public readonly actorId: string
  public readonly team: CombatTeam
  public readonly sprite: Phaser.Physics.Arcade.Sprite
  public readonly hurtboxes: HurtboxSystem
  public readonly stats: CombatStats
  private alive = true
  private dead = false

  public isInvincible = false
  public isStaggered = false
  public isFlinching = false

  public isHyperArmoring = false
  private hyperArmorAbsorption = 0

  public poiseDepleted = false
  public hasStaggerImmunity = false

  private invincibleTimer = 0
  private staggerTimer = 0
  private flinchTimer = 0
  private staggerImmunityTimer = 0
  private poiseRegenTimer = 0
  private staminaRegenTimer = 0

  constructor(config: CombatActorConfig) {
    this.actorId = config.actorId
    this.team = config.team
    this.stats = config.stats
    this.hurtboxes = config.hurtboxes
    this.sprite = config.sprite

    this.stats.poise = config.stats.maxPoise
    this.stats.stamina = config.stats.maxStamina
  }

  isAlive(): boolean {
    return this.alive
  }

  isDead(): boolean {
    return this.dead
  }

  kill(): void {
    this.alive = false
    this.dead = true
  }

  damage(value: number): void {
    if (this.isInvincible || this.dead) {
      return
    }

    this.stats.health -= value

    if (this.stats.health <= 0) {
      this.stats.health = 0
      this.kill()
    }
  }

  heal(value: number): void {
    this.stats.health += value

    if (this.stats.health > this.stats.maxHealth) {
      this.stats.health = this.stats.maxHealth
    }
  }

  receiveHit(result: HitResult): void {
    if (this.isInvincible || this.dead) {
      return
    }

    this.stats.health -= result.damage

    if (result.knockbackX !== 0 || result.knockbackY !== 0) {
      const body = this.sprite.body as Phaser.Physics.Arcade.Body
      if (body) {
        body.setVelocity(result.knockbackX, result.knockbackY)
      }
    }

    if (this.stats.health <= 0) {
      this.stats.health = 0
      this.die()
      return
    }

    this.applyPoiseDamage(result)
  }

  setHyperArmor(active: boolean, absorption?: number): void {
    this.isHyperArmoring = active
    if (active && absorption !== undefined) {
      this.hyperArmorAbsorption = absorption
    }
    if (!active) {
      this.hyperArmorAbsorption = 0
    }
  }

  private applyPoiseDamage(result: HitResult): void {
    if (this.hasStaggerImmunity) {
      return
    }

    if (result.staggerLevel === StaggerLevel.NONE) {
      return
    }

    if (this.isHyperArmoring) {
      const absorbed = result.poiseDamage * Math.min(1, Math.max(0, this.hyperArmorAbsorption))
      const actualPoiseDamage = Math.round(result.poiseDamage - absorbed)

      this.stats.poise = Math.max(0, this.stats.poise - actualPoiseDamage)
      this.poiseRegenTimer = 0

      if (this.poiseDepleted) {
        this.stats.poise = 0
        this.handleStaggerByLevel(result)
        return
      }

      if (this.stats.poise <= 0) {
        this.stats.poise = 0
        this.poiseDepleted = true
      }
      return
    }

    // NOT hyperArmoring — poise irrelevant, always apply stagger level directly
    this.handleStaggerByLevel(result)
  }

  private handleStaggerByLevel(result: HitResult): void {
    switch (result.staggerLevel) {
      case StaggerLevel.NONE:
        return
      case StaggerLevel.FLINCH:
        this.enterFlinch(result.staggerDuration)
        return
      case StaggerLevel.STAGGER:
        this.enterStagger(result.staggerDuration)
        return
      case StaggerLevel.KNOCKDOWN:
        this.enterStagger(result.staggerDuration)
        return
      case StaggerLevel.LAUNCH:
        this.enterStagger(result.staggerDuration)
        if (result.knockbackY === 0) {
          const body = this.sprite.body as Phaser.Physics.Arcade.Body
          if (body) {
            body.setVelocityY(-300)
          }
        }
        return
    }
  }

  enterFlinch(duration: number): void {
    if (this.isStaggered || this.dead || this.isFlinching) {
      return
    }

    this.isFlinching = true
    this.flinchTimer = duration

    EventBus.emit("stagger-start", {
      actorId: this.actorId,
      duration,
      isFlinch: true,
    })
  }

  enterStagger(duration: number): void {
    if (this.isStaggered || this.dead) {
      return
    }

    this.isStaggered = true
    this.isFlinching = false
    this.staggerTimer = duration

    const body = this.sprite.body as Phaser.Physics.Arcade.Body
    if (body) {
      body.setVelocity(0, 0)
    }

    EventBus.emit("stagger-start", {
      actorId: this.actorId,
      duration,
      isFlinch: false,
    })
  }

  exitStagger(): void {
    this.isStaggered = false
    this.staggerTimer = 0
    this.poiseDepleted = false
    this.stats.poise = this.stats.maxPoise

    if (this.stats.staggerImmunity > 0) {
      this.hasStaggerImmunity = true
      this.staggerImmunityTimer = this.stats.staggerImmunity
    }

    EventBus.emit("stagger-end", {
      actorId: this.actorId,
    })
  }

  exitFlinch(): void {
    this.isFlinching = false
    this.flinchTimer = 0

    EventBus.emit("flinch-end", {
      actorId: this.actorId,
    })
  }

  setInvincible(duration: number): void {
    this.isInvincible = true
    this.invincibleTimer = duration
  }

  consumeStamina(cost: number): boolean {
    if (this.stats.stamina < Math.abs(cost)) {
      return false
    }
    this.stats.stamina = Math.max(0, this.stats.stamina - Math.abs(cost))
    this.staminaRegenTimer = 0
    return true
  }

  private die(): void {
    this.alive = false
    this.dead = true
    this.isStaggered = false
    this.isFlinching = false

    const body = this.sprite.body as Phaser.Physics.Arcade.Body
    if (body) {
      body.setVelocity(0, 0)
      body.setAllowGravity(false)
    }

    EventBus.emit("actor-died", {
      actorId: this.actorId,
      team: this.team,
    })

    if (this.team === "enemy") {
      EventBus.emit("enemy-defeated", {
        enemyId: this.actorId,
      })
    }
  }

  update(delta: number): void {
    if (this.dead) {
      return
    }

    if (this.isInvincible) {
      this.invincibleTimer -= delta
      if (this.invincibleTimer <= 0) {
        this.isInvincible = false
        this.invincibleTimer = 0
      }
    }

    if (this.isStaggered) {
      this.staggerTimer -= delta
      if (this.staggerTimer <= 0) {
        this.exitStagger()
      }
      return
    }

    if (this.isFlinching) {
      this.flinchTimer -= delta
      if (this.flinchTimer <= 0) {
        this.exitFlinch()
      }
    }

    if (this.hasStaggerImmunity) {
      this.staggerImmunityTimer -= delta
      if (this.staggerImmunityTimer <= 0) {
        this.hasStaggerImmunity = false
        this.staggerImmunityTimer = 0
      }
    }

    if (this.stats.poise < this.stats.maxPoise) {
      this.poiseRegenTimer += delta
      if (this.poiseRegenTimer >= this.stats.poiseRegenDelay) {
        this.stats.poise = Math.min(
          this.stats.maxPoise,
          this.stats.poise + this.stats.poiseRegenRate * (delta / 1000),
        )
        if (this.stats.poise >= this.stats.maxPoise) {
          this.poiseDepleted = false
        }
      }
    }

    if (this.stats.stamina < this.stats.maxStamina) {
      this.staminaRegenTimer += delta
      if (this.staminaRegenTimer >= this.stats.staminaRegenDelay) {
        this.stats.stamina = Math.min(
          this.stats.maxStamina,
          this.stats.stamina + this.stats.staminaRegenRate * (delta / 1000),
        )
      }
    }
  }
}
