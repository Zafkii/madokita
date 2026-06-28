import Phaser from "phaser"
import { Player } from "../entities/playable/player/Player"
import { GameSettingsManager } from "../settings/GameSettingsManager"
import { INPUT_ACTIONS } from "../input/InputAction"
import { DEFAULT_INPUT_BINDINGS } from "../input/InputBindings"
import type { InputAction } from "../input/InputAction"

function resolveKey(action: InputAction): string {
  const saved = GameSettingsManager.getData().keyBindings[action]
  if (saved && saved.toUpperCase() in Phaser.Input.Keyboard.KeyCodes) {
    return saved
  }
  return DEFAULT_INPUT_BINDINGS[action]
}

export class ControlsManager {
  private left: Phaser.Input.Keyboard.Key
  private right: Phaser.Input.Keyboard.Key
  private jump: Phaser.Input.Keyboard.Key
  private dodge: Phaser.Input.Keyboard.Key
  private attack: Phaser.Input.Keyboard.Key
  private skill: Phaser.Input.Keyboard.Key
  private burst: Phaser.Input.Keyboard.Key

  constructor(scene: Phaser.Scene) {
    this.left = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.MOVE_LEFT))
    this.right = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.MOVE_RIGHT))
    this.jump = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.JUMP))
    this.dodge = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.DODGE))
    this.attack = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.ATTACK))
    this.skill = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.SKILL))
    this.burst = scene.input.keyboard!.addKey(resolveKey(INPUT_ACTIONS.TRANSFORM))
  }

  update(player: Player): void {
    // BLOCK MOVEMENT
    if (player.playerState.isBursting || player.playerState.isStaggered || player.playerState.isDead) {
      player.idle()
      return
    }
    // HORIZONTAL MOVEMENT
    if (this.left.isDown) {
      player.moveLeft()
    } else if (this.right.isDown) {
      player.moveRight()
    } else {
      player.idle()
    }
    // ACTIONS
    if (Phaser.Input.Keyboard.JustDown(this.jump)) {
      player.jump()
    }
    if (Phaser.Input.Keyboard.JustDown(this.dodge)) {
      player.dodge()
    }
    if (Phaser.Input.Keyboard.JustDown(this.attack)) {
      if (player.hasAttack("vertical")) {
        player.attack("vertical")
      } else {
        player.attack("basic")
      }
    }
    if (Phaser.Input.Keyboard.JustDown(this.skill)) {
      if (player.hasAttack("horizontal")) {
        player.attack("horizontal")
      } else {
        player.attack("basic")
      }
    }
    if (Phaser.Input.Keyboard.JustDown(this.burst)) {
      player.burst()
    }
  }
}
