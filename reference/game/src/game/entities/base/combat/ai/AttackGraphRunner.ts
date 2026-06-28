import Phaser from "phaser"

import type { AttackConfig, AttackGraph } from "../../../../types/CombatTypes"

import { AttackController } from "../controllers/AttackControllers"

import { CombatTargetSystem } from "../CombatTargetSystem"

export class AttackGraphRunner {
  private scene: Phaser.Scene

  private controllers = new Map<string, AttackController>()

  private graph: AttackGraph

  private currentNode: string

  private waiting = false

  constructor(
    scene: Phaser.Scene,

    sprite: Phaser.Physics.Arcade.Sprite,

    targetSystem: CombatTargetSystem,

    attacks: AttackConfig[],

    graph: AttackGraph,
  ) {
    this.scene = scene

    this.graph = graph

    this.currentNode = graph.start

    attacks.forEach((attack) => {
      const controller = new AttackController(
        scene,
        sprite,
        targetSystem,
        attack,
      )

      this.controllers.set(attack.id, controller)
    })
  }

  update(): void {
    const node = this.graph.nodes[this.currentNode]

    if (!node) {
      return
    }

    // =========================
    // WAIT NODE
    // =========================

    if (node.wait && !this.waiting) {
      this.waiting = true

      this.scene.time.delayedCall(node.wait, () => {
        this.waiting = false

        this.goNext(node)
      })

      return
    }

    // =========================
    // ATTACK NODE
    // =========================

    if (node.attackId) {
      const controller = this.controllers.get(node.attackId)

      if (!controller) {
        return
      }

      if (!controller.isRunning()) {
        controller.tryAttack()
      }

      controller.update()

      if (!controller.isRunning()) {
        this.goNext(node)
      }
    }
  }

  private goNext(node: any): void {
    if (!node.next || node.next.length === 0) {
      return
    }

    const nextIndex = Phaser.Math.Between(0, node.next.length - 1)

    this.currentNode = node.next[nextIndex]
  }
}
