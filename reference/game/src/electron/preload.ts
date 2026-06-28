const { contextBridge, ipcRenderer } = require("electron")

contextBridge.exposeInMainWorld("electronAPI", {
  setResolution: (width: number, height: number) =>
    ipcRenderer.invoke("set-resolution", width, height),

  toggleFullscreen: () => ipcRenderer.invoke("toggle-fullscreen"),

  onResolutionChanged: (callback: () => void) => {
    ipcRenderer.on("resolution-changed", () => {
      callback()
    })
  },

  onFullscreenChanged: (callback: (fs: boolean) => void) => {
    ipcRenderer.on("fullscreen-changed", (_event: any, fs: boolean) => {
      callback(fs)
    })
  },

  // =====================================
  // SETTINGS
  // =====================================

  saveSetting: (key: string, value: string) =>
    ipcRenderer.invoke("save-setting", key, value),

  getSetting: (key: string) => ipcRenderer.invoke("get-setting", key),

  getAllSettings: () => ipcRenderer.invoke("get-all-settings"),

  // =====================================
  // SAVE
  // =====================================

  saveGame: (data: string) => ipcRenderer.invoke("save-game", data),

  loadSave: () => ipcRenderer.invoke("load-save"),
})
