import Phaser from "phaser"
import { MainMenuView, type FocusableEntry } from "../MainMenuView"
import { MainMenuState } from "../MainMenuState"
import { SubMenuLayout } from "../SubMenuLayout"
import { GameSettingsManager } from "../../../../game/settings/GameSettingsManager"
import { INPUT_ACTIONS } from "../../../../game/input/InputAction"
import type { InputAction } from "../../../../game/input/InputAction"

type SettingsCallbacks = {
  onBack: () => void
}

const NAV_ITEMS = ["PROFILE", "DISPLAY", "CONTROLLERS", "VOLUME"] as const
type Category = (typeof NAV_ITEMS)[number]

const RESOLUTION_OPTIONS = [
  { label: "1920x1080", width: 1920, height: 1080 },
  { label: "1600x900", width: 1600, height: 900 },
  { label: "1280x720", width: 1280, height: 720 },
  { label: "854x480", width: 854, height: 480 },
  { label: "427x240", width: 427, height: 240 },
]

const FPS_OPTIONS = [120, 60, 30]

const GAME_ACTION_ENTRIES: { action: InputAction; label: string }[] = [
  { action: INPUT_ACTIONS.MOVE_LEFT, label: "MOVE LEFT" },
  { action: INPUT_ACTIONS.MOVE_RIGHT, label: "MOVE RIGHT" },
  { action: INPUT_ACTIONS.JUMP, label: "JUMP" },
  { action: INPUT_ACTIONS.DODGE, label: "DODGE" },
  { action: INPUT_ACTIONS.ATTACK, label: "ATTACK" },
  { action: INPUT_ACTIONS.SKILL, label: "SKILL" },
  { action: INPUT_ACTIONS.TRANSFORM, label: "TRANSFORM" },
]

const UI_ACTION_ENTRIES: { action: InputAction; label: string }[] = [
  { action: INPUT_ACTIONS.MENU_UP, label: "MENU UP" },
  { action: INPUT_ACTIONS.MENU_DOWN, label: "MENU DOWN" },
  { action: INPUT_ACTIONS.MENU_CONFIRM, label: "CONFIRM" },
  { action: INPUT_ACTIONS.MENU_BACK, label: "BACK" },
]

function normalizeKey(eventKey: string): string {
  if (eventKey === " ") return "SPACE"
  const upper = eventKey.toUpperCase()
  const map: Record<string, string> = {
    ARROWLEFT: "LEFT",
    ARROWRIGHT: "RIGHT",
    ARROWUP: "UP",
    ARROWDOWN: "DOWN",
    ESCAPE: "ESC",
  }
  return map[upper] ?? upper
}

export class SettingsView extends MainMenuView {
  private callbacks: SettingsCallbacks
  private state: MainMenuState
  private selectedCategory: Category = "PROFILE"
  private displayContent: "resolution" | "fps" | "mode" = "resolution"
  private controllersTab: "ui" | "game" = "ui"

  private allObjects: Phaser.GameObjects.GameObject[] = []
  private contentObjects: Phaser.GameObjects.GameObject[] = []
  private navTexts: Phaser.GameObjects.Text[] = []
  private scrollItems: { text: Phaser.GameObjects.Text; baseY: number }[] = []
  private scrollY = 0
  private maxScroll = 0
  private wheelHandler:
    | ((
        pointer: Phaser.Input.Pointer,
        gameObjects: Phaser.GameObjects.GameObject[],
        dx: number,
        dy: number,
      ) => void)
    | null = null
  private allFocusItems: FocusableEntry[] = []
  private contentFocusCount = 0

  constructor(
    scene: Phaser.Scene,
    state: MainMenuState,
    callbacks: SettingsCallbacks,
  ) {
    super(scene)
    this.state = state
    this.callbacks = callbacks
  }

