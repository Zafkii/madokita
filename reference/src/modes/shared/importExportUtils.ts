export function normalizeExportPath(raw: string): string {
  let p = raw.replace(/\\/g, "/")
  const publicIdx = p.toLowerCase().indexOf("/public/")
  if (publicIdx >= 0) {
    p = p.slice(publicIdx + 8)
  }
  if (p && !p.startsWith("./") && !p.startsWith("/")) {
    p = "./" + p
  }
  return p
}

export function pathToUrl(raw: string): string | null {
  if (!raw) return null
  if (raw.startsWith("http://") || raw.startsWith("https://")) return raw
  let p = raw.replace(/\\/g, "/").replace(/^\.\//, "")
  const publicIdx = p.toLowerCase().indexOf("/public/")
  if (publicIdx >= 0) {
    p = p.slice(publicIdx + 8)
  }
  p = p.replace(/^[a-zA-Z]:\//, "")
  if (p && !p.startsWith("/")) {
    p = "/" + p
  }
  return p
}

export function downloadFile(filename: string, content: string): void {
  const blob = new Blob([content], { type: "text/typescript" })
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function extractFilename(path: string): string {
  return path.replace(/\\/g, "/").split("/").pop() || path
}

export function parseLoadImagePaths(text: string): Record<string, string> {
  const result: Record<string, string> = {}
  const regex =
    /(?:\/\/\s*)?this\.load\.image\(\s*"([^"]+)"[\s\S]*?"([^"]+)"[\s\S]*?\)/g
  let match: RegExpExecArray | null
  while ((match = regex.exec(text)) !== null) {
    const key = match[1]
    const path = match[2]
    result[key] = path
  }
  if (Object.keys(result).length > 0) {
    console.log("[parseLoadImagePaths] found paths:", result)
  } else {
    console.warn("[parseLoadImagePaths] no load.image() calls found in text")
  }
  return result
}
