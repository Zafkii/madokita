import type {
  EditorFrame,
  MovementAnimDef,
  Movement,
} from "../../types/MovementTypes"
import { CanvasView } from "../../ui/CanvasView"
import { PropertyInspector } from "../../ui/PropertyInspector"
import { MovementPreview } from "../../preview/MovementPreview"
import { BaseEditor } from "../shared/BaseEditor"
import {
  exportMovements,
  parseMovementImportText,
  parseMovementLoadImagePaths,
} from "./MovementImportExport"
import type { MovementEntry } from "./movementTypes"

export type { MovementEntry } from "./movementTypes"

export class MovementEditor extends BaseEditor {
  movements: MovementEntry[] = []
  private currentMovementIndex = 0

  constructor(
    canvasView: CanvasView,
    propertyInspector: PropertyInspector,
    tablesContainer: HTMLElement,
    rightPanel: HTMLElement,
  ) {
    super(canvasView, propertyInspector, tablesContainer, rightPanel)
    this.currentAnimName = "walk"
    this.initDefaultMovement()
    this.tableGrid.init()
    this.ensureAnimExists("walk")
    this.refreshUI()
  }

  protected createPreview(container: HTMLElement): MovementPreview {
    return new MovementPreview(container)
  }

  // ── ToolbarHost / abstract overrides ──

  protected getCurrentDef(): Movement {
    if (this.movements.length === 0) {
      this.movements.push({
        name: "default",
        def: {
          assetKey: "idle",
          defaultOriginX: 0.5,
          defaultOriginY: 0.5,
          animations: {},
        },
      })
      this.currentMovementIndex = 0
    }
    return this.movements[this.currentMovementIndex].def
  }

  protected getEntries(): MovementEntry[] {
    return this.movements
  }

  getCurrentAnim(): MovementAnimDef | null {
    return this.getCurrentDef().animations[this.currentAnimName] ?? null
  }

  getAnimations(): Record<string, any> {
    return this.getCurrentDef().animations
  }

  ensureAnimExists(name: string): void {
    const cur = this.getCurrentDef()
    if (!cur.animations[name]) {
      cur.animations[name] = {
        frames: [],
        fps: 14,
        loop: true,
      }
    }
  }

  protected createDefaultAnimObj(_name: string): MovementAnimDef {
    return { frames: [], fps: 14, loop: true }
  }

  protected getRemoveAnimFallbackName(): string {
    return "walk"
  }

  protected getExtraFrameProps(): Record<string, any> {
    return {}
  }

