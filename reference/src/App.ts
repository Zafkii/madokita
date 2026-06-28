declare global {
  interface Window {
    showOpenFilePicker(options?: any): Promise<FileSystemFileHandle[]>
    showSaveFilePicker(options?: any): Promise<FileSystemFileHandle>
  }
}

import { CanvasView } from "./ui/CanvasView"
import { PropertyInspector } from "./ui/PropertyInspector"
import { ThemeManager } from "./ui/ThemeManager"
import { AttackingEditor } from "./modes/Attacking/AttackingEditor"
import { MovementEditor } from "./modes/Movement/MovementEditor"

type EditorMode = "attack" | "movement"

export class App {
  private container: HTMLElement
  private canvasView!: CanvasView
  private propertyInspector!: PropertyInspector
  private attackingEditor!: AttackingEditor
  private movementEditor!: MovementEditor
  private themeManager = new ThemeManager()
  private modeIndicator: HTMLElement
  private zoomDisplay: HTMLElement
  private animationTables: HTMLElement
  private rightPanel: HTMLElement
  private topBar: HTMLElement
  private statusbarEl: HTMLElement
  private btnTheme!: HTMLButtonElement
  private resetViewBtn!: HTMLButtonElement
  private btnImport!: HTMLButtonElement
  private btnSave!: HTMLButtonElement
  private modeSelect!: HTMLSelectElement
  private currentMode: EditorMode = "attack"
  private currentFileHandle: FileSystemFileHandle | null = null

  constructor(container: HTMLElement) {
    this.container = container
    this.container.innerHTML = ""
    this.buildLayout()
    this.modeIndicator = document.getElementById("mode-indicator")!
    this.zoomDisplay = document.getElementById("zoom-display")!
    this.animationTables = this.container.querySelector("#animation-tables")!
    this.rightPanel = this.container.querySelector("#right-panel")!
    this.topBar = this.container.querySelector("#top-bar")!
    this.statusbarEl = this.container.querySelector("#statusbar")!
    this.btnTheme = this.container.querySelector("#btn-theme")!
    this.btnImport = this.container.querySelector("#btn-import")!
    this.btnSave = this.container.querySelector("#btn-save")!
    this.modeSelect = this.container.querySelector("#mode-select")!
    this.resetViewBtn = this.container.querySelector("#btn-reset-view")!

    this.propertyInspector = new PropertyInspector(this.rightPanel, {
      onTransformChange: () => {},
      onOriginChange: () => {},
      onBaseRotationChange: () => {},
      mode: "attack",
    })

    this.attackingEditor = new AttackingEditor(
      this.canvasView,
      this.propertyInspector,
      this.animationTables,
      this.rightPanel,
    )
    this.movementEditor = new MovementEditor(
      this.canvasView,
      this.propertyInspector,
      this.animationTables,
      this.rightPanel,
    )

    this.bindEvents()
    this.switchMode("attack")
    this.applyTheme()
  }

