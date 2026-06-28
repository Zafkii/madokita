import Phaser from "phaser"
import { GameEffect } from "./GameEffect"
import { MadokaBurstEffect } from "./madoka/BurstEffect"

export type EffectRegistry = {
  burst: new (scene: Phaser.Scene, owner?: Phaser.Physics.Arcade.Sprite) => GameEffect
  ultimate?: new (scene: Phaser.Scene, owner?: Phaser.Physics.Arcade.Sprite) => GameEffect
}

const DefaultEffects: Record<string, EffectRegistry> = {
  madoka: {
    burst: MadokaBurstEffect,
  },
  kyouko: {
    burst: MadokaBurstEffect,
  },
}

export class EffectManager {
  private scene: Phaser.Scene
  private activeEffects: Map<string, GameEffect> = new Map()
  private registries: Record<string, EffectRegistry>

  constructor(scene: Phaser.Scene, customRegistries?: Record<string, EffectRegistry>) {
    this.scene = scene
    this.registries = { ...DefaultEffects, ...customRegistries }
  }

  registerCharacterEffects(
    characterKey: string,
    registry: EffectRegistry,
  ): void {
    this.registries[characterKey] = registry
  }

  playBurst(
    characterKey: string,
    owner?: Phaser.Physics.Arcade.Sprite,
  ): void {
    this.stop("burst")
    const registry = this.registries[characterKey]
    if (!registry?.burst) return

    const effect = new registry.burst(this.scene, owner)
    this.activeEffects.set("burst", effect)
    effect.play()
  }

  playUltimate(
    characterKey: string,
    owner?: Phaser.Physics.Arcade.Sprite,
  ): void {
    this.stop("ultimate")
    const registry = this.registries[characterKey]
    if (!registry?.ultimate) return

    const effect = new registry.ultimate(this.scene, owner)
    this.activeEffects.set("ultimate", effect)
    effect.play()
  }

  stop(effectId: string): void {
    const existing = this.activeEffects.get(effectId)
    if (existing) {
      existing.stop()
      existing.destroy()
      this.activeEffects.delete(effectId)
    }
  }

  stopAll(): void {
    this.activeEffects.forEach((effect) => {
      effect.stop()
      effect.destroy()
    })
    this.activeEffects.clear()
  }

  update(delta: number): void {
    this.activeEffects.forEach((effect) => {
      effect.update(delta)
    })
  }

  destroy(): void {
    this.stopAll()
  }
}
