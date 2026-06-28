import { CombatActor } from "../CombatActor"
import { CombatTargetSystem } from "../CombatTargetSystem"
import { EnemyCombatController } from "../controllers/EnemyCombatController"
import { EnemyState } from "./EnemyState"
import { MultiSpriteAnimator } from "../MultiSpriteAnimator"
import { EventBus } from "../../../../core/EventBus"
import { subscribeToCombatEvents } from "../CombatHelpers"

export class EnemyAIController {
  private actor: CombatActor
  private targetSystem: CombatTargetSystem
  private combat: EnemyCombatController
  private visionRange: number
  private attackRange: number
  private moveSpeed: number
  private animPrefix?: string
  private animator?: MultiSpriteAnimator
  private state: EnemyState = EnemyState.IDLE
  private retreatStartTime = 0
  private retreatDuration = 600
  private previousState: EnemyState = EnemyState.IDLE
  private unsubscribers: (() => void)[] = []

  constructor(
    actor: CombatActor,
    targetSystem: CombatTargetSystem,
    combat: EnemyCombatController,
    config: {
      visionRange: number
      attackRange: number
      moveSpeed: number
      animPrefix?: string
      animator?: MultiSpriteAnimator
    },
  ) {
    this.actor = actor
    this.targetSystem = targetSystem
    this.combat = combat
    this.visionRange = config.visionRange
    this.attackRange = config.attackRange
    this.moveSpeed = config.moveSpeed
    this.animPrefix = config.animPrefix
    this.animator = config.animator

    this.unsubscribers.push(
      ...subscribeToCombatEvents(EventBus, this.actor.actorId, {
        onStaggerStart: (isFlinch) => {
          if (isFlinch) {
            this.onFlinchStart()
          } else {
            this.onStaggerStart()
          }
        },
        onStaggerEnd: () => this.onStaggerEnd(),
        onFlinchEnd: () => this.onFlinchEnd(),
        onDeath: () => this.onDeath(),
      }),
    )
  }

  destroy(): void {
    this.unsubscribers.forEach((fn) => fn())
    this.unsubscribers = []
  }

  update(): void {
    if (this.state === EnemyState.DEAD) return

    switch (this.state) {
      case EnemyState.IDLE:
        this.updateIdle()
        break
      case EnemyState.CHASE:
        this.updateChase()
        break
      case EnemyState.ATTACK:
        this.updateAttack()
        break
      case EnemyState.RETREAT:
        this.updateRetreat()
        break
      case EnemyState.STAGGERED:
        this.updateStaggered()
        break
      case EnemyState.FLINCH:
        this.updateFlinch()
        break
    }
  }

  private getTarget(): CombatActor | undefined {
    return this.targetSystem.getClosestEnemy(this.actor)
  }

  private updateIdle(): void {
    const target = this.getTarget()
    if (!target) return

    const dist = this.getAbsDistance(target)
    this.actor.sprite.setVelocityX(0)
    this.playAnim("idle")

    if (dist <= this.visionRange) {
      this.state = EnemyState.CHASE
    }
  }

  private updateChase(): void {
    const target = this.getTarget()
    if (!target) {
      this.state = EnemyState.IDLE
      return
    }

    const dist = this.getDistance(target)
    const absDist = Math.abs(dist)

    if (absDist > this.visionRange) {
      this.state = EnemyState.IDLE
      return
    }

    if (absDist <= this.attackRange) {
      const started = this.combat.tryAttack("sting")
      if (started) {
        this.state = EnemyState.ATTACK
      }
      return
    }

    const dir = Math.sign(dist)
    this.actor.sprite.setVelocityX(dir * this.moveSpeed)
    this.actor.sprite.setFlipX(dir < 0)
    this.playAnim("walk")
  }

  private updateAttack(): void {
    this.actor.sprite.setVelocityX(0)

    if (this.combat.isAttacking()) {
      return
    }

    this.retreatStartTime = this.actor.sprite.scene.time.now
    this.state = EnemyState.RETREAT
  }

  private updateRetreat(): void {
    const target = this.getTarget()
    if (!target) {
      this.state = EnemyState.IDLE
      return
    }

    const dist = this.actor.sprite.x - target.sprite.x
    const dir = Math.sign(dist)

    this.actor.sprite.setVelocityX(dir * this.moveSpeed)
    this.actor.sprite.setFlipX(dir < 0)
    this.playAnim("walk")

    const time = this.actor.sprite.scene.time.now
    if (time >= this.retreatStartTime + this.retreatDuration) {
      this.state = EnemyState.CHASE
    }
  }

  private updateStaggered(): void {
    this.actor.sprite.setVelocityX(0)
    this.playAnim("idle")
  }

  private updateFlinch(): void {
    this.actor.sprite.setVelocityX(0)
  }

  private onFlinchStart(): void {
    if (this.state === EnemyState.DEAD) return

    this.previousState = this.state
    this.state = EnemyState.FLINCH
  }

  private onFlinchEnd(): void {
    if (this.state !== EnemyState.FLINCH) return
    this.state = this.previousState
  }

  private onStaggerStart(): void {
    if (this.state === EnemyState.DEAD) return

    this.combat.interruptAll()
    this.state = EnemyState.STAGGERED
  }

  private onStaggerEnd(): void {
    if (this.state !== EnemyState.STAGGERED) return
    this.state = EnemyState.IDLE
  }

  private onDeath(): void {
    this.state = EnemyState.DEAD
    this.combat.interruptAll()
    this.actor.sprite.setVelocity(0, 0)
    this.actor.sprite.setAlpha(0.4)

    this.actor.sprite.scene.time.delayedCall(1500, () => {
      this.actor.sprite.destroy()
    })
  }

  private getDistance(target: CombatActor): number {
    return target.sprite.x - this.actor.sprite.x
  }

  private getAbsDistance(target: CombatActor): number {
    return Math.abs(this.getDistance(target))
  }

  private playAnim(name: string): void {
    if (this.animator) {
      this.animator.playAnimation(name)
      return
    }
    if (!this.animPrefix) return
    const key = `${this.animPrefix}_${name}`
    if (this.actor.sprite.anims.currentAnim?.key !== key) {
      if (this.actor.sprite.scene.anims.exists(key)) {
        this.actor.sprite.play(key)
      }
    }
  }
}