  private buildLayout(): void {
    const tm = this.themeManager
    this.container.innerHTML = `
      <div id="mode-indicator" style="
        font-size:var(--font-sm);color:${tm.textMuted};padding:2px 12px;
        background:#12121f; border-bottom:1px solid #2a2a2a; flex-shrink:0;
      "></div>
      <div id="top-bar" style="
        display:grid; grid-template-columns:auto 1fr auto; align-items:center; gap:var(--gap-lg); padding:6px 12px;
        background:#16162a; border-bottom:1px solid #333; flex-shrink:0;
      ">
        <div style="display:flex;flex-direction:column;justify-content:center;gap:var(--gap-sm);">
          <select id="mode-select" style="
            padding:3px 6px; background:#1a1a2e; color:#e0e0e0;
            border:1px solid #444; border-radius:var(--radius-md); font-size:var(--font-lg); font-weight:600;
          ">
            <option value="attack">Attack Editor</option>
            <option value="movement">Movement Editor</option>
          </select>
        </div>
        <div id="animation-tables" style="display:flex;justify-content:center;align-items:center;gap:var(--gap-md);flex-wrap:wrap;"></div>
        <div id="header-actions" style="display:flex;flex-direction:column;justify-content:center;gap:var(--gap-sm);">
          <button id="btn-theme" style="
            padding:var(--btn-pad-md); background:#2a2a4a; color:var(--btn-color,${tm.textPrimary});
            border:1px solid #555; border-radius:var(--radius-md); cursor:pointer; font-size:var(--font-sm);
          ">☀ Light</button>
          <button id="btn-import" style="
            padding:var(--btn-pad-md); background:#a82; color:var(--btn-color,${tm.textOnColor});
            border:none; border-radius:var(--radius-lg); cursor:pointer; font-size:var(--font-sm); font-weight:600;
          ">Open</button>
          <div style="position:relative;">
            <button id="btn-save" style="
              padding:var(--btn-pad-md); background:#2a5; color:var(--btn-color,${tm.textOnColor});
              border:none; border-radius:var(--radius-lg); cursor:pointer; font-size:var(--font-sm); font-weight:600; width:100%;
            ">Save ▼</button>
            <div id="save-dropdown" style="
              display:none; position:absolute; bottom:100%; left:0; width:100%;
              background:#1a1a2e; border:1px solid #444; border-radius:var(--radius-lg); z-index:999; overflow:hidden;
            ">
              <div id="save-option" style="
                padding:6px 12px; cursor:pointer; color:var(--btn-color,#e0e0e0); font-size:var(--font-sm); border-bottom:1px solid #333;
              ">Save</div>
              <div id="save-as-option" style="
                padding:6px 12px; cursor:pointer; color:var(--btn-color,#e0e0e0); font-size:var(--font-sm);
              ">Save As</div>
            </div>
          </div>
        </div>
      </div>
      <div id="main-area" style="display:flex;flex:1;overflow:hidden;min-height:0;">
        <div id="canvas-container" style="flex:1;position:relative;overflow:hidden;background:#12121f;"></div>
        <div id="right-panel" style="width:280px;flex-shrink:0;border-left:1px solid #333;overflow-y:auto;display:flex;flex-direction:column;background:#1a1a2e;"></div>
      </div>
      <div id="statusbar" style="
        display:flex; align-items:center; gap:16px; padding:4px 12px;
        background:#16162a; border-top:1px solid #333; flex-shrink:0;
        font-size:var(--font-md); color:${tm.textMuted};
      ">
        <span id="zoom-display">Zoom: 100%</span>
        <button id="btn-reset-view" style="
          padding:var(--btn-pad-sm); background:transparent; color:${tm.textMuted};
          border:1px solid #444; border-radius:var(--radius-md); cursor:pointer; font-size:var(--font-sm);
        ">Reset View</button>
      </div>
      <div id="toast-container" style="
        position:fixed; bottom:60px; right:16px; z-index:9999;
        display:flex; flex-direction:column; gap:8px; pointer-events:none;
      "></div>
    `

    const canvasContainer = this.container.querySelector(
      "#canvas-container",
    ) as HTMLElement

    this.canvasView = new CanvasView({
      container: canvasContainer,
      onChange: () => {},
      onOriginChange: () => {},
      onZoomChange: (z) => this.updateZoomDisplay(z),
    })
  }

  private bindEvents(): void {
    const resetBtn = this.container.querySelector(
      "#btn-reset-view",
    ) as HTMLButtonElement
    const saveDropdown = this.container.querySelector(
      "#save-dropdown",
    ) as HTMLElement
    const saveOption = this.container.querySelector(
      "#save-option",
    ) as HTMLElement
    const saveAsOption = this.container.querySelector(
      "#save-as-option",
    ) as HTMLElement

    this.modeSelect.addEventListener("change", () => {
      this.switchMode(this.modeSelect.value as EditorMode)
    })

    this.btnSave.addEventListener("click", (e) => {
      e.stopPropagation()
      saveDropdown.style.display =
        saveDropdown.style.display === "none" ? "block" : "none"
    })

    saveOption.addEventListener("click", () => {
      saveDropdown.style.display = "none"
      this.save(false)
    })

    saveAsOption.addEventListener("click", () => {
      saveDropdown.style.display = "none"
      this.save(true)
    })

    document.addEventListener("click", () => {
      saveDropdown.style.display = "none"
    })

    this.btnImport.addEventListener("click", async () => {
      if (!this.getEditor().canImport()) {
        this.showToast("Save or clear current work before opening")
        return
      }
      try {
        const [fileHandle] = await window.showOpenFilePicker({
          types: [{ description: "TypeScript", accept: { "text/typescript": [".ts"] } }],
        })
        const file = await fileHandle.getFile()
        const text = await file.text()
        const firstLine = text.split("\n")[0]?.trim()
        console.log("[App.import] firstLine:", firstLine)

        let ok = false
        if (firstLine === "// Attacking Editor") {
          console.log("[App.import] detected attacking editor file")
          this.switchMode("attack")
          ok = this.attackingEditor.importFromText(text)
        } else if (firstLine === "// Movement Editor") {
          console.log("[App.import] detected movement editor file")
          this.switchMode("movement")
          ok = this.movementEditor.importFromText(text)
        } else {
          console.log("[App.import] no editor header, trying current mode:", this.currentMode)
          ok = this.getEditor().importFromText(text)
        }

        if (!ok) {
          console.warn("[App.import] importFromText returned false")
          this.showToast("Failed to parse .ts file")
          return
        }
        console.log("[App.import] file loaded successfully")
        this.currentFileHandle = fileHandle
      } catch (err) {
        if ((err as DOMException).name === "AbortError") return
        console.error("[App.import] open failed", err)
        this.showToast("Failed to open file")
      }
    })

    resetBtn.addEventListener("click", () => {
      this.canvasView.resetView()
    })

    this.btnTheme.addEventListener("click", () => {
      this.themeManager.setTheme(this.themeManager.isLight ? "dark" : "light")
      this.applyTheme()
    })
  }

