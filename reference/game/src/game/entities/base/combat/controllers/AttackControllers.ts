import Phaser from "phaser"
import type { AttackConfig } from "../../../../types/CombatTypes"
import { Hitbox } from "../Hitbox"
import { CombatTargetSystem } from "../CombatTargetSystem"
import { CombatActor } from "../CombatActor"

export class AttackController {
  private scene: Phaser.Scene
  private sprite: Phaser.Physics.Arcade.Sprite
  private targetSystem: CombatTargetSystem
  private attack: AttackConfig
  private hitbox: Hitbox
  private lastAttackTime = 0
  private attacking = false
  private owner: CombatActor

  private activeStartTime = 0
  private activeTimer: Phaser.Time.TimerEvent | null = null
  private recoverTimer: Phaser.Time.TimerEvent | null = null
  private windupTimer: Phaser.Time.TimerEvent | null = null
  private haStartTimer: Phaser.Time.TimerEvent | null = null
  private haEndTimer: Phaser.Time.TimerEvent | null = null

  // TRACKING
  private attackVector = new Phaser.Math.Vector2()
  private animationsEnabled = true
  private isInWindupOrActive = false
  private spacingStartX = 0

  constructor(
    scene: Phaser.Scene,
    sprite: Phaser.Physics.Arcade.Sprite,
    targetSystem: CombatTargetSystem,
    attack: AttackConfig,
    owner?: CombatActor,
  ) {
    this.scene = scene
    this.sprite = sprite
    this.targetSystem = targetSystem
    this.attack = attack
    this.hitbox = new Hitbox(scene, sprite, attack.hitbox)
    this.owner = owner!
  }

  setOwner(actor: CombatActor): void {
    this.owner = actor
  }

  setAnimationsEnabled(enabled: boolean): void {
    this.animationsEnabled = enabled
  }

  setTiming(windup: number, activeTime: number, recover: number): void {
    this.attack.windup = windup
    this.attack.activeTime = activeTime
    this.attack.recover = recover
  }

  getHitbox(): Hitbox {
    return this.hitbox
  }

  getConfig(): AttackConfig {
    return this.attack
  }

  update(): void {
    const frameIndex = this.activeStartTime > 0
      ? Math.floor((this.scene.time.now - this.activeStartTime) / 50)
      : 0
    this.hitbox.update(frameIndex)

    if (this.isInWindupOrActive && this.attack.spacingDistance && this.attack.spacingSpeed) {
      const dist = Math.abs(this.sprite.x - this.spacingStartX)
      if (dist < this.attack.spacingDistance) {
        const body = this.sprite.body as Phaser.Physics.Arcade.Body
        const dir = this.sprite.flipX ? -1 : 1
        body.setVelocityX(dir * this.attack.spacingSpeed)
      }
    }

    if (this.attacking && this.owner && !this.owner.isAlive()) {
      this.interrupt()
    }

    if (
      this.attacking &&
      this.owner &&
      this.owner.isStaggered &&
      !this.owner.isHyperArmoring
    ) {
      this.interrupt()
    }
  }

  tryAttack(): boolean {
    const time = this.scene.time.now

    if (this.attacking) {
      return false
    }

    if (time < this.lastAttackTime + this.attack.cooldown) {
      return false
    }

    if (this.owner && this.owner.isStaggered) {
      return false
    }

    if (this.owner && !this.owner.isAlive()) {
      return false
    }

    // Stamina check
    const staminaCost = this.attack.staminaCost ?? 0
    if (staminaCost > 0 && this.owner && !this.owner.consumeStamina(staminaCost)) {
      return false
    }

    this.attacking = true
    this.isInWindupOrActive = true
    this.lastAttackTime = time

    this.captureAttackVector()

    if (this.animationsEnabled) {
      if (this.attack.texture) {
        this.sprite.setTexture(this.attack.texture)
      }
    }

    this.spacingStartX = this.sprite.x

    const windup = this.attack.windup ?? 0
    this.windupTimer = this.scene.time.delayedCall(windup, () => {
      if (!this.attacking) return
      this.activateAttack()
    })

    return true
  }

