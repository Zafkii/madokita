import { App } from "./App"

const appEl = document.getElementById("app")
if (!appEl) throw new Error("Root element #app not found")

const app = new App(appEl)

;(window as any).__editorApp = app