  private async save(forceSaveAs: boolean): Promise<void> {
    const editor = this.currentMode === "attack"
      ? this.attackingEditor
      : this.movementEditor

    let handle = this.currentFileHandle

    if (!handle || forceSaveAs) {
      try {
        handle = await window.showSaveFilePicker({
          suggestedName: `${editor.getExportFileName()}.ts`,
          types: [{
            description: "TypeScript",
            accept: { "text/typescript": [".ts"] },
          }],
        })
      } catch (err) {
        if ((err as DOMException).name === "AbortError") return
        // fallback: download instead
        editor.export()
        return
      }
      if (!forceSaveAs) this.currentFileHandle = handle
    }

    if (!handle) {
      editor.export()
      return
    }
    try {
      const writable = await handle.createWritable()
      await writable.write(editor.getExportContent())
      await writable.close()
    } catch {
      editor.export()
    }
  }

  private getEditor():
    | AttackingEditor
    | MovementEditor
    | {
        canImport(): boolean
        importFromText(_: string): boolean
        export(): void
      } {
    return this.currentMode === "attack"
      ? this.attackingEditor
      : this.movementEditor
  }

  private switchMode(mode: EditorMode): void {
    // Stop previous mode's preview
    if (this.currentMode === "attack") {
      this.attackingEditor.stopPreview()
    } else {
      this.movementEditor.stopPreview()
    }

    this.currentMode = mode
    this.modeSelect.value = mode
    this.canvasView.setBackgroundImage(null)
    this.canvasView.setOverlays([], 0)

    this.rightPanel.innerHTML = ""
    this.animationTables.innerHTML = ""

    if (mode === "attack") {
      this.propertyInspector = new PropertyInspector(this.rightPanel, {
        mode: "attack",
        onTransformChange: (s) => this.attackingEditor.onTransformChange(s),
        onOriginChange: (x, y) => this.attackingEditor.onOriginChange(x, y),
        onBaseRotationChange: (d) =>
          this.attackingEditor.onBaseRotationChange(d),
        onPhaseChange: (p) => this.attackingEditor.onPhaseChange(p),
        onGuardFpsChange: (fps) => this.attackingEditor.onGuardFpsChange(fps),
        onWindupChange: (ms) => this.attackingEditor.onWindupChange(ms),
        onActiveTimeChange: (ms) => this.attackingEditor.onActiveTimeChange(ms),
        onRecoverChange: (ms) => this.attackingEditor.onRecoverChange(ms),
        onGuardDurationChange: (ms) =>
          this.attackingEditor.onGuardDurationChange(ms),
        onHurtboxChange: (hb) => this.attackingEditor.onHurtboxChange(hb),
      })
      this.attackingEditor.activate(
        this.animationTables,
        this.rightPanel,
        this.propertyInspector,
      )
      this.attackingEditor.onSpriteChange = () => this.refreshImportState()
      this.canvasView.updateCallbacks({
        onChange: (s) => this.attackingEditor.onTransformChange(s),
        onOriginChange: (x, y) => this.attackingEditor.onOriginChange(x, y),
        onHurtboxSelect: (i) => this.attackingEditor.selectHurtbox(i),
        onHurtboxChange: (hb) => this.attackingEditor.onHurtboxChange(hb),
      })
    } else {
      this.propertyInspector = new PropertyInspector(this.rightPanel, {
        mode: "movement",
        onTransformChange: (s) => this.movementEditor.onTransformChange(s),
        onOriginChange: (x, y) => this.movementEditor.onOriginChange(x, y),
        onBaseRotationChange: (d) =>
          this.movementEditor.onBaseRotationChange(d),
        onFpsChange: (fps) => this.movementEditor.onFpsChange(fps),
        onHurtboxChange: (hb) => this.movementEditor.onHurtboxChange(hb),
      })
      this.movementEditor.activate(
        this.animationTables,
        this.rightPanel,
        this.propertyInspector,
      )
      this.movementEditor.onSpriteChange = () => this.refreshImportState()
      this.canvasView.updateCallbacks({
        onChange: (s) => this.movementEditor.onTransformChange(s),
        onOriginChange: (x, y) => this.movementEditor.onOriginChange(x, y),
        onHurtboxSelect: (i) => this.movementEditor.selectHurtbox(i),
        onHurtboxChange: (hb) => this.movementEditor.onHurtboxChange(hb),
      })
    }

    this.updateModeIndicator()
    this.refreshImportState()
  }

