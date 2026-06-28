import Phaser from "phaser"
import { Hurtbox } from "./Hurtbox"
import type { HurtboxData } from "./MovementTypes"

export class HurtboxSystem {
  private hurtbox: Hurtbox

  constructor(scene: Phaser.Scene, sprite: Phaser.Physics.Arcade.Sprite) {
    this.hurtbox = new Hurtbox(scene, sprite)
  }

  update(data: HurtboxData[], frameIndex: number, charScale: number): void {
    this.hurtbox.update(data, frameIndex, charScale)
  }

  getRects(): { rect: Phaser.Geom.Rectangle; damageMultiplier: number }[] {
    return this.hurtbox.getBodyRects()
  }

  getRect(): Phaser.Geom.Rectangle {
    return this.hurtbox.getRect()
  }

  getBodyCount(): number {
    return this.hurtbox.getBodyCount()
  }
}
