import type { GameSettings } from "./GameSettings"
import type { SettingsRepository } from "./SettingsRepository"

export class GameSettingsManager {
  private static repository: SettingsRepository

  private static data: GameSettings

  static async initialize(repository: SettingsRepository): Promise<void> {
    this.repository = repository

    this.data = await repository.load()
  }

  static getData(): Readonly<GameSettings> {
    return this.data
  }

  static get<K extends keyof GameSettings>(key: K): Readonly<GameSettings[K]> {
    return this.data[key]
  }

  static async save(): Promise<void> {
    await this.repository.save(this.data)
  }

  static async setResolution(width: number, height: number): Promise<void> {
    this.data.resolution.width = width
    this.data.resolution.height = height
    await this.save()
  }

  static applyFullscreen(enabled: boolean): void {
    this.data.fullscreen = enabled
  }

  static async setFullscreen(enabled: boolean): Promise<void> {
    this.applyFullscreen(enabled)
    await this.save()
  }

  static async setLanguage(language: string): Promise<void> {
    this.data.language = language
    await this.save()
  }

  static async setFpsLimit(fps: number): Promise<void> {
    this.data.fpsLimit = fps
    await this.save()
  }

  static async setKeyBinding(action: string, key: string): Promise<void> {
    this.data.keyBindings[action] = key
    await this.save()
  }

  static async setSensitivity(value: number): Promise<void> {
    this.data.sensitivity = value
    await this.save()
  }
}
