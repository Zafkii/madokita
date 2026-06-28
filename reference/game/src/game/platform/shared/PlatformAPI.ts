export interface PlatformAPI {
  isElectron(): boolean
  getPlatformName(): string
  closeGame(): void
  setResolution(width: number, height: number): Promise<void>
  toggleFullscreen(): Promise<void>
  setupResolutionListener?(): void
  setupFullscreenListener?(callback: (fs: boolean) => void): void
}
