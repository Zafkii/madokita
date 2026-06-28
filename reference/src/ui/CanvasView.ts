import { ThemeManager } from "./ThemeManager"
import { CanvasController } from "./CanvasController"
import type { EditorHurtboxData } from "../types/MovementTypes"

export type TransformState = {
  offsetX: number
  offsetY: number
  rotation: number
  scaleX: number
  scaleY: number
}

export type OverlayData = {
  image: HTMLImageElement
  frameW: number
  frameH: number
  frameIndex: number
  transform: TransformState
  originX: number
  originY: number
}

export type CanvasViewOptions = {
  container: HTMLElement
  onChange?: (state: TransformState) => void
  onOriginChange?: (originX: number, originY: number) => void
  onZoomChange?: (zoom: number) => void
  onHurtboxSelect?: (index: number) => void
  onHurtboxChange?: (hb: EditorHurtboxData) => void
}

const HANDLE_SIZE = 8
const HANDLE_HIT = 12

export class CanvasView {
  readonly canvas: HTMLCanvasElement
  private ctx: CanvasRenderingContext2D
  private container: HTMLElement

  cameraX = 0
  cameraY = 0
  zoom = 1

  private themeManager = new ThemeManager()

  boundaryW = 1280
  boundaryH = 720
  boundaryVisible = true

  private resizeObserver: ResizeObserver | null = null
  private bgImage: HTMLImageElement | null = null
  overlays: OverlayData[] = []
  activeOverlayIndex = 0
  hurtboxData: EditorHurtboxData[] = []
  activeHurtboxIndex = -1

  transform: TransformState = {
    offsetX: 0,
    offsetY: 0,
    rotation: 0,
    scaleX: 1,
    scaleY: 1,
  }

  baseRotation = 0

  private restPoseW = 0
  private restPoseH = 0
  private restPoseOriginX = 0.5
  private restPoseOriginY = 0.5

  onChange?: (state: TransformState) => void
  onOriginChange?: (originX: number, originY: number) => void
  onZoomChange?: (zoom: number) => void
  onHurtboxSelect?: (index: number) => void
  onHurtboxChange?: (hb: EditorHurtboxData) => void

  private animFrameId = 0
  private needsRender = true
  private controller: CanvasController

  constructor(options: CanvasViewOptions) {
    this.container = options.container
    this.onChange = options.onChange
    this.onOriginChange = options.onOriginChange
    this.onZoomChange = options.onZoomChange
    this.onHurtboxSelect = options.onHurtboxSelect
    this.onHurtboxChange = options.onHurtboxChange

    this.canvas = document.createElement("canvas")
    this.canvas.style.width = "100%"
    this.canvas.style.height = "100%"
    this.canvas.style.display = "block"
    this.canvas.tabIndex = 0
    this.canvas.style.outline = "none"
    this.canvas.style.cursor = "grab"
    this.container.appendChild(this.canvas)

    const ctx = this.canvas.getContext("2d")
    if (!ctx) throw new Error("Canvas 2D context not available")
    this.ctx = ctx

    this.controller = new CanvasController(this)
    this.resize()
    this.setupResizeObserver()
    this.loop()
  }

  private setupResizeObserver(): void {
    this.resizeObserver = new ResizeObserver(() => {
      this.resize()
    })
    this.resizeObserver.observe(this.container)
  }

  resize(): void {
    const rect = this.container.getBoundingClientRect()
    const dpr = window.devicePixelRatio || 1
    this.canvas.width = rect.width * dpr
    this.canvas.height = rect.height * dpr
    this.needsRender = true
  }

  private get canvasCenterX(): number {
    return this.canvas.width / 2
  }
  private get canvasCenterY(): number {
    return this.canvas.height / 2
  }

  worldToCanvas(worldX: number, worldY: number): { x: number; y: number } {
    return {
      x: worldX * this.zoom + this.canvasCenterX + this.cameraX,
      y: worldY * this.zoom + this.canvasCenterY + this.cameraY,
    }
  }

  canvasToWorld(canvasX: number, canvasY: number): { x: number; y: number } {
    return {
      x: (canvasX - this.canvasCenterX - this.cameraX) / this.zoom,
      y: (canvasY - this.canvasCenterY - this.cameraY) / this.zoom,
    }
  }

  activeTransform(): TransformState {
    const ov = this.overlays[this.activeOverlayIndex]
    return ov ? ov.transform : this.transform
  }

