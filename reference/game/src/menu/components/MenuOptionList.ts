import Phaser from "phaser"
import { SubMenuLayout } from "./mainmenuoptions/SubMenuLayout"
import { MENU_LAYOUT } from "../config/menuLayoutConfig"

export type MenuOption = {
  label: string
  onSelect: () => void
}

export class MenuOptionList {
  private scene: Phaser.Scene
  private options: MenuOption[]
  private texts: Phaser.GameObjects.Text[] = []
  private startY!: number
  private spacing!: number
  private origStartY?: number
  private origSpacing?: number
  private focusIndex = -1
  constructor(
    scene: Phaser.Scene,
    options: MenuOption[],
    startY?: number,
    spacing?: number,
  ) {
    this.origStartY = startY
    this.origSpacing = spacing
    this.scene = scene
    this.options = options
    this.computeLayout()
  }

  private computeLayout(): void {
    const scaleRatio = this.scene.scale.height / MENU_LAYOUT.DESIGN_HEIGHT
    this.spacing = this.origSpacing ?? Math.round(MENU_LAYOUT.LIST_SPACING * scaleRatio)
    const totalHeight = (this.options.length - 1) * this.spacing
    this.startY = this.origStartY ?? this.scene.scale.height / 2 + Math.round(MENU_LAYOUT.LIST_OFFSET * scaleRatio) - totalHeight / 2
  }
  create(): void {
    this.focusIndex = -1
    this.options.forEach((option, index) => {
      const finalY = this.startY + index * this.spacing
      const text = this.scene.add.text(
        this.scene.scale.width / 2,
        finalY - 8,
        option.label,
        SubMenuLayout.style({
          fontSize: SubMenuLayout.optionFont(this.scene.scale.height),
          color: SubMenuLayout.C_NORMAL,
        }),
      )

      text.setOrigin(0.5)
      text.setAlpha(0)
      text.setInteractive({ useHandCursor: true })

      // =====================================
      // HOVER
      // =====================================
      text.on("pointerover", () => {
        text.setScale(1.08)
        text.setColor(SubMenuLayout.C_ACTIVE)
      })
      text.on("pointerout", () => {
        text.setScale(1)
        text.setColor(SubMenuLayout.C_NORMAL)
      })
      // =====================================
      // CLICK
      // =====================================
      text.on("pointerdown", () => {
        option.onSelect()
      })
      // =====================================
      // APPEAR
      // =====================================

      this.scene.tweens.add({
        targets: text,
        alpha: 1,
        y: finalY,
        duration: 400,
        delay: index * 120,
        ease: "Sine.easeOut",
      })
      this.texts.push(text)
    })
  }

  handleResize(): void {
    this.destroy()
    this.computeLayout()
    this.create()
  }

  destroy(): void {
    this.texts.forEach((text) => {
      text.destroy()
    })
    this.texts = []
    this.focusIndex = -1
  }

  focusNext(): void {
    if (this.texts.length === 0) return
    this.unfocusCurrent()
    this.focusIndex = (this.focusIndex + 1) % this.texts.length
    this.focusCurrent()
  }

  focusPrev(): void {
    if (this.texts.length === 0) return
    this.unfocusCurrent()
    this.focusIndex =
      (this.focusIndex - 1 + this.texts.length) % this.texts.length
    this.focusCurrent()
  }

  focusConfirm(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.options.length) return
    this.options[this.focusIndex].onSelect()
  }

  private unfocusCurrent(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.texts.length) return
    this.texts[this.focusIndex].setColor(MENU_LAYOUT.COLOR_NORMAL)
    this.texts[this.focusIndex].setScale(1)
  }

  private focusCurrent(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.texts.length) return
    this.texts[this.focusIndex].setColor(MENU_LAYOUT.COLOR_ACTIVE)
    this.texts[this.focusIndex].setScale(1.08)
  }
}
