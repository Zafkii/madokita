import type { AttackPhase } from "../../types/AttackTypes"
import type {
  AttackAnimDef,
  Attack,
} from "../../types/AttackTypes"
import { CanvasView } from "../../ui/CanvasView"
import { PropertyInspector } from "../../ui/PropertyInspector"
import { AttackPreview } from "../../preview/AttackPreview"
import { BaseEditor } from "../shared/BaseEditor"
import {
  exportAttacks,
  parseImportText,
  parseLoadImagePaths,
} from "./attackingImportExport"
import type { AttackingEntry } from "./attackingTypes"

export type { AttackingEntry } from "./attackingTypes"

export class AttackingEditor extends BaseEditor {
  attacks: AttackingEntry[] = []
  private currentAttackIndex = 0

  constructor(
    canvasView: CanvasView,
    propertyInspector: PropertyInspector,
    tablesContainer: HTMLElement,
    rightPanel: HTMLElement,
  ) {
    super(canvasView, propertyInspector, tablesContainer, rightPanel)
    this.currentAnimName = "slash"
    this.initDefaultAttack()
    this.preview.setAnimation(this.getCurrentAnim(), (frame: any, idx: number) =>
      this.onPreviewFrame(frame, idx),
    )
  }

  protected createPreview(container: HTMLElement): AttackPreview {
    return new AttackPreview(container)
  }

  // ── ToolbarHost / abstract overrides ──

  protected getCurrentDef(): Attack {
    if (this.attacks.length === 0) {
      this.attacks.push({
        name: "default",
        def: {
          assetKey: "weapon",
          defaultOriginX: 0.5,
          defaultOriginY: 0.5,
          animations: {},
        },
      })
      this.currentAttackIndex = 0
    }
    return this.attacks[this.currentAttackIndex].def
  }

  protected getEntries(): AttackingEntry[] {
    return this.attacks
  }

  getCurrentAnim(): AttackAnimDef | null {
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
        guardFps: 14,
        windup: 200,
        activeTime: 250,
        recover: 800,
        guard: 3000,
      }
    }
  }

  protected createDefaultAnimObj(_name: string): AttackAnimDef {
    return {
      frames: [],
      guardFps: 14,
      windup: 200,
      activeTime: 250,
      recover: 800,
      guard: 3000,
    }
  }

  protected getRemoveAnimFallbackName(): string {
    return "slash"
  }

  protected getExtraFrameProps(last: any): Record<string, any> {
    return { phase: last?.phase ?? "wu" }
  }

  protected syncEditorUI(): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      this.propertyInspector.setPhaseDurations(
        anim.windup ?? 200,
        anim.activeTime ?? 250,
        anim.recover ?? 800,
        anim.guard ?? 3000,
      )
      this.propertyInspector.setGuardFps(anim.guardFps ?? 14)
    }
  }

  protected refreshUI(): void {
    this.tableGrid.rebuildAnimRows()
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

  private initDefaultAttack(): void {
    if (this.attacks.length === 0) {
      this.attacks.push({
        name: "sayaka-attack",
        def: {
          assetKey: "weapon",
          defaultOriginX: 0.5,
          defaultOriginY: 0.5,
          animations: {},
        },
      })
    }
    this.ensureAnimExists(this.currentAnimName)
  }

  // ── Phase ──

  onPhaseChange(phase: string): void {
    const anim = this.getCurrentDef().animations[this.currentAnimName]
    if (!anim || anim.frames.length === 0) return
    if (!this.isPhaseChangeAllowed(anim, this.currentFrameIndex, phase)) {
      const curPhase = anim.frames[this.currentFrameIndex].phase ?? "atk"
      this.propertyInspector.setPhase(curPhase)
      return
    }
    anim.frames[this.currentFrameIndex].phase = phase as AttackPhase
    for (let i = this.currentFrameIndex + 1; i < anim.frames.length; i++) {
      anim.frames[i].phase = phase as AttackPhase
    }
    this.propertyInspector.setPhase(phase)
  }

  private isPhaseChangeAllowed(
    anim: AttackAnimDef,
    changeIdx: number,
    newPhase: string,
  ): boolean {
    const order: AttackPhase[] = ["wu", "atk", "rc", "guard"]
    const newOrder = order.indexOf(newPhase as AttackPhase)
    if (newOrder < 0) return true
    for (const p of order) {
      if (p === newPhase) continue
      const pOrder = order.indexOf(p)
      if (pOrder >= newOrder) continue
      let hasBefore = false
      for (let i = 0; i < changeIdx; i++) {
        if ((anim.frames[i].phase ?? "atk") === p) {
          hasBefore = true
          break
        }
      }
      if (!hasBefore) return false
    }
    return true
  }

  // ── Preview extra (phase + hurtboxes) ──

  protected onPreviewFrameExtra(frame: any): void {
    this.propertyInspector.setPhase(frame.phase ?? "atk")
    this.canvasView.setHurtboxes(
      frame.hurtboxes ?? [],
      this.selectedHurtboxIndex,
    )
  }

  // ── Duration handlers ──

  onGuardFpsChange(fps: number): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      anim.guardFps = fps
      ;(this.preview as AttackPreview).setGuardParams(fps, anim.guard ?? 3000)
    }
  }

  onWindupChange(ms: number): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      anim.windup = ms
      ;(this.preview as AttackPreview).setPhaseDurations(
        ms,
        anim.activeTime ?? 250,
        anim.recover ?? 800,
      )
    }
  }

  onActiveTimeChange(ms: number): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      anim.activeTime = ms
      ;(this.preview as AttackPreview).setPhaseDurations(
        anim.windup ?? 200,
        ms,
        anim.recover ?? 800,
      )
    }
  }

  onRecoverChange(ms: number): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      anim.recover = ms
      ;(this.preview as AttackPreview).setPhaseDurations(
        anim.windup ?? 200,
        anim.activeTime ?? 250,
        ms,
      )
    }
  }

  onGuardDurationChange(ms: number): void {
    const anim = this.getCurrentAnim()
    if (anim) {
      anim.guard = ms
      ;(this.preview as AttackPreview).setGuardParams(anim.guardFps ?? 14, ms)
    }
  }

  // ── Import / Export ──

  importFromText(text: string): boolean {
    const def = parseImportText(text)
    if (!def) return false

    const paths = parseLoadImagePaths(text)

    this.setupAdditionalSpritesFromPaths(paths)

    const name = def.assetKey.replace(/(Def|def)$/, "")
    this.attacks[0] = { name, def: { ...def } }
    const firstAnimName = Object.keys(def.animations)[0] || "slash"
    this.currentAnimName = firstAnimName
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
    return exportAttacks(this.attacks, spritePaths)
  }
}
