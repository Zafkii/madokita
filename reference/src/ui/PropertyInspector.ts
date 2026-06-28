import type { TransformState } from "./CanvasView"
import type { EditorHurtboxData } from "../types/MovementTypes"

export type PropertyInspectorConfig = {
  onTransformChange: (state: TransformState) => void
  onOriginChange: (x: number, y: number) => void
  onBaseRotationChange: (deg: number) => void
  onPhaseChange?: (phase: string) => void
  onFpsChange?: (fps: number) => void
  onWindupChange?: (ms: number) => void
  onActiveTimeChange?: (ms: number) => void
  onRecoverChange?: (ms: number) => void
  onGuardDurationChange?: (ms: number) => void
  onGuardFpsChange?: (fps: number) => void
  onHurtboxChange?: (hb: EditorHurtboxData) => void
  mode?: "attack" | "movement"
}

export class PropertyInspector {
  private el: HTMLElement
  private config: PropertyInspectorConfig
  private mode: "attack" | "movement"

  private offsetXInput!: HTMLInputElement
  private offsetYInput!: HTMLInputElement
  private rotationInput!: HTMLInputElement
  private scaleXInput!: HTMLInputElement
  private scaleYInput!: HTMLInputElement
  private originXInput!: HTMLInputElement
  private originYInput!: HTMLInputElement
  private baseRotationInput!: HTMLInputElement
  private originSectionEl!: HTMLElement
  private transformSectionEl!: HTMLElement
  private fpsInput?: HTMLInputElement
  private frameIndexDisplay!: HTMLElement
  private phaseSelect?: HTMLSelectElement
  private phaseGroupEl?: HTMLElement
  private activeElementType: "sprite" | "hurtbox" = "sprite"
  private currentHurtboxW = 0
  private currentHurtboxH = 0
  private currentHurtboxDmgMult = 1

  private ignoreNextChange = false

  constructor(container: HTMLElement, config: PropertyInspectorConfig) {
    this.config = config
    this.mode = config.mode ?? "attack"
    this.el = document.createElement("div")
    this.el.className = "property-inspector"
    this.el.style.cssText =
      "padding:12px;display:flex;flex-direction:column;gap:var(--gap-md);flex-shrink:0;box-sizing:border-box;"
    container.appendChild(this.el)
    this.render()
    this.bindEvents()
  }

  private render(): void {
    this.el.innerHTML = ""
    this.renderStyles()
    this.renderProperties()
    this.renderOrigin()
    this.renderAnimation()
    if (this.mode === "attack") {
      this.renderPhaseDurations()
    }
  }

  private renderStyles(): void {
    const style = document.createElement("style")
    style.textContent = `
      .prop-group { display:flex; align-items:center; gap:var(--gap-md); margin-bottom:4px; }
      .prop-label { font-size:var(--font-md); color:var(--label-color,#aaa); min-width:85px; flex-shrink:0; }
      .prop-input {
        flex:1; padding:4px 6px; background:#1a1a2e; color:#e0e0e0;
        border:1px solid #444; border-radius:var(--radius-md); font-size:var(--font-lg); min-width:0;
      }
      .prop-input:focus { outline:none; border-color:#6af; }
      .prop-input::-webkit-outer-spin-button,
      .prop-input::-webkit-inner-spin-button { -webkit-appearance:none; margin:0; }
      .prop-input[type=number] { -moz-appearance:textfield; }
      select.prop-input { cursor:pointer; }
      .section-title { font-weight:600; font-size:var(--font-lg); margin:8px 0 4px; color:var(--label-color,#8cf); border-bottom:1px solid #333; padding-bottom:2px; }
    `
    this.el.appendChild(style)
  }

  private renderProperties(): void {
    const section = document.createElement("div")
    section.innerHTML = `
      <div class="section-title">Selected Element Properties</div>
      <div class="prop-group">
        <div class="prop-label">Offset X</div>
        <input id="prop-offset-x" type="number" value="0" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Offset Y</div>
        <input id="prop-offset-y" type="number" value="0" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Rotation (°)</div>
        <input id="prop-rotation" type="number" value="0" step="1" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Scale X</div>
        <input id="prop-scale-x" type="number" value="1" step="0.05" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Scale Y</div>
        <input id="prop-scale-y" type="number" value="1" step="0.05" class="prop-input" />
      </div>
    `
    if (this.mode === "attack") {
      this.phaseGroupEl = document.createElement("div")
      this.phaseGroupEl.className = "prop-group"
      this.phaseGroupEl.innerHTML = `
        <div class="prop-label">Phase</div>
        <select id="prop-phase" class="prop-input">
          <option value="wu">WU</option>
          <option value="atk">ATK</option>
          <option value="rc">RC</option>
          <option value="guard">GUARD</option>
        </select>
      `
      section.appendChild(this.phaseGroupEl)
    }
    this.el.appendChild(section)
    this.transformSectionEl = section

    this.offsetXInput = this.el.querySelector("#prop-offset-x")!
    this.offsetYInput = this.el.querySelector("#prop-offset-y")!
    this.rotationInput = this.el.querySelector("#prop-rotation")!
    this.scaleXInput = this.el.querySelector("#prop-scale-x")!
    this.scaleYInput = this.el.querySelector("#prop-scale-y")!
    if (this.mode === "attack") {
      this.phaseSelect = this.el.querySelector("#prop-phase")!
    }
  }

