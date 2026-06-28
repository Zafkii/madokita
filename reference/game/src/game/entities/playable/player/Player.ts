import Phaser from "phaser"
import { PlayerAnimations } from "./PlayerAnimations"
import { PlayerMovement } from "./PlayerMovement"
import { PlayerState } from "./PlayerState"
import { PlayerCombat } from "./PlayerCombat"
import { HurtboxSystem } from "../../base/combat/HurtboxSystem"
import { CharacterRegistry } from "../../../data/characters/CharacterRegistry"
import { CombatActor } from "../../base/combat/CombatActor"
import { CombatTeam } from "../../base/combat/CombatTeam"
import { CombatTargetSystem } from "../../base/combat/CombatTargetSystem"
import { CollisionSystem } from "../../../systems/CollisionSystem"
import { EffectManager } from "../../../effects/EffectManager"
import { EventBus } from "../../../core/EventBus"
import { DepthLayers } from "../../../config/DepthLayers"
import {
  resolveHurtboxData,
  subscribeToCombatEvents,
} from "../../base/combat/CombatHelpers"
import type { HurtboxData } from "../../base/combat/MovementTypes"

export class Player extends Phaser.Physics.Arcade.Sprite {
  public readonly playerState: PlayerState
  public readonly animationsController: PlayerAnimations
  public readonly movement: PlayerMovement
  public readonly hurtboxes: HurtboxSystem
  public readonly combatActor: CombatActor
  public readonly combat: PlayerCombat
  public readonly characterKey: string
  private effectManager: EffectManager
  private lastHurtboxData: HurtboxData[] | null = null
  private attackSeq = 0

  constructor(
    scene: Phaser.Scene,
    x: number,
    y: number,
    texture: string,
    targetSystem: CombatTargetSystem,
    collisionSystem: CollisionSystem,
    effectManager: EffectManager,
  ) {
    super(scene, x, y, texture)

    scene.add.existing(this)
    scene.physics.add.existing(this)

    this.setCollideWorldBounds(true)
    this.setScale(0.5)

    this.characterKey = texture

    const characterData = CharacterRegistry[texture]

    if (!characterData) {
      throw new Error(`Character data not found for texture: ${texture}`)
    }

    this.playerState = new PlayerState()
    this.effectManager = effectManager

    this.hurtboxes = new HurtboxSystem(scene, this)

    this.combatActor = new CombatActor({
      actorId: crypto.randomUUID(),
      team: CombatTeam.PLAYER,
      sprite: this,
      hurtboxes: this.hurtboxes,
      stats: {
        maxHealth: 100,
        health: 100,
        maxStamina: 100,
        stamina: 100,
        staminaRegenDelay: 500,
        staminaRegenRate: 40,
        attack: 10,
        defense: 5,
        maxPoise: 50,
        poise: 50,
        poiseRegenDelay: 2000,
        poiseRegenRate: 30,
        staggerImmunity: 500,
        hyperArmorPoiseAbsorption: 0.4,
      },
    })

    this.animationsController = new PlayerAnimations(this, this.playerState)
    this.animationsController.setDefaultTextureKey(texture)
    this.animationsController.setOverrideAttackCallback(() => {
      this.combat.getAnimator()?.stop()
    })
    if (characterData.movement) {
      this.animationsController.initBodyAnimator(characterData.movement)
    }

    this.movement = new PlayerMovement(
      this,
      this.animationsController,
      this.playerState,
      this.combatActor,
    )

    this.combat = new PlayerCombat(
      scene,
      this,
      targetSystem,
      collisionSystem,
      this.combatActor,
      characterData.attackConfigs,
    )

    if (characterData.attack && this.combat) {
      this.combat.initAnimator(characterData.attack, characterData.additionalTextures)
    }

    collisionSystem.registerHurtbox(this.hurtboxes, this.combatActor)

    this.unsubscribers.push(
      ...subscribeToCombatEvents(EventBus, this.combatActor.actorId, {
        onStaggerStart: (isFlinch) => this.onStaggerStart(isFlinch),
        onStaggerEnd: () => this.onStaggerEnd(),
        onFlinchEnd: () => this.onFlinchEnd(),
        onDeath: () => this.onDeath(),
      }),
    )
  }

