export const DepthLayers = {
  SKY: 0,
  BG_FAR: 1,
  BG_MID: 2,
  BG_NEAR: 3,
  STRUCTURES_FAR: 4,
  STRUCTURES_NEAR: 5,
  PLATFORMS: 6,
  ENEMIES: 7,
  PLAYERS_SECONDARY: 8,
  PLAYERS_MAIN: 9,
  EFFECTS_BG: 10,
  EFFECTS_FG: 11,
  UI_BG: 12,
  UI_FG: 13,
} as const

export type DepthLayer = (typeof DepthLayers)[keyof typeof DepthLayers]
