import type { GameEventMap } from "./GameEventMap"

type EventHandler<T> = (payload: T) => void

export class GameEventBus {
  private listeners = new Map<string, Set<EventHandler<any>>>()

  on<K extends keyof GameEventMap>(
    event: K,
    handler: EventHandler<GameEventMap[K]>,
  ): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }

    const handlers = this.listeners.get(event)!

    handlers.add(handler)

    return () => {
      handlers.delete(handler)
    }
  }

  emit<K extends keyof GameEventMap>(event: K, payload: GameEventMap[K]): void {
    const handlers = this.listeners.get(event)

    if (!handlers) {
      return
    }

    handlers.forEach((handler) => {
      handler(payload)
    })
  }

  clear(): void {
    this.listeners.clear()
  }
}
