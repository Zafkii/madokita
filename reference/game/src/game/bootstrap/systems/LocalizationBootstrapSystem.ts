import type { PlatformAPI } from "../../platform/shared/PlatformAPI"
import type { RuntimeSystem } from "../../runtime/RuntimeSystem"
import { LocalizationManager } from "../../localization/LocalizationManager"
import { GameRuntime } from "../../runtime/GameRuntime"

export class LocalizationBootstrapSystem implements RuntimeSystem {
  readonly id = "localization"

  async initialize(_: PlatformAPI): Promise<void> {
    await LocalizationManager.load("en")
    GameRuntime.getServices().register("localization", LocalizationManager)
  }
}
