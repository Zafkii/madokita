import Phaser from "phaser"
import type { Frame, HurtboxData, Movement, MovementAnimDef } from "./MovementTypes"
import type { Attack, AttackAnimDef } from "./AttackTypes"

export const AnimPhase = {
  None: "none",
  Windup: "windup",
  Active: "active",
  Recover: "recover",
} as const

export type AnimPhase = (typeof AnimPhase)[keyof typeof AnimPhase]

export type PhaseDurations = {
  windup: number
  active: number
  recover: number
}

export class MultiSpriteAnimator {
  private scene: Phaser.Scene
  readonly sprites: Phaser.GameObjects.Sprite[]
  private def: Movement | Attack
  private currentAnim: MovementAnimDef | AttackAnimDef | null = null
  private currentFrameHurtboxes: HurtboxData[] | null = null
  private currentPhase: AnimPhase = AnimPhase.None
  private startTime = 0
  private isVisible = false
  private isGuard = false
  private guardStartTime = 0
  private phaseDurations: PhaseDurations | null = null
  private guardDuration: number | null = null
  private guardFps: number | null = null
  private phaseFrameOffsets: { phase: AnimPhase; start: number; count: number }[] = []

  private static phaseMarkerMap: Record<string, AnimPhase> = {
    wu: AnimPhase.Windup,
    atk: AnimPhase.Active,
    rc: AnimPhase.Recover,
  }

  constructor(
    scene: Phaser.Scene,
    def: Movement | Attack,
    baseSprite: Phaser.GameObjects.Sprite,
    additionalTextures?: string[],
  ) {
    this.scene = scene
    this.def = def
    this.sprites = [baseSprite]
    if (additionalTextures) {
      for (const tex of additionalTextures) {
        const s = scene.add.sprite(0, 0, tex)
        s.setOrigin(def.defaultOriginX ?? 0.5, def.defaultOriginY ?? 0.5)
        s.setVisible(false)
        this.sprites.push(s)
      }
    }
  }

  playAnimation(name: string, durations?: PhaseDurations): void {
    const anim = this.def.animations[name]
    if (!anim || anim.frames.length === 0) return

    this.isGuard = false
    this.currentAnim = anim
    this.startTime = this.scene.time.now
    this.phaseDurations = durations ?? null
    this.isVisible = true
    this.showAdditionalSprites()

    if (durations) {
      this.buildPhaseFrameLookup(anim as AttackAnimDef)
    } else {
      this.phaseFrameOffsets = []
    }
  }

  private buildPhaseFrameLookup(anim: AttackAnimDef): void {
    const groups: { phase: AnimPhase; start: number; count: number }[] = []
    let current: string | null = null
    let startIdx = 0
    const map = MultiSpriteAnimator.phaseMarkerMap
    for (let i = 0; i < anim.frames.length; i++) {
      const p = anim.frames[i]?.phase ?? null
      if (p !== current) {
        if (current !== null && current !== "guard" && map[current]) {
          groups.push({ phase: map[current], start: startIdx, count: i - startIdx })
        }
        current = p
        startIdx = i
      }
    }
    if (current !== null && current !== "guard" && map[current]) {
      groups.push({ phase: map[current], start: startIdx, count: anim.frames.length - startIdx })
    }
    this.phaseFrameOffsets = groups
  }

  playGuard(name: string, duration?: number, fps?: number): void {
    const anim = this.def.animations[name]
    if (!anim) return
    const guardDuration = (anim as AttackAnimDef).guard
    if (!guardDuration || guardDuration <= 0) return
    this.isGuard = true
    this.currentAnim = anim
    this.isVisible = true
    this.currentPhase = AnimPhase.None
    this.guardStartTime = this.scene.time.now
    this.guardDuration = duration ?? null
    this.guardFps = fps ?? null
    this.showAdditionalSprites()
  }

  stop(): void {
    this.isVisible = false
    this.isGuard = false
    this.currentAnim = null
    this.currentFrameHurtboxes = null
    this.currentPhase = AnimPhase.None
    this.phaseDurations = null
    this.guardDuration = null
    this.guardFps = null
    this.phaseFrameOffsets = []
    this.hideAdditionalSprites()
  }

  isPlaying(): boolean {
    return this.isVisible && !this.isGuard
  }

  getPhase(): AnimPhase {
    return this.currentPhase
  }

  getAnimName(): string | null {
    if (!this.currentAnim) return null
    for (const [name, anim] of Object.entries(this.def.animations)) {
      if (anim === this.currentAnim) return name
    }
    return null
  }

  getAnimDef(name: string): MovementAnimDef | AttackAnimDef | null {
    return this.def.animations[name] ?? null
  }

  getAttackAnimDef(name: string): AttackAnimDef | null {
    const anim = this.def.animations[name]
    return (anim && "guard" in anim) ? (anim as AttackAnimDef) : null
  }

  getCurrentFrameHurtboxes(): HurtboxData[] | null {
    return this.currentFrameHurtboxes
  }

  getDefaultHurtboxes(): HurtboxData[] | undefined {
    return this.def.defaultHurtboxes
  }

