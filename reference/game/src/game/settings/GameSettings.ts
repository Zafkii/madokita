export interface GameSettings {
  resolution: {
    width: number
    height: number
  }

  fullscreen: boolean

  language: string

  fpsLimit: number

  sensitivity: number

  keyBindings: Record<string, string>
}
