import type { PlatformAPI } from "../shared/PlatformAPI"

export class BrowserPlatform implements PlatformAPI {
  isElectron(): boolean {
    return false
  }

  getPlatformName(): string {
    return "browser"
  }

  closeGame(): void {
    console.log("closeGame not available in browser")
  }

  async setResolution(width: number, height: number): Promise<void> {
    console.log(`Resolution ignored in browser: ${width}x${height}`)
  }

  async toggleFullscreen(): Promise<void> {
    console.log("fullscreen not available in browser")
  }

  setupFullscreenListener(_callback: (fs: boolean) => void): void {
    // no-op
  }
}
