import type { PlatformAPI } from "../platform/shared/PlatformAPI"
import type { RuntimeSystem } from "./RuntimeSystem"

export class RuntimeSystemManager {
  private readonly systems: RuntimeSystem[] = []

  register(system: RuntimeSystem): void {
    this.systems.push(system)
  }

  async initialize(platform: PlatformAPI): Promise<void> {
    for (const system of this.systems) {
      if (!system.initialize) {
        continue
      }
      console.log(`[Runtime] Initializing ${system.id}`)
      await system.initialize(platform)
    }
  }

  async start(platform: PlatformAPI): Promise<void> {
    for (const system of this.systems) {
      if (!system.start) {
        continue
      }
      console.log(`[Runtime] Starting ${system.id}`)
      await system.start(platform)
    }
  }

  async ready(platform: PlatformAPI): Promise<void> {
    for (const system of this.systems) {
      if (!system.ready) {
        continue
      }
      console.log(`[Runtime] Ready ${system.id}`)
      await system.ready(platform)
    }
  }

  async shutdown(platform: PlatformAPI): Promise<void> {
    for (const system of [...this.systems].reverse()) {
      if (!system.shutdown) {
        continue
      }
      console.log(`[Runtime] Shutting down ${system.id}`)
      await system.shutdown(platform)
    }
  }
}
