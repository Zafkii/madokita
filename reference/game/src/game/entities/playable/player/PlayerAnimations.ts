import Phaser from "phaser"
import type { PlayerState } from "./PlayerState"
import type { HurtboxData, Movement } from "../../base/combat/MovementTypes"
import { MultiSpriteAnimator } from "../../base/combat/MultiSpriteAnimator"

export class PlayerAnimations {
  private sprite: Phaser.Physics.Arcade.Sprite
  private playerState: PlayerState
  private defaultTextureKey: string = ""
  private guardTimer: Phaser.Time.TimerEvent | null = null
  private bodyAnimator: MultiSpriteAnimator | null = null
  private currentBodyAnim: string | null = null
  private onOverrideAttack: (() => void) | null = null

  constructor(
    sprite: Phaser.Physics.Arcade.Sprite,
    playerState: PlayerState,
  ) {
    this.sprite = sprite
    this.playerState = playerState
  }

  setOverrideAttackCallback(fn: () => void): void {
    this.onOverrideAttack = fn
  }

  setDefaultTextureKey(key: string): void {
    this.defaultTextureKey = key
  }

  initBodyAnimator(movement: Movement): void {
    this.bodyAnimator = new MultiSpriteAnimator(this.sprite.scene, movement, this.sprite)
  }

  isOnDefaultTexture(): boolean {
    return (
      !this.defaultTextureKey ||
      this.sprite.texture.key === this.defaultTextureKey
    )
  }

  restoreDefaultTexture(): void {
    if (
      this.defaultTextureKey &&
      this.sprite.texture.key !== this.defaultTextureKey
    ) {
      this.cancelGuard()
      this.sprite.setTexture(this.defaultTextureKey)
    }
  }

  playIdle(): void {
    if (this.guardTimer !== null) return
    if (this.playerState.isAttacking) return
    if (this.playerState.isAnimationLocked) return
    if (!this.canAnimate()) return
    if (this.playerState.isJumping) return
    if (this.playerState.isBursting) return

    this.restoreDefaultTexture()
    this.playBodyAnimation("idle")
  }

  playWalk(): void {
    if (this.playerState.isAnimationLocked) return
    if (!this.canAnimate()) return
    if (this.playerState.isJumping) return
    if (this.playerState.isBursting) return

    const wasAttacking = this.playerState.isAttacking
    this.restoreDefaultTexture()
    if (wasAttacking) {
      this.currentBodyAnim = ""
      this.onOverrideAttack?.()
    }
    this.playBodyAnimation("walk")
  }

  playJump(): void {
    if (!this.canAnimate()) return
    if (this.playerState.isBursting) return

    this.restoreDefaultTexture()
    this.playBodyAnimation("jump")
  }

  playDodge(): void {
    if (!this.canAnimate()) return

    this.restoreDefaultTexture()
    this.lockAnimation(200)
    this.playBodyAnimation("dodge")
  }

  playSkill(): void {
    if (!this.canAnimate()) return

    this.restoreDefaultTexture()
    this.lockAnimation(1000)
    this.playBodyAnimation("skill")
  }

  playGuard(duration?: number, _fps?: number): void {
    if (!this.canAnimate()) return

    this.cancelGuard()

    this.guardTimer = this.sprite.scene.time.delayedCall(duration ?? 1500, () => {
      this.resetFromGuard()
    })
  }

  cancelGuard(): void {
    if (this.guardTimer) {
      this.guardTimer.destroy()
      this.guardTimer = null
    }
  }

  private resetFromGuard(): void {
    this.cancelGuard()
    this.restoreDefaultTexture()
    this.playIdle()
  }

  playBurst(): void {
    if (!this.canAnimate()) return
    this.restoreDefaultTexture()
    this.playerState.isBursting = true
    this.lockAnimation(1200)
    this.playBodyAnimation("burst")
    this.sprite.scene.time.delayedCall(1200, () => {
      this.playerState.isBursting = false
      this.playIdle()
    })
  }

  private canAnimate(): boolean {
    if (this.playerState.isStaggered) return false
    if (this.playerState.isDead) return false
    return true
  }

  private lockAnimation(duration: number): void {
    this.playerState.isAnimationLocked = true
    this.sprite.scene.time.delayedCall(duration, () => {
      this.playerState.isAnimationLocked = false
    })
  }

  private playBodyAnimation(animation: string): void {
    if (!this.bodyAnimator) return
    if (this.currentBodyAnim === animation) return

    this.currentBodyAnim = animation
    this.bodyAnimator.playAnimation(animation)
  }

  getCurrentFrameHurtboxes(): HurtboxData[] | null {
    return this.bodyAnimator?.getCurrentFrameHurtboxes() ?? null
  }

  getDefaultHurtboxes(): HurtboxData[] | undefined {
    if (!this.bodyAnimator) return undefined
    return this.bodyAnimator.getDefaultHurtboxes()
  }

  update(charX: number, charY: number, charFlipX: boolean, charScale: number): void {
    if (this.playerState.isAnimationLocked) return
    if (this.isOnDefaultTexture() && this.bodyAnimator) {
      this.bodyAnimator.update(charX, charY, charFlipX, charScale)
    }
  }
}
