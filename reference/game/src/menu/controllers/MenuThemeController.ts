import Phaser from "phaser"
import { menuTimings } from "../config/menuTimings"

type EffectCallbacks = {
  onStart?: () => void
  onEnd?: () => void
}

export class MenuThemeController {
  private scene: Phaser.Scene
  private music?: Phaser.Sound.BaseSound
  private effectCallbacks = new Map<string, EffectCallbacks>()
  private effectActive = new Map<string, boolean>()

  constructor(scene: Phaser.Scene) {
    this.scene = scene
  }

  registerEffect(name: string, callbacks: EffectCallbacks): void {
    this.effectCallbacks.set(name, callbacks)
  }

  play(): void {
    if (this.music?.isPlaying) {
      return
    }

    this.music = this.scene.sound.add("menu-theme", {
      loop: true,
      volume: 0.7,
    })
    this.music.play()

    this.scene.time.addEvent({
      delay: 250,
      loop: true,
      callback: () => {
        if (!this.music?.isPlaying) {
          return
        }

        const webAudioSound = this.music as Phaser.Sound.WebAudioSound
        const currentTime = webAudioSound.source?.context.currentTime ?? 0
        const duration = this.music.duration
        if (duration <= 0) {
          return
        }

        const songTime = currentTime % duration

        for (const [name, timing] of Object.entries(menuTimings.effects)) {
          if (timing.enabled === false) {
            continue
          }

          const cb = this.effectCallbacks.get(name)
          if (!cb) {
            continue
          }

          const effectiveEnd =
            timing.end === "loop" ? menuTimings.trackDuration : timing.end
          const fadeStart = effectiveEnd - timing.fadeOutDuration
          const isActive = this.effectActive.get(name) ?? false

          if (songTime >= timing.start && !isActive) {
            this.effectActive.set(name, true)
            cb.onStart?.()
          }

          if (songTime >= fadeStart && isActive) {
            this.effectActive.set(name, false)
            cb.onEnd?.()
          }
        }
      },
    })
  }

  stop(): void {
    if (!this.music) {
      return
    }

    this.music.stop()
    this.music.destroy()
    this.music = undefined
  }

  fadeOut(duration = 1000): void {
    if (!this.music) {
      return
    }

    this.scene.tweens.add({
      targets: this.music,
      volume: 0,
      duration,
      onComplete: () => {
        this.stop()
      },
    })
  }
}
