export type ColorPalette = {
  lightPink: { enabled: boolean; rgb: string; alpha: number }
  red: { enabled: boolean; rgb: string; alpha: number }
  lightBlue: { enabled: boolean; rgb: string; alpha: number }
  yellow: { enabled: boolean; rgb: string; alpha: number }
  purple: { enabled: boolean; rgb: string; alpha: number }
}

export type EffectTiming = {
  enabled: boolean
  start: number
  end: "loop" | number
  fadeOutDuration: number
  initialColor?: { rgb: string; alpha: number }
  colors?: ColorPalette
  cycleInterval?: number
  colorTransition?: number
}

export const menuTimings = {
  trackDuration: 173,
  effects: {
    titleFill: {
      enabled: true,
      start: 40,
      end: "loop" as const,
      fadeOutDuration: 8,
    },
    borderTint: {
      enabled: true,
      start: 40,
      end: "loop" as const,
      fadeOutDuration: 8,
      initialColor: { rgb: "rgb(255,255,255)", alpha: 0.7 },
      cycleInterval: 10,
      colorTransition: 2,
      colors: {
        lightPink: { enabled: true, rgb: "rgb(253, 128, 255)", alpha: 0.7 },
        red: { enabled: true, rgb: "rgb(255, 8, 8)", alpha: 0.7 },
        yellow: { enabled: true, rgb: "rgb(255, 238, 0)", alpha: 0.8 },
        lightBlue: { enabled: true, rgb: "rgb(0, 217, 255)", alpha: 0.7 },
        purple: { enabled: true, rgb: "rgb(198, 83, 255)", alpha: 0.7 },
      },
    },
  } as Record<string, EffectTiming>,
}
