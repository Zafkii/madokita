import Phaser from "phaser"
import { gameConfig } from "../config/gameConfig"
import type { PlatformAPI } from "../platform/shared/PlatformAPI"
import { GameSettingsManager } from "../settings/GameSettingsManager"

declare global {
  interface Window {
    gamePlatform: PlatformAPI
  }
}

export async function createGame(platform: PlatformAPI): Promise<Phaser.Game> {
  console.log("Platform:", platform.getPlatformName())
  window.gamePlatform = platform

  const settings = GameSettingsManager.getData()
  const config: Phaser.Types.Core.GameConfig = {
    ...gameConfig,
    width: settings.resolution.width,
    height: settings.resolution.height,
    fps: {
      target: settings.fpsLimit,
    },
  }
  console.log(
    "[CreateGame] Phaser config:",
    config.width,
    "x",
    config.height,
    "| Scale mode:",
    config.scale?.mode,
  )

  let fsSaveTimer: ReturnType<typeof setTimeout> | null = null

  platform.setupFullscreenListener?.((fs) => {
    GameSettingsManager.applyFullscreen(fs)

    if (fsSaveTimer) clearTimeout(fsSaveTimer)
    fsSaveTimer = setTimeout(() => {
      GameSettingsManager.save().catch((e) =>
        console.error("Failed to persist fullscreen setting:", e),
      )
    }, 400)
  })

  const game = new Phaser.Game(config)
  // PREVENT AUTO PAUSE ON WINDOW BLUR
  game.events.off(Phaser.Core.Events.BLUR)
  // ALT + ENTER FULLSCREEN
  window.addEventListener("keydown", async (event) => {
    if (event.altKey && event.key === "Enter") {
      event.preventDefault()
      await platform.toggleFullscreen()
    }
  })

  return game
}
