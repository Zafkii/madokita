import type { PlatformAPI } from "../platform/shared/PlatformAPI"
import { RuntimePhase } from "./RuntimePhase"
import { RuntimeSystemManager } from "./RuntimeSystemManager"
import { LocalizationBootstrapSystem } from "../bootstrap/systems/LocalizationBootstrapSystem"
import { SaveBootstrapSystem } from "../bootstrap/systems/SaveBootstrapSystem"
import { SettingsBootstrapSystem } from "../bootstrap/systems/SettingsBootstrapSystem"
import { GameServiceContainer } from "./services/GameServiceContainer"
import { EventBus } from "../core/EventBus"

export class GameRuntime {
  private static phase: RuntimePhase = RuntimePhase.PRE_INIT
  private static readonly systems = new RuntimeSystemManager()
  private static readonly services = new GameServiceContainer()
  static getPhase(): RuntimePhase {
    return this.phase
  }

  static getServices(): GameServiceContainer {
    return this.services
  }

  static async initialize(platform: PlatformAPI): Promise<void> {
    this.registerCoreServices()
    this.registerSystems()
    this.phase = RuntimePhase.INIT
    await this.systems.initialize(platform)
    this.phase = RuntimePhase.POST_INIT
    await this.systems.start(platform)
    this.phase = RuntimePhase.READY
    await this.systems.ready(platform)
    this.phase = RuntimePhase.RUNNING
  }

  private static registerCoreServices(): void {
    this.services.register("events", EventBus)
  }

  private static registerSystems(): void {
    this.systems.register(new LocalizationBootstrapSystem())
    this.systems.register(new SaveBootstrapSystem())
    this.systems.register(new SettingsBootstrapSystem())
  }

  static async shutdown(platform: PlatformAPI): Promise<void> {
    this.phase = RuntimePhase.SHUTDOWN
    EventBus.clear()

    await this.systems.shutdown(platform)
  }
}
