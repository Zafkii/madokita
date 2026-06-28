import Phaser from "phaser"
import { BootScene } from "../scenes/BootScene"
import { PreloadScene } from "../scenes/PreloadScene"
import { GameScene } from "../scenes/GameScene"
import { MenuIntroScene } from "../scenes/MenuIntroScene"
import { MainMenuScene } from "../../menu/scenes/MainMenuScene"

export const BASE_WIDTH = 1280

export const BASE_HEIGHT = 720

export const gameConfig: Phaser.Types.Core.GameConfig = {
  type: Phaser.AUTO,
  width: BASE_WIDTH,
  height: BASE_HEIGHT,
  backgroundColor: "#181818",

  parent: "game-container",

  scale: {
    mode: Phaser.Scale.FIT,
    autoCenter: Phaser.Scale.CENTER_BOTH,
  },

  fps: {
    target: 60,
  },

  audio: {
    disableWebAudio: false,
    noAudio: false,
  },

  physics: {
    default: "arcade",
    arcade: {
      gravity: {
        x: 0,
        y: 1000,
      },
      debug: false,
    },
  },

  scene: [BootScene, PreloadScene, MenuIntroScene, MainMenuScene, GameScene],
}
