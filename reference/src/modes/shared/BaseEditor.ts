import type { TransformState, OverlayData } from "../../ui/CanvasView"
import {
  DEFAULT_EDITOR_HURTBOX,
  type EditorHurtboxData,
} from "../../types/MovementTypes"
import { CanvasView } from "../../ui/CanvasView"
import { PropertyInspector } from "../../ui/PropertyInspector"
import { loadImageFromUrl } from "../../ui/FileLoader"
import { extractFrame } from "../frameUtils"
import { AnimationTableGrid } from "./AnimationTableGrid"
import type { ToolbarHost } from "./ToolbarHost"
import { pathToUrl, downloadFile, extractFilename } from "./importExportUtils"

function defaultCount(arr: number[] | undefined, len: number): number[] {
  if (!arr) return Array(len).fill(0)
  while (arr.length < len) arr.push(0)
  return arr
}

interface BasePreview {
  setAnimation(anim: any, cb: (frame: any, idx: number) => void): void
  stop(): void
  destroy(): void
}

export abstract class BaseEditor implements ToolbarHost {
  protected canvasView: CanvasView
  protected propertyInspector: PropertyInspector
  protected tablesContainer: HTMLElement
  tableGrid: AnimationTableGrid

  currentAnimName = ""
  currentFrameIndex = 0
  currentSpriteIndex = 0
  selectedHurtboxIndex = -1

  baseSpritesheet: HTMLImageElement | null = null
  baseSpritePath = ""
  baseSpriteFrameIndex = 0
  baseSpriteFrameW = 256
  baseSpriteFrameH = 256
  baseSpriteTotalFrames = 1

  additionalSprites: ToolbarHost["additionalSprites"] = []

  get boundaryW(): number {
    return this.canvasView.boundaryW
  }
  get boundaryH(): number {
    return this.canvasView.boundaryH
  }

  protected preview: BasePreview
  protected isPreviewing = false

  onSpriteChange: (() => void) | null = null

  // ── Abstract methods (editor-specific) ──

  protected abstract getCurrentDef(): {
    defaultOriginX?: number
    defaultOriginY?: number
    animations: Record<string, any>
  }
  protected abstract getEntries(): { name: string }[]
  abstract getCurrentAnim(): any | null
  abstract getAnimations(): Record<string, any>
  protected abstract syncEditorUI(): void
  abstract ensureAnimExists(name: string): void
  protected abstract createDefaultAnimObj(name: string): any
  protected abstract refreshUI(): void
  protected abstract getRemoveAnimFallbackName(): string
  protected abstract getExtraFrameProps(last: any): Record<string, any>
  abstract importFromText(text: string): boolean
  abstract getExportContent(): string

  // ── Constructor ──

  constructor(
    canvasView: CanvasView,
    propertyInspector: PropertyInspector,
    tablesContainer: HTMLElement,
    rightPanel: HTMLElement,
  ) {
    this.canvasView = canvasView
    this.propertyInspector = propertyInspector
    this.tablesContainer = tablesContainer
    this.tableGrid = new AnimationTableGrid(this, this.tablesContainer)
    this.preview = this.createPreview(rightPanel)
  }

  protected abstract createPreview(container: HTMLElement): BasePreview

  // ── Activation ──

  activate(
    tablesContainer: HTMLElement,
    rightPanel: HTMLElement,
    propertyInspector?: PropertyInspector,
  ): void {
    this.tablesContainer = tablesContainer
    this.tablesContainer.innerHTML = ""
    if (propertyInspector) this.propertyInspector = propertyInspector
    this.tableGrid = new AnimationTableGrid(this, this.tablesContainer)
    this.tableGrid.init()
    this.selectAnimation(this.currentAnimName)

    this.preview.destroy()
    this.preview = this.createPreview(rightPanel)
    this.preview.setAnimation(this.getCurrentAnim(), (frame: any, idx: number) =>
      this.onPreviewFrame(frame, idx),
    )
    this.refreshUI()
  }

