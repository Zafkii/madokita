import Phaser from "phaser"
import { DEBUG_CONFIG } from "../../../../config/debugConfig"

type CombatDebugViewConfig = {
  fillColor: number
  fillAlpha?: number

  strokeColor?: number
  strokeAlpha?: number

  lineWidth?: number

  visibleWhenInactive?: boolean
}

export class CombatDebugView {
  private graphics?: Phaser.GameObjects.Graphics

  private config: CombatDebugViewConfig

  constructor(scene: Phaser.Scene, config: CombatDebugViewConfig) {
    this.config = config

    if (DEBUG_CONFIG.combat.enabled) {
      this.graphics = scene.add.graphics()

      this.graphics.setDepth(9999)
    }
  }

  update(rects: Phaser.Geom.Rectangle[], active: boolean = true): void {
    if (!this.graphics) {
      return
    }

    this.graphics.clear()

    if (!active && !this.config.visibleWhenInactive) {
      return
    }

    const fillAlpha = this.config.fillAlpha ?? 0.2

    const strokeColor = this.config.strokeColor ?? this.config.fillColor

    const strokeAlpha = this.config.strokeAlpha ?? 1

    const lineWidth = this.config.lineWidth ?? 2

    for (const rect of rects) {
      this.graphics.fillStyle(this.config.fillColor, fillAlpha)
      this.graphics.fillRect(rect.x, rect.y, rect.width, rect.height)
      this.graphics.lineStyle(lineWidth, strokeColor, strokeAlpha)
      this.graphics.strokeRect(rect.x, rect.y, rect.width, rect.height)
    }
  }

  destroy(): void {
    this.graphics?.destroy()
  }
}
