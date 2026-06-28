import Phaser from "phaser"
import { MainMenuView } from "../MainMenuView"
import { MainMenuState } from "../MainMenuState"
import { SubMenuLayout } from "../SubMenuLayout"
import { GameSettingsManager } from "../../../../game/settings/GameSettingsManager"
import { INPUT_ACTIONS } from "../../../../game/input/InputAction"
import type { InputAction } from "../../../../game/input/InputAction"

const ACTION_ENTRIES: { action: InputAction; label: string }[] = [
  { action: INPUT_ACTIONS.MOVE_LEFT, label: "MOVE LEFT" },
  { action: INPUT_ACTIONS.MOVE_RIGHT, label: "MOVE RIGHT" },
  { action: INPUT_ACTIONS.JUMP, label: "JUMP" },
  { action: INPUT_ACTIONS.DODGE, label: "DODGE" },
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

export class NavigationView extends MainMenuView {
  private state: MainMenuState
  private callbacks: Callbacks
  private _keyCaptureHandler: ((event: KeyboardEvent) => void) | null = null
  private scrollItems: { text: Phaser.GameObjects.Text; baseY: number }[] = []
  private scrollY = 0
  private maxScroll = 0

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
        "NAVIGATION",
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

    const backY = cy + (ACTION_ENTRIES.length + 1) * sp
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
