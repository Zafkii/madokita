import type { CanvasView, TransformState } from "./CanvasView"
import type { EditorHurtboxData } from "../types/MovementTypes"
import {
  canvasKeybindings,
  matchesModifier,
  matchesBinding,
} from "./CanvasKeybindings"

const cfg = canvasKeybindings

export class CanvasController {
  isDraggingOverlay = false
  isDraggingHandle = false
  handleIndex = -1
  isPanning = false
  isDraggingHurtboxBody = false
  isDraggingHurtboxHandle = false
  hurtboxHandleIndex = -1
  lastPointerX = 0
  lastPointerY = 0
  dragStartOffset = { x: 0, y: 0 }
  dragStartLocalDragX = 0
  dragStartLocalDragY = 0
  dragStartTransform: TransformState = {
    offsetX: 0,
    offsetY: 0,
    rotation: 0,
    scaleX: 1,
    scaleY: 1,
  }
  dragStartHurtbox: EditorHurtboxData | null = null
  dragStartOppWorld = { x: 0, y: 0 }

  undoStack: UndoSnapshot[] = []
  redoStack: UndoSnapshot[] = []
  snapBefore: UndoSnapshot | null = null
  wheelDebounceTimer: ReturnType<typeof setTimeout> | null = null

  private static makeHandlePositions(
    ox: number,
    oy: number,
    hw: number,
    hh: number,
  ): { x: number; y: number }[] {
    return [
      { x: ox, y: oy },
      { x: ox + hw, y: oy },
      { x: ox + hw, y: oy + hh },
      { x: ox, y: oy + hh },
      { x: ox + hw / 2, y: oy },
      { x: ox + hw, y: oy + hh / 2 },
      { x: ox + hw / 2, y: oy + hh },
      { x: ox, y: oy + hh / 2 },
    ]
  }
  view: CanvasView

  constructor(view: CanvasView) {
    this.view = view
    this.bindEvents()
  }

  // ── Events ──

  onContextMenu = (e: MouseEvent): void => {
    e.preventDefault()
  }

  bindEvents(): void {
    const c = this.view.canvas
    c.addEventListener("mousedown", this.onPointerDown)
    c.addEventListener("mousemove", this.onPointerMove)
    c.addEventListener("mouseup", this.onPointerUp)
    c.addEventListener("wheel", this.onWheel, { passive: false })
    c.addEventListener("contextmenu", this.onContextMenu)
    c.addEventListener("dblclick", this.onDoubleClick)
    c.addEventListener("keydown", this.onKeyDown)
    window.addEventListener("resize", this.onResize)
  }

  destroy(): void {
    const c = this.view.canvas
    c.removeEventListener("mousedown", this.onPointerDown)
    c.removeEventListener("mousemove", this.onPointerMove)
    c.removeEventListener("mouseup", this.onPointerUp)
    c.removeEventListener("wheel", this.onWheel)
    c.removeEventListener("contextmenu", this.onContextMenu)
    c.removeEventListener("dblclick", this.onDoubleClick)
    c.removeEventListener("keydown", this.onKeyDown)
    window.removeEventListener("resize", this.onResize)
  }

  onResize = (): void => {
    this.view.resize()
  }