  // ── Animation management ──

  selectAnimation(name: string): void {
    this.currentAnimName = name
    this.currentFrameIndex = 0
    this.selectedHurtboxIndex = -1
    this.currentSpriteIndex = 0
    this.isPreviewing = false
    this.preview.stop()
    this.canvasView.clearUndo()
    this.tableGrid.syncAnimName(this.currentAnimName)
    const total = this.getCurrentAnim()?.frames.length ?? 0
    this.tableGrid.updateFrameDisplay(0, total)
    this.loadCurrentFrame()
    this.syncEditorUI()
    this.preview.setAnimation(this.getCurrentAnim(), (frame: any, idx: number) =>
      this.onPreviewFrame(frame, idx),
    )
  }

  removeAnimation(name: string): void {
    const cur = this.getCurrentDef()
    delete cur.animations[name]
    const keys = Object.keys(cur.animations)
    if (keys.length === 0) {
      cur.animations[this.getRemoveAnimFallbackName()] =
        this.createDefaultAnimObj(this.getRemoveAnimFallbackName())
      this.currentAnimName = this.getRemoveAnimFallbackName()
    } else {
      this.currentAnimName = keys[0]
    }
    this.selectAnimation(this.currentAnimName)
  }

  renameAnimation(oldName: string, newName: string): boolean {
    const cur = this.getCurrentDef()
    if (!newName || newName === oldName || !!cur.animations[newName]) {
      return false
    }
    const anim = cur.animations[oldName]
    delete cur.animations[oldName]
    cur.animations[newName] = anim
    this.currentAnimName = newName
    this.selectedHurtboxIndex = -1
    this.tableGrid.syncAnimName(this.currentAnimName)
    this.loadCurrentFrame()
    this.syncEditorUI()
    return true
  }

  // ── Sprite count ──

  protected spriteCount(): number {
    return 1 + this.additionalSprites.length
  }

  protected ensureFrameArrays(frame: {
    spriteFrames: number[]
    offsetX: number[]
    offsetY: number[]
    rotation: number[]
    scaleX?: number[]
    scaleY?: number[]
  }): void {
    const n = this.spriteCount()
    frame.spriteFrames = defaultCount(frame.spriteFrames, n)
    frame.offsetX = defaultCount(frame.offsetX, n)
    frame.offsetY = defaultCount(frame.offsetY, n)
    frame.rotation = defaultCount(frame.rotation, n)
    if (!frame.scaleX) frame.scaleX = Array(n).fill(1)
    while (frame.scaleX.length < n) frame.scaleX.push(1)
    if (!frame.scaleY) frame.scaleY = Array(n).fill(1)
    while (frame.scaleY.length < n) frame.scaleY.push(1)
  }

  // ── Frame loading ──

  protected makeOverlayData(): OverlayData[] {
    const result: OverlayData[] = []
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    const frame = anim?.frames[this.currentFrameIndex] ?? null
    if (frame) this.ensureFrameArrays(frame)

    if (this.baseSpritesheet) {
      result.push({
        image: this.baseSpritesheet,
        frameW: this.baseSpriteFrameW,
        frameH: this.baseSpriteFrameH,
        frameIndex: frame
          ? (frame.spriteFrames[0] ?? this.baseSpriteFrameIndex)
          : this.baseSpriteFrameIndex,
        transform: {
          offsetX: frame ? (frame.offsetX[0] ?? 0) : 0,
          offsetY: frame ? (frame.offsetY[0] ?? 0) : 0,
          rotation: frame ? (frame.rotation[0] ?? 0) : 0,
          scaleX: frame ? (frame.scaleX?.[0] ?? 1) : 1,
          scaleY: frame ? (frame.scaleY?.[0] ?? 1) : 1,
        },
        originX: this.getCurrentDef().defaultOriginX ?? 0.5,
        originY: this.getCurrentDef().defaultOriginY ?? 0.5,
      })
    }

    for (let i = 1; i < this.spriteCount(); i++) {
      const sp = this.additionalSprites[i - 1]
      if (!sp.image) continue
      result.push({
        image: sp.image,
        frameW: sp.frameW,
        frameH: sp.frameH,
        frameIndex: frame ? (frame.spriteFrames[i] ?? 0) : sp.frameIdx,
        transform: {
          offsetX: frame ? (frame.offsetX[i] ?? 0) : 0,
          offsetY: frame ? (frame.offsetY[i] ?? 0) : 0,
          rotation: frame ? (frame.rotation[i] ?? 0) : 0,
          scaleX: frame ? (frame.scaleX?.[i] ?? 1) : 1,
          scaleY: frame ? (frame.scaleY?.[i] ?? 1) : 1,
        },
        originX: sp.originX,
        originY: sp.originY,
      })
    }

    return result
  }

