import Phaser from "phaser"
import { AssetResolver } from "../platform/shared/AssetResolver"

type SpritesheetManifest = {
  type: "spritesheet"
  key: string
  path: string
  frameWidth: number
  frameHeight: number
}

type ImageManifest = {
  type: "image"
  key: string
  path: string
}

type AudioManifest = {
  type: "audio"
  key: string
  path: string
}

type AssetManifestEntry = SpritesheetManifest | ImageManifest | AudioManifest

type CharacterManifest = {
  assetKey: string
  spritesheet: SpritesheetManifest
}

export const AssetManifest: Record<string, AssetManifestEntry> = {
  "menu-prev-1": {
    type: "image",
    key: "menu-prev-1",
    path: "menu/prevmenu1.png",
  },
  "menu-prev-2": {
    type: "image",
    key: "menu-prev-2",
    path: "menu/prevmenu2.jpg",
  },
  "madokita-title-mask": {
    type: "image",
    key: "madokita-title-mask",
    path: "menu/madokita-title.png",
  },
  "madokita-title-overlay": {
    type: "image",
    key: "madokita-title-overlay",
    path: "menu/madokita-title-top.png",
  },
  underline: {
    type: "image",
    key: "underline",
    path: "menu/underline.png",
  },
  "menu-theme": {
    type: "audio",
    key: "menu-theme",
    path: "menu/menutheme.ogg",
  },
  "texture-bg": {
    type: "image",
    key: "texture-bg",
    path: "menu/cosmic-effect.png",
  },
  labyrinth: {
    type: "image",
    key: "labyrinth",
    path: "images/labyrinth.png",
  },
  school_storage: {
    type: "image",
    key: "school_storage",
    path: "images/school_storage.png",
  },
  charlotte_phase_1: {
    type: "spritesheet",
    key: "charlotte_phase_1",
    path: "sprites/charlotte_phase_1.png",
    frameWidth: 256,
    frameHeight: 256,
  },
  mobbutterfly: {
    type: "spritesheet",
    key: "mobbutterfly",
    path: "sprites/enemies/mob1butterfly/mobbutterfly.png",
    frameWidth: 256,
    frameHeight: 256,
  },
  sayaka_sword: {
    type: "spritesheet",
    key: "sayaka_sword",
    path: "sprites/weapons/sayaka_sword.png",
    frameWidth: 145,
    frameHeight: 145,
  },
  testextension: {
    type: "spritesheet",
    key: "testextension",
    path: "sprites/enemies/testextension/testextension.png",
    frameWidth: 256,
    frameHeight: 256,
  },
  testextension_weapon: {
    type: "spritesheet",
    key: "testextension_weapon",
    path: "sprites/weapons/testextension_weapon.png",
    frameWidth: 145,
    frameHeight: 145,
  },
  testextension_chain: {
    type: "spritesheet",
    key: "testextension_chain",
    path: "sprites/extensions/testextension_chain.png",
    frameWidth: 64,
    frameHeight: 64,
  },
}

export const CharacterSpritesheets: Record<string, CharacterManifest> = {
  // ── Madoka ──
  madoka: {
    assetKey: "madoka",
    spritesheet: {
      type: "spritesheet",
      key: "madoka",
      path: "sprites/players/madoka_kaname/madoka_kaname.png",
      frameWidth: 256,
      frameHeight: 256,
    },
  },

  // ── Kyouko ──
  kyouko: {
    assetKey: "kyouko",
    spritesheet: {
      type: "spritesheet",
      key: "kyouko",
      path: "sprites/players/kyouko_sakura/kyouko_sakura.png",
      frameWidth: 256,
      frameHeight: 256,
    },
  },

  // ── Sayaka ──
  sayaka: {
    assetKey: "sayaka",
    spritesheet: {
      type: "spritesheet",
      key: "sayaka",
      path: "sprites/players/sayaka_miki/sayaka_miki.png",
      frameWidth: 256,
      frameHeight: 256,
    },
  },

  // ── Sayaka — Attack spritesheets ──
  sayaka_vertical_attack: {
    assetKey: "sayaka_vertical_attack",
    spritesheet: {
      type: "spritesheet",
      key: "sayaka_vertical_attack",
      path: "sprites/players/sayaka_miki/vertical_attack.png",
      frameWidth: 256,
      frameHeight: 256,
    },
  },
  sayaka_horizontal_attack: {
    assetKey: "sayaka_horizontal_attack",
    spritesheet: {
      type: "spritesheet",
      key: "sayaka_horizontal_attack",
      path: "sprites/players/sayaka_miki/horizontal_attack.png",
      frameWidth: 256,
      frameHeight: 256,
    },
  },
  sayaka_sword_throwing: {
    assetKey: "sayaka_sword_throwing",
    spritesheet: {
      type: "spritesheet",
      key: "sayaka_sword_throwing",
      path: "sprites/players/sayaka_miki/sword_throwing.png",
      frameWidth: 256,
      frameHeight: 256,
    },
  },
}