  onPointerDown = (e: MouseEvent): void => {
    const rect = this.view.canvas.getBoundingClientRect()
    const dpr = window.devicePixelRatio || 1
    const px = (e.clientX - rect.left) * dpr
    const py = (e.clientY - rect.top) * dpr
    this.lastPointerX = px
    this.lastPointerY = py

    if (cfg.panWithRightClick && e.button === 2) {
      this.isPanning = true
      this.view.canvas.style.cursor = "grabbing"
      return
    }
    if (e.button !== 0) return

    if (matchesModifier(e, cfg.panModifier)) {
      this.isPanning = true
      this.view.canvas.style.cursor = "grabbing"
      return
    }

    const world = this.view.canvasToWorld(px, py)

    // ── Hurtbox handle drag (with modifier) ── PRIORITIZE HURTBOX IF SELECTED
    const hbIdx = this.view.activeHurtboxIndex
    const hasHurtbox = hbIdx >= 0 && hbIdx < this.view.hurtboxData.length
    const activeImg = this.view.activeImage()
    if (hasHurtbox && matchesModifier(e, cfg.handleDragModifier)) {
      const hhi = this.view.getHurtboxHandleAt(world.x, world.y)
      if (hhi >= 0) {
        this.isDraggingHurtboxHandle = true
        this.hurtboxHandleIndex = hhi
        this.dragStartHurtbox = [
          ...this.view.hurtboxData[hbIdx],
        ] as EditorHurtboxData
        const all = this.view.getHurtboxHandlePositions(
          this.view.hurtboxData[hbIdx],
        )
        const oppIdx = hhi < 4 ? (hhi + 2) % 4 : ((hhi - 4 + 2) % 4) + 4
        this.dragStartOppWorld = { ...all[oppIdx] }
        this.snapBefore = this.captureSnapshot()
        this.view.canvas.style.cursor =
          hhi < 4
            ? "nwse-resize"
            : hhi === 4 || hhi === 6
              ? "ns-resize"
              : "ew-resize"
        return
      }
    }

    // ── Sprite handle drag (only if no hurtbox selected) ──
    if (
      activeImg &&
      !hasHurtbox &&
      matchesModifier(e, cfg.handleDragModifier)
    ) {
      const hi = this.view.getHandleAt(world.x, world.y)
      if (hi >= 0) {
        this.snapBefore = this.captureSnapshot()
        this.isDraggingHandle = true
        this.handleIndex = hi
        const t = this.view.activeTransform()
        this.dragStartTransform = { ...t }

        const ov = this.view.overlays[this.view.activeOverlayIndex]
        if (ov) {
          const size = this.view.getOverlaySize(ov)
          const hw = size.w * t.scaleX
          const hh = size.h * t.scaleY
          const cos = Math.cos(
            ((t.rotation + this.view.baseRotation) * Math.PI) / 180,
          )
          const sin = Math.sin(
            ((t.rotation + this.view.baseRotation) * Math.PI) / 180,
          )
          const ox = -ov.originX * hw
          const oy = -ov.originY * hh
          const positions = CanvasController.makeHandlePositions(ox, oy, hw, hh)
          const oppIdx = hi < 4 ? (hi + 2) % 4 : ((hi - 4 + 2) % 4) + 4
          const opposite = positions[oppIdx]
          const oppWorldX = t.offsetX + opposite.x * cos - opposite.y * sin
          const oppWorldY = t.offsetY + opposite.x * sin + opposite.y * cos
          this.dragStartLocalDragX =
            (world.x - oppWorldX) * cos + (world.y - oppWorldY) * sin
          this.dragStartLocalDragY =
            -(world.x - oppWorldX) * sin + (world.y - oppWorldY) * cos
        }

        this.view.canvas.style.cursor =
          hi < 4
            ? "nwse-resize"
            : hi === 4 || hi === 6
              ? "ns-resize"
              : "ew-resize"
        return
      }
    }

    // ── Hurtbox body drag (only active/selected hurtbox) ──
    if (hasHurtbox && this.hitTestHurtboxAt(world.x, world.y, hbIdx)) {
      this.isDraggingHurtboxBody = true
      this.dragStartHurtbox = [
        ...this.view.hurtboxData[hbIdx],
      ] as EditorHurtboxData
      const hb = this.view.hurtboxData[hbIdx]
      this.dragStartOffset.x = world.x - hb[2]
      this.dragStartOffset.y = world.y - hb[3]
      this.snapBefore = this.captureSnapshot()
      this.view.canvas.style.cursor = "grabbing"
      return
    }

    // ── Sprite drag (only if no hurtbox selected) ──
    if (e.button === 0 && activeImg && !hasHurtbox) {
      const t = this.view.activeTransform()
      const dx = world.x - t.offsetX
      const dy = world.y - t.offsetY
      const ov = this.view.overlays[this.view.activeOverlayIndex]
      const size = ov ? this.view.getOverlaySize(ov) : { w: 0, h: 0 }
      const maxDim =
        Math.max(size.w, size.h) *
        Math.max(Math.abs(t.scaleX), Math.abs(t.scaleY)) *
        0.6
      if (dx * dx + dy * dy < maxDim * maxDim) {
        this.snapBefore = this.captureSnapshot()
        this.isDraggingOverlay = true
        this.dragStartOffset.x = dx
        this.dragStartOffset.y = dy
        this.view.canvas.style.cursor = "grabbing"
        return
      }
    }
  }

