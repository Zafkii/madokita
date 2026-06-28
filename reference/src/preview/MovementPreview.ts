import type { EditorFrame, MovementAnimDef } from "../types/MovementTypes"
import {
  HEADER_HTML, PLAY_BTN_HTML, PLAY_BTN_TEXT, STOP_BTN_TEXT, STOP_BTN_BG,
  LOOP_HTML, SPEED_HTML, createPreviewContainer,
} from "./previewShared"

export class MovementPreview {
  private el: HTMLElement
  private playBtn: HTMLButtonElement
  private loopCheck: HTMLInputElement
  private speedInput: HTMLInputElement

  private playing = false
  private anim: MovementAnimDef | null = null
  private onPreviewFrame: ((frame: EditorFrame, idx: number) => void) | null = null

  private startTime = 0
  private timerId = 0

  constructor(container: HTMLElement) {
    this.el = createPreviewContainer(container)
    this.el.innerHTML = HEADER_HTML + `
      <div style="display:flex;align-items:center;gap:8px;">
        ${PLAY_BTN_HTML}
        ${LOOP_HTML}
        ${SPEED_HTML}
      </div>
    `

    this.playBtn = this.el.querySelector<HTMLButtonElement>(".preview-play-btn")!
    this.loopCheck = this.el.querySelector("#preview-loop")! as HTMLInputElement
    this.speedInput = this.el.querySelector("#preview-speed")! as HTMLInputElement

    const speedVal = this.el.querySelector("#preview-speed-val")!
    this.speedInput.addEventListener("input", () => {
      speedVal.textContent = `${parseFloat(this.speedInput.value).toFixed(1)}x`
    })
    this.playBtn.addEventListener("click", () => this.togglePlay())
  }

  setAnimation(
    anim: MovementAnimDef | null,
    onFrame: (frame: EditorFrame, idx: number) => void,
  ): void {
    this.anim = anim
    this.onPreviewFrame = onFrame
    this.stop()
  }

  togglePlay(): void {
    if (this.playing) this.stop()
    else this.play()
  }

  play(): void {
    if (!this.anim || this.anim.frames.length === 0) return
    this.playing = true
    this.playBtn.textContent = STOP_BTN_TEXT
    this.playBtn.style.background = STOP_BTN_BG
    this.startTime = performance.now()
    this.tick()
  }

  stop(): void {
    this.playing = false
    this.playBtn.textContent = PLAY_BTN_TEXT
    this.playBtn.style.background = "#2a5"
    cancelAnimationFrame(this.timerId)
    if (this.anim && this.anim.frames.length > 0) {
      this.onPreviewFrame?.(this.anim.frames[0], 0)
    }
  }

  private tick = (): void => {
    if (!this.playing || !this.anim || this.anim.frames.length === 0) return

    const speed = parseFloat(this.speedInput.value) || 1
    const fps = this.anim.fps ?? 14
    const frameDuration = 1000 / fps
    const elapsed = (performance.now() - this.startTime) * speed

    const totalFrames = this.anim.frames.length
    let frameIndex = Math.floor(elapsed / frameDuration)

    if (this.loopCheck.checked) {
      frameIndex = frameIndex % totalFrames
    } else if (frameIndex >= totalFrames) {
      this.stop()
      return
    }

    const frame = this.anim.frames[frameIndex]
    this.onPreviewFrame?.(frame, frameIndex)
    this.timerId = requestAnimationFrame(this.tick)
  }

  destroy(): void {
    cancelAnimationFrame(this.timerId)
    if (this.el.parentElement) {
      this.el.parentElement.removeChild(this.el)
    }
  }
}
