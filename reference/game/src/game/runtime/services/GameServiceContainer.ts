import type { GameServiceMap } from "./GameServiceMap"

export class GameServiceContainer {
  private readonly services = new Map<keyof GameServiceMap, unknown>()

  register<K extends keyof GameServiceMap>(
    key: K,
    service: GameServiceMap[K],
  ): void {
    this.services.set(key, service)
  }

  get<K extends keyof GameServiceMap>(key: K): GameServiceMap[K] {
    const service = this.services.get(key)

    if (!service) {
      throw new Error(`Service "${key}" is not registered`)
    }

    return service as GameServiceMap[K]
  }

  has(key: keyof GameServiceMap): boolean {
    return this.services.has(key)
  }
}