  activeImage(): HTMLImageElement | null {
    const ov = this.overlays[this.activeOverlayIndex]
    return ov?.image ?? null
  }

  getOverlaySize(ov: OverlayData): { w: number; h: number } {
    if (ov.frameW > 0 && ov.frameH > 0) {
      return { w: ov.frameW, h: ov.frameH }
    }
    return { w: ov.image.width, h: ov.image.height }
  }

  getHandlePositions(): { x: number; y: number }[] {
    const ov = this.overlays[this.activeOverlayIndex]
    if (!ov) return []
    const t = ov.transform
    const size = this.getOverlaySize(ov)
    const hw = size.w * t.scaleX
    const hh = size.h * t.scaleY
    const cos = Math.cos(((t.rotation + this.baseRotation) * Math.PI) / 180)
    const sin = Math.sin(((t.rotation + this.baseRotation) * Math.PI) / 180)
    const ox = -ov.originX * hw
    const oy = -ov.originY * hh
    const positions = [
      { x: ox, y: oy },
      { x: ox + hw, y: oy },
      { x: ox + hw, y: oy + hh },
      { x: ox, y: oy + hh },
      { x: ox + hw / 2, y: oy },
      { x: ox + hw, y: oy + hh / 2 },
      { x: ox + hw / 2, y: oy + hh },
      { x: ox, y: oy + hh / 2 },
    ]
    return positions.map((c) => ({
      x: t.offsetX + c.x * cos - c.y * sin,
      y: t.offsetY + c.x * sin + c.y * cos,
    }))
  }

  getHandleAt(worldX: number, worldY: number): number {
    const handles = this.getHandlePositions()
    if (handles.length === 0) return -1
    const threshold = HANDLE_HIT / this.zoom
    for (let i = 0; i < handles.length; i++) {
      const dx = worldX - handles[i].x
      const dy = worldY - handles[i].y
      if (dx * dx + dy * dy < threshold * threshold) return i
    }
    return -1
  }

  snapToStep(v: number, s: number): number {
    const snapped = Math.round(v / s) * s
    const decimals = Math.max(0, Math.ceil(-Math.log10(s)))
    const factor = 10 ** decimals
    return Math.round(snapped * factor) / factor
  }

  requestRender(): void {
    this.needsRender = true
  }

  // ── Render loop ──

  private loop = (): void => {
    if (this.needsRender) {
      this.render()
      this.needsRender = false
    }
    this.animFrameId = requestAnimationFrame(this.loop)
  }

  private render(): void {
    const ctx = this.ctx
    const w = this.canvas.width
    const h = this.canvas.height
    ctx.clearRect(0, 0, w, h)
    this.drawGrid()
    ctx.save()
    ctx.translate(
      this.canvasCenterX + this.cameraX,
      this.canvasCenterY + this.cameraY,
    )
    ctx.scale(this.zoom, this.zoom)
    if (this.overlays.length === 0) {
      this.drawBackground()
    }
    this.drawBoundary()
    this.drawRestPoseOutline()
    for (let i = 0; i < this.overlays.length; i++) {
      this.drawOverlay(this.overlays[i], i === this.activeOverlayIndex)
    }
    for (let i = 0; i < this.hurtboxData.length; i++) {
      this.drawHurtbox(this.hurtboxData[i], i === this.activeHurtboxIndex)
    }
    ctx.restore()
    if (this.activeHurtboxIndex >= 0) {
      ctx.save()
      ctx.setTransform(1, 0, 0, 1, 0, 0)
      this.drawHurtboxSelection()
      ctx.restore()
    }
  }

  setBaseRestPose(w: number, h: number, originX: number, originY: number): void {
    this.restPoseW = w
    this.restPoseH = h
    this.restPoseOriginX = originX
    this.restPoseOriginY = originY
    this.needsRender = true
  }

  clearBaseRestPose(): void {
    this.restPoseW = 0
    this.restPoseH = 0
    this.needsRender = true
  }

  // ── Drawing ──

