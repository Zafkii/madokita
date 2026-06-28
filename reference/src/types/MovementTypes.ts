export type EditorHurtboxData = [
  width: number,
  height: number,
  offsetX: number,
  offsetY: number,
  scaleX: number,
  scaleY: number,
  rotation: number,
  damageMultiplier: number,
]

export const DEFAULT_HURTBOX_SIZE = 100

export const DEFAULT_EDITOR_HURTBOX: EditorHurtboxData = [
  DEFAULT_HURTBOX_SIZE,
  DEFAULT_HURTBOX_SIZE,
  0,
  0,
  1,
  1,
  0,
  1,
]

export type EditorFrame = {
  spriteFrames: number[]
  offsetX: number[]
  offsetY: number[]
  rotation: number[]
  scaleX?: number[]
  scaleY?: number[]
  hurtboxes?: EditorHurtboxData[]
}

export type MovementAnimDef = {
  frames: EditorFrame[]
  fps?: number
  loop?: boolean
}

export type Movement = {
  assetKey: string
  defaultOriginX?: number
  defaultOriginY?: number
  defaultHurtboxes?: EditorHurtboxData[]
  animations: Record<string, MovementAnimDef>
}
