import Phaser from "phaser"
import { MainMenuView } from "../MainMenuView"
import { MainMenuState } from "../MainMenuState"
import { MenuOptionList } from "../../MenuOptionList"

type Callbacks = {
  onNavigation: () => void
  onGame: () => void
  onBack: () => void
}

export class ControlsView extends MainMenuView {
  private callbacks: Callbacks
  constructor(scene: Phaser.Scene, _state: MainMenuState, callbacks: Callbacks) {
    super(scene)
    this.callbacks = callbacks
  }

  create(): void {
    const list = new MenuOptionList(this.scene, [
      {
        label: "NAVIGATION",
        onSelect: () => this.callbacks.onNavigation(),
      },
      {
        label: "GAME",
        onSelect: () => this.callbacks.onGame(),
      },
      {
        label: "BACK",
        onSelect: () => this.callbacks.onBack(),
      },
    ])

    list.create()
    this.track({
      destroy: () => list.destroy(),
    } as Phaser.GameObjects.GameObject)
  }
}
