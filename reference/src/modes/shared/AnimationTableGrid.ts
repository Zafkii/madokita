import type { ToolbarHost } from "./ToolbarHost"
import type { EditorHurtboxData } from "../../types/MovementTypes"

export class AnimationTableGrid {
  private host: ToolbarHost
  private containerEl: HTMLElement

  private spriteSelect!: HTMLSelectElement

  private animTbody!: HTMLTableSectionElement

  private spriteTbody!: HTMLTableSectionElement
  private baseSpriteRow!: HTMLTableRowElement
  private spriteRows: HTMLTableRowElement[] = []
  private baseBrowseBtn!: HTMLButtonElement
  private baseSpriteFrameInput!: HTMLInputElement
  private baseSpriteFrameTotal!: HTMLSpanElement
  private spriteFrameInputs: HTMLInputElement[] = []
  private spriteFrameTotals: HTMLSpanElement[] = []

  private hurtboxTbody!: HTMLTableSectionElement
  private hurtboxRows: HTMLTableRowElement[] = []

  private animFrameInputs: Map<string, HTMLInputElement> = new Map()
  private animFrameTotals: Map<string, HTMLSpanElement> = new Map()
  private animRowEls: Map<string, HTMLTableRowElement> = new Map()

  constructor(host: ToolbarHost, containerEl: HTMLElement) {
    this.host = host
    this.containerEl = containerEl
  }

  init(): void {
    this.spriteFrameInputs = []
    this.spriteFrameTotals = []
    this.animFrameInputs.clear()
    this.animFrameTotals.clear()
    this.animRowEls.clear()
    this.containerEl.innerHTML = ""
    this.containerEl.style.cssText = "display:flex;flex-direction:column;gap:var(--gap-sm);"

    const noSpin = document.createElement("style")
    noSpin.textContent =
      "input[type=number]::-webkit-outer-spin-button,input[type=number]::-webkit-inner-spin-button{-webkit-appearance:none;margin:0}input[type=number]{-moz-appearance:textfield}"
    this.containerEl.appendChild(noSpin)

    this.spriteSelect = document.createElement("select")
    this.spriteSelect.style.display = "none"

    const cols = document.createElement("div")
    cols.style.cssText = "display:grid;grid-template-columns:1fr 1fr;gap:var(--gap-md);"

    cols.appendChild(this.buildAnimationsSection())

    cols.appendChild(this.buildHurtboxesSection())

    cols.appendChild(this.buildSpritesSection())

    cols.appendChild(this.buildHitboxesSection())

    this.containerEl.appendChild(cols)

    for (let i = 0; i < this.host.additionalSprites.length; i++) {
      this.addSpriteRow(i)
    }
    this.updateActiveSpriteRow(this.host.currentSpriteIndex)

    this.syncBaseSpriteTotalFrames(this.host.baseSpriteTotalFrames)
    this.syncBaseSpriteFrameIdx(this.host.baseSpriteFrameIndex)
  }

