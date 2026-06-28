import { createGame } from "./game/bootstrap/createGame"
import { BrowserPlatform } from "./game/platform/browser/BrowserPlatform"
import { ElectronPlatform } from "./game/platform/electron/ElectronPlatform"
import { GameRuntime } from "./game/runtime/GameRuntime"
import { GameSettingsManager } from "./game/settings/GameSettingsManager"
import { SaveManager } from "./game/save/SaveManager"

function isElectronEnvironment(): boolean {
  return navigator.userAgent.toLowerCase().includes("electron")
}

async function bootstrap(): Promise<void> {
  const platform = isElectronEnvironment()
    ? new ElectronPlatform()
    : new BrowserPlatform()

  await GameRuntime.initialize(platform)
  console.log(
    "[Main] Resolution from settings:",
    GameSettingsManager.getData().resolution,
  )

  if (platform instanceof ElectronPlatform) {
    platform.setupResolutionListener()
  }

  const game = await createGame(platform)

  if (platform instanceof ElectronPlatform) {
    window.electronAPI.onResolutionChanged(() => {
      const s = GameSettingsManager.getData()
      game.scale.resize(s.resolution.width, s.resolution.height)
      window.dispatchEvent(new Event("resize"))
    })
  }

  window.addEventListener("beforeunload", () => {
    SaveManager.save()
    GameRuntime.shutdown(platform)
  })
}

bootstrap()
