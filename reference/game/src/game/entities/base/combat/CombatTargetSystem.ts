import Phaser from "phaser"
import { CombatActor } from "./CombatActor"

export class CombatTargetSystem {
  private actors: CombatActor[] = []

  register(actor: CombatActor): void {
    this.actors.push(actor)
  }

  unregister(actor: CombatActor): void {
    this.actors = this.actors.filter((a) => a !== actor)
  }

  getClosestEnemy(source: CombatActor): CombatActor | undefined {
    const enemies = this.actors.filter((actor) => {
      return actor.team !== source.team && actor.isAlive()
    })

    if (enemies.length === 0) {
      return undefined
    }

    enemies.sort((a, b) => {
      const distA = Phaser.Math.Distance.Between(
        source.sprite.x,
        source.sprite.y,
        a.sprite.x,
        a.sprite.y,
      )

      const distB = Phaser.Math.Distance.Between(
        source.sprite.x,
        source.sprite.y,
        b.sprite.x,
        b.sprite.y,
      )

      return distA - distB
    })

    return enemies[0]
  }

  getEnemiesForPosition(x: number, y: number): CombatActor[] {
    return this.actors
      .filter((actor) => actor.isAlive())
      .sort((a, b) => {
        const distA = Phaser.Math.Distance.Between(x, y, a.sprite.x, a.sprite.y)

        const distB = Phaser.Math.Distance.Between(x, y, b.sprite.x, b.sprite.y)

        return distA - distB
      })
  }
}