  private renderOrigin(): void {
    const section = document.createElement("div")
    section.innerHTML = `
      <div class="section-title">Base Sprite Properties</div>
      <div class="prop-group">
        <div class="prop-label">Origin X</div>
        <input id="prop-origin-x" type="number" value="0.5" step="0.01" min="0" max="1" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Origin Y</div>
        <input id="prop-origin-y" type="number" value="0.5" step="0.01" min="0" max="1" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Base Rot (°)</div>
        <input id="prop-base-rotation" type="number" value="0" class="prop-input" />
      </div>
    `
    this.el.appendChild(section)
    this.originSectionEl = section

    this.originXInput = this.el.querySelector("#prop-origin-x")!
    this.originYInput = this.el.querySelector("#prop-origin-y")!
    this.baseRotationInput = this.el.querySelector("#prop-base-rotation")!
  }

  private renderAnimation(): void {
    const section = document.createElement("div")
    section.innerHTML = `
      <div class="section-title">Animation</div>
      <div class="prop-group">
        <div class="prop-label">Frame</div>
        <span id="prop-frame-index" style="font-size:var(--font-lg);color:#e0e0e0;">0 / 0</span>
      </div>
    `
    if (this.mode === "movement") {
      const fpsGroup = document.createElement("div")
      fpsGroup.className = "prop-group"
      fpsGroup.innerHTML = `
        <div class="prop-label">FPS</div>
        <input id="prop-fps" type="number" value="14" min="1" class="prop-input" />
      `
      section.appendChild(fpsGroup)
    }
    this.el.appendChild(section)

    this.frameIndexDisplay = this.el.querySelector("#prop-frame-index")!
    if (this.mode === "movement") {
      this.fpsInput = this.el.querySelector("#prop-fps")!
    }
  }

  private renderPhaseDurations(): void {
    const section = document.createElement("div")
    section.innerHTML = `
      <div class="section-title">Phase Durations (ms)</div>
      <div class="prop-group">
        <div class="prop-label">Windup</div>
        <input id="prop-windup" type="number" value="200" min="0" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Active</div>
        <input id="prop-active-time" type="number" value="250" min="0" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Recover</div>
        <input id="prop-recover" type="number" value="800" min="0" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Guard</div>
        <input id="prop-guard" type="number" value="3000" min="0" class="prop-input" />
      </div>
      <div class="prop-group">
        <div class="prop-label">Guard FPS</div>
        <input id="prop-guard-fps" type="number" value="14" min="1" class="prop-input" />
      </div>
    `
    this.el.appendChild(section)
  }

