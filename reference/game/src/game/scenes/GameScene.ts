import Phaser from "phaser"
import { ControlsManager } from "../controls/ControlsManager"
import { Player } from "../entities/playable/player/Player"
import { CharlottePhase1 } from "../entities/enemies/charlottePhase1"
import { MobButterfly } from "../entities/enemies/mobButterfly"
import { AISystem } from "../systems/AISystem"
import { CombatTargetSystem } from "../entities/base/combat/CombatTargetSystem"
import { CollisionSystem } from "../systems/CollisionSystem"
import { CombatSystem } from "../systems/CombatSystem"
import { EffectManager } from "../effects/EffectManager"
import { DepthLayers } from "../config/DepthLayers"
import { SaveManager } from "../save/SaveManager"
import { BASE_WIDTH } from "../config/gameConfig"

const STAGE_ID = "stage1"

const DEFAULT_POSITIONS: Record<string, { x: number; y: number }> = {
  sayaka: { x: 2000, y: 400 },
  kyouko: { x: 800, y: 400 },
}

export class GameScene extends Phaser.Scene {
  private controls!: ControlsManager
  private aiSystem!: AISystem
  private players: Player[] = []
  private witches: CharlottePhase1[] = []
  private butterflies: MobButterfly[] = []
  private currentPlayer!: Player
  private platforms!: Phaser.Physics.Arcade.StaticGroup
  private combatTargetSystem!: CombatTargetSystem
  private collisionSystem!: CollisionSystem
  private combatSystem!: CombatSystem
  private effectManager!: EffectManager
  private bg!: Phaser.GameObjects.Image

  constructor() {
    super("game")
  }
  create(): void {
    // WORLD
    this.physics.world.setBounds(0, 0, 4000, 720)
    this.cameras.main.setBounds(0, 0, 4000, 720)

    // SYSTEMS
    this.combatTargetSystem = new CombatTargetSystem()
    this.collisionSystem = new CollisionSystem()
    this.combatSystem = new CombatSystem()
    this.effectManager = new EffectManager(this)

    // BACKGROUND
    this.bg = this.add.image(
      this.cameras.main.width / 2,
      this.cameras.main.height / 2,
      "labyrinth",
    )
    this.bg.setOrigin(0.5, 0.5)
    this.bg.setScrollFactor(0)
    this.bg.setDepth(DepthLayers.SKY)

    // PLATFORM
    this.platforms = this.physics.add.staticGroup()
    const floor = this.platforms.create(2000, 810, "school_storage")
    floor.setDisplaySize(4000, 200)
    floor.refreshBody()
    floor.setDepth(DepthLayers.PLATFORMS)

    // RESTORE POSITIONS FROM SAVE
    const savedPos = SaveManager.getData().currentLocation.players
    const sayakaPos = savedPos.sayaka ?? DEFAULT_POSITIONS.sayaka
    const kyoukoPos = savedPos.kyouko ?? DEFAULT_POSITIONS.kyouko

    // PLAYERS
    const sayaka = new Player(
      this,
      sayakaPos.x,
      sayakaPos.y,
      "sayaka",
      this.combatTargetSystem,
      this.collisionSystem,
      this.effectManager,
    )
    const kyouko = new Player(
      this,
      kyoukoPos.x,
      kyoukoPos.y,
      "kyouko",
      this.combatTargetSystem,
      this.collisionSystem,
      this.effectManager,
    )
    this.players.push(sayaka, kyouko)
    this.players.forEach((player) => {
      this.combatTargetSystem.register(player.combatActor)
      this.combatSystem.register(player.combatActor)
    })

    // CONTROL
    sayaka.setControlled(true)
    sayaka.setMainPlayer(true)
    this.currentPlayer = sayaka

    // ENEMIES
    const witch = new CharlottePhase1(this, 600, 300)
    witch.setDepth(DepthLayers.ENEMIES)
    this.witches.push(witch)

    const butterfly = new MobButterfly(
      this,
      1400,
      580,
      this.combatTargetSystem,
      this.collisionSystem,
    )
    butterfly.setDepth(DepthLayers.ENEMIES)
    this.butterflies.push(butterfly)
    this.combatTargetSystem.register(butterfly.combatActor)
    this.combatSystem.register(butterfly.combatActor)

    // COLLISIONS
    this.players.forEach((player) => {
      this.physics.add.collider(player, this.platforms)
    })

    // CAMERA
    this.cameras.main.setZoom(this.scale.width / BASE_WIDTH)
    this.cameras.main.startFollow(this.currentPlayer)
    this.cameras.main.setLerp(0.08, 0.08)

    // SYSTEMS
    this.controls = new ControlsManager(this)
    this.aiSystem = new AISystem(this.players, () => this.currentPlayer)

    // AUTOSAVE every 5s
    this.time.addEvent({
      delay: 5000,
      callback: () => this.savePositions(),
      loop: true,
    })
  }

  private savePositions(): void {
    const players: Record<string, { x: number; y: number }> = {}
    for (const player of this.players) {
      players[player.characterKey] = {
        x: Math.round(player.x),
        y: Math.round(player.y),
      }
    }
    SaveManager.update({
      currentLocation: {
        stageId: STAGE_ID,
        players,
      },
      lastSavedAt: new Date().toISOString(),
    })

    const saveText = this.add
      .text(this.cameras.main.width / 2, 16, "Saving...", {
        fontSize: "14px",
        color: "#ffffff",
      })
      .setOrigin(0.5, 0)
      .setAlpha(0.7)
      .setDepth(999)

    this.tweens.add({
      targets: saveText,
      alpha: 0,
      delay: 800,
      duration: 200,
      onComplete: () => saveText.destroy(),
    })
  }

  update(_time: number, delta: number): void {
    const offsetX = (this.cameras.main.scrollX - 2000) * 0.08
    const offsetY = (this.cameras.main.scrollY - 360) * 0.08
    this.bg.setPosition(
      this.cameras.main.width / 2 - offsetX,
      this.cameras.main.height / 2 - offsetY,
    )

    this.controls.update(this.currentPlayer)
    this.aiSystem.update()

    this.players.forEach((player) => {
      player.update()
    })

    this.witches.forEach((witch) => {
      witch.update()
    })

    this.butterflies.forEach((b) => {
      b.update()
    })

    this.collisionSystem.update()
    this.combatSystem.update(delta)
    this.effectManager.update(delta)
  }
}
