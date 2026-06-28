import Phaser from "phaser"
import {
  PLAYER_DODGE_FORCE,
  PLAYER_DODGE_IFRAME_DURATION,
  PLAYER_DODGE_STAMINA_COST,
  PLAYER_JUMP_STAMINA_COST,
  PLAYER_JUMP_FORCE,
  PLAYER_MOVE_SPEED,
} from "./PlayerConstants"
import type { FacingDirection } from "./PlayerTypes"
import type { PlayerAnimations } from "./PlayerAnimations"
import type { PlayerState } from "./PlayerState"
import { CombatActor } from "../../base/combat/CombatActor"
export class PlayerMovement {
  private facingDirection: FacingDirection
  private sprite: Phaser.Physics.Arcade.Sprite
  private animations: PlayerAnimations
  private playerState: PlayerState
  private combatActor?: CombatActor
  constructor(
    sprite: Phaser.Physics.Arcade.Sprite,
    animations: PlayerAnimations,
    playerState: PlayerState,
    combatActor?: CombatActor,
  ) {
    this.sprite = sprite
    this.animations = animations
    this.playerState = playerState
    this.combatActor = combatActor
    // INITIAL DIRECTION
    const worldCenter = this.sprite.scene.physics.world.bounds.width / 2
    this.facingDirection = this.sprite.x < worldCenter ? "right" : "left"
    this.sprite.setFlipX(this.facingDirection === "left")
  }
  moveLeft(): void {
    if (this.playerState.isBursting || this.playerState.isStaggered || this.playerState.isDead) {
      return
    }
    if (this.playerState.isMovementLocked && !this.playerState.isJumping) {
      this.sprite.setVelocityX(0)
      return
    }
    this.sprite.setVelocityX(-PLAYER_MOVE_SPEED)
    this.sprite.setFlipX(true)
    this.facingDirection = "left"
    if (this.playerState.isJumping) {
      return
    }
    this.animations.playWalk()
  }

  moveRight(): void {
    if (this.playerState.isBursting || this.playerState.isStaggered || this.playerState.isDead) {
      return
    }
    if (this.playerState.isMovementLocked && !this.playerState.isJumping) {
      this.sprite.setVelocityX(0)
      return
    }
    this.sprite.setVelocityX(PLAYER_MOVE_SPEED)
    this.sprite.setFlipX(false)
    this.facingDirection = "right"

    if (this.playerState.isJumping) {
      return
    }

    this.animations.playWalk()
  }

  idle(): void {
    if (this.playerState.isBursting || this.playerState.isDead) {
      this.sprite.setVelocityX(0)
      return
    }

    this.sprite.setVelocityX(0)
    if (this.playerState.isJumping) {
      return
    }
    this.animations.playIdle()
  }

  jump(): void {
    if (this.playerState.isStaggered || this.playerState.isDead) {
      return
    }
    const body = this.sprite.body as Phaser.Physics.Arcade.Body
    if (!body.onFloor()) {
      return
    }
    if (this.combatActor && !this.combatActor.consumeStamina(PLAYER_JUMP_STAMINA_COST)) {
      return
    }
    // SET STATE FIRST
    this.playerState.isJumping = true
    // JUMP FORCE
    this.sprite.setVelocityY(-PLAYER_JUMP_FORCE)
    // FORCE JUMP ANIMATION
    this.animations.playJump()
  }

  dodge(): void {
    if (this.playerState.isStaggered || this.playerState.isDead) {
      return
    }
    if (this.combatActor && !this.combatActor.consumeStamina(PLAYER_DODGE_STAMINA_COST)) {
      return
    }
    const direction = this.facingDirection === "right" ? 1 : -1
    this.sprite.setVelocityX(direction * PLAYER_DODGE_FORCE)
    this.animations.playDodge()

    if (this.combatActor) {
      this.combatActor.setInvincible(PLAYER_DODGE_IFRAME_DURATION)
      this.playerState.isInvincible = true
      this.sprite.setAlpha(0.6)
      this.sprite.scene.time.delayedCall(PLAYER_DODGE_IFRAME_DURATION, () => {
        this.playerState.isInvincible = false
        this.sprite.setAlpha(1)
      })
    }
  }

  update(): void {
    const body = this.sprite.body as Phaser.Physics.Arcade.Body
    // FALL DETECTION
    if (!body.onFloor() && body.velocity.y !== 0) {
      this.playerState.isJumping = true
    }
    // LAND DETECTION
    if (body.onFloor() && body.velocity.y >= 0) {
      if (this.playerState.isJumping) {
        this.playerState.isJumping = false

        this.animations.playIdle()
      }
    }
  }
}
