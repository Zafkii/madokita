import Phaser from "phaser"
import type { HurtboxData } from "./MovementTypes"
import { CombatDebugView } from "./debug/CombatDebugView"
import { DEBUG_COMBAT_CONFIG } from "../../../config/debugCombatConfig"

type HurtboxBody = {
  zone: Phaser.GameObjects.Zone
  body: Phaser.Physics.Arcade.Body
  damageMultiplier: number
}

export class Hurtbox {
  private sprite: Phaser.Physics.Arcade.Sprite
  private bodies: HurtboxBody[] = []
  private scene: Phaser.Scene
  private debugView: CombatDebugView

  constructor(scene: Phaser.Scene, sprite: Phaser.Physics.Arcade.Sprite) {
    this.scene = scene
    this.sprite = sprite
    this.debugView = new CombatDebugView(scene, DEBUG_COMBAT_CONFIG.hurtbox)
  }

  update(data: HurtboxData[], _frameIndex: number, charScale: number): void {
    const spriteBody = this.sprite.body as Phaser.Physics.Arcade.Body
    const facing = this.sprite.flipX ? -1 : 1

    while (this.bodies.length < data.length) {
      const zone = this.scene.add.zone(this.sprite.x, this.sprite.y, 1, 1)
      this.scene.physics.add.existing(zone)
      const body = zone.body as Phaser.Physics.Arcade.Body
      body.setAllowGravity(false)
      body.setImmovable(true)
      this.bodies.push({ zone, body, damageMultiplier: 1 })
    }
    while (this.bodies.length > data.length) {
      const b = this.bodies.pop()!
      b.zone.destroy()
    }

    for (let i = 0; i < data.length; i++) {
      const [w, h, ox, oy, sx, sy, rot, dmgMult] = data[i]
      const sw = w * Math.abs(sx) * charScale
      const sh = h * Math.abs(sy) * charScale
      const sox = ox * facing * charScale
      const soy = oy * charScale
      const cx = spriteBody.center.x + sox
      const cy = spriteBody.center.y + soy
      this.bodies[i].damageMultiplier = dmgMult

      if (rot !== 0) {
        const rad = rot * Math.PI / 180
        const cos = Math.cos(rad)
        const sin = Math.sin(rad)
        const hw = sw / 2
        const hh = sh / 2
        const corners = [
          { x: -hw, y: -hh },
          { x:  hw, y: -hh },
          { x:  hw, y:  hh },
          { x: -hw, y:  hh },
        ]
        let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
        for (const c of corners) {
          const rx = c.x * cos - c.y * sin + cx
          const ry = c.x * sin + c.y * cos + cy
          if (rx < minX) minX = rx
          if (ry < minY) minY = ry
          if (rx > maxX) maxX = rx
          if (ry > maxY) maxY = ry
        }
        const rcx = (minX + maxX) / 2
        const rcy = (minY + maxY) / 2
        const bw = maxX - minX
        const bh = maxY - minY
        this.bodies[i].body.setSize(bw, bh, false)
        this.bodies[i].body.setOffset(0, 0)
        this.bodies[i].body.reset(
          rcx + this.bodies[i].zone.displayOriginX - this.bodies[i].body.halfWidth,
          rcy + this.bodies[i].zone.displayOriginY - this.bodies[i].body.halfHeight,
        )
      } else {
        this.bodies[i].body.setSize(sw, sh, false)
        this.bodies[i].body.setOffset(0, 0)
        this.bodies[i].body.reset(
          cx + this.bodies[i].zone.displayOriginX - this.bodies[i].body.halfWidth,
          cy + this.bodies[i].zone.displayOriginY - this.bodies[i].body.halfHeight,
        )
      }
    }

    this.debugView.update(this.getBodyRects().map(r => r.rect))
  }

  getRect(): Phaser.Geom.Rectangle {
    const b = this.bodies[0]
    if (!b) return new Phaser.Geom.Rectangle(0, 0, 0, 0)
    return new Phaser.Geom.Rectangle(b.body.x, b.body.y, b.body.width, b.body.height)
  }

  getBodyRects(): { rect: Phaser.Geom.Rectangle; damageMultiplier: number }[] {
    return this.bodies.map((b) => ({
      rect: new Phaser.Geom.Rectangle(
        b.body.x,
        b.body.y,
        b.body.width,
        b.body.height,
      ),
      damageMultiplier: b.damageMultiplier,
    }))
  }

  getBodyCount(): number {
    return this.bodies.length
  }
}
