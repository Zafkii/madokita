import Phaser from "phaser"
import { AssetManifest, CharacterSpritesheets } from "../core/AssetLoader"

export class PreloadScene extends Phaser.Scene {
  constructor() {
    super("preload")
  }

  preload() {
    const menuKeys = [
      "menu-prev-1",
      "menu-prev-2",
      "madokita-title-mask",
      "madokita-title-overlay",
      "menu-theme",
      "texture-bg",
    ]

    menuKeys.forEach((key) => {
      const entry = AssetManifest[key]
      if (entry) {
        switch (entry.type) {
          case "image":
            this.load.image(entry.key, `./assets/${entry.path}`)
            break
          case "audio":
            this.load.audio(entry.key, `./assets/${entry.path}`)
            break
        }
      }
    })

    this.load.image("labyrinth", "./assets/images/labyrinth.png")
    this.load.image("school_storage", "./assets/images/school_storage.png")

    this.load.spritesheet(
      "charlotte_phase_1",
      "./assets/sprites/charlotte_phase_1.png",
      { frameWidth: 256, frameHeight: 256 },
    )

    this.load.spritesheet(
      "mobbutterfly",
      "./assets/sprites/enemies/mob1butterfly/mobbutterfly.png",
      { frameWidth: 256, frameHeight: 256 },
    )

    this.load.spritesheet(
      "sayaka_sword",
      "./assets/sprites/weapons/sayaka_sword.png",
      { frameWidth: 145, frameHeight: 145 },
    )

    const playerKeys = Object.keys(CharacterSpritesheets)
    for (const key of playerKeys) {
      const manifest = CharacterSpritesheets[key]
      this.load.spritesheet(
        manifest.spritesheet.key,
        `./assets/${manifest.spritesheet.path}`,
        { frameWidth: 256, frameHeight: 256 },
      )
    }
  }
  // scenes: ("game", "main-menu", "menu-intro")
  create() {
    this.scene.start("main-menu")
  }
}
