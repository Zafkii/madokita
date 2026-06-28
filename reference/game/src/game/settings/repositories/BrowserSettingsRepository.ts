import type { GameSettings } from "../GameSettings"
import type { SettingsRepository } from "../SettingsRepository"
import { defaultSettings } from "../defaultSettings"

export class BrowserSettingsRepository implements SettingsRepository {
  async load(): Promise<GameSettings> {
    return structuredClone(defaultSettings)
  }

  async save(_settings: GameSettings): Promise<void> {}
}
