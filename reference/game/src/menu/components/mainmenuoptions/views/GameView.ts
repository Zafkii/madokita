import Phaser from "phaser"
import { MainMenuView } from "../MainMenuView"
import { MainMenuState } from "../MainMenuState"
import { SubMenuLayout } from "../SubMenuLayout"
import { GameSettingsManager } from "../../../../game/settings/GameSettingsManager"
import { INPUT_ACTIONS } from "../../../../game/input/InputAction"
import type { InputAction } from "../../../../game/input/InputAction"

const ACTION_ENTRIES: { action: InputAction; label: string }[] = [
  { action: INPUT_ACTIONS.ATTACK, label: "ATTACK" },
  { action: INPUT_ACTIONS.SKILL, label: "SKILL" },
  { action: INPUT_ACTIONS.TRANSFORM, label: "TRANSFORM" },
]

function normalizeKey(eventKey: string): string {
  if (eventKey === " ") return "SPACE"
  const upper = eventKey.toUpperCase()
  const map: Record<string, string> = {
    ARROWLEFT: "LEFT",
    ARROWRIGHT: "RIGHT",
    ARROWUP: "UP",
    ARROWDOWN: "DOWN",
    ESCAPE: "ESC",
  }
  return map[upper] ?? upper
}

type Callbacks = {
  onBack: () => void
}

export class GameView extends MainMenuView {
  private state: MainMenuState
  private callbacks: Callbacks
  private _keyCaptureHandler: ((event: KeyboardEvent) => void) | null = null
  private scrollItems: { text: Phaser.GameObjects.Text; baseY: number }[] = []
  private scrollY = 0
  private maxScroll = 0
  private sensValueText?: Phaser.GameObjects.Text

  constructor(scene: Phaser.Scene, state: MainMenuState, callbacks: Callbacks) {
    super(scene)
    this.state = state
    this.callbacks = callbacks
  }

