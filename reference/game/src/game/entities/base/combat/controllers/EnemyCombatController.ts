import Phaser from "phaser"
import type { AttackConfig } from "../../../../types/CombatTypes"
import { AttackController } from "./AttackControllers"
import { CombatTargetSystem } from "../CombatTargetSystem"

export class EnemyCombatController {
  private attacks = new Map<string, AttackController>()
  private currentAttack?: AttackController

  constructor(
    scene: Phaser.Scene,
    sprite: Phaser.Physics.Arcade.Sprite,
    targetSystem: CombatTargetSystem,
    attackConfigs: AttackConfig[],
  ) {
    attackConfigs.forEach((config) => {
      const controller = new AttackController(
        scene,
        sprite,
        targetSystem,
        config,
      )

      this.attacks.set(config.id, controller)
    })
  }

  update(): void {
    this.attacks.forEach((attack) => {
      attack.update()
    })
  }

  tryAttack(id: string): boolean {
    const attack = this.attacks.get(id)

    if (!attack) {
      return false
    }

    const started = attack.tryAttack()

    if (started) {
      this.currentAttack = attack
    }

    return started
  }

  isAttacking(): boolean {
    if (!this.currentAttack) {
      return false
    }

    return this.currentAttack.isRunning()
  }

  interruptAll(): void {
    this.attacks.forEach((attack) => {
      attack.interrupt()
    })
    this.currentAttack = undefined
  }
}
