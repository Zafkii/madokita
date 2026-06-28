import type { GameSettings } from "../GameSettings"
import { defaultSettings } from "../defaultSettings"

export function validateSettings(
  settings: Partial<GameSettings>,
): GameSettings {
  return {
    resolution: {
      width:
        typeof settings.resolution?.width === "number" && Number.isFinite(settings.resolution.width)
          ? settings.resolution.width
          : defaultSettings.resolution.width,

      height:
        typeof settings.resolution?.height === "number" && Number.isFinite(settings.resolution.height)
          ? settings.resolution.height
          : defaultSettings.resolution.height,
    },

    fullscreen:
      typeof settings.fullscreen === "boolean"
        ? settings.fullscreen
        : defaultSettings.fullscreen,

    language:
      typeof settings.language === "string"
        ? settings.language
        : defaultSettings.language,

    fpsLimit:
      typeof settings.fpsLimit === "number" && Number.isFinite(settings.fpsLimit)
        ? settings.fpsLimit
        : defaultSettings.fpsLimit,

    sensitivity:
      typeof settings.sensitivity === "number" && Number.isFinite(settings.sensitivity)
        ? settings.sensitivity
        : defaultSettings.sensitivity,

    keyBindings: {
      ...defaultSettings.keyBindings,
      ...(typeof settings.keyBindings === "object" && settings.keyBindings !== null
        ? settings.keyBindings
        : {}),
    },
  }
}
