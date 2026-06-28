import { Player } from "../entities/playable/player/Player"

export class AISystem {
  private players: Player[]

  private controlledPlayer: () => Player

  constructor(
    players: Player[],

    controlledPlayer: () => Player,
  ) {
    this.players = players

    this.controlledPlayer = controlledPlayer
  }

  update(): void {
    const mainPlayer = this.controlledPlayer()

    this.players.forEach((player) => {
      if (player.isControlled()) {
        return
      }

      this.handleCompanionAI(player, mainPlayer)
    })
  }

  private handleCompanionAI(
    player: Player,

    mainPlayer: Player,
  ): void {
    const distance = mainPlayer.x - player.x

    // =========================
    // FOLLOW PLAYER
    // =========================

    if (Math.abs(distance) > 250) {
      if (distance > 0) {
        player.moveRight()
      } else {
        player.moveLeft()
      }
    } else {
      player.idle()
    }

    // =========================
    // RANDOM JUMP
    // =========================

    if (Math.random() < 0.002) {
      player.jump()
    }

    // =========================
    // RANDOM DASH
    // =========================

    if (Math.random() < 0.001) {
      player.dodge()
    }
  }
}
