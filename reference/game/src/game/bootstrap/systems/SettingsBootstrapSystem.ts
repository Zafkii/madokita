import type { PlatformAPI } from "../../platform/shared/PlatformAPI"
import type { RuntimeSystem } from "../../runtime/RuntimeSystem"
import { GameRuntime } from "../../runtime/GameRuntime"
import { GameSettingsManager } from "../../settings/GameSettingsManager"
import { BrowserSettingsRepository } from "../../settings/repositories/BrowserSettingsRepository"
import { ElectronSettingsRepository } from "../../settings/repositories/ElectronSettingsRepository"

export class SettingsBootstrapSystem implements RuntimeSystem {
  readonly id = "settings-system"

  async initialize(platform: PlatformAPI): Promise<void> {
    GameRuntime.getServices().register("settings", GameSettingsManager)

    const repository = platform.isElectron()
      ? new ElectronSettingsRepository()
      : new BrowserSettingsRepository()

    await GameSettingsManager.initialize(repository)
  }
}