  onPointerMove = (e: MouseEvent): void => {
    const rect = this.view.canvas.getBoundingClientRect()
    const dpr = window.devicePixelRatio || 1
    const px = (e.clientX - rect.left) * dpr
    const py = (e.clientY - rect.top) * dpr
    const dx = px - this.lastPointerX
    const dy = py - this.lastPointerY
    this.lastPointerX = px
    this.lastPointerY = py

    if (this.isPanning) {
      this.view.cameraX += dx
      this.view.cameraY += dy
      this.view.requestRender()
      return
    }

    if (this.isDraggingHandle) {
      const world = this.view.canvasToWorld(px, py)
      const ov = this.view.overlays[this.view.activeOverlayIndex]
      if (!ov) return
      const t = ov.transform
      const size = this.view.getOverlaySize(ov)
      const hw = size.w * this.dragStartTransform.scaleX
      const hh = size.h * this.dragStartTransform.scaleY
      const cos = Math.cos(
        ((this.dragStartTransform.rotation + this.view.baseRotation) *
          Math.PI) /
          180,
      )
      const sin = Math.sin(
        ((this.dragStartTransform.rotation + this.view.baseRotation) *
          Math.PI) /
          180,
      )
      const ox = -ov.originX * hw
      const oy = -ov.originY * hh
      const positions = CanvasController.makeHandlePositions(ox, oy, hw, hh)
      const hi = this.handleIndex
      const oppIdx = hi < 4 ? (hi + 2) % 4 : ((hi - 4 + 2) % 4) + 4
      const opposite = positions[oppIdx]
      const oppWorldX =
        this.dragStartTransform.offsetX + opposite.x * cos - opposite.y * sin
      const oppWorldY =
        this.dragStartTransform.offsetY + opposite.x * sin + opposite.y * cos

      const localDragX =
        (world.x - oppWorldX) * cos + (world.y - oppWorldY) * sin
      const localDragY =
        -(world.x - oppWorldX) * sin + (world.y - oppWorldY) * cos

      if (hi < 4) {
        // Non-uniform scaling on corners: each axis scales independently
        if (Math.abs(this.dragStartLocalDragX) > 0.01) {
          const factorX = localDragX / this.dragStartLocalDragX
          t.scaleX = this.view.snapToStep(
            Math.max(0.01, this.dragStartTransform.scaleX * factorX),
            0.01,
          )
        }
        if (Math.abs(this.dragStartLocalDragY) > 0.01) {
          const factorY = localDragY / this.dragStartLocalDragY
          t.scaleY = this.view.snapToStep(
            Math.max(0.01, this.dragStartTransform.scaleY * factorY),
            0.01,
          )
        }
      } else {
        const isTopBottom = hi === 4 || hi === 6
        if (isTopBottom) {
          if (Math.abs(this.dragStartLocalDragY) > 0.01) {
            const factorY = localDragY / this.dragStartLocalDragY
            t.scaleY = this.view.snapToStep(
              Math.max(0.01, this.dragStartTransform.scaleY * factorY),
              0.01,
            )
          }
        } else {
          if (Math.abs(this.dragStartLocalDragX) > 0.01) {
            const factorX = localDragX / this.dragStartLocalDragX
            t.scaleX = this.view.snapToStep(
              Math.max(0.01, this.dragStartTransform.scaleX * factorX),
              0.01,
            )
          }
        }
      }

      const newHw = size.w * t.scaleX
      const newHh = size.h * t.scaleY
      const newOx = -ov.originX * newHw
      const newOy = -ov.originY * newHh

      let newOppX = 0
      let newOppY = 0
      if (hi < 4) {
        const cornerIdx = oppIdx
        newOppX = cornerIdx === 1 || cornerIdx === 2 ? newOx + newHw : newOx
        newOppY = cornerIdx === 2 || cornerIdx === 3 ? newOy + newHh : newOy
      } else {
        if (oppIdx === 4) {
          newOppX = newOx + newHw / 2
          newOppY = newOy
        } else if (oppIdx === 5) {
          newOppX = newOx + newHw
          newOppY = newOy + newHh / 2
        } else if (oppIdx === 6) {
          newOppX = newOx + newHw / 2
          newOppY = newOy + newHh
        } else {
          newOppX = newOx
          newOppY = newOy + newHh / 2
        }
      }

      t.offsetX = this.view.snapToStep(
        oppWorldX - (newOppX * cos - newOppY * sin),
        0.05,
      )
      t.offsetY = this.view.snapToStep(
        oppWorldY - (newOppX * sin + newOppY * cos),
        0.05,
      )

      this.view.requestRender()
      this.view.onChange?.(t)
      return
    }

    // ── Hurtbox handle drag ──
    if (this.isDraggingHurtboxHandle) {
      const world = this.view.canvasToWorld(px, py)
      const i = this.view.activeHurtboxIndex
      if (i < 0 || !this.dragStartHurtbox) return
      const hb = this.view.hurtboxData[i]
      const src = this.dragStartHurtbox
      const rot = (src[6] * Math.PI) / 180
      const cos = Math.cos(rot)
      const sin = Math.sin(rot)

      // Convert world mouse to hurtbox local space (unrotate around offset, unscale)
      const dx = world.x - src[2]
      const dy = world.y - src[3]
      const localX = (dx * cos + dy * sin) / (src[4] || 0.001)
      const localY = (-dx * sin + dy * cos) / (src[5] || 0.001)

      // Opposite corner local position
      const oppIdx =
        this.hurtboxHandleIndex < 4
          ? (this.hurtboxHandleIndex + 2) % 4
          : ((this.hurtboxHandleIndex - 4 + 2) % 4) + 4
      const halfW = src[0] / 2
      const halfH = src[1] / 2
      const oppLocalPositions = [
        { x: -halfW, y: -halfH },
        { x: halfW, y: -halfH },
        { x: halfW, y: halfH },
        { x: -halfW, y: halfH },
        { x: 0, y: -halfH },
        { x: halfW, y: 0 },
        { x: 0, y: halfH },
        { x: -halfW, y: 0 },
      ]
      const opp = oppLocalPositions[oppIdx]

      let newW = src[0]
      let newH = src[1]
      if (this.hurtboxHandleIndex < 4) {
        newW = Math.max(1, Math.abs(localX - opp.x))
        newH = Math.max(1, Math.abs(localY - opp.y))
      } else if (
        this.hurtboxHandleIndex === 4 ||
        this.hurtboxHandleIndex === 6
      ) {
        newH = Math.max(1, Math.abs(localY - opp.y))
      } else {
        newW = Math.max(1, Math.abs(localX - opp.x))
      }

      hb[0] = this.view.snapToStep(newW, 1)
      hb[1] = this.view.snapToStep(newH, 1)

      // Opposite-anchored: recompute offset so the opposite corner stays fixed in world space
      const newHalfW = hb[0] / 2
      const newHalfH = hb[1] / 2
      const newOppLocalPositions = [
        { x: -newHalfW, y: -newHalfH },
        { x: newHalfW, y: -newHalfH },
        { x: newHalfW, y: newHalfH },
        { x: -newHalfW, y: newHalfH },
        { x: 0, y: -newHalfH },
        { x: newHalfW, y: 0 },
        { x: 0, y: newHalfH },
        { x: -newHalfW, y: 0 },
      ]
      const newOpp = newOppLocalPositions[oppIdx]
      hb[2] = this.view.snapToStep(
        this.dragStartOppWorld.x -
          (newOpp.x * hb[4] * cos - newOpp.y * hb[5] * sin),
        0.05,
      )
      hb[3] = this.view.snapToStep(
        this.dragStartOppWorld.y -
          (newOpp.x * hb[4] * sin + newOpp.y * hb[5] * cos),
        0.05,
      )

      this.view.requestRender()
      this.view.onHurtboxChange?.([...hb] as EditorHurtboxData)
      return
    }

    // ── Hurtbox body drag ──
    if (this.isDraggingHurtboxBody) {
      const world = this.view.canvasToWorld(px, py)
      const i = this.view.activeHurtboxIndex
      if (i < 0 || !this.dragStartHurtbox) return
      const hb = this.view.hurtboxData[i]
      const snap = (v: number) => Math.round(v)
      hb[2] = snap(world.x - this.dragStartOffset.x)
      hb[3] = snap(world.y - this.dragStartOffset.y)
      this.view.requestRender()
      this.view.onHurtboxChange?.([...hb] as EditorHurtboxData)
      return
    }

    if (this.isDraggingOverlay) {
      const world = this.view.canvasToWorld(px, py)
      const t = this.view.activeTransform()
      t.offsetX = Math.round(world.x - this.dragStartOffset.x)
      t.offsetY = Math.round(world.y - this.dragStartOffset.y)
      this.view.requestRender()
      this.view.onChange?.(t)
      return
    }

    const activeImg = this.view.activeImage()
    if (activeImg) {
      const world = this.view.canvasToWorld(px, py)
      if (matchesModifier(e, cfg.handleDragModifier)) {
        // Sprite handles
        const hi = this.view.getHandleAt(world.x, world.y)
        if (hi >= 0) {
          this.view.canvas.style.cursor =
            hi < 4
              ? "nwse-resize"
              : hi === 4 || hi === 6
                ? "ns-resize"
                : "ew-resize"
          return
        }
        // Hurtbox handles
        if (this.view.activeHurtboxIndex >= 0) {
          const hhi = this.view.getHurtboxHandleAt(world.x, world.y)
          if (hhi >= 0) {
            this.view.canvas.style.cursor =
              hhi < 4
                ? "nwse-resize"
                : hhi === 4 || hhi === 6
                  ? "ns-resize"
                  : "ew-resize"
            return
          }
        }
      }
      // ── Hurtbox hover (only active hurtbox) ──
      const hbIdx = this.view.activeHurtboxIndex
      if (hbIdx >= 0 && this.hitTestHurtboxAt(world.x, world.y, hbIdx)) {
        this.view.canvas.style.cursor = "grab"
        return
      }
      // ── Sprite hover ──
      const t = this.view.activeTransform()
      const dx2 = world.x - t.offsetX
      const dy2 = world.y - t.offsetY
      const ov = this.view.overlays[this.view.activeOverlayIndex]
      const size = ov ? this.view.getOverlaySize(ov) : { w: 0, h: 0 }
      const maxDim =
        Math.max(size.w, size.h) *
        Math.max(Math.abs(t.scaleX), Math.abs(t.scaleY)) *
        0.6
      this.view.canvas.style.cursor =
        dx2 * dx2 + dy2 * dy2 < maxDim * maxDim ? "grab" : "default"
    }
  }

