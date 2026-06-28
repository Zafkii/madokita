import Phaser from "phaser"
import { GameEffect } from "../GameEffect"
import { DepthLayers } from "../../config/DepthLayers"

export class SayakaUltimateEffect extends GameEffect {
  private overlay: Phaser.GameObjects.Rectangle | null = null

  play(): void {
    if (this.isPlaying) return
    this.isPlaying = true

    // Placeholder: white bg overlay + black silhouette effect
    // Future: full-screen white flash, all non-essential elements turn white,
    // characters/enemies become black silhouettes with blue outline

    this.overlay = this.scene.add.rectangle(
      this.scene.cameras.main.centerX,
      this.scene.cameras.main.centerY,
      this.scene.cameras.main.width * 2,
      this.scene.cameras.main.height * 2,
      0xffffff,
      0,
    )
    this.overlay.setScrollFactor(0)
    this.overlay.setDepth(DepthLayers.EFFECTS_BG)

    this.scene.tweens.add({
      targets: this.overlay,
      alpha: { from: 0, to: 0.9 },
      duration: 800,
      ease: "Power2",
    })

    this.scene.cameras.main.shake(400, 0.008)
  }

  stop(): void {
    if (!this.isPlaying) return
    this.isPlaying = false

    if (this.overlay) {
      this.scene.tweens.add({
        targets: this.overlay,
        alpha: { from: this.overlay.alpha, to: 0 },
        duration: 600,
        onComplete: () => {
          this.overlay?.destroy()
          this.overlay = null
        },
      })
    }
  }

  destroy(): void {
    if (this.overlay) {
      this.overlay.destroy()
      this.overlay = null
    }
  }
}
