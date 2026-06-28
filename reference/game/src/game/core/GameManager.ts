import type { GameState } from "./GameState"

export class GameManager {
  private static gameState: GameState = "MENU"
  static getState(): GameState {
    return this.gameState
  }

  static setState(state: GameState): void {
    this.gameState = state
  }
}
