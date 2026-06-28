export type AttackPhase = "wu" | "atk" | "rc" | "guard"

import type { HurtboxData } from "./MovementTypes"

export type AttackFrame = {
  spriteFrames: number[]
  offsetX: number[]
  offsetY: number[]
  rotation: number[]
  scaleX?: number[]
  scaleY?: number[]
  phase?: AttackPhase
  hurtboxes?: HurtboxData[]
}

export type AttackAnimDef = {
  frames: AttackFrame[]
  fps?: number
  loop?: boolean
  guard?: number
  guardFps?: number
  windupFrames?: number
  activeFrames?: number
  recoverFrames?: number
}

export type Attack = {
  assetKey: string
  defaultOriginX?: number
  defaultOriginY?: number
  defaultHurtboxes?: HurtboxData[]
  animations: Record<string, AttackAnimDef>
}
