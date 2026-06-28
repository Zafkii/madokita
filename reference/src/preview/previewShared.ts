export const WRAPPER_CSS = "padding:12px 0;border-top:1px solid #333;margin-top:auto;"
export const HEADER_HTML = `<div style="font-weight:600;font-size:13px;margin-bottom:6px;color:var(--label-color,#8cf);">Preview</div>`
const PLAY_BTN_CLASS = "preview-play-btn"
export const PLAY_BTN_HTML = `<button class="${PLAY_BTN_CLASS}" style="padding:4px 14px;background:#2a5;color:var(--btn-color,#fff);border:none;border-radius:3px;cursor:pointer;font-size:13px;font-weight:600;">▶ Play</button>`
export const PLAY_BTN_TEXT = "▶ Play"
export const STOP_BTN_TEXT = "⏹ Stop"
export const STOP_BTN_BG = "#c33"
export const LOOP_HTML = `<label style="font-size:var(--font-md);color:#aaa;display:flex;align-items:center;gap:var(--gap-sm);"><input type="checkbox" id="preview-loop" checked /> Loop</label>`
export const SPEED_HTML = `<label style="font-size:var(--font-md);color:#aaa;display:flex;align-items:center;gap:var(--gap-sm);">Speed:<input id="preview-speed" type="range" min="0.1" max="3" step="0.1" value="1" style="width:60px;" /><span id="preview-speed-val">1x</span></label>`

export function createPreviewContainer(container: HTMLElement): HTMLElement {
  const el = document.createElement("div")
  el.style.cssText = WRAPPER_CSS
  container.appendChild(el)
  return el
}
