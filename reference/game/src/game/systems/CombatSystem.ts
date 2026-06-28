import { CombatActor } from "../entities/base/combat/CombatActor"

export class CombatSystem {
  private actors: CombatActor[] = []

  register(actor: CombatActor): void {
    this.actors.push(actor)
  }

  unregister(actor: CombatActor): void {
    this.actors = this.actors.filter((a) => a !== actor)
  }

  update(delta: number): void {
    for (const actor of this.actors) {
      if (actor.isAlive() || actor.isDead()) {
        actor.update(delta)
      }
    }
  }
}