export class AssetLoader {
  private scene: Phaser.Scene
  private loaded = new Set<string>()

  constructor(scene: Phaser.Scene) {
    this.scene = scene
  }

  isLoaded(key: string): boolean {
    return this.loaded.has(key) || this.scene.textures.exists(key)
  }

  loadSingle(key: string): Promise<void> {
    if (this.isLoaded(key)) {
      return Promise.resolve()
    }

    const entry = AssetManifest[key]
    if (!entry) {
      return Promise.resolve()
    }

    return new Promise((resolve) => {
      this.loadEntry(entry, () => {
        this.loaded.add(key)
        resolve()
      })
    })
  }

  loadBatch(keys: string[]): Promise<void> {
    const unloaded = keys.filter((k) => !this.isLoaded(k))
    if (unloaded.length === 0) {
      return Promise.resolve()
    }

    return new Promise((resolve) => {
      this.scene.load.once("complete", () => {
        unloaded.forEach((k) => this.loaded.add(k))
        resolve()
      })

      unloaded.forEach((key) => {
        const entry = AssetManifest[key]
        if (entry) {
          this.loadEntry(entry)
        }
      })

      if (unloaded.length > 0) {
        this.scene.load.start()
      } else {
        resolve()
      }
    })
  }

  loadCharacterSpritesheet(characterKey: string): Promise<void> {
    if (this.isLoaded(characterKey)) {
      return Promise.resolve()
    }

    const manifest = CharacterSpritesheets[characterKey]
    if (!manifest) {
      return Promise.resolve()
    }

    return new Promise((resolve) => {
      this.scene.load.once("complete", () => {
        this.loaded.add(characterKey)
        resolve()
      })

      const s = manifest.spritesheet
      this.scene.load.spritesheet(s.key, AssetResolver.resolve(s.path), {
        frameWidth: s.frameWidth,
        frameHeight: s.frameHeight,
      })

      this.scene.load.start()
    })
  }

  loadAll(entries: AssetManifestEntry[]): Promise<void> {
    const unloaded = entries.filter((e) => !this.isLoaded(e.key))

    return new Promise((resolve) => {
      this.scene.load.once("complete", () => {
        unloaded.forEach((e) => this.loaded.add(e.key))
        resolve()
      })

      unloaded.forEach((entry) => this.loadEntry(entry))

      if (unloaded.length > 0) {
        this.scene.load.start()
      } else {
        resolve()
      }
    })
  }

  loadCoreAssets(): Promise<void> {
    const coreKeys = [
      "labyrinth",
      "school_storage",
      "charlotte_phase_1",
      "mobbutterfly",
    ]

    const playerKeys = Object.keys(CharacterSpritesheets)
    for (const key of playerKeys) {
      const manifest = CharacterSpritesheets[key]
      const unloaded = [manifest.spritesheet].filter(
        (e) => !this.isLoaded(e.key),
      )
      if (unloaded.length > 0) {
        unloaded.forEach((e) => this.loadEntry(e))
      }
    }

    return new Promise((resolve) => {
      this.scene.load.once("complete", () => {
        coreKeys.forEach((k) => this.loaded.add(k))
        playerKeys.forEach((k) => this.loaded.add(k))
        resolve()
      })

      coreKeys.forEach((key) => {
        const entry = AssetManifest[key]
        if (entry) this.loadEntry(entry)
      })

      this.scene.load.start()
    })
  }

  private loadEntry(entry: AssetManifestEntry, onLoad?: () => void): void {
    if (onLoad) {
      this.scene.load.once(`filecomplete-${entry.key}`, onLoad)
    }

    switch (entry.type) {
      case "image":
        this.scene.load.image(entry.key, AssetResolver.resolve(entry.path))
        break
      case "spritesheet":
        this.scene.load.spritesheet(
          entry.key,
          AssetResolver.resolve(entry.path),
          { frameWidth: entry.frameWidth, frameHeight: entry.frameHeight },
        )
        break
      case "audio":
        this.scene.load.audio(entry.key, AssetResolver.resolve(entry.path))
        break
    }
  }
}
