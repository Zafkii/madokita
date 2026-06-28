import Phaser from "phaser"

import type { HitboxConfig } from "../../../types/CombatTypes"

import { CombatDebugView } from "./debug/CombatDebugView"

import { DEBUG_COMBAT_CONFIG } from "../../../config/debugCombatConfig"

export class Hitbox {
  private sprite: Phaser.Physics.Arcade.Sprite
  private zones: Phaser.GameObjects.Zone[]
  private bodies: Phaser.Physics.Arcade.Body[]
  private config: HitboxConfig

  private active: boolean

  // =========================
  // DEBUG
  // =========================

  private debugView: CombatDebugView

  constructor(
    scene: Phaser.Scene,
    sprite: Phaser.Physics.Arcade.Sprite,
    config: HitboxConfig,
  ) {
    this.sprite = sprite
    this.config = config

    this.active = config.active ?? false

    this.zones = []
    this.bodies = []

    for (let i = 0; i < config.subHitboxes.length; i++) {
      const zone = scene.add.zone(sprite.x, sprite.y, 1, 1)

      scene.physics.add.existing(zone)

      const body = zone.body as Phaser.Physics.Arcade.Body

      body.setAllowGravity(false)
      body.setImmovable(true)

      zone.setActive(this.active)
      zone.setVisible(false)

      this.zones.push(zone)
      this.bodies.push(body)
    }

    // =========================
    // DEBUG
    // =========================

    this.debugView = new CombatDebugView(scene, DEBUG_COMBAT_CONFIG.hitbox)
  }

  update(frameIndex = 0): void {
    const spriteBody = this.sprite.body as Phaser.Physics.Arcade.Body
    const facing = this.sprite.flipX ? -1 : 1
    const charScale = this.sprite.scaleX

    for (let i = 0; i < this.config.subHitboxes.length; i++) {
      const sh = this.config.subHitboxes[i]
      const idx = Math.min(frameIndex, sh.width.length - 1)

      const w = sh.width[idx]
      const h = sh.height[idx]
      const ox = sh.offsetX[idx] ?? 0
      const oy = sh.offsetY[idx] ?? 0

      const x = spriteBody.center.x + ox * facing * charScale
      const y = spriteBody.center.y + oy * charScale

      this.bodies[i].setSize(w, h, false)
      this.bodies[i].setOffset(0, 0)
      this.bodies[i].reset(
        x + this.zones[i].displayOriginX - this.bodies[i].halfWidth,
        y + this.zones[i].displayOriginY - this.bodies[i].halfHeight,
      )
    }

    this.debugView.update(this.getRects(), this.active)
  }

  activate(): void {
    this.active = true
    for (const zone of this.zones) {
      zone.setActive(true)
    }
  }

  deactivate(): void {
    this.active = false
    for (const zone of this.zones) {
      zone.setActive(false)
    }
  }

  isActive(): boolean {
    return this.active
  }

  getBody(): Phaser.Physics.Arcade.Body {
    return this.bodies[0]
  }

  getGameObject(): Phaser.GameObjects.Zone {
    return this.zones[0]
  }

  getConfig(): HitboxConfig {
    return this.config
  }

  getRect(): Phaser.Geom.Rectangle {
    const body = this.bodies[0]
    return new Phaser.Geom.Rectangle(body.x, body.y, body.width, body.height)
  }

  getRects(): Phaser.Geom.Rectangle[] {
    return this.bodies.map(body => new Phaser.Geom.Rectangle(body.x, body.y, body.width, body.height))
  }
}
