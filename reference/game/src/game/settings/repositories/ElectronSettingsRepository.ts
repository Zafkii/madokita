import type { GameSettings } from "../GameSettings"
import type { SettingsRepository } from "../SettingsRepository"
import { validateSettings } from "../validators/validateSettings"

export class ElectronSettingsRepository implements SettingsRepository {
  async load(): Promise<GameSettings> {
    const rawSettings = await window.electronAPI.getAllSettings()

    const parsed: Partial<GameSettings> = {
      resolution: {
        width: Number(rawSettings.resolution_width),
        height: Number(rawSettings.resolution_height),
      },

      fullscreen: rawSettings.fullscreen === "true",

      language: rawSettings.language,

      fpsLimit: Number(rawSettings.fps_limit),

      sensitivity: rawSettings.sensitivity
        ? Number(rawSettings.sensitivity)
        : undefined,

      keyBindings: rawSettings.key_bindings
        ? JSON.parse(rawSettings.key_bindings)
        : undefined,
    }

    return validateSettings(parsed)
  }

  async save(settings: GameSettings): Promise<void> {
    await window.electronAPI.saveSetting(
      "resolution_width",
      settings.resolution.width.toString(),
    )

    await window.electronAPI.saveSetting(
      "resolution_height",
      settings.resolution.height.toString(),
    )

    await window.electronAPI.saveSetting(
      "fullscreen",
      settings.fullscreen.toString(),
    )

    await window.electronAPI.saveSetting("language", settings.language)

    await window.electronAPI.saveSetting(
      "fps_limit",
      settings.fpsLimit.toString(),
    )

    await window.electronAPI.saveSetting(
      "key_bindings",
      JSON.stringify(settings.keyBindings),
    )

    await window.electronAPI.saveSetting(
      "sensitivity",
      settings.sensitivity.toString(),
    )
  }
}
