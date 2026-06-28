import type { PlatformAPI } from "../../platform/shared/PlatformAPI"
import type { RuntimeSystem } from "../../runtime/RuntimeSystem"
import { SaveManager } from "../../save/SaveManager"
import { BrowserSaveRepository } from "../../save/repositories/BrowserSaveRepository"
import { ElectronSaveRepository } from "../../save/repositories/ElectronSaveRepository"
import { GameRuntime } from "../../runtime/GameRuntime"

export class SaveBootstrapSystem implements RuntimeSystem {
  readonly id = "save-system"
  async initialize(platform: PlatformAPI): Promise<void> {
    const repository = platform.isElectron()
      ? new ElectronSaveRepository()
      : new BrowserSaveRepository()
    await SaveManager.initialize(repository)
    GameRuntime.getServices().register("save", SaveManager)
  }
}
