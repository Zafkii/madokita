import Phaser from "phaser"

import { HurtboxSystem } from "../base/combat/HurtboxSystem"
import { EnemyCombatController } from "../base/combat/controllers/EnemyCombatController"
import { EnemyAIController } from "../base/combat/ai/EnemyAIController"
import { MultiSpriteAnimator } from "../base/combat/MultiSpriteAnimator"
import { resolveHurtboxData } from "../base/combat/CombatHelpers"

import { mobButterflyAttacks } from "../../data/enemies/mobButterfly/mobButterflyAttacks"
import { mobButterflyMovement, MobButterflyData } from "../../data/enemies/mobButterfly/mobButterfly"

import { CombatActor } from "../base/combat/CombatActor"
import { CombatTeam } from "../base/combat/CombatTeam"
import { CombatTargetSystem } from "../base/combat/CombatTargetSystem"
import { CollisionSystem } from "../../systems/CollisionSystem"

export class MobButterfly extends Phaser.Physics.Arcade.Sprite {
  public hurtboxSystem: HurtboxSystem
  private combat: EnemyCombatController
  private ai: EnemyAIController
  private animator: MultiSpriteAnimator
  public readonly combatActor: CombatActor

  constructor(
    scene: Phaser.Scene,
    x: number,
    y: number,
    targetSystem: CombatTargetSystem,
    collisionSystem: CollisionSystem,
  ) {
    super(scene, x, y, "mobbutterfly")

    scene.add.existing(this)
    scene.physics.add.existing(this)

    this.setScale(0.4)
    this.setDepth(5)

    const body = this.body as Phaser.Physics.Arcade.Body
    body.allowGravity = false

    this.animator = new MultiSpriteAnimator(scene, mobButterflyMovement, this)

    this.hurtboxSystem = new HurtboxSystem(scene, this)
    this.combatActor = new CombatActor({
      actorId: crypto.randomUUID(),

      team: CombatTeam.ENEMY,

      sprite: this,

      hurtboxes: this.hurtboxSystem,

      stats: {
        maxHealth: 40,
        health: 40,

        maxStamina: 100,
        stamina: 100,
        staminaRegenDelay: 1000,
        staminaRegenRate: 20,

        attack: 5,
        defense: 1,

        maxPoise: 20,
        poise: 20,
        poiseRegenDelay: 3000,
        poiseRegenRate: 10,

        staggerImmunity: 400,
        hyperArmorPoiseAbsorption: 0.3,
      },
    })

    collisionSystem.registerHurtbox(this.hurtboxSystem, this.combatActor)

    this.combat = new EnemyCombatController(
      scene,
      this,
      targetSystem,
      mobButterflyAttacks,
    )

    this.ai = new EnemyAIController(
      this.combatActor,
      targetSystem,
      this.combat,
      {
        visionRange: MobButterflyData.visionRange,
        attackRange: MobButterflyData.attackRange,
        moveSpeed: MobButterflyData.moveSpeed,
        animator: this.animator,
      },
    )
  }

  update(): void {
    this.ai.update()
    this.combat.update()
    this.animator.update(this.x, this.y, this.flipX, this.scaleX)
    const hurtboxData = resolveHurtboxData(null, this.animator)
    if (hurtboxData) {
      this.hurtboxSystem.update(hurtboxData, 0, this.scaleX)
    }
  }
}
