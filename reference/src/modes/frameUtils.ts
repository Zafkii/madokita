export function extractFrame(
  source: HTMLImageElement,
  frameW: number,
  frameH: number,
  frameIndex: number,
): HTMLCanvasElement {
  const cols = Math.floor(source.width / frameW)
  const row = Math.floor(frameIndex / cols)
  const col = frameIndex % cols
  const c = document.createElement("canvas")
  c.width = frameW
  c.height = frameH
  const ctx = c.getContext("2d")!
  ctx.drawImage(
    source,
    col * frameW,
    row * frameH,
    frameW,
    frameH,
    0,
    0,
    frameW,
    frameH,
  )
  return c
}