  protected loadCurrentFrame(): void {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    const n = this.spriteCount()

    this.canvasView.setOverlays(
      this.makeOverlayData(),
      this.getActiveOverlayIndex(),
    )

    if (!anim || anim.frames.length === 0) {
      const t: TransformState = { offsetX: 0, offsetY: 0, rotation: 0, scaleX: 1, scaleY: 1 }
      this.canvasView.setTransform(t)
      this.propertyInspector.setTransform(t)
      const def = this.getCurrentDef()
      this.canvasView.setOrigin(def.defaultOriginX ?? 0.5, def.defaultOriginY ?? 0.5)
      this.propertyInspector.setOrigin(def.defaultOriginX ?? 0.5, def.defaultOriginY ?? 0.5)
      this.propertyInspector.setFrameIndex(0, 0)
      this.canvasView.setHurtboxes([], -1)
      this.propertyInspector.setHurtboxData(null)
      this.propertyInspector.setActiveElement("sprite")
      this.tableGrid.refreshHurtboxes()
      this.syncEditorUI()
      this.updateBaseRestPose()
      return
    }

    const frame = anim.frames[this.currentFrameIndex]
    this.ensureFrameArrays(frame)

    this.baseSpriteFrameIndex = Math.min(
      this.baseSpriteTotalFrames - 1,
      frame.spriteFrames[0] ?? this.currentFrameIndex,
    )
    this.updateBaseSpriteFrame()

    for (let i = 0; i < this.additionalSprites.length; i++) {
      const sp = this.additionalSprites[i]
      if (sp.totalFrames <= 0) continue
      const sf = frame.spriteFrames[i + 1] ?? 0
      sp.frameIdx = Math.min(sp.totalFrames - 1, sf)
      this.tableGrid.syncSpriteFrameIdx(i, sp.frameIdx, sp.totalFrames)
    }

    const idx = this.currentSpriteIndex
    if (idx < n) {
      const t: TransformState = {
        offsetX: frame.offsetX[idx] ?? 0,
        offsetY: frame.offsetY[idx] ?? 0,
        rotation: frame.rotation[idx] ?? 0,
        scaleX: frame.scaleX?.[idx] ?? 1,
        scaleY: frame.scaleY?.[idx] ?? 1,
      }
      this.canvasView.setTransform(t)
      this.propertyInspector.setTransform(t)

      if (idx > 0) {
        const sp = this.additionalSprites[idx - 1]
        if (sp) {
          this.canvasView.setOrigin(sp.originX, sp.originY)
          this.propertyInspector.setOrigin(sp.originX, sp.originY)
        }
      } else {
        const def = this.getCurrentDef()
        const ox = def.defaultOriginX ?? 0.5
        const oy = def.defaultOriginY ?? 0.5
        this.canvasView.setOrigin(ox, oy)
        this.propertyInspector.setOrigin(ox, oy)
      }
    }
    this.propertyInspector.setFrameIndex(this.currentFrameIndex, anim.frames.length)
    this.syncEditorUI()

    const hb = frame.hurtboxes ?? []
    for (const h of hb) {
      h[4] = Math.round(Math.max(0.01, h[4]) * 100) / 100
      h[5] = Math.round(Math.max(0.01, h[5]) * 100) / 100
    }
    this.canvasView.setHurtboxes(hb, this.selectedHurtboxIndex)
    const selHb =
      this.selectedHurtboxIndex >= 0 && this.selectedHurtboxIndex < hb.length
        ? hb[this.selectedHurtboxIndex]
        : null
    this.propertyInspector.setHurtboxData(selHb)
    this.propertyInspector.setActiveElement(selHb ? "hurtbox" : "sprite")
    this.tableGrid.refreshHurtboxes()

    this.updateBaseRestPose()
  }