  update(
    charX: number,
    charY: number,
    charFlipX: boolean,
    charScale: number = 1,
  ): void {
    if (!this.isVisible || !this.currentAnim) {
      this.hideAdditionalSprites()
      return
    }

    if (this.isGuard) {
      this.updateGuard(charX, charY, charFlipX, charScale)
      return
    }

    const elapsed = this.scene.time.now - this.startTime
    const anim = this.currentAnim

    let frameIndex: number
    let totalFrames: number

    if (this.phaseDurations !== null && this.phaseFrameOffsets.length > 0) {
      let accumulatedTime = 0
      frameIndex = -1
      for (const group of this.phaseFrameOffsets) {
        let groupDuration: number
        if (group.phase === AnimPhase.Windup) groupDuration = this.phaseDurations.windup
        else if (group.phase === AnimPhase.Active) groupDuration = this.phaseDurations.active
        else groupDuration = this.phaseDurations.recover
        const perFrame = groupDuration / group.count
        const groupEnd = accumulatedTime + groupDuration
        if (elapsed < groupEnd) {
          frameIndex = group.start + Math.min(Math.floor((elapsed - accumulatedTime) / perFrame), group.count - 1)
          break
        }
        accumulatedTime = groupEnd
      }
      if (frameIndex === -1) {
        this.stop()
        return
      }
      totalFrames = anim.frames.length
    } else {
      totalFrames = anim.frames.length
      const frameDuration = 1000 / (anim.fps || 14)
      frameIndex = Math.floor(elapsed / frameDuration)
      if (anim.loop) {
        frameIndex = frameIndex % totalFrames
      } else if (frameIndex >= totalFrames) {
        this.stop()
        return
      }
    }

    const frame = anim.frames[frameIndex]
    if (!frame) return

    this.currentFrameHurtboxes = frame.hurtboxes ?? null

    if ("phase" in frame && frame.phase) {
      const phaseMap: Record<string, AnimPhase> = {
        wu: AnimPhase.Windup,
        atk: AnimPhase.Active,
        rc: AnimPhase.Recover,
        guard: AnimPhase.None,
      }
      this.currentPhase = phaseMap[frame.phase] ?? AnimPhase.Active
    } else {
      this.resolvePhaseFromFrameCount(anim, frameIndex, totalFrames)
    }

    this.applyFrame(frame, charX, charY, charFlipX, charScale)
  }

  private resolvePhaseFromFrameCount(anim: MovementAnimDef | AttackAnimDef, frameIndex: number, totalFrames: number): void {
    const wf = (anim as AttackAnimDef).windupFrames ?? 0
    const af = (anim as AttackAnimDef).activeFrames
    if (af !== undefined) {
      if (frameIndex < wf) this.currentPhase = AnimPhase.Windup
      else if (frameIndex < wf + af) this.currentPhase = AnimPhase.Active
      else this.currentPhase = AnimPhase.Recover
    } else {
      const rf = (anim as AttackAnimDef).recoverFrames ?? 0
      if (frameIndex < wf) this.currentPhase = AnimPhase.Windup
      else if (frameIndex >= totalFrames - rf) this.currentPhase = AnimPhase.Recover
      else this.currentPhase = AnimPhase.Active
    }
  }

  private updateGuard(charX: number, charY: number, charFlipX: boolean, charScale: number): void {
    const elapsed = this.scene.time.now - this.guardStartTime
    const anim = this.currentAnim!
    const totalGuardDuration = this.guardDuration ?? (anim as AttackAnimDef).guard ?? 1500
    if (elapsed >= totalGuardDuration) {
      this.stop()
      return
    }
    const frames = anim.frames
    const guardFrames = frames.filter(f => "phase" in f && f.phase === "guard")
    const loopFrames = guardFrames.length > 0 ? guardFrames : frames.slice(-3)
    const effectiveFps = this.guardFps ?? (anim as AttackAnimDef).guardFps ?? anim.fps ?? 14
    const idx = Math.floor(elapsed / (1000 / effectiveFps)) % loopFrames.length
    const frame = loopFrames[idx] ?? frames[frames.length - 1]
    if (frame) this.applyFrame(frame, charX, charY, charFlipX, charScale)
  }

  private applyFrame(
    frame: Frame,
    charX: number,
    charY: number,
    charFlipX: boolean,
    charScale: number,
  ): void {
    for (let i = 0; i < this.sprites.length; i++) {
      const sprite = this.sprites[i]

      const sf = frame.spriteFrames[i]
      const ox = frame.offsetX[i]
      const oy = frame.offsetY[i]
      const rot = frame.rotation[i]
      const sx = frame.scaleX?.[i] ?? 1
      const sy = frame.scaleY?.[i] ?? 1

      if (sf !== undefined) sprite.setFrame(sf)
      if (ox !== undefined && oy !== undefined) {
        const offsetX = (charFlipX ? -ox : ox) * charScale
        const offsetY = oy * charScale
        sprite.setPosition(charX + offsetX, charY + offsetY)
      }
      if (rot !== undefined) {
        sprite.setRotation(Phaser.Math.DegToRad(charFlipX ? -rot : rot))
      }
      if (i === 0) {
        sprite.setScale(Math.abs(sx) * charScale, sy * charScale)
      } else {
        const flip = charFlipX ? -1 : 1
        sprite.setScale(sx * flip * charScale, sy * charScale)
      }
    }
  }

  setDepth(depth: number): void {
    for (let i = 1; i < this.sprites.length; i++) {
      this.sprites[i].setDepth(depth)
    }
  }

  private showAdditionalSprites(): void {
    for (let i = 1; i < this.sprites.length; i++) {
      this.sprites[i].setVisible(true)
    }
  }

  private hideAdditionalSprites(): void {
    for (let i = 1; i < this.sprites.length; i++) {
      this.sprites[i].setVisible(false)
    }
  }

  destroy(): void {
    for (let i = 1; i < this.sprites.length; i++) {
      this.sprites[i].destroy()
    }
  }
}