  protected syncEditorUI(): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      this.propertyInspector.setFps(anim.fps ?? 14)
    }
  }

  protected refreshUI(): void {
    this.tableGrid.syncAnimName(this.currentAnimName)
    this.syncEditorUI()
    const total = this.getCurrentAnim()?.frames.length ?? 0
    this.tableGrid.updateFrameDisplay(this.currentFrameIndex, total)
    this.autoSelectSprite()
    this.loadCurrentFrame()
    this.preview.setAnimation(this.getCurrentAnim(), (frame: any, idx: number) =>
      this.onPreviewFrame(frame, idx),
    )
    this.tableGrid.rebuildSpriteGroups()
  }

  // ── Init ──

  private initDefaultMovement(): void {
    if (this.movements.length === 0) {
      this.movements.push({
        name: "player-movement",
        def: {
          assetKey: "idle",
          defaultOriginX: 0.5,
          defaultOriginY: 0.5,
          animations: {},
        },
      })
    }
    this.ensureAnimExists(this.currentAnimName)
  }

  // ── FPS ──

  onFpsChange(fps: number): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      anim.fps = fps
    }
  }

  // ── Legacy add/remove (kept for compatibility) ──

  addFrame(): void {
    this.isPreviewing = false
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim) return

    const last =
      anim.frames.length > 0 ? anim.frames[anim.frames.length - 1] : null
    const n = this.spriteCount()
    const cur =
      this.canvasView.getActiveOverlayTransform() ??
      this.canvasView.getTransform()

    const newFrame: EditorFrame = {
      spriteFrames: last ? [...last.spriteFrames] : Array(n).fill(0),
      offsetX: Array(n).fill(0),
      offsetY: Array(n).fill(0),
      rotation: Array(n).fill(0),
    }

    if (newFrame.offsetX.length > this.currentSpriteIndex) {
      newFrame.offsetX[this.currentSpriteIndex] = cur.offsetX
      newFrame.offsetY[this.currentSpriteIndex] = cur.offsetY
      newFrame.rotation[this.currentSpriteIndex] = cur.rotation
    }
    if (last) {
      for (let i = 0; i < n; i++) {
        newFrame.offsetX[i] = last.offsetX[i] ?? 0
        newFrame.offsetY[i] = last.offsetY[i] ?? 0
        newFrame.rotation[i] = last.rotation[i] ?? 0
      }
      newFrame.offsetX[this.currentSpriteIndex] = cur.offsetX
      newFrame.offsetY[this.currentSpriteIndex] = cur.offsetY
      newFrame.rotation[this.currentSpriteIndex] = cur.rotation
      newFrame.scaleX = last.scaleX ? [...last.scaleX] : Array(n).fill(1)
      newFrame.scaleY = last.scaleY ? [...last.scaleY] : Array(n).fill(1)
    }

    anim.frames.push(newFrame)
    const targetIdx = anim.frames.length - 1
    this.currentFrameIndex = targetIdx
    this.tableGrid.updateFrameDisplay(targetIdx, anim.frames.length)
    this.preview.setAnimation(this.getCurrentAnim(), (frame: any, idx: number) =>
      this.onPreviewFrame(frame, idx),
    )
    this.selectFrame(targetIdx)
  }

  removeFrame(): void {
    this.isPreviewing = false
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim || anim.frames.length <= 1) return
    anim.frames.splice(this.currentFrameIndex, 1)
    const targetIdx = Math.min(this.currentFrameIndex, anim.frames.length - 1)
    this.currentFrameIndex = targetIdx
    this.tableGrid.updateFrameDisplay(targetIdx, anim.frames.length)
    this.preview.setAnimation(this.getCurrentAnim(), (frame: any, idx: number) =>
      this.onPreviewFrame(frame, idx),
    )
    this.selectFrame(targetIdx)
  }

  // ── Import / Export ──

  importFromText(text: string): boolean {
    console.log("[MovementEditor.importFromText] parsing...")
    const def = parseMovementImportText(text)
    if (!def) {
      console.warn("[MovementEditor.importFromText] parseMovementImportText returned null")
      return false
    }
    console.log("[MovementEditor.importFromText] parsed def, assetKey:", def.assetKey)

    const paths = parseMovementLoadImagePaths(text)

    this.setupAdditionalSpritesFromPaths(paths)

    const name = def.assetKey.replace(/(Def|def)$/, "")
    this.movements[0] = { name, def: { ...def } }
    this.currentAnimName = Object.keys(def.animations)[0] || "walk"
    this.currentFrameIndex = 0
    this.canvasView.setOrigin(
      def.defaultOriginX ?? 0.5,
      def.defaultOriginY ?? 0.5,
    )
    this.tableGrid.rebuildAnimRows()
    this.refreshUI()

    this.tryLoadImagesFromPaths(paths)

    return true
  }

  getExportContent(): string {
    const spritePaths: Record<string, string> = {}
    if (this.baseSpritePath) spritePaths["char"] = this.baseSpritePath
    for (let i = 0; i < this.additionalSprites.length; i++) {
      if (this.additionalSprites[i].path) {
        spritePaths[String(i + 1)] = this.additionalSprites[i].path
      }
    }
    return exportMovements(this.movements, spritePaths)
  }
}
