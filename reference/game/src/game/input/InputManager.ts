import Phaser from "phaser"
import { DEFAULT_INPUT_BINDINGS } from "./InputBindings"
import { INPUT_ACTIONS } from "./InputAction"
import type { InputAction } from "./InputAction"
export class InputManager {
  private scene: Phaser.Scene
  private keys = new Map<InputAction, Phaser.Input.Keyboard.Key>()

  constructor(scene: Phaser.Scene) {
    this.scene = scene
    this.createBindings()
  }

  private createBindings(): void {
    const actions = Object.values(INPUT_ACTIONS) as InputAction[]
    for (const action of actions) {
      const key = DEFAULT_INPUT_BINDINGS[action]
      this.keys.set(action, this.scene.input.keyboard!.addKey(key))
    }
  }

  get(action: InputAction): Phaser.Input.Keyboard.Key {
    const key = this.keys.get(action)
    if (!key) {
      throw new Error(`Missing key binding for ${action}`)
    }
    return key
  }
}
