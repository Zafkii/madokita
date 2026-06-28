import Phaser from "phaser"
import { MENU_LAYOUT } from "../../config/menuLayoutConfig"

export interface FocusableEntry {
  text: Phaser.GameObjects.Text
  callback: () => void
}

export abstract class MainMenuView {
  protected scene: Phaser.Scene
  protected objects: Phaser.GameObjects.GameObject[] = []
  protected focusableItems: FocusableEntry[] = []
  protected focusIndex = -1
  private _captureHandler: ((event: KeyboardEvent) => void) | null = null

  constructor(scene: Phaser.Scene) {
    this.scene = scene
  }

  abstract create(): void

  static normalizeKey(eventKey: string): string {
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

  setCaptureHandler(handler: ((event: KeyboardEvent) => void) | null): void {
    this._captureHandler = handler
  }

  get captureHandler(): ((event: KeyboardEvent) => void) | null {
    return this._captureHandler
  }

  registerFocus(text: Phaser.GameObjects.Text, callback: () => void): void {
    this.focusableItems.push({ text, callback })
  }

  clearFocusableItems(): void {
    this.focusableItems = []
    this.focusIndex = -1
  }

  focusNext(): void {
    if (this.focusableItems.length === 0) return
    this.unfocusCurrent()
    this.focusIndex = (this.focusIndex + 1) % this.focusableItems.length
    this.focusCurrent()
  }

  focusPrev(): void {
    if (this.focusableItems.length === 0) return
    this.unfocusCurrent()
    this.focusIndex =
      (this.focusIndex - 1 + this.focusableItems.length) % this.focusableItems.length
    this.focusCurrent()
  }

  focusConfirm(): void {
    if (this._captureHandler) return
    if (this.focusIndex < 0 || this.focusIndex >= this.focusableItems.length)
      return
    this.focusableItems[this.focusIndex].callback()
  }

  focusBack(): void {}

  handleResize(_width: number, _height: number): void {
    // Override in subclasses that need to re-layout on resolution change
  }

  destroy(): void {
    this.objects.forEach((obj) => {
      obj.destroy()
    })
    this.objects = []
  }

  protected track<T extends Phaser.GameObjects.GameObject>(obj: T): T {
    this.objects.push(obj)
    return obj
  }

  protected unfocusCurrent(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.focusableItems.length)
      return
    const item = this.focusableItems[this.focusIndex]
    item.text.setColor(MENU_LAYOUT.COLOR_NORMAL)
    item.text.setScale(1)
  }

  protected focusCurrent(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.focusableItems.length)
      return
    const item = this.focusableItems[this.focusIndex]
    item.text.setColor(MENU_LAYOUT.COLOR_ACTIVE)
    item.text.setScale(1.05)
  }
}