  // ── Preview callback ──

  protected onPreviewFrame(frame: any, idx: number): void {
    this.isPreviewing = true
    this.ensureFrameArrays(frame)
    const n = this.spriteCount()
    const activeIdx = this.currentSpriteIndex

    if (activeIdx < n) {
      const t: TransformState = {
        offsetX: frame.offsetX[activeIdx] ?? 0,
        offsetY: frame.offsetY[activeIdx] ?? 0,
        rotation: frame.rotation[activeIdx] ?? 0,
        scaleX: frame.scaleX?.[activeIdx] ?? 1,
        scaleY: frame.scaleY?.[activeIdx] ?? 1,
      }
      this.canvasView.setTransform(t)
      this.propertyInspector.setTransform(t)
    }

    this.currentFrameIndex = idx
    const total = this.getCurrentAnim()?.frames.length ?? 0
    this.tableGrid.updateFrameDisplay(idx, total)
    this.propertyInspector.setFrameIndex(idx, total)

    this.baseSpriteFrameIndex = Math.min(
      this.baseSpriteTotalFrames - 1,
      frame.spriteFrames[0] ?? idx,
    )
    this.updateBaseSpriteFrame()

    for (let i = 0; i < this.additionalSprites.length; i++) {
      const sp = this.additionalSprites[i]
      const sf = frame.spriteFrames[i + 1] ?? 0
      sp.frameIdx = Math.min(sp.totalFrames - 1, sf)
      this.tableGrid.syncSpriteFrameIdx(i, sp.frameIdx, sp.totalFrames)
    }

    this.canvasView.setOverlays(
      this.makeOverlayData(),
      this.getActiveOverlayIndex(),
    )
    this.onPreviewFrameExtra(frame)
  }

  protected onPreviewFrameExtra(_frame: any): void {
    // Subclasses can override for editor-specific preview updates
  }

  // ── Frame operations ──

