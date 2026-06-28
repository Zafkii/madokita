import type { PlatformAPI } from "../shared/PlatformAPI"

declare global {
  interface Window {
    electronAPI: {
      setResolution(width: number, height: number): Promise<void>
      toggleFullscreen(): Promise<void>
      onResolutionChanged(callback: () => void): void
      onFullscreenChanged(callback: (fs: boolean) => void): void
      // SETTINGS
      saveSetting(key: string, value: string): Promise<void>
      getSetting(key: string): Promise<string | null>
      getAllSettings(): Promise<Record<string, string>>
      // SAVE
      saveGame(data: string): Promise<void>
      loadSave(): Promise<string | null>
    }
  }
}

export class ElectronPlatform implements PlatformAPI {
  isElectron(): boolean {
    return true
  }

  getPlatformName(): string {
    return "electron"
  }

  closeGame(): void {
    window.close()
  }

  async setResolution(width: number, height: number): Promise<void> {
    await window.electronAPI.setResolution(width, height)
  }

  async toggleFullscreen(): Promise<void> {
    await window.electronAPI.toggleFullscreen()
  }

  setupResolutionListener(): void {
    window.electronAPI.onResolutionChanged(() => {
      window.dispatchEvent(new Event("resize"))
    })
  }

  setupFullscreenListener(callback: (fs: boolean) => void): void {
    window.electronAPI.onFullscreenChanged((fs) => {
      callback(fs)
    })
  }
}