  onPointerUp = (): void => {
    const wasDragging =
      this.isDraggingOverlay ||
      this.isDraggingHandle ||
      this.isDraggingHurtboxBody ||
      this.isDraggingHurtboxHandle
    this.isDraggingOverlay = false
    this.isDraggingHandle = false
    this.isDraggingHurtboxBody = false
    this.isDraggingHurtboxHandle = false
    this.isPanning = false
    this.dragStartHurtbox = null
    this.view.canvas.style.cursor = this.view.activeImage() ? "grab" : "default"

    if (this.snapBefore) {
      if (wasDragging) {
        this.pushUndo(this.snapBefore)
      }
      this.snapBefore = null
    }
  }

  onWheel = (e: WheelEvent): void => {
    e.preventDefault()
    const rect = this.view.canvas.getBoundingClientRect()
    const dpr = window.devicePixelRatio || 1
    const px = (e.clientX - rect.left) * dpr
    const py = (e.clientY - rect.top) * dpr
    const worldBefore = this.view.canvasToWorld(px, py)

    // ── Hurtbox wheel interactions ──
    const hbIdx = this.view.activeHurtboxIndex
    const hasHurtbox = hbIdx >= 0 && hbIdx < this.view.hurtboxData.length
    if (hasHurtbox) {
      const isZoom = matchesModifier(e, cfg.zoomModifier)

      if (matchesModifier(e, cfg.scaleModifier)) {
        // Shift + wheel: uniform scale (matches sprite behavior)
        if (!this.snapBefore) this.snapBefore = this.captureSnapshot()
        const hb = this.view.hurtboxData[hbIdx]
        const delta = e.deltaY > 0 ? -0.01 : 0.01
        const raw = Math.max(0.01, hb[4] + delta)
        hb[4] = Math.round(raw * 100) / 100
        hb[5] = Math.round(raw * 100) / 100
        this.view.requestRender()
        this.view.onHurtboxChange?.([...hb] as EditorHurtboxData)
      } else if (!isZoom) {
        // No modifier or Alt: rotate
        if (!this.snapBefore) this.snapBefore = this.captureSnapshot()
        const hb = this.view.hurtboxData[hbIdx]
        hb[6] = (hb[6] + (e.deltaY > 0 ? -5 : 5)) % 360
        this.view.requestRender()
        this.view.onHurtboxChange?.([...hb] as EditorHurtboxData)
      }

      if (!isZoom) {
        if (this.wheelDebounceTimer) clearTimeout(this.wheelDebounceTimer)
        this.wheelDebounceTimer = setTimeout(() => {
          if (this.snapBefore) {
            const cur = this.captureSnapshot()
            if (this.snapshotsDiffer(this.snapBefore, cur)) {
              this.pushUndo(this.snapBefore)
            }
            this.snapBefore = null
          }
        }, 500)
        return
      }
      // Ctrl + wheel: fall through to zoom handler
    }

    if (matchesModifier(e, cfg.scaleModifier)) {
      if (!this.snapBefore) this.snapBefore = this.captureSnapshot()
      const delta = e.deltaY > 0 ? -0.01 : 0.01
      const t = this.view.activeTransform()
      const raw = Math.max(0.01, t.scaleX + delta)
      t.scaleX = Math.round(raw * 100) / 100
      t.scaleY = Math.round(raw * 100) / 100
      this.view.requestRender()
      this.view.onChange?.(this.view.activeTransform())
    } else if (matchesModifier(e, cfg.rotateModifier)) {
      if (!this.snapBefore) this.snapBefore = this.captureSnapshot()
      const delta = e.deltaY > 0 ? -0.05 : 0.05
      const t = this.view.activeTransform()
      t.rotation = (t.rotation + delta * 180) % 360
      this.view.requestRender()
      this.view.onChange?.(this.view.activeTransform())
    } else if (matchesModifier(e, cfg.zoomModifier)) {
      const factor = e.deltaY > 0 ? 0.9 : 1.1
      this.view.zoom = Math.max(0.05, Math.min(20, this.view.zoom * factor))
      const worldAfter = this.view.canvasToWorld(px, py)
      this.view.cameraX += (worldAfter.x - worldBefore.x) * this.view.zoom
      this.view.cameraY += (worldAfter.y - worldBefore.y) * this.view.zoom
      this.view.requestRender()
      this.view.onZoomChange?.(this.view.zoom)
    } else {
      if (!this.snapBefore) this.snapBefore = this.captureSnapshot()
      const t = this.view.activeTransform()
      t.rotation = (t.rotation + (e.deltaY > 0 ? -5 : 5)) % 360
      this.view.requestRender()
      this.view.onChange?.(t)
    }

    const hasTransformMod =
      matchesModifier(e, cfg.scaleModifier) ||
      matchesModifier(e, cfg.rotateModifier)
    if (hasTransformMod || !matchesModifier(e, cfg.zoomModifier)) {
      if (this.wheelDebounceTimer) clearTimeout(this.wheelDebounceTimer)
      this.wheelDebounceTimer = setTimeout(() => {
        if (this.snapBefore) {
          const cur = this.captureSnapshot()
          if (this.snapshotsDiffer(this.snapBefore, cur)) {
            this.pushUndo(this.snapBefore)
          }
          this.snapBefore = null
        }
      }, 400)
    }
  }