  private drawGrid(): void {
    const ctx = this.ctx
    const w = this.canvas.width
    const h = this.canvas.height
    const gridSize = 50 * this.zoom
    if (gridSize < 4) return
    const offX = (this.cameraX + this.canvasCenterX) % gridSize
    const offY = (this.cameraY + this.canvasCenterY) % gridSize
    const tm = this.themeManager
    ctx.strokeStyle = tm.gridLine
    ctx.lineWidth = 1
    ctx.beginPath()
    for (let x = offX; x < w; x += gridSize) {
      ctx.moveTo(x, 0)
      ctx.lineTo(x, h)
    }
    for (let y = offY; y < h; y += gridSize) {
      ctx.moveTo(0, y)
      ctx.lineTo(w, y)
    }
    ctx.stroke()
    ctx.strokeStyle = tm.gridAxis
    ctx.lineWidth = 1
    const origin = this.worldToCanvas(0, 0)
    ctx.beginPath()
    ctx.moveTo(origin.x - 8, origin.y)
    ctx.lineTo(origin.x + 8, origin.y)
    ctx.moveTo(origin.x, origin.y - 8)
    ctx.lineTo(origin.x, origin.y + 8)
    ctx.stroke()
  }

  private drawBackground(): void {
    if (!this.bgImage) return
    const ctx = this.ctx
    ctx.save()
    ctx.globalAlpha = 0.5
    ctx.drawImage(
      this.bgImage,
      -this.bgImage.width / 2,
      -this.bgImage.height / 2,
    )
    ctx.restore()
    ctx.strokeStyle = this.themeManager.bgBorderStroke
    ctx.lineWidth = 1
    ctx.strokeRect(
      -this.bgImage.width / 2,
      -this.bgImage.height / 2,
      this.bgImage.width,
      this.bgImage.height,
    )
  }

  private drawBoundary(): void {
    if (!this.boundaryVisible || this.boundaryW <= 0 || this.boundaryH <= 0)
      return
    const ctx = this.ctx
    ctx.strokeStyle = this.themeManager.boundaryStroke
    ctx.lineWidth = 1.5
    ctx.setLineDash([6, 4])
    ctx.strokeRect(
      -this.boundaryW / 2,
      -this.boundaryH / 2,
      this.boundaryW,
      this.boundaryH,
    )
    ctx.setLineDash([])
  }

  private drawRestPoseOutline(): void {
    if (this.restPoseW <= 0 || this.restPoseH <= 0) return
    const ctx = this.ctx
    const w = this.restPoseW
    const h = this.restPoseH
    const ox = this.restPoseOriginX
    const oy = this.restPoseOriginY

    ctx.save()
    ctx.strokeStyle = this.themeManager.boundaryStroke
    ctx.lineWidth = 1.5
    ctx.setLineDash([6, 4])
    ctx.strokeRect(-ox * w, -oy * h, w, h)
    ctx.setLineDash([])
    ctx.restore()
  }

  private drawHurtbox(hb: EditorHurtboxData, isActive: boolean): void {
    const ctx = this.ctx
    const w = hb[0]
    const h = hb[1]
    ctx.save()
    ctx.translate(hb[2], hb[3])
    ctx.rotate((hb[6] * Math.PI) / 180)
    ctx.scale(hb[4], hb[5])
    if (isActive) {
      ctx.fillStyle = "rgba(255,200,50,0.3)"
      ctx.strokeStyle = "#ffcc00"
      ctx.lineWidth = 2
    } else {
      ctx.fillStyle = "rgba(255,80,80,0.2)"
      ctx.strokeStyle = "#ff5555"
      ctx.lineWidth = 1.5
    }
    ctx.fillRect(-w / 2, -h / 2, w, h)
    ctx.strokeRect(-w / 2, -h / 2, w, h)
    ctx.restore()
  }

  private drawOverlay(ov: OverlayData, isActive: boolean): void {
    const ctx = this.ctx
    const img = ov.image
    const t = ov.transform
    const totalRotation = t.rotation + this.baseRotation

    ctx.save()
    ctx.translate(t.offsetX, t.offsetY)
    ctx.rotate((totalRotation * Math.PI) / 180)
    ctx.scale(t.scaleX, t.scaleY)

    let sx = 0,
      sy = 0,
      sw = img.width,
      sh = img.height
    if (ov.frameW > 0 && ov.frameH > 0) {
      const cols = Math.max(1, Math.floor(img.width / ov.frameW))
      const row = Math.floor(ov.frameIndex / cols)
      const col = ov.frameIndex % cols
      sx = col * ov.frameW
      sy = row * ov.frameH
      sw = ov.frameW
      sh = ov.frameH
    }
    ctx.drawImage(
      img,
      sx,
      sy,
      sw,
      sh,
      -ov.originX * sw,
      -ov.originY * sh,
      sw,
      sh,
    )
    ctx.restore()

    if (isActive && this.activeHurtboxIndex < 0) {
      ctx.save()
      ctx.setTransform(1, 0, 0, 1, 0, 0)
      this.drawCrosshair(t)
      this.drawHandles()
      ctx.restore()
    }
  }

