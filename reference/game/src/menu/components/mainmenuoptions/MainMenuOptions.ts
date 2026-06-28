import Phaser from "phaser"
import { AnimatedTitle } from "../../entities/title/AnimatedTitle"
import { MainMenuState } from "./MainMenuState"
import { MainMenuView } from "./MainMenuView"
import { MainOptionsView } from "./views/MainOptionsView"
import { SettingsView } from "./views/SettingsView"
import { GameSettingsManager } from "../../../game/settings/GameSettingsManager"

export class MainMenuOptions {
  private scene: Phaser.Scene
  private animatedTitle: AnimatedTitle
  private state: MainMenuState
  private currentView?: MainMenuView

  constructor(scene: Phaser.Scene, animatedTitle: AnimatedTitle) {
    this.scene = scene
    this.animatedTitle = animatedTitle
    const settings = GameSettingsManager.getData()
    this.state = new MainMenuState()
    this.state.resolution = `${settings.resolution.width}x${settings.resolution.height}`
    this.state.fps = settings.fpsLimit
    this.state.displayMode = settings.fullscreen ? "Fullscreen" : "Windowed"
    this.state.keyBindings = { ...settings.keyBindings }
    this.state.sensitivity = settings.sensitivity
  }

  create(): void {
    this.showMainOptions()
  }

  private setView(view: MainMenuView): void {
    this.currentView?.destroy()
    this.currentView = view
    view.create()
  }

  showMainOptions(): void {
    this.setView(
      new MainOptionsView(this.scene, this.animatedTitle, {
        onSettings: () => this.showSettings(),
      }),
    )
  }

  showSettings(): void {
    this.setView(
      new SettingsView(this.scene, this.state, {
        onBack: () => this.showMainOptions(),
      }),
    )
  }

  handleResize(width: number, height: number): void {
    this.currentView?.handleResize(width, height)
  }

  focusNext(): void {
    this.currentView?.focusNext()
  }

  focusPrev(): void {
    this.currentView?.focusPrev()
  }

  focusConfirm(): void {
    this.currentView?.focusConfirm()
  }

  focusBack(): void {
    this.currentView?.focusBack()
  }

  destroy(): void {
    this.currentView?.destroy()
  }
}
