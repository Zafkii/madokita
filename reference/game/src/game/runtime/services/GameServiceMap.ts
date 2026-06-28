import type { SaveManager } from "../../save/SaveManager"
import type { LocalizationManager } from "../../localization/LocalizationManager"
import type { GameSettingsManager } from "../../settings/GameSettingsManager"
import type { GameEventBus } from "../../events/GameEventBus"

export interface GameServiceMap {
  save: typeof SaveManager
  localization: typeof LocalizationManager
  settings: typeof GameSettingsManager
  events: GameEventBus
}