  onDoubleClick = (e: MouseEvent): void => {
    if (e.button !== 0) return
    const ov = this.view.overlays[this.view.activeOverlayIndex]
    if (!ov) return
    const before = this.captureSnapshot()
    const rect = this.view.canvas.getBoundingClientRect()
    const dpr = window.devicePixelRatio || 1
    const px = (e.clientX - rect.left) * dpr
    const py = (e.clientY - rect.top) * dpr
    const world = this.view.canvasToWorld(px, py)

    const size = this.view.getOverlaySize(ov)
    const imgW = size.w
    const imgH = size.h
    const t = ov.transform
    const cos = Math.cos(
      ((t.rotation + this.view.baseRotation) * Math.PI) / 180,
    )
    const sin = Math.sin(
      ((t.rotation + this.view.baseRotation) * Math.PI) / 180,
    )
    const relX = (world.x - t.offsetX) / t.scaleX
    const relY = (world.y - t.offsetY) / t.scaleY
    const localX = relX * cos + relY * sin
    const localY = -relX * sin + relY * cos
    const ix = localX + imgW * ov.originX
    const iy = localY + imgH * ov.originY

    if (ix >= 0 && ix <= imgW && iy >= 0 && iy <= imgH) {
      ov.originX = Math.max(0, Math.min(1, ix / imgW))
      ov.originY = Math.max(0, Math.min(1, iy / imgH))
      this.view.requestRender()
      this.view.onOriginChange?.(ov.originX, ov.originY)
      this.pushUndo(before)
    }
  }

