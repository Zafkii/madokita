import Phaser from "phaser"
import { MainMenuView } from "../MainMenuView"
import { MenuOptionList } from "../../MenuOptionList"
import { AnimatedTitle } from "../../../entities/title/AnimatedTitle"
import { TranslationKeys } from "../../../../game/localization/TranslationKeys"
import { SaveManager } from "../../../../game/save/SaveManager"

type Callbacks = {
  onSettings: () => void
}

export class MainOptionsView extends MainMenuView {
  private animatedTitle: AnimatedTitle
  private callbacks: Callbacks
  private optionList?: MenuOptionList

  constructor(
    scene: Phaser.Scene,
    animatedTitle: AnimatedTitle,
    callbacks: Callbacks,
  ) {
    super(scene)
    this.animatedTitle = animatedTitle
    this.callbacks = callbacks
  }

  create(): void {
    this.optionList?.destroy()
    this.optionList = new MenuOptionList(this.scene, [
      {
        label: TranslationKeys.MENU.NEW_GAME,
        onSelect: () => {
          SaveManager.update({
            currentLocation: {
              stageId: "stage1",
              players: { sayaka: { x: 2000, y: 400 } },
            },
            progress: {
              stagesUnlocked: ["stage1"],
              charactersUnlocked: ["madoka"],
              endingsUnlocked: [],
            },
            player: {
              upgrades: {},
              unlockedSkills: [],
            },
          })
          this.animatedTitle.destroy()
          this.scene.scene.start("game")
        },
      },

      {
        label: TranslationKeys.MENU.CONTINUE,
        onSelect: () => {
          this.animatedTitle.destroy()
          this.scene.scene.start("game")
        },
      },

      {
        label: TranslationKeys.MENU.SETTINGS,
        onSelect: () => {
          this.callbacks.onSettings()
        },
      },

      {
        label: TranslationKeys.MENU.EXIT,
        onSelect: () => {
          window.gamePlatform.closeGame()
        },
      },
    ])

    this.optionList.create()

    this.track({
      destroy: () => this.optionList?.destroy(),
    } as Phaser.GameObjects.GameObject)
  }

  handleResize(_width: number, _height: number): void {
    this.optionList?.handleResize()
  }

  focusNext(): void {
    this.optionList?.focusNext()
  }

  focusPrev(): void {
    this.optionList?.focusPrev()
  }

  focusConfirm(): void {
    this.optionList?.focusConfirm()
  }

  focusBack(): void {}
}
