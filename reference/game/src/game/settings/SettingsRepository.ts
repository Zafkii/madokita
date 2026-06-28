import type { GameSettings } from "./GameSettings"

export interface SettingsRepository {
  load(): Promise<GameSettings>
  save(settings: GameSettings): Promise<void>
}
