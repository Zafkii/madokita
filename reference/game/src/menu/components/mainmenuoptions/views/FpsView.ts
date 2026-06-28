import Phaser from "phaser"
import { MainMenuView } from "../MainMenuView"
import { MainMenuState } from "../MainMenuState"
import { SubMenuLayout } from "../SubMenuLayout"
import { GameSettingsManager } from "../../../../game/settings/GameSettingsManager"

type Callbacks = {
  onBack: () => void
}

export class FpsView extends MainMenuView {
  private state: MainMenuState
  private callbacks: Callbacks
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

    const title = this.track(this.scene.add.text(
      cx,
      SubMenuLayout.titleY(h),
      "FPS LIMIT",
      SubMenuLayout.style({ fontSize: SubMenuLayout.titleFont(h), color: SubMenuLayout.C_NORMAL }),
    ))
    title.setOrigin(0.5)

    const currentText = this.track(this.scene.add.text(
      lx,
      cy,
      `Current: ${this.state.fps} FPS`,
      SubMenuLayout.style({ fontSize: SubMenuLayout.infoFont(h), color: SubMenuLayout.C_MUTED }),
    ))
    currentText.setOrigin(0.5)

    const fpsOptions = [120, 60, 30]

    fpsOptions.forEach((fps, i) => {
      const isActive = fps === this.state.fps
      const text = this.track(this.scene.add.text(
        rx,
        cy + i * sp,
        `${fps} FPS`,
        SubMenuLayout.style({ fontSize: SubMenuLayout.optionFont(h), color: isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL }),
      ))
      text.setOrigin(0.5)
      text.setInteractive({ useHandCursor: true })

      if (isActive) {
        text.setAlpha(0.6)
      }

      text.on("pointerover", () => {
        if (!isActive) { text.setColor(SubMenuLayout.C_ACTIVE); text.setScale(1.05) }
      })
      text.on("pointerout", () => {
        text.setColor(isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL); text.setScale(1)
      })
      text.on("pointerdown", async () => {
        if (isActive) return
        this.state.fps = fps
        this.scene.game.loop.targetFps = fps
        try {
          await GameSettingsManager.setFpsLimit(fps)
        } catch (e) {
          console.error("Failed to set FPS limit:", e)
        }
        this.callbacks.onBack()
      })
    })

    const backText = this.track(this.scene.add.text(
      rx,
      cy + (fpsOptions.length + 1) * sp,
      "BACK",
      SubMenuLayout.style({ fontSize: SubMenuLayout.optionFont(h), color: SubMenuLayout.C_NORMAL }),
    ))
    backText.setOrigin(0.5)
    backText.setInteractive({ useHandCursor: true })
    backText.on("pointerover", () => backText.setColor(SubMenuLayout.C_ACTIVE))
    backText.on("pointerout", () => backText.setColor(SubMenuLayout.C_NORMAL))
    backText.on("pointerdown", () => this.callbacks.onBack())
  }
}