  onKeyDown = (e: KeyboardEvent): void => {
    if (e.repeat) return
    if (matchesBinding(e, cfg.undo)) {
      e.preventDefault()
      this.undo()
    } else if (matchesBinding(e, cfg.redo) || matchesBinding(e, cfg.redoAlt)) {
      e.preventDefault()
      this.redo()
    }
  }

  // ── Undo / Redo ──

  captureSnapshot(): UndoSnapshot {
    const ov = this.view.overlays[this.view.activeOverlayIndex]
    if (ov) {
      return {
        transform: { ...ov.transform },
        originX: ov.originX,
        originY: ov.originY,
      }
    }
    return {
      transform: { ...this.view.transform },
      originX: 0.5,
      originY: 0.5,
    }
  }

  applySnapshot(snap: UndoSnapshot): void {
    const ov = this.view.overlays[this.view.activeOverlayIndex]
    if (ov) {
      ov.transform = { ...snap.transform }
      ov.originX = snap.originX
      ov.originY = snap.originY
    } else {
      this.view.transform = { ...snap.transform }
    }
    this.view.requestRender()
    this.view.onChange?.(snap.transform)
    this.view.onOriginChange?.(snap.originX, snap.originY)
  }

  snapshotsDiffer(a: UndoSnapshot, b: UndoSnapshot): boolean {
    return (
      a.transform.offsetX !== b.transform.offsetX ||
      a.transform.offsetY !== b.transform.offsetY ||
      a.transform.rotation !== b.transform.rotation ||
      a.transform.scaleX !== b.transform.scaleX ||
      a.transform.scaleY !== b.transform.scaleY ||
      a.originX !== b.originX ||
      a.originY !== b.originY
    )
  }