  create(): void {
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const cx = w / 2
    const lx = SubMenuLayout.leftX(w)
    const rx = SubMenuLayout.rightX(w)
    const cy = SubMenuLayout.contentY(h)
    const sp = SubMenuLayout.spacing(h)

    const title = this.track(
      this.scene.add.text(
        cx,
        SubMenuLayout.titleY(h),
        "GAME",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.titleFont(h),
          color: SubMenuLayout.C_NORMAL,
        }),
      ),
    )
    title.setOrigin(0.5)

    const infoTexts: Phaser.GameObjects.Text[] = []

    ACTION_ENTRIES.forEach((entry, i) => {
      const y = cy + i * sp
      const info = this.track(
        this.scene.add.text(
          lx,
          y,
          `${entry.label}: ${this.state.keyBindings[entry.action]}`,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.infoFont(h),
            color: SubMenuLayout.C_ACTIVE,
          }),
        ),
      ) as Phaser.GameObjects.Text
      info.setOrigin(0.5)
      infoTexts.push(info)
      this.scrollItems.push({ text: info, baseY: y })
    })

    ACTION_ENTRIES.forEach((entry, i) => {
      const y = cy + i * sp
      const opt = this.track(
        this.scene.add.text(
          rx,
          y,
          entry.label,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.optionFont(h),
            color: SubMenuLayout.C_NORMAL,
          }),
        ),
      ) as Phaser.GameObjects.Text
      opt.setOrigin(0.5)
      opt.setInteractive({ useHandCursor: true })
      this.scrollItems.push({ text: opt, baseY: y })

      opt.on("pointerover", () => {
        if (this._keyCaptureHandler) return
        opt.setColor(SubMenuLayout.C_ACTIVE)
        opt.setScale(1.05)
      })
      opt.on("pointerout", () => {
        if (this._keyCaptureHandler) return
        opt.setColor(SubMenuLayout.C_NORMAL)
        opt.setScale(1)
      })
      opt.on("pointerdown", () => {
        if (this._keyCaptureHandler) return

        this._keyCaptureHandler = (event: KeyboardEvent) => {
          event.preventDefault()
          const key = normalizeKey(event.key)

          GameSettingsManager.setKeyBinding(entry.action, key)
          this.state.keyBindings[entry.action] = key
          infoTexts[i].setText(`${entry.label}: ${key}`)
          opt.setText(entry.label)
          opt.setColor(SubMenuLayout.C_NORMAL)
          opt.setScale(1)
          window.removeEventListener("keydown", this._keyCaptureHandler!)
          this._keyCaptureHandler = null
        }

        opt.setText("...")
        opt.setColor(SubMenuLayout.C_ACTIVE)
        opt.setScale(1.05)
        window.addEventListener("keydown", this._keyCaptureHandler)
      })
    })

    const sensRow = ACTION_ENTRIES.length
    const sensY = cy + sensRow * sp
    const sensInfo = this.track(
      this.scene.add.text(
        lx,
        sensY,
        `SENSITIVITY: ${this.state.sensitivity}`,
        SubMenuLayout.style({
          fontSize: SubMenuLayout.infoFont(h),
          color: SubMenuLayout.C_ACTIVE,
        }),
      ),
    ) as Phaser.GameObjects.Text
    sensInfo.setOrigin(0.5)
    this.scrollItems.push({ text: sensInfo, baseY: sensY })

    const sensLabel = this.track(
      this.scene.add.text(
        rx,
        sensY,
        "SENSITIVITY",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.optionFont(h),
          color: SubMenuLayout.C_NORMAL,
        }),
      ),
    ) as Phaser.GameObjects.Text
    sensLabel.setOrigin(0.5)
    this.scrollItems.push({ text: sensLabel, baseY: sensY })

    const sensValStyle = SubMenuLayout.style({
      fontSize: SubMenuLayout.optionFont(h),
      color: SubMenuLayout.C_ACTIVE,
    })
    const minusBtn = this.track(
      this.scene.add.text(rx - 60, sensY + sp, "[-]", sensValStyle),
    ) as Phaser.GameObjects.Text
    minusBtn.setOrigin(0.5)
    minusBtn.setInteractive({ useHandCursor: true })
    this.scrollItems.push({ text: minusBtn, baseY: sensY + sp })

    this.sensValueText = this.track(
      this.scene.add.text(rx, sensY + sp, `${this.state.sensitivity}`, sensValStyle),
    ) as Phaser.GameObjects.Text
    this.sensValueText.setOrigin(0.5)
    this.scrollItems.push({ text: this.sensValueText, baseY: sensY + sp })

    const plusBtn = this.track(
      this.scene.add.text(rx + 60, sensY + sp, "[+]", sensValStyle),
    ) as Phaser.GameObjects.Text
    plusBtn.setOrigin(0.5)
    plusBtn.setInteractive({ useHandCursor: true })
    this.scrollItems.push({ text: plusBtn, baseY: sensY + sp })

    minusBtn.on("pointerdown", () => {
      if (this._keyCaptureHandler) return
      const newVal = Math.max(5, this.state.sensitivity - 5)
      this.state.sensitivity = newVal
      GameSettingsManager.setSensitivity(newVal)
      sensInfo.setText(`SENSITIVITY: ${newVal}`)
      this.sensValueText?.setText(`${newVal}`)
    })

    plusBtn.on("pointerdown", () => {
      if (this._keyCaptureHandler) return
      const newVal = Math.min(100, this.state.sensitivity + 5)
      this.state.sensitivity = newVal
      GameSettingsManager.setSensitivity(newVal)
      sensInfo.setText(`SENSITIVITY: ${newVal}`)
      this.sensValueText?.setText(`${newVal}`)
    })

    const backY = sensY + 2 * sp
    const backText = this.track(
      this.scene.add.text(
        rx,
        backY,
        "BACK",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.optionFont(h),
          color: SubMenuLayout.C_NORMAL,
        }),
      ),
    ) as Phaser.GameObjects.Text
    backText.setOrigin(0.5)
    backText.setInteractive({ useHandCursor: true })
    this.scrollItems.push({ text: backText, baseY: backY })
    backText.on("pointerover", () => backText.setColor(SubMenuLayout.C_ACTIVE))
    backText.on("pointerout", () => backText.setColor(SubMenuLayout.C_NORMAL))
    backText.on("pointerdown", () => {
      if (this._keyCaptureHandler) {
        window.removeEventListener("keydown", this._keyCaptureHandler)
        this._keyCaptureHandler = null
      }
      this.callbacks.onBack()
    })

    const lastBottom = backY + (backText.height || 20)
    this.maxScroll = Math.max(0, lastBottom - (h - 20))

    if (this.maxScroll > 0) {
      this.scene.input.on("wheel", (_pointer: Phaser.Input.Pointer, _gameObjects: Phaser.GameObjects.GameObject[], _dx: number, dy: number) => {
        this.scrollY = Phaser.Math.Clamp(this.scrollY - dy, -this.maxScroll, 0)
        this.scrollItems.forEach(item => {
          item.text.y = item.baseY + this.scrollY
        })
      })
    }
  }
}