  private captureAttackVector(): void {
    const target = this.findClosestTarget()

    if (!target) {
      const facing = this.sprite.flipX ? -1 : 1
      this.attackVector = new Phaser.Math.Vector2(facing, 0)
      return
    }

    const dx = target.sprite.x - this.sprite.x
    const dy = target.sprite.y - this.sprite.y
    const vector = new Phaser.Math.Vector2(dx, dy)

    if (!this.attack.allowVerticalTracking) {
      vector.y = 0
    }

    vector.normalize()

    const tracking = this.attack.trackingStrength ?? 0
    const facing = this.sprite.flipX ? -1 : 1
    const baseVector = new Phaser.Math.Vector2(facing, 0)

    this.attackVector = baseVector.lerp(vector, tracking).normalize()
    this.sprite.setFlipX(this.attackVector.x < 0)
  }

  private findClosestTarget(): CombatActor | undefined {
    const targets = this.targetSystem.getEnemiesForPosition(
      this.sprite.x,
      this.sprite.y,
    )

    return targets.find((target) => target.sprite !== this.sprite)
  }

  private activateAttack(): void {
    if (!this.attacking) return

    this.activeStartTime = this.scene.time.now
    this.hitbox.activate()

    // Schedule hyper armor window
    if (this.attack.hyperArmor && this.owner) {
      const haStart = this.attack.hyperArmorStart ?? 0
      const haEnd = this.attack.hyperArmorEnd ?? this.attack.activeTime ?? 250

      this.haStartTimer = this.scene.time.delayedCall(haStart, () => {
        if (this.attacking && this.owner) {
          this.owner.setHyperArmor(true, this.owner.stats.hyperArmorPoiseAbsorption)
        }
      })

      this.haEndTimer = this.scene.time.delayedCall(haEnd, () => {
        this.owner?.setHyperArmor(false)
      })
    }

    if (this.attack.type === "lunge" || this.attack.type === "dash") {
      const speed = this.attack.lungeSpeed ?? 200
      const body = this.sprite.body as Phaser.Physics.Arcade.Body
      body.setVelocity(this.attackVector.x * speed, this.attackVector.y * speed)
    }

    this.activeTimer = this.scene.time.delayedCall(this.attack.activeTime ?? 250, () => {
      if (!this.attacking) return
      this.endAttack()
    })
  }

  private endAttack(): void {
    if (!this.attacking) return

    this.isInWindupOrActive = false
    this.hitbox.deactivate()
    this.owner?.setHyperArmor(false)

    const body = this.sprite.body as Phaser.Physics.Arcade.Body
    body.setVelocity(0, 0)

    const recoverDuration = this.attack.recover ?? 400
    this.recoverTimer = this.scene.time.delayedCall(recoverDuration, () => {
      this.attacking = false
    })
  }

  interrupt(): void {
    this.attacking = false
    this.isInWindupOrActive = false
    this.activeStartTime = 0
    this.spacingStartX = 0
    this.hitbox.deactivate()
    this.owner?.setHyperArmor(false)

    this.destroyTimers()

    const body = this.sprite.body as Phaser.Physics.Arcade.Body
    body.setVelocity(0, 0)
  }

  private destroyTimers(): void {
    if (this.windupTimer) { this.windupTimer.destroy(); this.windupTimer = null }
    if (this.activeTimer) { this.activeTimer.destroy(); this.activeTimer = null }
    if (this.recoverTimer) { this.recoverTimer.destroy(); this.recoverTimer = null }
    if (this.haStartTimer) { this.haStartTimer.destroy(); this.haStartTimer = null }
    if (this.haEndTimer) { this.haEndTimer.destroy(); this.haEndTimer = null }
  }

  isRunning(): boolean {
    return this.attacking
  }
}
