import Phaser from "phaser"
import { GameEffect } from "../GameEffect"
import { DepthLayers } from "../../config/DepthLayers"

export class MadokaBurstEffect extends GameEffect {
  private flashSprite: Phaser.GameObjects.Rectangle | null = null
  private glowSprite: Phaser.GameObjects.Arc | null = null
  private particles: Phaser.GameObjects.Rectangle[] = []
  private duration = 1200
  private elapsed = 0

  play(): void {
    if (this.isPlaying) return
    this.isPlaying = true
    this.elapsed = 0

    const cx = this.owner?.x ?? this.scene.cameras.main.centerX
    const cy = this.owner?.y ?? this.scene.cameras.main.centerY

    // Screen flash
    this.flashSprite = this.scene.add.rectangle(
      this.scene.cameras.main.centerX,
      this.scene.cameras.main.centerY,
      this.scene.cameras.main.width * 2,
      this.scene.cameras.main.height * 2,
      0xffffff,
      0.6,
    )
    this.flashSprite.setScrollFactor(0)
    this.flashSprite.setDepth(DepthLayers.EFFECTS_FG)

    // Glow burst around player
    this.glowSprite = this.scene.add.arc(cx, cy, 10, 0, 360, false, 0xff88ff, 0.8)
    this.glowSprite.setDepth(DepthLayers.EFFECTS_FG)
    this.glowSprite.setScale(0.5)

    this.scene.tweens.add({
      targets: this.flashSprite,
      alpha: { from: 0.6, to: 0 },
      duration: 600,
      ease: "Power2",
    })

    this.scene.tweens.add({
      targets: this.glowSprite,
      radius: { from: 10, to: 120 },
      alpha: { from: 0.8, to: 0 },
      scaleX: { from: 1, to: 3 },
      scaleY: { from: 1, to: 3 },
      duration: 400,
      ease: "Power2",
    })

    // Sparkle particles
    for (let i = 0; i < 8; i++) {
      const angle = (i / 8) * Math.PI * 2
      const dist = Phaser.Math.Between(30, 80)
      const px = cx + Math.cos(angle) * dist
      const py = cy + Math.sin(angle) * dist

      const spark = this.scene.add.rectangle(px, py, 4, 4, 0xffaaff, 1)
      spark.setDepth(DepthLayers.EFFECTS_FG)
      this.particles.push(spark)

      this.scene.tweens.add({
        targets: spark,
        x: px + Math.cos(angle) * 60,
        y: py + Math.sin(angle) * 60 - 30,
        alpha: { from: 1, to: 0 },
        scaleX: { from: 1, to: 0 },
        scaleY: { from: 1, to: 0 },
        duration: Phaser.Math.Between(300, 600),
        delay: Phaser.Math.Between(0, 200),
        ease: "Power2",
        onComplete: () => spark.destroy(),
      })
    }

    // Camera shake
    this.scene.cameras.main.shake(200, 0.005)
  }

  stop(): void {
    if (!this.isPlaying) return
    this.isPlaying = false

    if (this.flashSprite) {
      this.flashSprite.destroy()
      this.flashSprite = null
    }
    if (this.glowSprite) {
      this.glowSprite.destroy()
      this.glowSprite = null
    }
    this.particles.forEach((p) => {
      if (p.active) p.destroy()
    })
    this.particles = []
  }

  update(_delta: number): void {
    if (!this.isPlaying) return

    this.elapsed += _delta
    if (this.elapsed >= this.duration) {
      this.stop()
    }
  }

  destroy(): void {
    this.stop()
  }
}