  selectFrame(index: number): void {
    this.isPreviewing = false
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim || anim.frames.length === 0) return
    this.currentFrameIndex = Math.max(
      0,
      Math.min(anim.frames.length - 1, index),
    )
    this.tableGrid.updateFrameDisplay(this.currentFrameIndex, anim.frames.length)
    this.loadCurrentFrame()
  }

  // ── Character sprite ──

  updateBaseSpriteFrame(): void {
    if (!this.baseSpritesheet) return
    const canvas = extractFrame(
      this.baseSpritesheet,
      this.baseSpriteFrameW,
      this.baseSpriteFrameH,
      this.baseSpriteFrameIndex,
    )
    const img = new Image()
    img.onload = () => this.canvasView.setBackgroundImage(img)
    img.src = canvas.toDataURL()
    this.tableGrid.syncBaseSpriteFrameIdx(this.baseSpriteFrameIndex)
  }

  setBaseSpriteFrame(index: number): void {
    this.baseSpriteFrameIndex = index
    const anim = this.getCurrentAnim()
    if (anim && anim.frames.length > 0) {
      const frame = anim.frames[this.currentFrameIndex]
      this.ensureFrameArrays(frame)
      frame.spriteFrames[0] = index
    }
    this.loadCurrentFrame()
  }

  calcBaseSpriteTotalFrames(): void {
    if (!this.baseSpritesheet) return
    const cols = Math.floor(this.baseSpritesheet.width / this.baseSpriteFrameW)
    const rows = Math.floor(this.baseSpritesheet.height / this.baseSpriteFrameH)
    this.baseSpriteTotalFrames = cols * rows
    this.baseSpriteFrameIndex = Math.min(
      this.baseSpriteFrameIndex,
      this.baseSpriteTotalFrames - 1,
    )
    this.tableGrid.syncBaseSpriteTotalFrames(this.baseSpriteTotalFrames)
    this.updateBaseSpriteFrame()
    this.updateBaseRestPose()
  }

  loadBaseSpritesheet(img: HTMLImageElement): void {
    this.baseSpritesheet = img
    this.baseSpriteFrameIndex = 0
    this.calcBaseSpriteTotalFrames()
    this.updateBaseRestPose()
    const anim = this.getCurrentAnim()
    if (anim && anim.frames.length > 0) {
      this.loadCurrentFrame()
    }
    this.onSpriteChange?.()
  }

  // ── Additional sprites ──

  addAdditionalSprite(img: HTMLImageElement, index: number): void {
    const sp = this.additionalSprites[index]
    if (!sp) return
    sp.image = img
    sp.frameIdx = 0
    this.calcSpriteTotalFrames(index)
    this.currentSpriteIndex = index + 1
    this.loadCurrentFrame()
    this.onSpriteChange?.()
  }

  calcSpriteTotalFrames(_index: number): void {
    const sp = this.additionalSprites[_index]
    if (!sp || !sp.image) return
    const cols = Math.floor(sp.image.width / sp.frameW)
    const rows = Math.floor(sp.image.height / sp.frameH)
    sp.totalFrames = cols * rows
    this.loadCurrentFrame()
    this.tableGrid.rebuildSpriteGroups()
  }

  updateSpriteFrame(_index: number): void {
    const sp = this.additionalSprites[_index]
    if (!sp) return
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (anim && anim.frames.length > 0) {
      const frame = anim.frames[this.currentFrameIndex]
      this.ensureFrameArrays(frame)
      frame.spriteFrames[_index + 1] = sp.frameIdx
    }
    this.loadCurrentFrame()
  }

  // ── Sprite selection helpers ──

  protected autoSelectSprite(): void {
    if (
      !this.baseSpritesheet &&
      this.currentSpriteIndex === 0 &&
      this.additionalSprites.length > 0
    ) {
      const firstLoaded = this.additionalSprites.findIndex(
        (sp) => sp.image !== null,
      )
      if (firstLoaded >= 0) {
        this.currentSpriteIndex = firstLoaded + 1
      }
    }
    this.tableGrid.rebuildSpriteSelect()
  }

  selectSprite(index: number): void {
    this.currentSpriteIndex = Math.max(
      0,
      Math.min(index, this.spriteCount() - 1),
    )
    this.selectedHurtboxIndex = -1
    this.tableGrid.syncSpriteIndex(this.currentSpriteIndex)
    this.loadCurrentFrame()
  }

  selectHurtbox(index: number): void {
    const hurtboxes = this.getCurrentFrameHurtboxes()
    this.selectedHurtboxIndex =
      index >= 0 && index < hurtboxes.length ? index : -1
    this.tableGrid.syncHurtboxSelection(this.selectedHurtboxIndex)
    this.propertyInspector.setActiveElement(
      this.selectedHurtboxIndex >= 0 ? "hurtbox" : "sprite",
    )
    this.loadCurrentFrame()
  }

  // ── Transform / origin handlers ──

  onTransformChange(state: TransformState): void {
    if (this.isPreviewing) {
      this.isPreviewing = false
      this.preview.stop()
    }
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim) return
    if (anim.frames.length === 0) {
      this.propertyInspector.setTransform(state)
      return
    }

    const frame = anim.frames[this.currentFrameIndex]
    this.ensureFrameArrays(frame)
    const idx = this.currentSpriteIndex

    if (idx < this.spriteCount()) {
      frame.offsetX[idx] = state.offsetX
      frame.offsetY[idx] = state.offsetY
      frame.rotation[idx] = state.rotation
      frame.scaleX![idx] = Math.round(state.scaleX * 100) / 100
      frame.scaleY![idx] = Math.round(state.scaleY * 100) / 100
    }

    this.canvasView.setTransform(state)
    this.propertyInspector.setTransform(state)

    this.canvasView.setOverlays(
      this.makeOverlayData(),
      this.getActiveOverlayIndex(),
    )
  }

  onOriginChange(x: number, y: number): void {
    if (this.currentSpriteIndex > 0) {
      const idx = this.currentSpriteIndex - 1
      const sp = this.additionalSprites[idx]
      if (sp) {
        sp.originX = x
        sp.originY = y
      }
    } else {
      this.getCurrentDef().defaultOriginX = x
      this.getCurrentDef().defaultOriginY = y
      this.updateBaseRestPose()
    }
    this.canvasView.setOrigin(x, y)
    this.propertyInspector.setOrigin(x, y)
  }

  onBaseRotationChange(deg: number): void {
    this.canvasView.setBaseRotation(deg)
    this.propertyInspector.setBaseRotation(deg)
  }

  onHurtboxChange(hb: EditorHurtboxData): void {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim) return
    const frame = anim.frames?.[this.currentFrameIndex]
    if (!frame?.hurtboxes) return
    const idx = this.selectedHurtboxIndex
    if (idx < 0 || idx >= frame.hurtboxes.length) return
    frame.hurtboxes[idx] = hb
    this.canvasView.setHurtboxes(frame.hurtboxes, idx)
    this.tableGrid.refreshHurtboxes()
    this.propertyInspector.setHurtboxData(hb)
    this.canvasView.requestRender()
  }

  // ── Active element ──

  protected getActiveOverlayIndex(): number {
    return this.baseSpritesheet
      ? this.currentSpriteIndex
      : Math.max(0, this.currentSpriteIndex - 1)
  }

  // ── Rest pose ──

  protected updateBaseRestPose(): void {
    if (
      this.baseSpritesheet &&
      this.baseSpriteFrameW > 0 &&
      this.baseSpriteFrameH > 0
    ) {
      const def = this.getCurrentDef()
      this.canvasView.setBaseRestPose(
        this.baseSpriteFrameW,
        this.baseSpriteFrameH,
        def.defaultOriginX ?? 0.5,
        def.defaultOriginY ?? 0.5,
      )
    } else {
      this.canvasView.clearBaseRestPose()
    }
  }

  stopPreview(): void {
    this.preview.stop()
  }

  // ── Animation list helpers ──

  getAllAnimationNames(): string[] {
    return Object.keys(this.getCurrentDef().animations)
  }

  getAnimFrameTotal(name: string): number {
    const anim = this.getCurrentDef().animations[name]
    if (!anim) return 0
    return anim.frames?.length ?? 0
  }

  // ── Per-anim frame ops ──

  addFrameToAnim(name: string): void {
    const anim = this.getCurrentDef().animations[name]
    if (!anim) return
    if (!anim.frames) anim.frames = []
    const last =
      anim.frames.length > 0 ? anim.frames[anim.frames.length - 1] : undefined
    const n = this.spriteCount()
    anim.frames.push({
      spriteFrames: last ? [...last.spriteFrames] : Array(n).fill(0),
      offsetX: last ? [...last.offsetX] : Array(n).fill(0),
      offsetY: last ? [...last.offsetY] : Array(n).fill(0),
      rotation: last ? [...last.rotation] : Array(n).fill(0),
      scaleX: last?.scaleX ? [...last.scaleX] : Array(n).fill(1),
      scaleY: last?.scaleY ? [...last.scaleY] : Array(n).fill(1),
      ...this.getExtraFrameProps(last),
      hurtboxes: last?.hurtboxes
        ? last.hurtboxes.map((h: EditorHurtboxData) => [...h] as EditorHurtboxData)
        : undefined,
    })
    this.selectFrame(anim.frames.length - 1)
  }

  removeFrameFromAnim(name: string): void {
    const anim = this.getCurrentDef().animations[name]
    if (!anim) return
    if (anim.frames && anim.frames.length > 0) {
      anim.frames.splice(this.currentFrameIndex, 1)
      if (
        this.currentFrameIndex >= anim.frames.length &&
        anim.frames.length > 0
      ) {
        this.currentFrameIndex = anim.frames.length - 1
      } else if (anim.frames.length === 0) {
        this.currentFrameIndex = 0
      }
      this.loadCurrentFrame()
      this.tableGrid.updateFrameDisplay(this.currentFrameIndex, anim.frames.length)
    }
  }

  // ── Hurtbox ops ──

  getCurrentFrameHurtboxes(): EditorHurtboxData[] {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim) return []
    const frame = anim.frames?.[this.currentFrameIndex]
    if (!frame) return []
    return frame.hurtboxes ?? []
  }

  addHurtbox(): void {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim || !anim.frames?.length) return
    const cur = anim.frames[this.currentFrameIndex]
    if (!cur) return
    let newIndex = -1
    for (const frame of anim.frames) {
      if (!frame.hurtboxes) frame.hurtboxes = []
      frame.hurtboxes.push([...DEFAULT_EDITOR_HURTBOX])
      if (frame === cur) {
        newIndex = frame.hurtboxes.length - 1
      }
    }
    this.selectedHurtboxIndex = newIndex
    this.tableGrid.rebuild()
    this.loadCurrentFrame()
  }

  removeHurtbox(index: number): void {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim || !anim.frames?.length) return
    for (const frame of anim.frames) {
      if (!frame?.hurtboxes) continue
      frame.hurtboxes.splice(index, 1)
    }
    if (this.selectedHurtboxIndex === index) {
      this.selectedHurtboxIndex = -1
    } else if (this.selectedHurtboxIndex > index) {
      this.selectedHurtboxIndex--
    }
    this.tableGrid.rebuild()
    this.loadCurrentFrame()
  }

  updateHurtbox(index: number, w: number, h: number, dmgMult: number): void {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim) return
    const frame = anim.frames?.[this.currentFrameIndex]
    if (!frame?.hurtboxes) return
    if (index >= frame.hurtboxes.length) return
    const hb = frame.hurtboxes[index]
    hb[0] = w
    hb[1] = h
    hb[7] = dmgMult
    this.tableGrid.rebuild()
    this.loadCurrentFrame()
  }

  repeatPreviousHurtbox(): void {
    const idx = this.selectedHurtboxIndex
    if (idx < 0 || this.currentFrameIndex <= 0) return
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim) return
    const prev = anim.frames[this.currentFrameIndex - 1]
    if (!prev?.hurtboxes || idx >= prev.hurtboxes.length) return
    const frame = anim.frames[this.currentFrameIndex]
    if (!frame) return
    if (!frame.hurtboxes) frame.hurtboxes = []
    while (frame.hurtboxes.length <= idx) {
      frame.hurtboxes.push([...DEFAULT_EDITOR_HURTBOX])
    }
    frame.hurtboxes[idx] = [...prev.hurtboxes[idx]] as EditorHurtboxData
    this.tableGrid.rebuild()
    this.loadCurrentFrame()
  }

  // ── Import / Export (shared) ──

  canImport(): boolean {
    return true
  }

  protected tryLoadImagesFromPaths(paths: Record<string, string>): void {
    const entries = Object.entries(paths).filter(([_, p]) => p)
    if (entries.length === 0) return

    const failed: { key: string; filename: string }[] = []
    let completed = 0
    const total = entries.length
    const onResolved = () => {
      completed++
      if (completed >= total && failed.length > 0) {
        this.promptForSpritesheets(failed)
      }
    }

    for (const [key, rawPath] of entries) {
      const url = pathToUrl(rawPath)
      if (!url) {
        console.warn("[tryLoadImagesFromPaths] could not resolve URL for", rawPath)
        failed.push({ key, filename: extractFilename(rawPath) })
        onResolved()
        continue
      }
      console.log("[tryLoadImagesFromPaths] loading", key, "from", url)
      loadImageFromUrl(url)
        .then((result: { image: HTMLImageElement }) => {
          console.log("[tryLoadImagesFromPaths] loaded", key, "successfully")
          this.applyLoadedSprite(key, rawPath, result.image)
          onResolved()
        })
        .catch((err) => {
          console.warn("[tryLoadImagesFromPaths] failed to load", url, err)
          failed.push({ key, filename: extractFilename(url) })
          onResolved()
        })
    }
  }

  protected setupAdditionalSpritesFromPaths(paths: Record<string, string>): void {
    const spriteKeys = Object.keys(paths).filter((k) => k !== "char")
    for (const key of spriteKeys) {
      const idx = parseInt(key) || 0
      if (idx >= 1) {
        while (this.additionalSprites.length < idx) {
          this.additionalSprites.push({
            image: null,
            path: "",
            frameW: 145,
            frameH: 145,
            totalFrames: 1,
            frameIdx: 0,
            originX: 0.5,
            originY: 0.5,
          })
          this.tableGrid.addSpriteRow(this.additionalSprites.length - 1)
        }
        this.additionalSprites[idx - 1].path = paths[key]
      }
    }
  }

  private applyLoadedSprite(key: string, rawPath: string, image: HTMLImageElement): void {
    if (key === "char") {
      this.baseSpritePath = rawPath
      this.tableGrid.syncBaseSpritePath(rawPath)
      this.loadBaseSpritesheet(image)
    } else {
      const idx = parseInt(key) || 0
      if (idx >= 1 && idx - 1 < this.additionalSprites.length) {
        this.additionalSprites[idx - 1].path = rawPath
        this.addAdditionalSprite(image, idx - 1)
      }
    }
  }

  private promptForSpritesheets(failed: { key: string; filename: string }[]): void {
    const names = failed.map(f => f.filename).join(", ")
    console.log("[tryLoadImagesFromPaths] prompting for spritesheets:", names)

    const input = document.createElement("input")
    input.type = "file"
    input.accept = ".png,.jpg,.jpeg,.gif,.webp"
    input.multiple = true
    input.style.display = "none"
    document.body.appendChild(input)

    input.addEventListener("change", () => {
      const files = input.files
      if (!files || files.length === 0) return

      const remaining = [...failed]

      for (const file of files) {
        const img = new Image()
        img.onload = () => {
          const fileBase = file.name.replace(/\.\w+$/, "").toLowerCase()
          let matchIdx = remaining.findIndex(f => {
            const expectedBase = f.filename.replace(/\.\w+$/, "").toLowerCase()
            return fileBase === expectedBase || fileBase.includes(expectedBase) || expectedBase.includes(fileBase)
          })
          if (matchIdx === -1) matchIdx = 0
          const entry = remaining.splice(matchIdx, 1)[0]
          if (entry) {
            console.log("[tryLoadImagesFromPaths] user selected", file.name, "for", entry.key)
            this.applyLoadedSprite(entry.key, file.name, img)
          }
        }
        img.src = URL.createObjectURL(file)
      }

      input.remove()
    })

    input.click()
  }

  getExportFileName(): string {
    return this.getEntries()[0]?.name.replace(/[^a-zA-Z0-9_-]/g, "") ?? ""
  }

  export(): void {
    downloadFile(`${this.getExportFileName()}.ts`, this.getExportContent())
  }
}
