import Phaser from "phaser"

export abstract class GameEffect {
  protected scene: Phaser.Scene
  protected owner?: Phaser.Physics.Arcade.Sprite
  protected isPlaying = false

  constructor(scene: Phaser.Scene, owner?: Phaser.Physics.Arcade.Sprite) {
    this.scene = scene
    this.owner = owner
  }

  abstract play(): void
  abstract stop(): void
  update(_delta: number): void {}
  abstract destroy(): void

  getIsPlaying(): boolean {
    return this.isPlaying
  }
}