  create(): void {
    this.clearAllFocus()
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const cx = w / 2

    const title = this.addPersistent(
      this.scene.add.text(
        cx,
        SubMenuLayout.titleY(h),
        "SETTINGS",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.titleFont(h),
          color: SubMenuLayout.C_NORMAL,
        }),
      ),
    )
    title.setOrigin(0.5)

    this.renderNav()
    this.renderContent()
  }

  // ─── FOCUS ─────────────────────────────────────────────────────────────────

  focusNext(): void {
    if (this.allFocusItems.length === 0) return
    this._unfocusCurrent()
    this.focusIndex = (this.focusIndex + 1) % this.allFocusItems.length
    this._focusCurrent()
  }

  focusPrev(): void {
    if (this.allFocusItems.length === 0) return
    this._unfocusCurrent()
    this.focusIndex =
      (this.focusIndex - 1 + this.allFocusItems.length) % this.allFocusItems.length
    this._focusCurrent()
  }

  focusConfirm(): void {
    if (this.captureHandler) return
    if (this.focusIndex < 0 || this.focusIndex >= this.allFocusItems.length)
      return
    this.allFocusItems[this.focusIndex].callback()
  }

  focusBack(): void {
    if (this.captureHandler) return
    this.callbacks.onBack()
  }

  private registerPersistentFocus(
    text: Phaser.GameObjects.Text,
    callback: () => void,
  ): void {
    this.allFocusItems.push({ text, callback })
  }

  private registerContentFocus(
    text: Phaser.GameObjects.Text,
    callback: () => void,
  ): void {
    this.allFocusItems.push({ text, callback })
    this.contentFocusCount++
  }

  private clearContentFocus(): void {
    if (this.contentFocusCount > 0) {
      this.allFocusItems.splice(
        this.allFocusItems.length - this.contentFocusCount,
      )
      this.contentFocusCount = 0
    }
  }

  private clearAllFocus(): void {
    this.allFocusItems = []
    this.contentFocusCount = 0
    this.focusIndex = -1
  }

  private _unfocusCurrent(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.allFocusItems.length)
      return
    const item = this.allFocusItems[this.focusIndex]
    item.text.setColor(SubMenuLayout.C_NORMAL)
    item.text.setScale(1)
  }

  private _focusCurrent(): void {
    if (this.focusIndex < 0 || this.focusIndex >= this.allFocusItems.length)
      return
    const item = this.allFocusItems[this.focusIndex]
    item.text.setColor(SubMenuLayout.C_ACTIVE)
    item.text.setScale(1.05)
  }

  handleResize(_width: number, _height: number): void {
    const scrollRatio = this.maxScroll > 0 ? -this.scrollY / this.maxScroll : 0

    this.allObjects.forEach((obj) => obj.destroy())
    this.allObjects = []
    this.contentObjects = []
    this.navTexts = []
    this.scrollItems = []
    if (this.captureHandler) {
      window.removeEventListener("keydown", this.captureHandler)
      this.setCaptureHandler(null)
    }
    if (this.wheelHandler) {
      this.scene.input.off("wheel", this.wheelHandler)
      this.wheelHandler = null
    }
    this.scrollY = 0
    this.maxScroll = 0
    this.clearAllFocus()

    this.create()

    if (this.maxScroll > 0 && scrollRatio > 0) {
      this.scrollY = -Math.round(scrollRatio * this.maxScroll)
      this.scrollItems.forEach((item) => {
        item.text.y = item.baseY + this.scrollY
      })
    }
  }

  destroy(): void {
    this.allObjects.forEach((obj) => obj.destroy())
    this.allObjects = []
    this.contentObjects = []
    this.navTexts = []
    this.scrollItems = []
    if (this.captureHandler) {
      window.removeEventListener("keydown", this.captureHandler)
      this.setCaptureHandler(null)
    }
    if (this.wheelHandler) {
      this.scene.input.off("wheel", this.wheelHandler)
      this.wheelHandler = null
    }
    super.destroy()
  }

  private addPersistent<T extends Phaser.GameObjects.GameObject>(obj: T): T {
    this.allObjects.push(obj)
    return obj
  }

  private addContent<T extends Phaser.GameObjects.GameObject>(obj: T): T {
    this.allObjects.push(obj)
    this.contentObjects.push(obj)
    return obj
  }

  private clearContent(): void {
    this.contentObjects.forEach((obj) => {
      const idx = this.allObjects.indexOf(obj)
      if (idx >= 0) this.allObjects.splice(idx, 1)
      obj.destroy()
    })
    this.contentObjects = []
    this.scrollItems = []
    this.scrollY = 0
    this.maxScroll = 0
    if (this.wheelHandler) {
      this.scene.input.off("wheel", this.wheelHandler)
      this.wheelHandler = null
    }
    this.clearContentFocus()
  }

  private trackScrollItem(text: Phaser.GameObjects.Text, baseY: number): void {
    this.scrollItems.push({ text, baseY })
  }

  private makeInteractive(
    text: Phaser.GameObjects.Text,
    onSelect: () => void,
    persistent = false,
  ): void {
    text.setInteractive({ useHandCursor: true })
    text.on("pointerover", () => {
      text.setColor(SubMenuLayout.C_ACTIVE)
      text.setScale(1.05)
    })
    text.on("pointerout", () => {
      text.setColor(SubMenuLayout.C_NORMAL)
      text.setScale(1)
    })
    text.on("pointerdown", onSelect)
    if (persistent) {
      this.registerPersistentFocus(text, onSelect)
    } else {
      this.registerContentFocus(text, onSelect)
    }
  }

  private selectCategory(category: Category): void {
    this.selectedCategory = category
    this.updateNavHighlight()
    this.renderContent()
  }

  private updateNavHighlight(): void {
    this.navTexts.forEach((text, i) => {
      if (i === NAV_ITEMS.indexOf(this.selectedCategory)) {
        text.setColor(SubMenuLayout.C_ACTIVE)
      } else {
        text.setColor(SubMenuLayout.C_NORMAL)
      }
    })
  }

  private setupScroll(lastY: number): void {
    const h = this.scene.scale.height
    const lastBottom = lastY + 20
    this.maxScroll = Math.max(0, lastBottom - (h - 20))
    if (this.maxScroll > 0) {
      this.wheelHandler = (
        _pointer: Phaser.Input.Pointer,
        _gameObjects: Phaser.GameObjects.GameObject[],
        _dx: number,
        dy: number,
      ) => {
        this.scrollY = Phaser.Math.Clamp(this.scrollY - dy, -this.maxScroll, 0)
        this.scrollItems.forEach((item) => {
          item.text.y = item.baseY + this.scrollY
        })
      }
      this.scene.input.on("wheel", this.wheelHandler)
    }
  }

  // ─── NAV ─────────────────────────────────────────────────────────────────

  private renderNav(): void {
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const nx = SubMenuLayout.navX(w)
    const cy = SubMenuLayout.contentY(h)
    const sp = SubMenuLayout.spacing(h)

    NAV_ITEMS.forEach((item, i) => {
      const text = this.addPersistent(
        this.scene.add.text(
          nx,
          cy + i * sp,
          item,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.optionFont(h),
            color:
              item === this.selectedCategory
                ? SubMenuLayout.C_ACTIVE
                : SubMenuLayout.C_NORMAL,
          }),
        ),
      )
      text.setOrigin(0.5)
      text.setInteractive({ useHandCursor: true })
      text.on("pointerover", () => {
        if (item !== this.selectedCategory)
          text.setColor(SubMenuLayout.C_ACTIVE)
        text.setScale(1.05)
      })
      text.on("pointerout", () => {
        if (item !== this.selectedCategory)
          text.setColor(SubMenuLayout.C_NORMAL)
        text.setScale(1)
      })
      text.on("pointerdown", () => this.selectCategory(item))
      this.navTexts.push(text)
      this.registerPersistentFocus(text, () => this.selectCategory(item))
    })

    const backText = this.addPersistent(
      this.scene.add.text(
        nx,
        cy + (NAV_ITEMS.length + 1) * sp,
        "BACK",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.optionFont(h),
          color: SubMenuLayout.C_NORMAL,
        }),
      ),
    )
    backText.setOrigin(0.5)
    this.makeInteractive(backText, () => this.callbacks.onBack(), true)
  }

  // ─── CONTENT DISPATCHER ───────────────────────────────────────────────────

  private renderContent(): void {
    this.clearContent()
    switch (this.selectedCategory) {
      case "PROFILE":
        this.renderProfileContent()
        break
      case "DISPLAY":
        this.renderDisplayContent()
        break
      case "CONTROLLERS":
        this.renderControllersContent()
        break
      case "VOLUME":
        this.renderVolumeContent()
        break
    }
  }

  // ─── PROFILE ──────────────────────────────────────────────────────────────

  private renderProfileContent(): void {
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const cy = SubMenuLayout.contentY(h)
    const sp = SubMenuLayout.spacing(h)
    const fontSize = SubMenuLayout.optionFont(h)

    const dataItems = [
      `NAME: Player`,
      `LEVEL: 1`,
      `LOCATION: Mitakihara`,
      `COMPLETION: 0%`,
      `FAVORITE CHARACTER: Sayaka`,
    ]

    const spriteX = Math.round(w * 0.38)
    const dataX = Math.round(w * 0.65)

    if (this.scene.textures.exists("sayaka")) {
      const sprite = this.addContent(
        this.scene.add.image(spriteX, cy, "sayaka", 0),
      )
      sprite.setOrigin(0.5)
      const scale = Math.min(120 / 256, (h * 0.4) / 256)
      sprite.setScale(scale)
    } else {
      const rect = this.addContent(
        this.scene.add.rectangle(spriteX, cy, 120, 120, 0x888888),
      )
      rect.setOrigin(0.5)
    }

    dataItems.forEach((item, i) => {
      const y = cy + i * sp
      const text = this.addContent(
        this.scene.add.text(
          dataX,
          y,
          item,
          SubMenuLayout.style({
            fontSize,
            color: item.startsWith("---")
              ? SubMenuLayout.C_MUTED
              : SubMenuLayout.C_NORMAL,
          }),
        ),
      )
      text.setOrigin(0, 0.5)
      this.trackScrollItem(text, y)
    })

    const lastY = cy + dataItems.length * sp
    this.setupScroll(lastY)
  }

  // ─── DISPLAY ──────────────────────────────────────────────────────────────

  private renderDisplayContent(): void {
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const ix = SubMenuLayout.infoX(w)
    const ax = SubMenuLayout.actionX(w)
    const cy = SubMenuLayout.contentY(h)
    const sp = SubMenuLayout.spacing(h)
    const fontSize = SubMenuLayout.optionFont(h)

    const labels = [
      {
        key: "resolution" as const,
        text: `RESOLUTION: ${this.state.resolution}`,
      },
      { key: "fps" as const, text: `FPS LIMIT: ${this.state.fps}` },
      { key: "mode" as const, text: `MODE: ${this.state.displayMode}` },
    ]

    labels.forEach((item, i) => {
      const isActive = item.key === this.displayContent
      const isClickable = item.key !== "resolution"
      const text = this.addContent(
        this.scene.add.text(
          ix,
          cy + i * sp,
          item.text,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.infoFont(h),
            color: isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL,
          }),
        ),
      )
      text.setOrigin(0.5)
      if (isClickable) {
        text.setInteractive({ useHandCursor: true })
        text.on("pointerover", () => {
          text.setColor(SubMenuLayout.C_ACTIVE)
          text.setScale(1.05)
        })
        text.on("pointerout", () => {
          text.setColor(
            isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL,
          )
          text.setScale(1)
        })
        text.on("pointerdown", () => {
          this.displayContent = item.key
          this.renderContent()
        })
        this.registerContentFocus(text, () => {
          this.displayContent = item.key
          this.renderContent()
        })
      }
    })

    if (this.displayContent === "resolution") {
      RESOLUTION_OPTIONS.forEach((opt, i) => {
        const isActive = opt.label === this.state.resolution
        const y = cy + i * sp
        const text = this.addContent(
          this.scene.add.text(
            ax,
            y,
            opt.label,
            SubMenuLayout.style({
              fontSize,
              color: isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL,
            }),
          ),
        )
        text.setOrigin(0.5)
        if (isActive) {
          text.setAlpha(0.6)
        } else {
          text.setInteractive({ useHandCursor: true })
          text.on("pointerover", () => {
            text.setColor(SubMenuLayout.C_ACTIVE)
            text.setScale(1.05)
          })
          text.on("pointerout", () => {
            text.setColor(SubMenuLayout.C_NORMAL)
            text.setScale(1)
          })
          text.on("pointerdown", async () => {
            this.state.resolution = opt.label
            try {
              await window.gamePlatform.setResolution(opt.width, opt.height)
              this.scene.game.scale.resize(opt.width, opt.height)
              window.dispatchEvent(new Event("resize"))
              await GameSettingsManager.setResolution(opt.width, opt.height)
            } catch (e) {
              console.error("Failed to set resolution:", e)
            }
            this.handleResize(this.scene.scale.width, this.scene.scale.height)
          })
          this.registerContentFocus(text, async () => {
            this.state.resolution = opt.label
            try {
              await window.gamePlatform.setResolution(opt.width, opt.height)
              this.scene.game.scale.resize(opt.width, opt.height)
              window.dispatchEvent(new Event("resize"))
              await GameSettingsManager.setResolution(opt.width, opt.height)
            } catch (e) {
              console.error("Failed to set resolution:", e)
            }
            this.handleResize(this.scene.scale.width, this.scene.scale.height)
          })
        }
      })
    } else if (this.displayContent === "fps") {
      FPS_OPTIONS.forEach((fps, i) => {
        const isActive = fps === this.state.fps
        const y = cy + i * sp
        const text = this.addContent(
          this.scene.add.text(
            ax,
            y,
            `${fps} FPS`,
            SubMenuLayout.style({
              fontSize,
              color: isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL,
            }),
          ),
        )
        text.setOrigin(0.5)
        if (isActive) {
          text.setAlpha(0.6)
        } else {
          text.setInteractive({ useHandCursor: true })
          text.on("pointerover", () => {
            text.setColor(SubMenuLayout.C_ACTIVE)
            text.setScale(1.05)
          })
          text.on("pointerout", () => {
            text.setColor(SubMenuLayout.C_NORMAL)
            text.setScale(1)
          })
          text.on("pointerdown", async () => {
            this.state.fps = fps
            this.scene.game.loop.targetFps = fps
            try {
              await GameSettingsManager.setFpsLimit(fps)
            } catch (e) {
              console.error("Failed to set FPS limit:", e)
            }
            this.displayContent = "resolution"
            this.renderContent()
          })
          this.registerContentFocus(text, async () => {
            this.state.fps = fps
            this.scene.game.loop.targetFps = fps
            try {
              await GameSettingsManager.setFpsLimit(fps)
            } catch (e) {
              console.error("Failed to set FPS limit:", e)
            }
            this.displayContent = "resolution"
            this.renderContent()
          })
        }
      })
    } else {
      const modes = [
        { label: "WINDOWED", value: false },
        { label: "FULLSCREEN", value: true },
      ]
      modes.forEach((mode, i) => {
        const isActive =
          mode.value === (this.state.displayMode === "Fullscreen")
        const y = cy + i * sp
        const text = this.addContent(
          this.scene.add.text(
            ax,
            y,
            mode.label,
            SubMenuLayout.style({
              fontSize,
              color: isActive ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL,
            }),
          ),
        )
        text.setOrigin(0.5)
        if (isActive) {
          text.setAlpha(0.6)
        } else {
          text.setInteractive({ useHandCursor: true })
          text.on("pointerover", () => {
            text.setColor(SubMenuLayout.C_ACTIVE)
            text.setScale(1.05)
          })
          text.on("pointerout", () => {
            text.setColor(SubMenuLayout.C_NORMAL)
            text.setScale(1)
          })
          text.on("pointerdown", async () => {
            const current = GameSettingsManager.getData().fullscreen
            if (current !== mode.value) {
              this.state.displayMode = mode.value ? "Fullscreen" : "Windowed"
              try {
                await GameSettingsManager.setFullscreen(mode.value)
                await window.gamePlatform.toggleFullscreen()
              } catch (e) {
                console.error("Failed to toggle display mode:", e)
              }
            }
            this.displayContent = "resolution"
            this.renderContent()
          })
          this.registerContentFocus(text, async () => {
            const current = GameSettingsManager.getData().fullscreen
            if (current !== mode.value) {
              this.state.displayMode = mode.value ? "Fullscreen" : "Windowed"
              try {
                await GameSettingsManager.setFullscreen(mode.value)
                await window.gamePlatform.toggleFullscreen()
              } catch (e) {
                console.error("Failed to toggle display mode:", e)
              }
            }
            this.displayContent = "resolution"
            this.renderContent()
          })
        }
      })
    }
  }

  // ─── CONTROLLERS ──────────────────────────────────────────────────────────

  private renderControllersContent(): void {
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const nx = SubMenuLayout.infoX(w)
    const cx = SubMenuLayout.actionX(w)
    const cy = SubMenuLayout.contentY(h)
    const sp = SubMenuLayout.spacing(h)
    const fontSize = SubMenuLayout.optionFont(h)

    const TABS = [
      { label: "UI", key: "ui" as const },
      { label: "GAME", key: "game" as const },
    ]

    TABS.forEach((tab, i) => {
      const isSelected = tab.key === this.controllersTab
      const text = this.addContent(
        this.scene.add.text(
          nx,
          cy + i * sp,
          tab.label,
          SubMenuLayout.style({
            fontSize,
            color: isSelected ? SubMenuLayout.C_ACTIVE : SubMenuLayout.C_NORMAL,
          }),
        ),
      )
      text.setOrigin(0.5)
      text.setInteractive({ useHandCursor: true })
      text.on("pointerover", () => {
        if (!isSelected) text.setColor(SubMenuLayout.C_ACTIVE)
        text.setScale(1.05)
      })
      text.on("pointerout", () => {
        if (!isSelected) text.setColor(SubMenuLayout.C_NORMAL)
        text.setScale(1)
      })
      text.on("pointerdown", () => {
        this.controllersTab = tab.key
        this.renderContent()
      })
      this.registerContentFocus(text, () => {
        this.controllersTab = tab.key
        this.renderContent()
      })
    })

    if (this.controllersTab === "ui") {
      this.renderKeyBindingContent(cx, cy, sp, h, UI_ACTION_ENTRIES, false)
    } else {
      this.renderKeyBindingContent(cx, cy, sp, h, GAME_ACTION_ENTRIES, false)

      const sensRow = GAME_ACTION_ENTRIES.length
      const sensY = cy + sensRow * sp

      const sensText = this.addContent(
        this.scene.add.text(
          cx,
          sensY,
          `SENSITIVITY: ${this.state.sensitivity}`,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.infoFont(h),
            color: SubMenuLayout.C_ACTIVE,
          }),
        ),
      )
      sensText.setOrigin(0.5)
      this.trackScrollItem(sensText, sensY)

      const valStyle = SubMenuLayout.style({
        fontSize,
        color: SubMenuLayout.C_ACTIVE,
      })
      const minusBtn = this.addContent(
        this.scene.add.text(cx - 60, sensY + sp, "[-]", valStyle),
      ) as Phaser.GameObjects.Text
      minusBtn.setOrigin(0.5)
      minusBtn.setInteractive({ useHandCursor: true })
      this.trackScrollItem(minusBtn, sensY + sp)

      const sensValue = this.addContent(
        this.scene.add.text(
          cx,
          sensY + sp,
          `${this.state.sensitivity}`,
          valStyle,
        ),
      ) as Phaser.GameObjects.Text
      sensValue.setOrigin(0.5)
      this.trackScrollItem(sensValue, sensY + sp)

      const plusBtn = this.addContent(
        this.scene.add.text(cx + 60, sensY + sp, "[+]", valStyle),
      ) as Phaser.GameObjects.Text
      plusBtn.setOrigin(0.5)
      plusBtn.setInteractive({ useHandCursor: true })
      this.trackScrollItem(plusBtn, sensY + sp)

      minusBtn.on("pointerdown", () => {
        const newVal = Math.max(5, this.state.sensitivity - 5)
        this.state.sensitivity = newVal
        GameSettingsManager.setSensitivity(newVal)
        sensText.setText(`SENSITIVITY: ${newVal}`)
        sensValue.setText(`${newVal}`)
      })
      this.registerContentFocus(minusBtn, () => {
        const newVal = Math.max(5, this.state.sensitivity - 5)
        this.state.sensitivity = newVal
        GameSettingsManager.setSensitivity(newVal)
        sensText.setText(`SENSITIVITY: ${newVal}`)
        sensValue.setText(`${newVal}`)
      })

      plusBtn.on("pointerdown", () => {
        const newVal = Math.min(100, this.state.sensitivity + 5)
        this.state.sensitivity = newVal
        GameSettingsManager.setSensitivity(newVal)
        sensText.setText(`SENSITIVITY: ${newVal}`)
        sensValue.setText(`${newVal}`)
      })
      this.registerContentFocus(plusBtn, () => {
        const newVal = Math.min(100, this.state.sensitivity + 5)
        this.state.sensitivity = newVal
        GameSettingsManager.setSensitivity(newVal)
        sensText.setText(`SENSITIVITY: ${newVal}`)
        sensValue.setText(`${newVal}`)
      })

      const lastY = sensY + 2 * sp
      this.setupScroll(lastY)
    }
  }

  private renderKeyBindingContent(
    x: number,
    cy: number,
    sp: number,
    h: number,
    entries: { action: InputAction; label: string }[],
    includeNavSensitivity: boolean,
  ): void {
    const infoTexts: Phaser.GameObjects.Text[] = []

    entries.forEach((entry, i) => {
      const y = cy + i * sp
      const key = this.state.keyBindings[entry.action] ?? "?"

      const text = this.addContent(
        this.scene.add.text(
          x,
          y,
          `${entry.label}: ${key}`,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.infoFont(h),
            color: SubMenuLayout.C_NORMAL,
          }),
        ),
      ) as Phaser.GameObjects.Text
      text.setOrigin(0.5)
      text.setInteractive({ useHandCursor: true })
      infoTexts.push(text)
      this.trackScrollItem(text, y)
      const idx = i

      text.on("pointerover", () => {
        if (this.captureHandler) return
        text.setColor(SubMenuLayout.C_ACTIVE)
        text.setScale(1.05)
      })
      text.on("pointerout", () => {
        if (this.captureHandler) return
        text.setColor(SubMenuLayout.C_NORMAL)
        text.setScale(1)
      })
      const onStartCapture = () => {
        if (this.captureHandler) return

        const handler = (event: KeyboardEvent) => {
          event.preventDefault()
          const key = normalizeKey(event.key)
          GameSettingsManager.setKeyBinding(entry.action, key)
          this.state.keyBindings[entry.action] = key
          infoTexts[idx].setText(`${entry.label}: ${key}`)
          text.setColor(SubMenuLayout.C_NORMAL)
          text.setScale(1)
          window.removeEventListener("keydown", handler)
          this.setCaptureHandler(null)
        }

        text.setText(`${entry.label}: ...`)
        text.setColor(SubMenuLayout.C_ACTIVE)
        text.setScale(1.05)
        this.setCaptureHandler(handler)
        window.addEventListener("keydown", handler)
      }

      text.on("pointerdown", onStartCapture)
      this.registerContentFocus(text, onStartCapture)
    })

    let lastY = cy + entries.length * sp

    if (includeNavSensitivity) {
      const sensY = cy + entries.length * sp
      const sensInfo = this.addContent(
        this.scene.add.text(
          x,
          sensY,
          `SENSITIVITY: ${this.state.sensitivity}`,
          SubMenuLayout.style({
            fontSize: SubMenuLayout.infoFont(h),
            color: SubMenuLayout.C_ACTIVE,
          }),
        ),
      )
      sensInfo.setOrigin(0.5)
      this.trackScrollItem(sensInfo, sensY)

      lastY = sensY + sp
    }

    this.setupScroll(lastY)
  }

  // ─── VOLUME ───────────────────────────────────────────────────────────────

  private renderVolumeContent(): void {
    const w = this.scene.scale.width
    const h = this.scene.scale.height
    const ix = SubMenuLayout.infoX(w)
    const cy = SubMenuLayout.contentY(h)

    const text = this.addContent(
      this.scene.add.text(
        ix,
        cy,
        "Volume controls coming soon",
        SubMenuLayout.style({
          fontSize: SubMenuLayout.infoFont(h),
          color: SubMenuLayout.C_MUTED,
        }),
      ),
    )
    text.setOrigin(0.5)
  }
}
