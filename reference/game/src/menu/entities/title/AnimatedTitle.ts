import Phaser from "phaser"
import { AnimatedTitleConfig } from "./AnimatedTitleConfig"
import type { ColorPalette } from "../../config/menuTimings"

export class AnimatedTitle {
  private scene: Phaser.Scene
  private container!: Phaser.GameObjects.Container
  private cosmic!: Phaser.GameObjects.TileSprite
  private overlay!: Phaser.GameObjects.Image
  private cosmicTween?: Phaser.Tweens.Tween
  private cycleTimer?: Phaser.Time.TimerEvent
  private colorTween?: Phaser.Tweens.Tween

  constructor(scene: Phaser.Scene) {
    this.scene = scene
  }

  create(): void {
    const scaleRatio =
      this.scene.scale.height / AnimatedTitleConfig.DESIGN_HEIGHT
    const centerX = this.scene.scale.width / 2
    const y =
      this.scene.scale.height *
      (AnimatedTitleConfig.Y / AnimatedTitleConfig.DESIGN_HEIGHT)
    this.container = this.scene.add.container(centerX, y)

    this.cosmic = this.scene.add.tileSprite(
      0,
      0,
      AnimatedTitleConfig.WIDTH,
      AnimatedTitleConfig.HEIGHT,
      "texture-bg",
    )
    this.cosmic.setOrigin(0.5)
    this.cosmic.setAlpha(0)

    this.cosmic.enableFilters()
    this.cosmic.filters!.internal.addMask("madokita-title-mask", false)

    this.overlay = this.scene.add.image(0, 0, "madokita-title-overlay")
    this.overlay.setOrigin(0.5)
    this.overlay.setTint(0xffffff).setTintMode(Phaser.TintModes.FILL)

    const s = AnimatedTitleConfig.BASE_SCALE * scaleRatio
    this.cosmic.setScale(s)
    this.overlay.setScale(s)

    this.container.add(this.cosmic)
    this.container.add(this.overlay)

    this.scene.scale.on("resize", this.handleResize, this)
  }

  showCosmic(duration = 2200): void {
    this.cosmicTween?.stop()

    this.cosmicTween = this.scene.tweens.add({
      targets: this.cosmic,
      alpha: 1,
      duration,
      ease: "Sine.easeOut",
    })
  }

  hideCosmic(duration = 6000): void {
    this.cosmicTween?.stop()
    this.cosmicTween = this.scene.tweens.add({
      targets: this.cosmic,
      alpha: 0,
      duration,
      ease: "Sine.easeInOut",
    })
  }

  update(delta: number): void {
    if (this.cosmic.alpha <= 0.001) {
      return
    }

    const time = this.scene.time.now * 0.00018
    this.cosmic.tilePositionX += AnimatedTitleConfig.COSMIC_TRAVEL_X * delta
    this.cosmic.tilePositionY += AnimatedTitleConfig.COSMIC_TRAVEL_Y * delta
    this.cosmic.tilePositionX +=
      Math.sin(time * 0.002) * AnimatedTitleConfig.COSMIC_WOBBLE_X
    this.cosmic.tilePositionY +=
      Math.sin(time * 0.002) * AnimatedTitleConfig.COSMIC_WOBBLE_Y
    this.cosmic.tilePositionX +=
      Math.cos(time * 0.02) * AnimatedTitleConfig.COSMIC_JITTER_X
    this.cosmic.tilePositionY +=
      Math.cos(time * 0.06) * AnimatedTitleConfig.COSMIC_JITTER_Y
  }

  startOverlayCycle(
    colors: ColorPalette,
    interval: number,
    transition: number,
  ): void {
    const activeColors = Object.entries(colors)
      .filter(([, c]) => c.enabled)
      .map(([name, c]) => ({
        name,
        hex: AnimatedTitle.rgbToHex(c.rgb),
        alpha: c.alpha,
      }))

    if (activeColors.length === 0) {
      return
    }

    let index = 0
    this.tweenToColor(activeColors[0].hex, activeColors[0].alpha, transition)

    this.cycleTimer = this.scene.time.addEvent({
      delay: interval * 1000,
      loop: true,
      callback: () => {
        index = (index + 1) % activeColors.length
        this.tweenToColor(
          activeColors[index].hex,
          activeColors[index].alpha,
          transition,
        )
      },
    })
  }

  private static rgbToHex(rgb: string): number {
    const match = rgb.match(/\d+/g)
    if (!match || match.length < 3) return 0xffffff
    const [r, g, b] = match.map(Number)
    return (r << 16) | (g << 8) | b
  }

  stopOverlayCycle(
    initialColor?: { rgb: string; alpha: number },
    duration?: number,
  ): void {
    this.cycleTimer?.destroy()
    this.cycleTimer = undefined
    this.colorTween?.stop()
    this.colorTween = undefined

    if (initialColor && duration !== undefined) {
      this.tweenToColor(
        AnimatedTitle.rgbToHex(initialColor.rgb),
        initialColor.alpha,
        duration,
      )
    }
  }

  private tweenToColor(
    targetHex: number,
    targetAlpha?: number,
    duration = 0,
  ): void {
    this.colorTween?.stop()

    if (duration <= 0 && targetAlpha === undefined) {
      this.overlay.setTint(targetHex).setTintMode(Phaser.TintModes.FILL)
      if (targetAlpha !== undefined) this.overlay.setAlpha(targetAlpha)
      return
    }

    const current = Phaser.Display.Color.IntegerToColor(
      this.overlay.tintTopLeft ?? 0xffffff,
    )
    const target = Phaser.Display.Color.IntegerToColor(targetHex)
    const start = {
      r: current.red,
      g: current.green,
      b: current.blue,
      a: this.overlay.alpha,
    }

    this.colorTween = this.scene.tweens.add({
      targets: start,
      r: target.red,
      g: target.green,
      b: target.blue,
      a: targetAlpha,
      duration: duration * 1000,
      ease: "Sine.easeInOut",
      onUpdate: () => {
        this.overlay
          .setTint(
            Phaser.Display.Color.GetColor(
              Math.round(start.r),
              Math.round(start.g),
              Math.round(start.b),
            ),
          )
          .setTintMode(Phaser.TintModes.FILL)
        this.overlay.setAlpha(start.a)
      },
    })
  }

  setOverlayColor(rgb: string, alpha: number): void {
    this.overlay
      .setTint(AnimatedTitle.rgbToHex(rgb))
      .setTintMode(Phaser.TintModes.FILL)
    this.overlay.setAlpha(alpha)
  }

  destroy(): void {
    this.scene.scale.off("resize", this.handleResize, this)
    this.cosmicTween?.stop()
    this.stopOverlayCycle()
    this.container.destroy()
  }

  handleResize(): void {
    const h = this.scene.scale.height
    const scaleRatio = h / AnimatedTitleConfig.DESIGN_HEIGHT
    this.container.setX(this.scene.scale.width / 2)
    this.container.setY(h * (AnimatedTitleConfig.Y / AnimatedTitleConfig.DESIGN_HEIGHT))
    const s = AnimatedTitleConfig.BASE_SCALE * scaleRatio
    this.cosmic.setScale(s)
    this.overlay.setScale(s)
  }
}
