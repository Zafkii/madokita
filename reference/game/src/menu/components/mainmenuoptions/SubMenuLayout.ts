import Phaser from "phaser"
import { MENU_LAYOUT } from "../../config/menuLayoutConfig"

export class SubMenuLayout {
  static readonly C_NORMAL = MENU_LAYOUT.COLOR_NORMAL
  static readonly C_ACTIVE = MENU_LAYOUT.COLOR_ACTIVE
  static readonly C_MUTED = MENU_LAYOUT.COLOR_MUTED

  static y(designY: number, h: number): number {
    return Math.round(designY * h / MENU_LAYOUT.DESIGN_HEIGHT)
  }

  static fontSize(designPx: number, h: number): string {
    return `${Math.round(designPx * h / MENU_LAYOUT.DESIGN_HEIGHT)}px`
  }

  static titleFont(h: number): string {
    return this.fontSize(MENU_LAYOUT.FONT_TITLE, h)
  }

  static optionFont(h: number): string {
    return this.fontSize(MENU_LAYOUT.FONT_OPTION, h)
  }

  static infoFont(h: number): string {
    return this.fontSize(MENU_LAYOUT.FONT_INFO, h)
  }

  static style(base: Phaser.Types.GameObjects.Text.TextStyle): Phaser.Types.GameObjects.Text.TextStyle {
    return {
      ...base,
      font: `${MENU_LAYOUT.FONT_WEIGHT} ${base.fontSize ?? "24px"} ${MENU_LAYOUT.FONT_FAMILY}`,
      resolution: MENU_LAYOUT.TEXT_RESOLUTION,
      padding: { x: MENU_LAYOUT.TEXT_PADDING_X, y: MENU_LAYOUT.TEXT_PADDING_Y },
    }
  }

  static titleY(h: number): number {
    return Math.round(h * MENU_LAYOUT.TITLE_Y_RATIO)
  }

  static contentY(h: number): number {
    return Math.round(h * MENU_LAYOUT.CONTENT_Y_RATIO)
  }

  static leftX(w: number): number {
    return Math.round(w * MENU_LAYOUT.LEFT_X_RATIO)
  }

  static rightX(w: number): number {
    return Math.round(w * MENU_LAYOUT.RIGHT_X_RATIO)
  }

  static navX(w: number): number {
    return Math.round(w * MENU_LAYOUT.NAV_X_RATIO)
  }

  static infoX(w: number): number {
    return Math.round(w * MENU_LAYOUT.INFO_X_RATIO)
  }

  static actionX(w: number): number {
    return Math.round(w * MENU_LAYOUT.ACTION_X_RATIO)
  }

  static spacing(h: number): number {
    const scaledMin = Math.round(MENU_LAYOUT.SPACING_MIN * h / MENU_LAYOUT.DESIGN_HEIGHT)
    return Math.round(Math.max(scaledMin, h * MENU_LAYOUT.SPACING_RATIO))
  }
}
