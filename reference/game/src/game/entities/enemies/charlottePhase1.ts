import Phaser from "phaser"
import { MultiSpriteAnimator } from "../base/combat/MultiSpriteAnimator"
import { charlottePhase1Movement } from "../../data/enemies/charlottePhase1"

export class CharlottePhase1 extends Phaser.Physics.Arcade.Sprite {
  private moveDirection = 1
  private moveSpeed = 80
  private floatOffset = 0
  private animator: MultiSpriteAnimator
  private hasStartedIdle = false

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y, "charlotte_phase_1")
    scene.add.existing(this)
    scene.physics.add.existing(this)
    this.setScale(0.5)
    this.setDepth(5)
    const body = this.body as Phaser.Physics.Arcade.Body
    body.allowGravity = false
    this.animator = new MultiSpriteAnimator(scene, charlottePhase1Movement, this)
  }

  update(): void {
    // HORIZONTAL MOVEMENT
    this.setVelocityX(this.moveDirection * this.moveSpeed)
    // FLOATING MOVEMENT
    this.floatOffset += 0.05
    this.setVelocityY(Math.sin(this.floatOffset) * 30)
    // TURN BACK
    if (this.x >= 2400) {
      this.moveDirection = -1
      this.setFlipX(true)
    }

    if (this.x <= 1600) {
      this.moveDirection = 1
      this.setFlipX(false)
    }

    if (!this.hasStartedIdle) {
      this.hasStartedIdle = true
      this.animator.playAnimation("idle")
    }
    this.animator.update(this.x, this.y, this.flipX, this.scaleX)
  }
}
