import { app, BrowserWindow, ipcMain } from "electron"

import path from "node:path"

import { initDatabase } from "../game/database/db"

import { SettingsRepository } from "../game/database/repositories/SettingsRepository"

import { GameSaveRepository } from "../game/save/repositories/GameSaveRepository"

const isDev = !!process.env.VITE_DEV_SERVER_URL

let mainWindow: BrowserWindow | null = null

async function bootstrap(): Promise<void> {
  await initDatabase()

  await SettingsRepository.initialize()

  await GameSaveRepository.initialize()
}

async function createWindow(): Promise<void> {
  const preloadPath = path.join(__dirname, "preload.cjs")

  mainWindow = new BrowserWindow({
    show: false,

    width: 1280,

    height: 720,

    minWidth: 480,

    minHeight: 270,

    backgroundColor: "#181818",

    autoHideMenuBar: true,
    icon: path.join(__dirname, "../../public/assets/madokita.ico"),

    webPreferences: {
      preload: preloadPath,

      contextIsolation: true,

      nodeIntegration: false,

      backgroundThrottling: false,
    },
  })

  mainWindow.setAspectRatio(16 / 9)

  const savedSettings = await SettingsRepository.getAllSettings()
  const width = Math.max(854, Number(savedSettings.resolution_width) || 1280)
  const height = Math.max(480, Number(savedSettings.resolution_height) || 720)
  mainWindow.setContentSize(width, height)
  console.log("[Main process] Window content size:", width, height)

  if (isDev) {
    mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL!)
  } else {
    const indexPath = path.join(app.getAppPath(), "dist/index.html")

    mainWindow.loadFile(indexPath)
  }

  mainWindow.once("ready-to-show", () => {
    mainWindow!.show()
  })
}

// =====================================
// WINDOW
// =====================================

ipcMain.handle("set-resolution", (_event, width: number, height: number) => {
  if (!mainWindow) {
    return
  }

  mainWindow.setAspectRatio(width / height)

  if (!mainWindow.isFullScreen()) {
    mainWindow.setContentSize(width, height)
  }

  mainWindow.webContents.send("resolution-changed")
})

ipcMain.handle("toggle-fullscreen", () => {
  if (!mainWindow) {
    return
  }

  const isFullscreen = mainWindow.isFullScreen()

  mainWindow.setFullScreen(!isFullscreen)

  mainWindow.webContents.send("fullscreen-changed", !isFullscreen)
})

// =====================================
// SETTINGS
// =====================================

ipcMain.handle("save-setting", async (_event, key: string, value: string) => {
  await SettingsRepository.setSetting(key, value)
})

ipcMain.handle("get-setting", async (_event, key: string) => {
  return await SettingsRepository.getSetting(key)
})

ipcMain.handle("get-all-settings", async () => {
  return await SettingsRepository.getAllSettings()
})

// =====================================
// SAVE SYSTEM
// =====================================

ipcMain.handle("save-game", async (_event, data: string) => {
  const parsed = JSON.parse(data)

  await GameSaveRepository.save(data, parsed.version)
})

ipcMain.handle("load-save", async () => {
  return await GameSaveRepository.load()
})

// =====================================
// APP
// =====================================

app.whenReady().then(async () => {
  await bootstrap()

  await createWindow()

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit()
  }
})
