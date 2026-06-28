import sqlite3 from "sqlite3"
import { open, type Database } from "sqlite"
import fs from "node:fs"
import path from "node:path"
import { app } from "electron"
let db: Database | null = null
export async function initDatabase(): Promise<Database> {
  if (db) {
    return db
  }
  // SAVE DIRECTORY
  const saveDir = path.join(app.getPath("userData"), "save")

  fs.mkdirSync(saveDir, {
    recursive: true,
  })
  // SAVE FILE
  const dbPath = path.join(saveDir, "madokita.sav")
  console.log("💾 Save File:", dbPath)
  // OPEN DATABASE
  db = await open({
    filename: dbPath,
    driver: sqlite3.Database,
  })

  return db
}

export function getDatabase(): Database {
  if (!db) {
    throw new Error("Database not initialized")
  }
  return db
}
