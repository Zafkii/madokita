import Phaser from "phaser"
import { AnimatedTitle } from "../entities/title/AnimatedTitle"
import { MenuThemeController } from "../controllers/MenuThemeController"
import { TouchToStart } from "../components/TouchToStart"
import { MainMenuOptions } from "../components/mainmenuoptions/MainMenuOptions"
import { MainMenuView } from "../components/mainmenuoptions/MainMenuView"
import { GameSettingsManager } from "../../game/settings/GameSettingsManager"
import { INPUT_ACTIONS } from "../../game/input/InputAction"
import { menuTimings } from "../config/menuTimings"

export class MainMenuScene extends Phaser.Scene {
  private started = false
  private menuTheme!: MenuThemeController
  private animatedTitle!: AnimatedTitle
  private touchToStart!: TouchToStart
  private menuOptions!: MainMenuOptions
  private keydownHandler!: (event: KeyboardEvent) => void

  constructor() {
    super("main-menu")
  }

  create(): void {
    this.setupBackground()
    this.setupMusic()
    this.setupTitle()
    this.setupTouchPrompt()
    this.setupInput()
    this.setupResizeListener()
    this.setupKeyboardNav()
    this.setupCleanup()
  }

  update(_time: number, delta: number): void {
    this.animatedTitle?.update(delta)
  }

  private setupBackground(): void {
    this.cameras.main.setBackgroundColor("#000000")
  }

  private setupMusic(): void {
    this.menuTheme = new MenuThemeController(this)
    this.menuTheme.registerEffect("titleFill", {
      onStart: () => this.animatedTitle.showCosmic(),
      onEnd: () =>
        this.animatedTitle.hideCosmic(
          menuTimings.effects.titleFill.fadeOutDuration * 1000,
        ),
    })
    this.menuTheme.registerEffect("borderTint", {
      onStart: () =>
        this.animatedTitle.startOverlayCycle(
          menuTimings.effects.borderTint.colors!,
          menuTimings.effects.borderTint.cycleInterval!,
          menuTimings.effects.borderTint.colorTransition!,
        ),
      onEnd: () => {
        const timing = menuTimings.effects.borderTint
        this.animatedTitle.stopOverlayCycle(
          timing.initialColor,
          timing.fadeOutDuration,
        )
      },
    })
    this.menuTheme.play()
  }

  private setupTitle(): void {
    this.animatedTitle = new AnimatedTitle(this)
    this.animatedTitle.create()

    const initial = menuTimings.effects.borderTint.initialColor
    if (initial) {
      this.animatedTitle.setOverlayColor(initial.rgb, initial.alpha)
    }
  }

  private setupTouchPrompt(): void {
    this.touchToStart = new TouchToStart(this)
    this.touchToStart.create()
  }

  private setupInput(): void {
    this.input.once("pointerdown", () => {
      this.activateMenu()
    })
  }

  private setupKeyboardNav(): void {
    this.keydownHandler = (event: KeyboardEvent) => {
      const key = MainMenuView.normalizeKey(event.key)
      const bindings = GameSettingsManager.getData().keyBindings

      if (key === bindings[INPUT_ACTIONS.MENU_CONFIRM]) {
        event.preventDefault()
        if (!this.started) {
          this.activateMenu()
          return
        }
        this.menuOptions?.focusConfirm()
        return
      }

      if (!this.started) return

      if (key === bindings[INPUT_ACTIONS.MENU_UP]) {
        event.preventDefault()
        this.menuOptions?.focusPrev()
      } else if (key === bindings[INPUT_ACTIONS.MENU_DOWN]) {
        event.preventDefault()
        this.menuOptions?.focusNext()
      } else if (key === bindings[INPUT_ACTIONS.MENU_BACK]) {
        event.preventDefault()
        this.menuOptions?.focusBack()
      }
    }

    window.addEventListener("keydown", this.keydownHandler)
  }

  private activateMenu(): void {
    if (this.started) {
      return
    }
    this.started = true
    this.touchToStart.hide(() => {
      this.showOptions()
    })
  }

  private showOptions(): void {
    this.menuOptions = new MainMenuOptions(this, this.animatedTitle)
    this.menuOptions.create()
  }

  private setupResizeListener(): void {
    this.scale.on("resize", () => {
      this.animatedTitle.handleResize()
      this.touchToStart?.handleResize()
      this.menuOptions?.handleResize(this.scale.width, this.scale.height)
    })
  }

  private setupCleanup(): void {
    this.events.once("shutdown", () => {
      this.cleanup()
    })

    this.events.once("destroy", () => {
      this.cleanup()
    })
  }

  private cleanup(): void {
    if (this.keydownHandler) {
      window.removeEventListener("keydown", this.keydownHandler)
    }
    this.menuTheme.stop()
    this.animatedTitle?.destroy()
    this.menuOptions?.destroy()
  }
}
