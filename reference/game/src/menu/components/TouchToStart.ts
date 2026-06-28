import Phaser from "phaser"
import { TranslationKeys } from "../../game/localization/TranslationKeys"
import { MENU_LAYOUT } from "../config/menuLayoutConfig"
import { SubMenuLayout } from "./mainmenuoptions/SubMenuLayout"

export class TouchToStart {
  private scene: Phaser.Scene
  private text!: Phaser.GameObjects.Text

  constructor(scene: Phaser.Scene) {
    this.scene = scene
  }

  create(): void {
    const h = this.scene.scale.height
    const scaleRatio = h / MENU_LAYOUT.DESIGN_HEIGHT
    const y = h * MENU_LAYOUT.TOUCH_Y_RATIO
    const fontSize = Math.round(MENU_LAYOUT.FONT_TOUCH * scaleRatio)

    this.text = this.scene.add.text(
      this.scene.scale.width / 2,
      y,
      TranslationKeys.MENU.TOUCH_TO_START,
      SubMenuLayout.style({
        fontSize: `${fontSize}px`,
        color: SubMenuLayout.C_NORMAL,
      }),
    )

    this.text.setOrigin(0.5)

    this.scene.tweens.add({
      targets: this.text,
      alpha: 0.3,
      duration: 900,
      yoyo: true,
      repeat: -1,
    })
  }

  handleResize(): void {
    const h = this.scene.scale.height
    const scaleRatio = h / MENU_LAYOUT.DESIGN_HEIGHT
    const y = h * MENU_LAYOUT.TOUCH_Y_RATIO
    const fontSize = Math.round(MENU_LAYOUT.FONT_TOUCH * scaleRatio)
    this.text.setPosition(this.scene.scale.width / 2, y)
    this.text.setFontSize(fontSize)
  }

  hide(onComplete?: () => void): void {
    this.scene.tweens.add({
      targets: this.text,
      alpha: 0,
      y: this.text.y - 20,
      duration: 350,
      ease: "Sine.easeInOut",
      onComplete: () => {
        this.text.destroy()
        onComplete?.()
      },
    })
  }
}
