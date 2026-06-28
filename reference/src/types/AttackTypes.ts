import type { EditorHurtboxData } from "./MovementTypes"

export type AttackPhase = "wu" | "atk" | "rc" | "guard"

export type EditorAttackFrame = {
  spriteFrames: number[]
  offsetX: number[]
  offsetY: number[]
  rotation: number[]
  scaleX?: number[]
  scaleY?: number[]
  phase?: AttackPhase
  hurtboxes?: EditorHurtboxData[]
}

export type AttackAnimDef = {
  frames: EditorAttackFrame[]
  fps?: number
  loop?: boolean
  windup?: number
  activeTime?: number
  recover?: number
  guard?: number
  guardFps?: number
}

export type Attack = {
  assetKey: string
  defaultOriginX?: number
  defaultOriginY?: number
  defaultHurtboxes?: EditorHurtboxData[]
  animations: Record<string, AttackAnimDef>
}