  private bindEvents(): void {
    const inputs = [
      this.offsetXInput,
      this.offsetYInput,
      this.rotationInput,
      this.scaleXInput,
      this.scaleYInput,
    ]
    for (const input of inputs) {
      input.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        if (
          this.activeElementType === "hurtbox" &&
          this.config.onHurtboxChange
        ) {
          const hbData: EditorHurtboxData = [
            this.currentHurtboxW,
            this.currentHurtboxH,
            parseFloat(this.offsetXInput.value) || 0,
            parseFloat(this.offsetYInput.value) || 0,
            parseFloat(this.scaleXInput.value) || 1,
            parseFloat(this.scaleYInput.value) || 1,
            parseFloat(this.rotationInput.value) || 0,
            this.currentHurtboxDmgMult,
          ]
          this.config.onHurtboxChange(hbData)
        } else {
          this.config.onTransformChange(this.getTransform())
        }
      })
    }

    if (this.phaseSelect) {
      this.phaseSelect.addEventListener("change", () => {
        if (this.ignoreNextChange) return
        this.config.onPhaseChange?.(this.phaseSelect!.value)
      })
    }

    const onOriginInp = () => {
      if (this.ignoreNextChange) return
      this.config.onOriginChange(
        parseFloat(this.originXInput.value) || 0.5,
        parseFloat(this.originYInput.value) || 0.5,
      )
    }
    this.originXInput.addEventListener("input", onOriginInp)
    this.originYInput.addEventListener("input", onOriginInp)

    this.baseRotationInput.addEventListener("input", () => {
      if (this.ignoreNextChange) return
      this.config.onBaseRotationChange(
        parseFloat(this.baseRotationInput.value) || 0,
      )
    })

    if (this.fpsInput) {
      this.fpsInput.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        this.config.onFpsChange?.(parseInt(this.fpsInput!.value) || 14)
      })
    }

    if (this.mode === "attack") {
      const windup = this.el.querySelector("#prop-windup") as HTMLInputElement
      const activeTime = this.el.querySelector(
        "#prop-active-time",
      ) as HTMLInputElement
      const recover = this.el.querySelector("#prop-recover") as HTMLInputElement
      const guardEl = this.el.querySelector("#prop-guard") as HTMLInputElement
      const guardFpsEl = this.el.querySelector(
        "#prop-guard-fps",
      ) as HTMLInputElement

      windup?.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        this.config.onWindupChange?.(parseInt(windup.value) || 0)
      })
      activeTime?.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        this.config.onActiveTimeChange?.(parseInt(activeTime.value) || 0)
      })
      recover?.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        this.config.onRecoverChange?.(parseInt(recover.value) || 0)
      })
      guardEl?.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        this.config.onGuardDurationChange?.(parseInt(guardEl.value) || 0)
      })
      guardFpsEl?.addEventListener("input", () => {
        if (this.ignoreNextChange) return
        this.config.onGuardFpsChange?.(parseInt(guardFpsEl.value) || 14)
      })
    }

    // Transform inputs for hurtbox are handled by the existing onTransformChange listener
    // which calls onHurtboxChange with correct cached W/H/DMG values
  }

  setTransform(state: TransformState): void {
    this.ignoreNextChange = true
    this.offsetXInput.value = String(state.offsetX)
    this.offsetYInput.value = String(state.offsetY)
    this.rotationInput.value = String(Math.round(state.rotation))
    this.scaleXInput.value = String(Math.round(state.scaleX * 100) / 100)
    this.scaleYInput.value = String(Math.round(state.scaleY * 100) / 100)
    this.ignoreNextChange = false
  }

  setOrigin(x: number, y: number): void {
    this.ignoreNextChange = true
    this.originXInput.value = String(x.toFixed(3))
    this.originYInput.value = String(y.toFixed(3))
    this.ignoreNextChange = false
  }

  setBaseRotation(deg: number): void {
    this.ignoreNextChange = true
    this.baseRotationInput.value = String(deg)
    this.ignoreNextChange = false
  }

  setActiveElement(type: "sprite" | "hurtbox"): void {
    this.activeElementType = type
    if (type === "hurtbox") {
      this.transformSectionEl.style.display = ""
      this.originSectionEl.style.display = "none"
      if (this.phaseGroupEl) {
        this.phaseGroupEl.style.display = "none"
      }
    } else {
      this.transformSectionEl.style.display = ""
      this.originSectionEl.style.display = ""
      if (this.phaseGroupEl) {
        this.phaseGroupEl.style.display = ""
      }
    }
  }

  setFrameIndex(current: number, total: number): void {
    this.frameIndexDisplay.textContent = `${current + 1} / ${total}`
  }

  setPhase(phase: string): void {
    if (!this.phaseSelect) return
    this.ignoreNextChange = true
    this.phaseSelect.value = phase
    this.ignoreNextChange = false
  }

  getTransform(): TransformState {
    return {
      offsetX: parseInt(this.offsetXInput.value) || 0,
      offsetY: parseInt(this.offsetYInput.value) || 0,
      rotation: Math.round(parseFloat(this.rotationInput.value) || 0),
      scaleX: Math.round((parseFloat(this.scaleXInput.value) || 1) * 100) / 100,
      scaleY: Math.round((parseFloat(this.scaleYInput.value) || 1) * 100) / 100,
    }
  }

  setGuardFps(fps: number): void {
    const el = this.el.querySelector("#prop-guard-fps") as HTMLInputElement
    if (el) {
      this.ignoreNextChange = true
      el.value = String(fps)
      this.ignoreNextChange = false
    }
  }

  setHurtboxData(hb: EditorHurtboxData | null): void {
    this.ignoreNextChange = true
    if (hb) {
      this.currentHurtboxW = hb[0]
      this.currentHurtboxH = hb[1]
      this.currentHurtboxDmgMult = hb[7]
      this.offsetXInput.value = String(hb[2])
      this.offsetYInput.value = String(hb[3])
      this.scaleXInput.value = String(Math.round(hb[4] * 100) / 100)
      this.scaleYInput.value = String(Math.round(hb[5] * 100) / 100)
      this.rotationInput.value = String(hb[6])
    }
    this.ignoreNextChange = false
  }

  setPhaseDurations(
    windup: number,
    activeTime: number,
    recover: number,
    guardDuration: number,
  ): void {
    const w = this.el.querySelector("#prop-windup") as HTMLInputElement
    const a = this.el.querySelector("#prop-active-time") as HTMLInputElement
    const r = this.el.querySelector("#prop-recover") as HTMLInputElement
    const g = this.el.querySelector("#prop-guard") as HTMLInputElement
    if (!w) return
    this.ignoreNextChange = true
    w.value = String(windup)
    a.value = String(activeTime)
    r.value = String(recover)
    g.value = String(guardDuration)
    this.ignoreNextChange = false
  }

  setFps(fps: number): void {
    if (!this.fpsInput) return
    this.ignoreNextChange = true
    this.fpsInput.value = String(fps)
    this.ignoreNextChange = false
  }

  getElement(): HTMLElement {
    return this.el
  }
}