  private unsubscribers: (() => void)[] = []

  destroy(): void {
    this.unsubscribers.forEach((fn) => fn())
    super.destroy()
  }

  setMainPlayer(value: boolean): void {
    this.setDepth(value ? DepthLayers.PLAYERS_MAIN : DepthLayers.PLAYERS_SECONDARY)
  }

  update(): void {
    if (this.playerState.isDead) return

    this.movement.update()
    this.animationsController.update(this.x, this.y, this.flipX, this.scaleX)
    this.combat.update()
    const hurtboxData = resolveHurtboxData(this.combat, this.animationsController)
    if (hurtboxData) {
      this.lastHurtboxData = hurtboxData
      this.hurtboxes.update(hurtboxData, 0, this.scaleX)
    } else if (this.lastHurtboxData && this.playerState.isJumping) {
      this.hurtboxes.update(this.lastHurtboxData, 0, this.scaleX)
    }
  }

  setControlled(value: boolean): void {
    this.playerState.isControlled = value
  }

  isControlled(): boolean {
    return this.playerState.isControlled
  }

  moveLeft(): void {
    this.movement.moveLeft()
  }

  moveRight(): void {
    this.movement.moveRight()
  }

  idle(): void {
    this.movement.idle()
  }

  jump(): void {
    this.movement.jump()
  }

  dodge(): void {
    this.movement.dodge()
  }

  attack(id = "basic"): void {
    if (this.playerState.isStaggered || this.playerState.isDead) return
    if (this.playerState.isAttacking) return

    const cfg = this.combat.getAttackConfig(id)
    const animDef = this.combat.getAnimator()?.getAttackAnimDef(id)
    const started = this.combat.attack(id, () => {
      this.playerState.isAttacking = false
      this.playerState.isAnimationLocked = false
      if (this.animationsController.isOnDefaultTexture()) {
        this.animationsController.playIdle()
      } else {
        const guardDuration = animDef?.guard ?? cfg?.guard
        const guardFps = animDef?.guardFps
        this.combat.getAnimator()?.playGuard(id, guardDuration)
        this.animationsController.playGuard(guardDuration, guardFps)
      }
    })
    if (started) {
      this.animationsController.cancelGuard()
      this.playerState.isAttacking = true
      this.playerState.isAnimationLocked = true
      this.playerState.isMovementLocked = true

      const w = cfg?.windup ?? 0
      const at = cfg?.activeTime ?? 250
      const r = cfg?.recover ?? 400
      const seq = ++this.attackSeq

      // end of activeTime: unlock movement (recover can move)
      this.scene.time.delayedCall(w + at, () => {
        if (this.attackSeq !== seq) return
        this.playerState.isMovementLocked = false
      })

      // recover 2/3: unlock animation (last third free)
      this.scene.time.delayedCall(w + at + r * 2 / 3, () => {
        if (this.attackSeq !== seq) return
        this.playerState.isAnimationLocked = false
      })
    }
  }

  hasAttack(id: string): boolean {
    return this.combat.hasAttack(id)
  }

  burst(): void {
    if (this.playerState.isStaggered || this.playerState.isDead) return
    this.animationsController.playBurst()
    this.effectManager.playBurst(this.characterKey, this)
  }

  private onStaggerStart(isFlinch: boolean): void {
    if (isFlinch) {
      this.setTint(0xffffff)
      this.scene.time.delayedCall(80, () => {
        if (!this.playerState.isStaggered) {
          this.clearTint()
        }
      })
      return
    }

    this.playerState.isStaggered = true
    this.setVelocity(0, 0)
    this.setTint(0xff8888)
    this.combat.interruptAll()
    this.animationsController.cancelGuard()
  }

  private onFlinchEnd(): void {
    if (!this.playerState.isStaggered) {
      this.clearTint()
    }
  }

  private onStaggerEnd(): void {
    this.playerState.isStaggered = false
    this.clearTint()
    this.combatActor.exitStagger()
    this.animationsController.restoreDefaultTexture()
  }

  private onDeath(): void {
    this.playerState.isDead = true
    this.setVelocity(0, 0)
    this.setAlpha(0.4)
    this.setTint(0x444444)
    this.body?.enable && ((this.body as Phaser.Physics.Arcade.Body).enable = false)
  }
}
