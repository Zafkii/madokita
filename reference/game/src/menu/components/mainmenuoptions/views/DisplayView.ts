import Phaser from "phaser"
import { MainMenuView } from "../MainMenuView"
import { MainMenuState } from "../MainMenuState"
import { SubMenuLayout } from "../SubMenuLayout"

type Callbacks = {
  onResolution: () => void
  onFps: () => void
  onToggleDisplayMode: () => void
  onBack: () => void
}

export class DisplayView extends MainMenuView {
  private state: MainMenuState
  private callbacks: Callbacks
  private modeText?: Phaser.GameObjects.Text
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
        "DISPLAY",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.titleFont(h),
          color: SubMenuLayout.C_NORMAL,
        }),
      ),
    )
    title.setOrigin(0.5)

    const resolutionText = this.track(
      this.scene.add.text(
        lx,
        cy,
        `RESOLUTION: ${this.state.resolution}`,
        SubMenuLayout.style({
          fontSize: SubMenuLayout.infoFont(h),
          color: SubMenuLayout.C_ACTIVE,
        }),
      ),
    )
    resolutionText.setOrigin(0.5)

    const fpsText = this.track(
      this.scene.add.text(
        lx,
        cy + sp,
        `FPS LIMIT: ${this.state.fps}`,
        SubMenuLayout.style({
          fontSize: SubMenuLayout.infoFont(h),
          color: SubMenuLayout.C_ACTIVE,
        }),
      ),
    )
    fpsText.setOrigin(0.5)

    this.modeText = this.track(
      this.scene.add.text(
        lx,
        cy + 2 * sp,
        `MODE: ${this.state.displayMode}`,
        SubMenuLayout.style({
          fontSize: SubMenuLayout.infoFont(h),
          color: SubMenuLayout.C_ACTIVE,
        }),
      ),
    ) as Phaser.GameObjects.Text
    this.modeText.setOrigin(0.5)

    const buttons = [
      { label: "RESOLUTION", action: () => this.callbacks.onResolution() },
      { label: "FPS LIMIT", action: () => this.callbacks.onFps() },
      { label: "DISPLAY MODE", action: () => { this.callbacks.onToggleDisplayMode(); this.modeText?.setText(`MODE: ${this.state.displayMode}`) } },
      { label: "BACK", action: () => this.callbacks.onBack() },
    ]

    buttons.forEach((btn, i) => {
      const text = this.track(
        this.scene.add.text(
          rx,
          cy + i * sp,
          btn.label,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.optionFont(h),
            color: SubMenuLayout.C_NORMAL,
          }),
        ),
      )
      text.setOrigin(0.5)
      text.setInteractive({ useHandCursor: true })
      text.on("pointerover", () => {
        text.setColor(SubMenuLayout.C_ACTIVE)
        text.setScale(1.05)
      })
      text.on("pointerout", () => {
        text.setColor(SubMenuLayout.C_NORMAL)
        text.setScale(1)
      })
      text.on("pointerdown", () => btn.action())
    })
  }
}