  private refreshImportState(): void {
    const editor = this.getEditor()
    const disabled = !editor.canImport()
    this.btnImport.disabled = disabled
    this.btnImport.style.background = disabled ? "#c33" : "#a82"
    this.btnImport.style.cursor = disabled ? "not-allowed" : "pointer"
  }

  private showToast(msg: string): void {
    const container = this.container.querySelector("#toast-container")
    if (!container) return
    const el = document.createElement("div")
    el.textContent = msg
    el.style.cssText =
      "padding:8px 14px;background:#333;color:#eee;border-radius:6px;font-size:var(--font-lg);" +
      "box-shadow:0 2px 8px rgba(0,0,0,0.4);transition:opacity 0.3s;opacity:0;pointer-events:auto;"
    container.appendChild(el)
    requestAnimationFrame(() => {
      el.style.opacity = "1"
    })
    setTimeout(() => {
      el.style.opacity = "0"
      setTimeout(() => el.remove(), 300)
    }, 2500)
  }

  private updateModeIndicator(): void {
    if (this.currentMode === "attack") {
      this.modeIndicator.textContent =
        "Editing attack frames — position, rotate & scale sprites per frame"
    } else {
      this.modeIndicator.textContent =
        "Editing movement frames — position, rotate & scale sprites per animation"
    }
  }

  private updateZoomDisplay(zoom: number): void {
    this.zoomDisplay.textContent = `Zoom: ${Math.round(zoom * 100)}%`
  }

  private applyTheme(): void {
    const tm = this.themeManager
    this.canvasView.setTheme(tm)
    this.modeIndicator.style.color = tm.textMuted
    this.resetViewBtn.style.color = tm.textMuted
    this.btnTheme.textContent = tm.isLight ? "🌙 Dark" : "☀ Light"
    this.btnTheme.style.color = tm.buttonText
    this.btnTheme.style.background = tm.buttonSwitchBg
    this.btnTheme.style.borderColor = tm.toolbarBorderBottom
    this.topBar.style.background = tm.toolbarBg
    this.topBar.style.borderBottomColor = tm.toolbarBorderBottom
    this.topBar.style.color = tm.toolbarColor
    this.statusbarEl.style.background = tm.statusbarBg
    this.statusbarEl.style.borderTopColor = tm.statusbarBorderTop
    this.statusbarEl.style.color = tm.labelColor
    const styleId = "anim-theme-style"
    let styleEl = document.getElementById(styleId) as HTMLStyleElement | null
    if (!styleEl) {
      styleEl = document.createElement("style")
      styleEl.id = styleId
      document.head.appendChild(styleEl)
    }
    styleEl.textContent = `
:root {
  --font-sm: ${tm.fontSm};
  --font-md: ${tm.fontMd};
  --font-lg: ${tm.fontLg};
  --radius-sm: ${tm.radiusSm};
  --radius-md: ${tm.radiusMd};
  --radius-lg: ${tm.radiusLg};
  --input-pad: ${tm.inputPad};
  --input-pad-sm: ${tm.inputPadSm};
  --btn-pad-sm: ${tm.btnPadSm};
  --btn-pad-md: ${tm.btnPadMd};
  --gap-sm: ${tm.gapSm};
  --gap-md: ${tm.gapMd};
  --gap-lg: ${tm.gapLg};
  --cell-pad: ${tm.cellPad};
  --nav-pad: ${tm.navPad};
  --section-pad: ${tm.sectionPad};
  --btn-color: ${tm.buttonText};
  --label-color: ${tm.labelColor};
}
.anim-row-highlight { background: ${tm.rowSelected}; }
`

    this.rightPanel.style.background = tm.panelBg
    this.rightPanel.style.borderLeftColor = tm.panelBorderLeft
    this.rightPanel.style.color = tm.textPrimary
  }
}