  private buildAnimationsSection(): HTMLElement {
    const section = this.group("rgba(255,120,150,0.1)")

    const headerRow = document.createElement("div")
    headerRow.style.cssText =
      "display:flex;align-items:center;gap:var(--gap-sm);margin-bottom:2px;"

    const addAnimBtn = document.createElement("button")
    addAnimBtn.textContent = "+ Add Animation"
    addAnimBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#2a4;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    addAnimBtn.addEventListener("click", () => {
      let name = "new"
      let i = 1
      while (this.host.getAnimations()[name]) name = `new_${++i}`
      this.host.ensureAnimExists(name)
      this.host.selectAnimation(name)
      this.rebuildAnimRows()
    })
    headerRow.appendChild(addAnimBtn)

    const removeAnimBtn = document.createElement("button")
    removeAnimBtn.textContent = "− Remove Animation"
    removeAnimBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#c33;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    removeAnimBtn.addEventListener("click", () => {
      const anims = this.host.getAnimations()
      if (Object.keys(anims).length <= 1) return
      this.host.removeAnimation(this.host.currentAnimName)
      this.rebuildAnimRows()
    })
    headerRow.appendChild(removeAnimBtn)

    section.appendChild(headerRow)

    const table = document.createElement("table")
    table.style.cssText = "border-collapse:collapse;font-size:var(--font-sm);width:100%;"
    const thead = document.createElement("thead")
    const thr = document.createElement("tr")
    const animTh1 = document.createElement("th")
    animTh1.textContent = "Animation"
    animTh1.style.cssText = this.thStyle
    thr.appendChild(animTh1)
    const animTh2 = document.createElement("th")
    animTh2.textContent = "Frame"
    animTh2.style.cssText = this.thStyle
    thr.appendChild(animTh2)
    thead.appendChild(thr)
    table.appendChild(thead)
    this.animTbody = document.createElement("tbody")
    table.appendChild(this.animTbody)
    section.appendChild(table)

    this.rebuildAnimRows()

    return section
  }

  private buildAnimRow(name: string): void {
    const tr = document.createElement("tr")
    tr.dataset["animName"] = name
    tr.style.cssText = "cursor:pointer;"
    tr.addEventListener("click", () => this.host.selectAnimation(name))
    this.animRowEls.set(name, tr)

    const nameCell = document.createElement("td")
    nameCell.style.cssText =
      "padding:var(--cell-pad);white-space:nowrap;overflow:hidden;max-width:80px;"
    const nameInput = document.createElement("input")
    nameInput.type = "text"
    nameInput.value = name
    nameInput.style.cssText =
      "width:100%;box-sizing:border-box;background:transparent;color:var(--btn-color,#e0e0e0);" +
      "border:none;outline:none;font-weight:600;font-size:var(--font-sm);font-family:inherit;padding:0;"
    nameInput.addEventListener("change", () => {
      const newName = nameInput.value.trim()
      if (!newName || newName === name) {
        nameInput.value = name
        return
      }
      if (!this.host.renameAnimation(name, newName)) {
        nameInput.value = name
      } else {
        this.rebuildAnimRows()
      }
    })
    nameInput.addEventListener("click", (e) => e.stopPropagation())
    nameInput.addEventListener("mousedown", (e) => e.stopPropagation())
    nameCell.appendChild(nameInput)
    tr.appendChild(nameCell)

    const frameCell = document.createElement("td")
    frameCell.style.cssText = "padding:var(--cell-pad);white-space:nowrap;"

    const prevBtn = document.createElement("button")
    prevBtn.textContent = "◀"
    prevBtn.style.cssText =
      "padding:var(--nav-pad);background:#2a2a4a;color:var(--btn-color,#e0e0e0);border:1px solid #444;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);line-height:1.4;"
    prevBtn.addEventListener("click", (e) => {
      e.stopPropagation()
      const total = this.host.getAnimFrameTotal(name)
      if (total <= 0) return
      if (this.host.currentAnimName !== name) this.host.selectAnimation(name)
      const cur = this.host.currentFrameIndex
      this.host.selectFrame(cur > 0 ? cur - 1 : 0)
    })
    frameCell.appendChild(prevBtn)

    const frameInput = document.createElement("input")
    frameInput.type = "number"
    frameInput.min = "0"
    frameInput.readOnly = true
    frameInput.value = "0"
    frameInput.style.cssText =
      "width:24px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);text-align:center;"
    this.animFrameInputs.set(name, frameInput)
    frameCell.appendChild(frameInput)

    const totalSpan = document.createElement("span")
    totalSpan.textContent = "/ 0"
    totalSpan.style.cssText =
      "font-size:var(--font-sm);color:var(--label-color,#888);margin:0 1px;"
    this.animFrameTotals.set(name, totalSpan)
    frameCell.appendChild(totalSpan)

    const nextBtn = document.createElement("button")
    nextBtn.textContent = "▶"
    nextBtn.style.cssText =
      "padding:var(--nav-pad);background:#2a2a4a;color:var(--btn-color,#e0e0e0);border:1px solid #444;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);line-height:1.4;"
    nextBtn.addEventListener("click", (e) => {
      e.stopPropagation()
      const total = this.host.getAnimFrameTotal(name)
      if (total <= 0) return
      if (this.host.currentAnimName !== name) this.host.selectAnimation(name)
      const cur = this.host.currentFrameIndex
      this.host.selectFrame(Math.min(total - 1, cur + 1))
    })
    frameCell.appendChild(nextBtn)

    frameCell.appendChild(document.createTextNode("\u00A0"))

    const addFBtn = document.createElement("button")
    addFBtn.textContent = "+ Add Frame"
    addFBtn.style.cssText =
      "padding:var(--nav-pad);background:#2a4;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);line-height:1.4;font-weight:600;white-space:nowrap;"
    addFBtn.addEventListener("click", (e) => {
      e.stopPropagation()
      this.host.addFrameToAnim(name)
      this.syncSingleAnimFrame(name)
    })
    frameCell.appendChild(addFBtn)

    const removeFBtn = document.createElement("button")
    removeFBtn.textContent = "− Remove Frame"
    removeFBtn.style.cssText =
      "padding:var(--nav-pad);background:#c33;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);line-height:1.4;font-weight:600;white-space:nowrap;"
    removeFBtn.addEventListener("click", (e) => {
      e.stopPropagation()
      const total = this.host.getAnimFrameTotal(name)
      if (total <= 1) return
      this.host.removeFrameFromAnim(name)
      this.syncSingleAnimFrame(name)
    })
    frameCell.appendChild(removeFBtn)

    tr.appendChild(frameCell)
    this.animTbody.appendChild(tr)

    this.syncSingleAnimFrame(name)
  }

  rebuildAnimRows(): void {
    this.animFrameInputs.clear()
    this.animFrameTotals.clear()
    this.animRowEls.clear()
    this.animTbody.innerHTML = ""
    for (const name of this.host.getAllAnimationNames()) {
      this.buildAnimRow(name)
    }
    this.highlightCurrentAnimRow()
  }

  private highlightCurrentAnimRow(): void {
    for (const [, tr] of this.animRowEls) {
      tr.classList.remove("anim-row-highlight")
      tr.style.fontWeight = "400"
    }
    const tr = this.animRowEls.get(this.host.currentAnimName)
    if (tr) {
      tr.classList.add("anim-row-highlight")
      tr.style.fontWeight = "600"
    }
  }

  private syncSingleAnimFrame(name: string): void {
    const input = this.animFrameInputs.get(name)
    const totalSpan = this.animFrameTotals.get(name)
    const total = this.host.getAnimFrameTotal(name)
    let frameIdx = 0
    if (name === this.host.currentAnimName) {
      frameIdx = this.host.currentFrameIndex
    }
    if (input) input.value = String(total > 0 ? frameIdx + 1 : 0)
    if (totalSpan) totalSpan.textContent = `/ ${total}`
  }

  private buildHurtboxesSection(): HTMLElement {
    const section = this.group("rgba(255,200,100,0.1)")

    const headerRow = document.createElement("div")
    headerRow.style.cssText =
      "display:flex;align-items:center;gap:var(--gap-sm);margin-bottom:2px;"

    const addHurtboxBtn = document.createElement("button")
    addHurtboxBtn.textContent = "+ Add Hurtbox"
    addHurtboxBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#2a4;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    addHurtboxBtn.addEventListener("click", () => {
      this.host.addHurtbox()
      this.rebuildHurtboxRows()
    })
    headerRow.appendChild(addHurtboxBtn)

    const removeHurtboxBtn = document.createElement("button")
    removeHurtboxBtn.textContent = "− Remove Hurtbox"
    removeHurtboxBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#c33;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    removeHurtboxBtn.addEventListener("click", () => {
      const hurtboxes = this.host.getCurrentFrameHurtboxes()
      if (hurtboxes.length <= 0) return
      this.host.removeHurtbox(hurtboxes.length - 1)
      this.rebuildHurtboxRows()
    })
    headerRow.appendChild(removeHurtboxBtn)

    const repeatBtn = document.createElement("button")
    repeatBtn.textContent = "◀ Copy Prev Hurtbox"
    repeatBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#48a;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    repeatBtn.addEventListener("click", () => {
      this.host.repeatPreviousHurtbox()
      this.rebuildHurtboxRows()
    })
    headerRow.appendChild(repeatBtn)

    section.appendChild(headerRow)

    const table = document.createElement("table")
    table.style.cssText = "border-collapse:collapse;font-size:var(--font-sm);width:100%;"
    const thead = document.createElement("thead")
    const thr = document.createElement("tr")
    for (const h of ["Width", "Height", "Damage Multiplier"]) {
      const th = document.createElement("th")
      th.textContent = h
      th.style.cssText = this.thStyle
      thr.appendChild(th)
    }
    thead.appendChild(thr)
    table.appendChild(thead)
    this.hurtboxTbody = document.createElement("tbody")
    table.appendChild(this.hurtboxTbody)
    section.appendChild(table)

    this.rebuildHurtboxRows()

    return section
  }

  private rebuildHurtboxRows(): void {
    this.refreshHurtboxes()
  }

  private buildHurtboxRow(idx: number, hb: EditorHurtboxData): void {
    const tr = document.createElement("tr")
    tr.style.cssText = "cursor:pointer;"
    tr.addEventListener("click", () => this.host.selectHurtbox(idx))

    const wCell = document.createElement("td")
    wCell.style.cssText = "padding:var(--input-pad-sm);text-align:center;"
    const wInput = document.createElement("input")
    wInput.type = "number"
    wInput.value = String(hb[0])
    wInput.style.cssText =
      "width:50px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);text-align:center;"
    wInput.addEventListener("change", () => {
      this.host.updateHurtbox(idx, parseInt(wInput.value) || 0, hb[1], hb[7])
    })
    wCell.appendChild(wInput)
    tr.appendChild(wCell)

    const hCell = document.createElement("td")
    hCell.style.cssText = "padding:var(--input-pad-sm);text-align:center;"
    const hInput = document.createElement("input")
    hInput.type = "number"
    hInput.value = String(hb[1])
    hInput.style.cssText =
      "width:50px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);text-align:center;"
    hInput.addEventListener("change", () => {
      this.host.updateHurtbox(idx, hb[0], parseInt(hInput.value) || 0, hb[7])
    })
    hCell.appendChild(hInput)
    tr.appendChild(hCell)

    const dCell = document.createElement("td")
    dCell.style.cssText = "padding:var(--input-pad-sm);text-align:center;"
    const dInput = document.createElement("input")
    dInput.type = "number"
    dInput.step = "0.1"
    dInput.min = "0"
    dInput.value = String(hb[7])
    dInput.style.cssText =
      "width:50px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);text-align:center;"
    dInput.addEventListener("change", () => {
      this.host.updateHurtbox(idx, hb[0], hb[1], parseFloat(dInput.value) || 1)
    })
    dCell.appendChild(dInput)
    tr.appendChild(dCell)

    this.hurtboxTbody.appendChild(tr)
    this.hurtboxRows.push(tr)
  }

  private buildSpritesSection(): HTMLElement {
    const section = this.group("rgba(200,120,255,0.1)")

    const headerRow = document.createElement("div")
    headerRow.style.cssText =
      "display:flex;align-items:center;gap:var(--gap-sm);margin-bottom:2px;"

    const addSpriteBtn = document.createElement("button")
    addSpriteBtn.textContent = "+ Add Sprite"
    addSpriteBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#2a4;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    addSpriteBtn.addEventListener("click", () => {
      const idx = this.host.additionalSprites.length
      this.host.additionalSprites.push({
        image: null,
        path: "",
        frameW: 145,
        frameH: 145,
        totalFrames: 1,
        frameIdx: 0,
        originX: 0.5,
        originY: 0.5,
      })
      this.addSpriteRow(idx)
      this.host.selectSprite(idx + 1)
    })
    headerRow.appendChild(addSpriteBtn)

    const removeSpriteBtn = document.createElement("button")
    removeSpriteBtn.textContent = "− Remove Sprite"
    removeSpriteBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#c33;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;"
    removeSpriteBtn.addEventListener("click", () => {
      const idx = this.host.currentSpriteIndex
      if (idx <= 0) return
      const spriteIdx = idx - 1
      if (spriteIdx >= this.host.additionalSprites.length) return
      this.host.additionalSprites.splice(spriteIdx, 1)
      this.reindexSpriteRows()
      this.host.selectSprite(0)
    })
    headerRow.appendChild(removeSpriteBtn)

    section.appendChild(headerRow)

    const table = document.createElement("table")
    table.style.cssText = "border-collapse:collapse;font-size:var(--font-sm);width:100%;"
    const thead = document.createElement("thead")
    const thr = document.createElement("tr")
    for (const h of ["Sprite", "File", "Width", "Height", "Frame"]) {
      const th = document.createElement("th")
      th.textContent = h
      th.style.cssText = this.thStyle
      thr.appendChild(th)
    }
    thead.appendChild(thr)
    table.appendChild(thead)
    this.spriteTbody = document.createElement("tbody")
    table.appendChild(this.spriteTbody)
    section.appendChild(table)

    const baseRow = document.createElement("tr")
    baseRow.dataset["spriteIdx"] = "0"
    baseRow.style.cssText = "cursor:pointer;"
    baseRow.addEventListener("click", () => this.host.selectSprite(0))

    const baseNameCell = document.createElement("td")
    baseNameCell.textContent = "Base"
    baseNameCell.style.cssText =
      "font-weight:600;padding:var(--cell-pad);white-space:nowrap;"
    baseRow.appendChild(baseNameCell)

    const baseFileCell = document.createElement("td")
    baseFileCell.style.cssText = "padding:var(--cell-pad);white-space:nowrap;"
    this.baseBrowseBtn = document.createElement("button")
    this.baseBrowseBtn.textContent = this.host.baseSpritePath
      ? this.host.baseSpritePath.split(/[/\\]/).pop()!
      : "Browse"
    this.baseBrowseBtn.style.cssText =
      "padding:var(--cell-pad);background:#2a2a4a;color:var(--btn-color,#e0e0e0);border:1px solid #444;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);max-width:90px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;"
    const baseFileInput = document.createElement("input")
    baseFileInput.type = "file"
    baseFileInput.accept = ".png"
    baseFileInput.style.display = "none"
    this.baseBrowseBtn.addEventListener("click", () => baseFileInput.click())
    baseFileInput.addEventListener("change", () => {
      const file = baseFileInput.files?.[0]
      if (!file) return
      const img = new Image()
      img.onload = () => {
        this.host.loadBaseSpritesheet(img)
        this.host.baseSpritePath = file.name
        this.baseBrowseBtn.textContent = file.name
      }
      img.src = URL.createObjectURL(file)
      baseFileInput.value = ""
    })
    baseFileCell.appendChild(this.baseBrowseBtn)
    baseFileCell.appendChild(baseFileInput)
    baseRow.appendChild(baseFileCell)

    const baseWC = document.createElement("td")
    baseWC.style.cssText = "padding:var(--cell-pad);"
    const baseW = document.createElement("input")
    baseW.type = "number"
    baseW.value = String(this.host.baseSpriteFrameW)
    baseW.style.cssText =
      "width:40px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);"
    baseW.addEventListener("change", () => {
      this.host.baseSpriteFrameW = Math.max(1, parseInt(baseW.value) || 256)
      if (this.host.baseSpritesheet) this.host.calcBaseSpriteTotalFrames()
    })
    baseWC.appendChild(baseW)
    baseRow.appendChild(baseWC)

    const baseHC = document.createElement("td")
    baseHC.style.cssText = "padding:var(--cell-pad);"
    const baseH = document.createElement("input")
    baseH.type = "number"
    baseH.value = String(this.host.baseSpriteFrameH)
    baseH.style.cssText =
      "width:40px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);"
    baseH.addEventListener("change", () => {
      this.host.baseSpriteFrameH = Math.max(1, parseInt(baseH.value) || 256)
      if (this.host.baseSpritesheet) this.host.calcBaseSpriteTotalFrames()
    })
    baseHC.appendChild(baseH)
    baseRow.appendChild(baseHC)

    const baseFC = document.createElement("td")
    baseFC.style.cssText = "padding:var(--cell-pad);white-space:nowrap;"
    const basePrev = this.btn("◀", () => {
      this.host.setBaseSpriteFrame(
        Math.max(0, this.host.baseSpriteFrameIndex - 1),
      )
    })
    this.baseSpriteFrameInput = document.createElement("input")
    this.baseSpriteFrameInput.type = "number"
    this.baseSpriteFrameInput.min = "0"
    this.baseSpriteFrameInput.value = "0"
    this.baseSpriteFrameInput.style.cssText =
      "width:24px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);text-align:center;"
    this.baseSpriteFrameInput.addEventListener("change", () => {
      this.host.setBaseSpriteFrame(
        Math.max(0, parseInt(this.baseSpriteFrameInput.value) || 0),
      )
    })
    this.baseSpriteFrameInput.addEventListener("wheel", (e) => {
      e.preventDefault()
      this.host.setBaseSpriteFrame(
        Math.max(
          0,
          Math.min(
            this.host.baseSpriteTotalFrames - 1,
            this.host.baseSpriteFrameIndex + (e.deltaY < 0 ? 1 : -1),
          ),
        ),
      )
    })
    this.baseSpriteFrameTotal = document.createElement("span")
    this.baseSpriteFrameTotal.textContent = "/ 1"
    this.baseSpriteFrameTotal.style.cssText =
      "font-size:var(--font-sm);color:var(--label-color,#888);margin:0 1px;"
    const baseNext = this.btn("▶", () => {
      this.host.setBaseSpriteFrame(
        Math.min(
          this.host.baseSpriteTotalFrames - 1,
          this.host.baseSpriteFrameIndex + 1,
        ),
      )
    })
    baseFC.appendChild(basePrev)
    baseFC.appendChild(this.baseSpriteFrameInput)
    baseFC.appendChild(this.baseSpriteFrameTotal)
    baseFC.appendChild(baseNext)
    baseRow.appendChild(baseFC)

    this.spriteTbody.appendChild(baseRow)
    this.baseSpriteRow = baseRow

    return section
  }

  private buildHitboxesSection(): HTMLElement {
    const section = this.group("rgba(100,200,255,0.1)")

    const headerRow = document.createElement("div")
    headerRow.style.cssText =
      "display:flex;align-items:center;gap:var(--gap-sm);margin-bottom:2px;"

    const addHitboxBtn = document.createElement("button")
    addHitboxBtn.textContent = "+ Add Hitbox"
    addHitboxBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#2a4;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;opacity:0.4;"
    headerRow.appendChild(addHitboxBtn)

    const removeHitboxBtn = document.createElement("button")
    removeHitboxBtn.textContent = "− Remove Hitbox"
    removeHitboxBtn.style.cssText =
      "padding:var(--btn-pad-sm);background:#c33;color:#fff;border:none;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);font-weight:600;white-space:nowrap;opacity:0.4;"
    headerRow.appendChild(removeHitboxBtn)

    section.appendChild(headerRow)

    const placeholder = document.createElement("div")
    placeholder.textContent = "nothing to show"
    placeholder.style.cssText =
      "font-size:var(--font-sm);color:var(--label-color,#666);padding:8px 4px;text-align:center;"
    section.appendChild(placeholder)

    return section
  }

  addSpriteRow(idx: number): void {
    this.rebuildSpriteSelect()
    const sp = this.host.additionalSprites[idx]
    const tr = document.createElement("tr")
    tr.dataset["spriteIdx"] = String(idx + 1)
    tr.style.cssText = "cursor:pointer;"
    tr.addEventListener("click", () => this.host.selectSprite(idx + 1))

    const nameCell = document.createElement("td")
    nameCell.textContent = `Added${idx + 1}`
    nameCell.style.cssText =
      "font-weight:600;padding:var(--cell-pad);white-space:nowrap;"
    tr.appendChild(nameCell)

    const fileCell = document.createElement("td")
    fileCell.style.cssText = "padding:var(--cell-pad);white-space:nowrap;"
    const browseBtn = document.createElement("button")
    browseBtn.textContent = sp.path ? sp.path.split(/[/\\]/).pop()! : "Browse"
    browseBtn.style.cssText =
      "padding:var(--cell-pad);background:#2a2a4a;color:var(--btn-color,#e0e0e0);border:1px solid #444;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);max-width:90px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;"
    const fileInput = document.createElement("input")
    fileInput.type = "file"
    fileInput.accept = ".png"
    fileInput.style.display = "none"
    browseBtn.addEventListener("click", () => fileInput.click())
    fileInput.addEventListener("change", () => {
      const file = fileInput.files?.[0]
      if (!file) return
      const img = new Image()
      img.onload = () => {
        sp.path = file.name
        this.host.addAdditionalSprite(img, idx)
        browseBtn.textContent = file.name
      }
      img.src = URL.createObjectURL(file)
      fileInput.value = ""
    })
    fileCell.appendChild(browseBtn)
    fileCell.appendChild(fileInput)
    tr.appendChild(fileCell)

    const wc = document.createElement("td")
    wc.style.cssText = "padding:var(--cell-pad);"
    const ww = document.createElement("input")
    ww.type = "number"
    ww.value = String(sp.frameW)
    ww.style.cssText =
      "width:40px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);"
    ww.addEventListener("change", () => {
      sp.frameW = Math.max(1, parseInt(ww.value) || 145)
      if (sp.image) this.host.calcSpriteTotalFrames(idx)
    })
    wc.appendChild(ww)
    tr.appendChild(wc)

    const hc = document.createElement("td")
    hc.style.cssText = "padding:var(--cell-pad);"
    const wh = document.createElement("input")
    wh.type = "number"
    wh.value = String(sp.frameH)
    wh.style.cssText =
      "width:40px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);"
    wh.addEventListener("change", () => {
      sp.frameH = Math.max(1, parseInt(wh.value) || 145)
      if (sp.image) this.host.calcSpriteTotalFrames(idx)
    })
    hc.appendChild(wh)
    tr.appendChild(hc)

    const fc = document.createElement("td")
    fc.style.cssText = "padding:var(--cell-pad);white-space:nowrap;"
    const prev = this.btn("◀", () => {
      sp.frameIdx = Math.max(0, sp.frameIdx - 1)
      this.host.updateSpriteFrame(idx)
      frameInput.value = String(sp.frameIdx)
    })
    const frameInput = document.createElement("input")
    frameInput.type = "number"
    frameInput.min = "0"
    frameInput.value = String(sp.frameIdx)
    frameInput.style.cssText =
      "width:24px;padding:var(--input-pad-sm);background:#1a1a2e;color:#e0e0e0;border:1px solid #444;border-radius:var(--radius-sm);font-size:var(--font-sm);text-align:center;"
    frameInput.dataset["spriteIdx"] = String(idx)
    frameInput.addEventListener("change", () => {
      sp.frameIdx = Math.max(0, parseInt(frameInput.value) || 0)
      this.host.updateSpriteFrame(idx)
      frameInput.value = String(sp.frameIdx)
    })
    const frameTotal = document.createElement("span")
    frameTotal.textContent = `/ ${sp.totalFrames}`
    frameTotal.style.cssText =
      "font-size:var(--font-sm);color:var(--label-color,#888);margin:0 1px;"
    const next = this.btn("▶", () => {
      sp.frameIdx = Math.min(sp.totalFrames - 1, sp.frameIdx + 1)
      this.host.updateSpriteFrame(idx)
      frameInput.value = String(sp.frameIdx)
    })
    fc.appendChild(prev)
    fc.appendChild(frameInput)
    fc.appendChild(frameTotal)
    fc.appendChild(next)
    tr.appendChild(fc)

    this.spriteFrameInputs[idx] = frameInput
    this.spriteFrameTotals[idx] = frameTotal

    this.spriteTbody.appendChild(tr)
    this.spriteRows.push(tr)
  }

  rebuildSpriteGroups(): void {
    this.spriteFrameInputs = []
    this.spriteFrameTotals = []
    this.reindexSpriteRows()
  }

  private reindexSpriteRows(): void {
    for (const tr of this.spriteRows) tr.remove()
    this.spriteRows = []
    for (let i = 0; i < this.host.additionalSprites.length; i++) {
      this.addSpriteRow(i)
    }
    this.rebuildSpriteSelect()
  }

  private updateActiveSpriteRow(idx: number): void {
    this.baseSpriteRow.classList.toggle("anim-row-highlight", idx === 0)
    this.highlightRow(this.spriteRows, idx - 1)
  }

  private highlightHurtboxRow(): void {
    this.highlightRow(
      this.hurtboxRows,
      this.host.selectedHurtboxIndex,
    )
  }

  private highlightRow(
    rows: HTMLTableRowElement[],
    idx: number,
  ): void {
    for (const tr of rows) tr.classList.remove("anim-row-highlight")
    if (idx >= 0 && idx < rows.length) rows[idx].classList.add("anim-row-highlight")
  }

  private get thStyle(): string {
    return "padding:var(--cell-pad);color:var(--label-color,#888);font-weight:600;text-align:left;font-size:var(--font-sm);border-bottom:1px solid #444;"
  }

  rebuild(): void {
    this.init()
  }

  syncAnimName(_name: string): void {
    this.highlightCurrentAnimRow()
  }

  syncHurtboxSelection(_index: number): void {
    this.highlightHurtboxRow()
  }

  refreshHurtboxes(): void {
    this.host.selectedHurtboxIndex = Math.min(
      this.host.selectedHurtboxIndex,
      this.host.getCurrentFrameHurtboxes().length - 1,
    )
    this.hurtboxRows = []
    this.hurtboxTbody.innerHTML = ""
    const hurtboxes = this.host.getCurrentFrameHurtboxes()
    for (let i = 0; i < hurtboxes.length; i++) {
      this.buildHurtboxRow(i, hurtboxes[i])
    }
    this.highlightHurtboxRow()
  }

  syncSpriteIndex(idx: number): void {
    this.spriteSelect.value = String(idx)
    this.updateActiveSpriteRow(idx)
  }

  syncBaseSpritePath(path: string): void {
    if (this.baseBrowseBtn)
      this.baseBrowseBtn.textContent = path
        ? path.split(/[/\\]/).pop()!
        : "Browse"
  }

  syncBaseSpriteFrameIdx(idx: number): void {
    this.baseSpriteFrameInput.value = String(idx)
  }

  syncBaseSpriteTotalFrames(total: number): void {
    this.baseSpriteFrameInput.max = String(total - 1)
    this.baseSpriteFrameTotal.textContent = `/ ${total}`
  }

  syncSpriteFrameIdx(idx: number, frameIdx: number, total: number): void {
    const input = this.spriteFrameInputs[idx]
    if (input) input.value = String(frameIdx)
    const totalEl = this.spriteFrameTotals[idx]
    if (totalEl) totalEl.textContent = `/ ${total}`
  }

  updateFrameDisplay(idx: number, total: number): void {
    const name = this.host.currentAnimName
    const input = this.animFrameInputs.get(name)
    const totalSpan = this.animFrameTotals.get(name)
    if (input) input.value = String(total > 0 ? idx + 1 : 0)
    if (totalSpan) totalSpan.textContent = `/ ${total}`
    this.highlightCurrentAnimRow()
  }

  rebuildSpriteSelect(): void {
    this.spriteSelect.innerHTML = ""
    const charOpt = document.createElement("option")
    charOpt.value = "0"
    charOpt.textContent = "Base Sprite"
    if (this.host.currentSpriteIndex === 0) charOpt.selected = true
    this.spriteSelect.appendChild(charOpt)
    for (let i = 0; i < this.host.additionalSprites.length; i++) {
      const opt = document.createElement("option")
      opt.value = String(i + 1)
      opt.textContent = `Sprite ${i + 1}`
      if (this.host.currentSpriteIndex === i + 1) opt.selected = true
      this.spriteSelect.appendChild(opt)
    }
  }

  private group(bg: string): HTMLElement {
    const g = document.createElement("div")
    g.style.cssText = `display:flex;flex-direction:column;padding:var(--section-pad);border-radius:var(--radius-lg);background:${bg};`
    return g
  }

  private btn(text: string, onClick: () => void): HTMLButtonElement {
    const b = document.createElement("button")
    b.textContent = text
    b.style.cssText =
      "padding:var(--btn-pad-sm);background:#2a2a4a;color:var(--btn-color,#e0e0e0);border:1px solid #444;border-radius:var(--radius-sm);cursor:pointer;font-size:var(--font-sm);white-space:nowrap;"
    b.addEventListener("click", onClick)
    return b
  }
}
