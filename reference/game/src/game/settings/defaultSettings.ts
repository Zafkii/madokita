import type { GameSettings } from "./GameSettings"
import { DEFAULT_INPUT_BINDINGS } from "../input/InputBindings"

export const defaultSettings: GameSettings = {
  resolution: {
    width: 1280,
    height: 720,
  },

  fullscreen: false,

  language: "en",

  fpsLimit: 60,

  sensitivity: 50,

  keyBindings: { ...DEFAULT_INPUT_BINDINGS },
}
