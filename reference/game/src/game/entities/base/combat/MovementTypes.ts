export type HurtboxData = [
  width: number,
  height: number,
  offsetX: number,
  offsetY: number,
  scaleX: number,
  scaleY: number,
  rotation: number,
  damageMultiplier: number,
]

export type Frame = {
  spriteFrames: number[]
  offsetX: number[]
  offsetY: number[]
  rotation: number[]
  scaleX?: number[]
  scaleY?: number[]
  hurtboxes?: HurtboxData[]
}

export type MovementAnimDef = {
  frames: Frame[]
  fps?: number
  loop?: boolean
}

export type Movement = {
  assetKey: string
  defaultOriginX?: number
  defaultOriginY?: number
  defaultHurtboxes?: HurtboxData[]
  animations: Record<string, MovementAnimDef>
}