  pushUndo(snap: UndoSnapshot): void {
    this.undoStack.push(snap)
    if (this.undoStack.length > 20) this.undoStack.shift()
    this.redoStack = []
  }

  clearUndo(): void {
    this.undoStack = []
    this.redoStack = []
  }

  undo(): void {
    const prev = this.undoStack.pop()
    if (!prev) return
    this.redoStack.push(this.captureSnapshot())
    this.applySnapshot(prev)
  }

  redo(): void {
    const next = this.redoStack.pop()
    if (!next) return
    this.undoStack.push(this.captureSnapshot())
    this.applySnapshot(next)
  }

  private hitTestSingleHurtbox(
    worldX: number,
    worldY: number,
    hb: EditorHurtboxData,
  ): boolean {
    const w = hb[0]
    const h = hb[1]
    const ox = hb[2]
    const oy = hb[3]
    const sx = hb[4]
    const sy = hb[5]
    const rot = hb[6]
    const dx = worldX - ox
    const dy = worldY - oy
    const cos = Math.cos((rot * Math.PI) / 180)
    const sin = Math.sin((rot * Math.PI) / 180)
    const localX = (dx * cos + dy * sin) / (sx || 0.001)
    const localY = (-dx * sin + dy * cos) / (sy || 0.001)
    const halfW = w / 2
    const halfH = h / 2
    return (
      localX >= -halfW && localX <= halfW && localY >= -halfH && localY <= halfH
    )
  }

  private hitTestHurtboxAt(
    worldX: number,
    worldY: number,
    idx: number,
  ): boolean {
    const hbs = this.view.hurtboxData
    if (idx < 0 || idx >= hbs.length) return false
    return this.hitTestSingleHurtbox(worldX, worldY, hbs[idx])
  }
}

type UndoSnapshot = {
  transform: TransformState
  originX: number
  originY: number
}