  private drawCrosshair(t: TransformState): void {
    const ctx = this.ctx
    const p = this.worldToCanvas(t.offsetX, t.offsetY)
    ctx.strokeStyle = this.themeManager.crosshairStroke
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(p.x - 6, p.y)
    ctx.lineTo(p.x + 6, p.y)
    ctx.moveTo(p.x, p.y - 6)
    ctx.lineTo(p.x, p.y + 6)
    ctx.stroke()
  }

  private drawHandles(): void {
    const ctx = this.ctx
    const all = this.getHandlePositions()
    const tm = this.themeManager
    for (let i = 0; i < all.length; i++) {
      const p = this.worldToCanvas(all[i].x, all[i].y)
      const isCorner = i < 4
      const size = isCorner ? HANDLE_SIZE : HANDLE_SIZE - 2
      ctx.fillStyle = tm.handleFill
      ctx.fillRect(p.x - size / 2, p.y - size / 2, size, size)
      ctx.strokeStyle = tm.handleStroke
      ctx.lineWidth = 1.5
      ctx.strokeRect(p.x - size / 2, p.y - size / 2, size, size)
    }
    const corners = all.slice(0, 4)
    ctx.strokeStyle = tm.handleDash
    ctx.lineWidth = 1
    ctx.setLineDash([4, 4])
    ctx.beginPath()
    for (let i = 0; i <= corners.length; i++) {
      const p = this.worldToCanvas(
        corners[i % corners.length].x,
        corners[i % corners.length].y,
      )
      i === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)
    }
    ctx.stroke()
    ctx.setLineDash([])
  }

  getHurtboxHandlePositions(
    hb: EditorHurtboxData,
  ): { x: number; y: number }[] {
    const w = hb[0]
    const h = hb[1]
    const ox = hb[2]
    const oy = hb[3]
    const sx = hb[4]
    const sy = hb[5]
    const rot = (hb[6] * Math.PI) / 180
    const cos = Math.cos(rot)
    const sin = Math.sin(rot)
    const halfW = w / 2
    const halfH = h / 2
    const locals = [
      { x: -halfW, y: -halfH },
      { x: halfW, y: -halfH },
      { x: halfW, y: halfH },
      { x: -halfW, y: halfH },
      { x: 0, y: -halfH },
      { x: halfW, y: 0 },
      { x: 0, y: halfH },
      { x: -halfW, y: 0 },
    ]
    return locals.map((c) => ({
      x: ox + c.x * sx * cos - c.y * sy * sin,
      y: oy + c.x * sx * sin + c.y * sy * cos,
    }))
  }

  getHurtboxHandleAt(worldX: number, worldY: number): number {
    const i = this.activeHurtboxIndex
    if (i < 0 || i >= this.hurtboxData.length) return -1
    const handles = this.getHurtboxHandlePositions(this.hurtboxData[i])
    const threshold = HANDLE_HIT / this.zoom
    for (let j = 0; j < handles.length; j++) {
      const dx = worldX - handles[j].x
      const dy = worldY - handles[j].y
      if (dx * dx + dy * dy < threshold * threshold) return j
    }
    return -1
  }

  private drawHurtboxSelection(): void {
    const i = this.activeHurtboxIndex
    if (i < 0 || i >= this.hurtboxData.length) return
    const hb = this.hurtboxData[i]
    const ctx = this.ctx
    const tm = this.themeManager

    // Crosshair at hurtbox center
    const cx = this.worldToCanvas(hb[2], hb[3])
    ctx.strokeStyle = tm.crosshairStroke
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(cx.x - 6, cx.y)
    ctx.lineTo(cx.x + 6, cx.y)
    ctx.moveTo(cx.x, cx.y - 6)
    ctx.lineTo(cx.x, cx.y + 6)
    ctx.stroke()

    // Handles
    const all = this.getHurtboxHandlePositions(hb)
    for (let j = 0; j < all.length; j++) {
      const p = this.worldToCanvas(all[j].x, all[j].y)
      const isCorner = j < 4
      const size = isCorner ? HANDLE_SIZE : HANDLE_SIZE - 2
      ctx.fillStyle = tm.handleFill
      ctx.fillRect(p.x - size / 2, p.y - size / 2, size, size)
      ctx.strokeStyle = tm.handleStroke
      ctx.lineWidth = 1.5
      ctx.strokeRect(p.x - size / 2, p.y - size / 2, size, size)
    }

    // Dashed border
    const corners = all.slice(0, 4)
    ctx.strokeStyle = tm.handleDash
    ctx.lineWidth = 1
    ctx.setLineDash([4, 4])
    ctx.beginPath()
    for (let j = 0; j <= corners.length; j++) {
      const p = this.worldToCanvas(
        corners[j % corners.length].x,
        corners[j % corners.length].y,
      )
      j === 0 ? ctx.moveTo(p.x, p.y) : ctx.lineTo(p.x, p.y)
    }
    ctx.stroke()
    ctx.setLineDash([])
  }

  // ── Public API ──

  setBackgroundImage(img: HTMLImageElement | null): void {
    this.bgImage = img
    this.needsRender = true
  }

  setHurtboxes(data: EditorHurtboxData[], activeIndex: number): void {
    this.hurtboxData = data
    this.activeHurtboxIndex = Math.min(activeIndex, data.length - 1)
    this.needsRender = true
  }

  setOverlays(overlays: OverlayData[], activeIndex: number): void {
    this.overlays = overlays
    this.activeOverlayIndex = Math.min(activeIndex, overlays.length - 1)
    if (this.overlays.length === 0) {
      this.activeOverlayIndex = 0
    }
    this.needsRender = true
  }

  setActiveOverlayIndex(idx: number): void {
    this.activeOverlayIndex = Math.min(idx, this.overlays.length - 1)
    this.needsRender = true
  }

  getActiveOverlayIndex(): number {
    return this.activeOverlayIndex
  }

  getActiveOverlayTransform(): TransformState | null {
    const ov = this.overlays[this.activeOverlayIndex]
    return ov ? { ...ov.transform } : null
  }

  setTransform(state: TransformState): void {
    this.transform = { ...state }
    this.needsRender = true
  }

  setOrigin(originX: number, originY: number): void {
    const ov = this.overlays[this.activeOverlayIndex]
    if (ov) {
      ov.originX = originX
      ov.originY = originY
    }
    this.needsRender = true
  }

  setBaseRotation(deg: number): void {
    this.baseRotation = deg
    this.needsRender = true
  }

  setBoundary(w: number, h: number): void {
    this.boundaryW = w
    this.boundaryH = h
    this.needsRender = true
  }

  setTheme(tm: ThemeManager): void {
    this.themeManager = tm
    this.container.style.background = tm.canvasBg
    this.needsRender = true
  }

  getTransform(): TransformState {
    return { ...this.transform }
  }

  setZoom(z: number): void {
    this.zoom = Math.max(0.05, Math.min(20, z))
    this.needsRender = true
    this.onZoomChange?.(this.zoom)
  }

  updateCallbacks(opts: {
    onChange?: (s: TransformState) => void
    onOriginChange?: (x: number, y: number) => void
    onHurtboxSelect?: (index: number) => void
    onHurtboxChange?: (hb: EditorHurtboxData) => void
  }): void {
    if (opts.onChange !== undefined) this.onChange = opts.onChange
    if (opts.onOriginChange !== undefined)
      this.onOriginChange = opts.onOriginChange
    if (opts.onHurtboxSelect !== undefined)
      this.onHurtboxSelect = opts.onHurtboxSelect
    if (opts.onHurtboxChange !== undefined)
      this.onHurtboxChange = opts.onHurtboxChange
  }

  resetView(): void {
    this.cameraX = 0
    this.cameraY = 0
    this.zoom = 1
    this.needsRender = true
    this.onZoomChange?.(this.zoom)
  }

  clearUndo(): void {
    this.controller.clearUndo()
  }

  destroy(): void {
    cancelAnimationFrame(this.animFrameId)
    this.resizeObserver?.disconnect()
    this.controller.destroy()
    this.container.removeChild(this.canvas)
  }
}
