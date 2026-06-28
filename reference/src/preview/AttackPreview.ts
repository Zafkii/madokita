import type { AttackAnimDef, EditorAttackFrame } from "../types/AttackTypes"
import {
  HEADER_HTML, PLAY_BTN_HTML, PLAY_BTN_TEXT, STOP_BTN_TEXT, STOP_BTN_BG,
  LOOP_HTML, SPEED_HTML, createPreviewContainer,
} from "./previewShared"

type PreviewCallback = (frame: EditorAttackFrame, frameIndex: number) => void

const PHASE_MAP: Record<string, string> = {
  wu: "windup",
  atk: "active",
  rc: "recover",
}

type PhaseGroup = {
  phase: string
  start: number
  count: number
}

export class AttackPreview {
  private el: HTMLElement
  private playBtn: HTMLButtonElement
  private loopCheck: HTMLInputElement
  private speedInput: HTMLInputElement

  private playing = false
  private anim: AttackAnimDef | null = null
  private callback: PreviewCallback | null = null

  private phaseGroups: PhaseGroup[] = []
  private phaseDurations: { windup: number; active: number; recover: number } | null = null
  private guardFrameIndices: number[] = []
  private guardDuration = 3000
  private guardFps = 14

  private isAttackIdle = false
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

  setAnimation(anim: AttackAnimDef | null, callback: PreviewCallback): void {
    this.anim = anim
    this.callback = callback
    this.buildPhaseData(anim)
    this.stop()
  }

  private buildPhaseData(anim: AttackAnimDef | null): void {
    this.phaseGroups = []
    this.guardFrameIndices = []
    this.phaseDurations = null
    this.guardDuration = 3000
    this.guardFps = 14

    if (!anim || anim.frames.length === 0) return

    if (anim.windup !== undefined && anim.activeTime !== undefined && anim.recover !== undefined) {
      this.phaseDurations = {
        windup: anim.windup,
        active: anim.activeTime,
        recover: anim.recover,
      }
    }

    let current: string | null = null
    let startIdx = 0
    const groups: PhaseGroup[] = []

    for (let i = 0; i < anim.frames.length; i++) {
      const p = anim.frames[i]?.phase ?? null
      if (p !== current) {
        if (current !== null && current !== "guard" && PHASE_MAP[current]) {
          groups.push({ phase: PHASE_MAP[current], start: startIdx, count: i - startIdx })
        }
        current = p
        startIdx = i
      }
    }
    if (current !== null && current !== "guard" && PHASE_MAP[current]) {
      groups.push({ phase: PHASE_MAP[current], start: startIdx, count: anim.frames.length - startIdx })
    }

    this.phaseGroups = groups

    for (let i = 0; i < anim.frames.length; i++) {
      if (anim.frames[i]?.phase === "guard") {
        this.guardFrameIndices.push(i)
      }
    }

    this.guardDuration = anim.guard ?? 3000
    this.guardFps = anim.guardFps ?? anim.fps ?? 14
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
    this.isAttackIdle = false
    this.startTime = performance.now()
    this.tick()
  }

  stop(): void {
    this.playing = false
    this.playBtn.textContent = PLAY_BTN_TEXT
    this.playBtn.style.background = "#2a5"
    cancelAnimationFrame(this.timerId)
    if (this.anim && this.anim.frames.length > 0) {
      this.callback?.(this.anim.frames[0], 0)
    }
  }

  private tick = (): void => {
    if (!this.playing || !this.anim || this.anim.frames.length === 0) return

    const speed = parseFloat(this.speedInput.value) || 1
    const elapsed = (performance.now() - this.startTime) * speed

    if (this.isAttackIdle) {
      this.tickIdle(elapsed)
    } else {
      this.tickMain(elapsed)
    }

    this.timerId = requestAnimationFrame(this.tick)
  }

  private tickMain(elapsed: number): void {
    const anim = this.anim!
    let frameIndex: number

    if (this.phaseDurations !== null && this.phaseGroups.length > 0) {
      let accumulatedTime = 0
      frameIndex = -1

      for (const group of this.phaseGroups) {
        const groupDuration =
          group.phase === "windup" ? this.phaseDurations.windup
            : group.phase === "active" ? this.phaseDurations.active
              : this.phaseDurations.recover

        const perFrame = groupDuration / group.count
        const groupEnd = accumulatedTime + groupDuration

        if (elapsed < groupEnd) {
          const localIdx = Math.floor((elapsed - accumulatedTime) / perFrame)
          frameIndex = group.start + Math.min(localIdx, group.count - 1)
          break
        }
        accumulatedTime = groupEnd
      }

      if (frameIndex === -1) {
        if (this.guardFrameIndices.length > 0) {
          this.isAttackIdle = true
          this.startTime = performance.now()
          this.tickIdle(0)
        } else if (this.loopCheck.checked) {
          this.startTime = performance.now()
          frameIndex = 0
          const frame = anim.frames[0]
          this.callback?.(frame, 0)
        } else {
          this.stop()
        }
        return
      }
    } else {
      const totalFrames = anim.frames.length
      const frameDuration = 1000 / (anim.fps || 14)
      frameIndex = Math.floor(elapsed / frameDuration)

      if (anim.loop || this.loopCheck.checked) {
        frameIndex = frameIndex % totalFrames
      } else if (frameIndex >= totalFrames) {
        this.stop()
        return
      }
    }

    const frame = anim.frames[frameIndex]
    this.callback?.(frame, frameIndex)
  }

  private tickIdle(elapsed: number): void {
    const anim = this.anim!
    const loopFrames = this.guardFrameIndices.length > 0
      ? this.guardFrameIndices
      : anim.frames.length >= 3
        ? [anim.frames.length - 3, anim.frames.length - 2, anim.frames.length - 1]
        : [anim.frames.length - 1]

    const totalIdleDuration = this.guardDuration

    if (elapsed >= totalIdleDuration) {
      if (this.loopCheck.checked) {
        this.isAttackIdle = false
        this.startTime = performance.now()
        return
      } else {
        this.stop()
        return
      }
    }

    const frameDuration = 1000 / this.guardFps
    const idx = Math.floor(elapsed / frameDuration) % loopFrames.length
    const frameIndex = loopFrames[idx]
    const frame = anim.frames[frameIndex]
    this.callback?.(frame, frameIndex)
  }

  setPhaseDurations(windup: number, active: number, recover: number): void {
    if (this.phaseDurations) {
      this.phaseDurations = { windup, active, recover }
    }
    if (this.playing) {
      this.startTime = performance.now()
    }
  }

  setGuardParams(fps: number, duration: number): void {
    this.guardFps = fps
    this.guardDuration = duration
  }

  destroy(): void {
    cancelAnimationFrame(this.timerId)
    if (this.el.parentElement) {
      this.el.parentElement.removeChild(this.el)
    }
  }
}
