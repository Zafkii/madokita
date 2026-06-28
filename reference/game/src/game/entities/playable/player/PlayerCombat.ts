import Phaser from "phaser"
import { AttackController } from "../../base/combat/controllers/AttackControllers"
import { CombatTargetSystem } from "../../base/combat/CombatTargetSystem"
import { CombatActor } from "../../base/combat/CombatActor"
import { CollisionSystem } from "../../../systems/CollisionSystem"
import { MultiSpriteAnimator } from "../../base/combat/MultiSpriteAnimator"
import type { AttackConfig } from "../../../types/CombatTypes"
import type { Attack } from "../../base/combat/AttackTypes"
import type { HurtboxData } from "../../base/combat/MovementTypes"

const PLAYER_BASIC_ATTACK: AttackConfig = {
  id: "basic",
  animation: "skill",
  type: "static",
  damage: 5,
  cooldown: 400,
  activeTime: 120,
  windup: 80,
  staminaCost: 10,
  poiseDamage: 15,
  recover: 400,
  guard: 1500,
  hitbox: {
    subHitboxes: [
      {
        width: [60],
        height: [40],
        offsetX: [40],
        offsetY: [0],
      },
    ],
    damage: 5,
    poiseDamage: 15,
    staggerDuration: 300,
    staggerLevel: "stagger",
    knockbackX: 200,
    knockbackY: 0,
  },
}

export class PlayerCombat {
  private collisionSystem: CollisionSystem
  private sprite: Phaser.Physics.Arcade.Sprite
  private attackControllers = new Map<string, AttackController>()
  private animator: MultiSpriteAnimator | null = null
  private attackSequence = 0
  private attackTimer: Phaser.Time.TimerEvent | null = null

  constructor(
    scene: Phaser.Scene,
    sprite: Phaser.Physics.Arcade.Sprite,
    targetSystem: CombatTargetSystem,
    collisionSystem: CollisionSystem,
    owner: CombatActor,
    attackConfigs?: AttackConfig[],
  ) {
    this.collisionSystem = collisionSystem
    this.sprite = sprite
    const configs = attackConfigs ?? [PLAYER_BASIC_ATTACK]
    for (const config of configs) {
      const controller = new AttackController(scene, sprite, targetSystem, config, owner)
      this.attackControllers.set(config.id, controller)
      this.collisionSystem.registerHitbox(controller.getHitbox(), owner)
    }
  }

  initAnimator(def: Attack, additionalTextures?: string[]): void {
    this.animator = new MultiSpriteAnimator(this.sprite.scene, def, this.sprite, additionalTextures)
  }

  getAnimator(): MultiSpriteAnimator | null {
    return this.animator
  }

  getCurrentFrameHurtboxes(): HurtboxData[] | null {
    return this.animator?.getCurrentFrameHurtboxes() ?? null
  }

  getDefaultHurtboxes(): HurtboxData[] | undefined {
    return this.animator?.getDefaultHurtboxes()
  }

  hasAttack(id: string): boolean {
    return this.attackControllers.has(id)
  }

  interruptAll(): void {
    this.attackControllers.forEach(c => c.interrupt())
    if (this.attackTimer) {
      this.attackTimer.destroy()
      this.attackTimer = null
    }
  }

  attack(id = "basic", onComplete?: () => void): boolean {
    const controller = this.attackControllers.get(id)
    if (!controller) return false
    if (this.isAttacking()) return false

    const cfg = controller.getConfig()
    const windup = cfg.windup ?? 0
    const activeTime = cfg.activeTime ?? 250
    const recover = cfg.recover ?? 400
    if (this.animator) {
      controller.setTiming(windup, activeTime, recover)
    }
    const started = controller.tryAttack()
    if (started) {
      const totalMs = windup + activeTime + recover
      if (this.animator) {
        this.animator.playAnimation(id, { windup, active: activeTime, recover })
      }
      const seq = ++this.attackSequence
      this.attackTimer = this.sprite.scene.time.delayedCall(totalMs, () => {
        if (this.attackSequence !== seq) return
        this.attackTimer = null
        onComplete?.()
      })
    }
    return started
  }

  getAttackConfig(id: string): AttackConfig | undefined {
    return this.attackControllers.get(id)?.getConfig()
  }

  isAttacking(id?: string): boolean {
    if (id) {
      return this.attackControllers.get(id)?.isRunning() ?? false
    }
    for (const c of this.attackControllers.values()) {
      if (c.isRunning()) return true
    }
    return false
  }

  update(_dt = 0): void {
    this.attackControllers.forEach((c) => c.update())
    if (this.animator) {
      this.animator.update(this.sprite.x, this.sprite.y, this.sprite.flipX, this.sprite.scaleX)
    }
  }

  destroy(): void {
    this.attackControllers.forEach((c) =>
      this.collisionSystem.unregisterHitbox(c.getHitbox()),
    )
    this.animator?.destroy()
  }
}
