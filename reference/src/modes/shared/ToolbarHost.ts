import type { EditorHurtboxData } from "../../types/MovementTypes"

export interface ToolbarHost {
  baseSpritesheet: HTMLImageElement | null
  baseSpritePath: string
  baseSpriteFrameW: number
  baseSpriteFrameH: number
  baseSpriteTotalFrames: number
  baseSpriteFrameIndex: number
  additionalSprites: {
    image: HTMLImageElement | null
    path: string
    frameW: number
    frameH: number
    totalFrames: number
    frameIdx: number
    originX: number
    originY: number
  }[]
  boundaryW: number
  boundaryH: number
  currentAnimName: string
  currentFrameIndex: number
  currentSpriteIndex: number
  selectedHurtboxIndex: number

  loadBaseSpritesheet(img: HTMLImageElement): void
  addAdditionalSprite(img: HTMLImageElement, index: number): void
  calcBaseSpriteTotalFrames(): void
  calcSpriteTotalFrames(index: number): void
  updateBaseSpriteFrame(): void
  updateSpriteFrame(index: number): void
  
  getAnimations(): Record<string, any>
  getCurrentAnim(): any | null
  
  ensureAnimExists(name: string): void
  selectAnimation(name: string): void
  selectFrame(index: number): void
  selectSprite(index: number): void
  selectHurtbox(index: number): void
  renameAnimation(oldName: string, newName: string): boolean
  removeAnimation(name: string): void
  setBaseSpriteFrame(index: number): void

  // Animation list
  getAllAnimationNames(): string[]
  getAnimFrameTotal(name: string): number

  // Per-anim frame ops
  addFrameToAnim(name: string): void
  removeFrameFromAnim(name: string): void

  // Hurtbox ops (current frame)
  getCurrentFrameHurtboxes(): EditorHurtboxData[]
  addHurtbox(): void
  removeHurtbox(index: number): void
  updateHurtbox(index: number, w: number, h: number, dmgMult: number): void
  repeatPreviousHurtbox(): void
}
