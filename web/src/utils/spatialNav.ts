// 电视遥控器方向键导航（预留）：按方向键时把焦点移动到该方向上最近的可聚焦元素。
const FOCUSABLE = 'a[href], button:not([disabled]), input, [tabindex="0"]'

type Dir = 'ArrowUp' | 'ArrowDown' | 'ArrowLeft' | 'ArrowRight'

function center(r: DOMRect) {
  return { x: r.left + r.width / 2, y: r.top + r.height / 2 }
}

function move(dir: Dir) {
  const cur = document.activeElement as HTMLElement | null
  const all = [...document.querySelectorAll<HTMLElement>(FOCUSABLE)].filter((el) => {
    const r = el.getBoundingClientRect()
    return r.width > 0 && r.height > 0 && el !== cur
  })
  if (!cur || cur === document.body) {
    all[0]?.focus()
    return
  }
  const from = center(cur.getBoundingClientRect())
  let best: HTMLElement | null = null
  let bestScore = Infinity
  for (const el of all) {
    const to = center(el.getBoundingClientRect())
    const dx = to.x - from.x
    const dy = to.y - from.y
    const ok = { ArrowUp: dy < -4, ArrowDown: dy > 4, ArrowLeft: dx < -4, ArrowRight: dx > 4 }[dir]
    if (!ok) continue
    const primary = dir === 'ArrowUp' || dir === 'ArrowDown' ? Math.abs(dy) : Math.abs(dx)
    const secondary = dir === 'ArrowUp' || dir === 'ArrowDown' ? Math.abs(dx) : Math.abs(dy)
    const score = primary + secondary * 2
    if (score < bestScore) {
      bestScore = score
      best = el
    }
  }
  if (best) {
    best.focus()
    best.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'smooth' })
  }
}

export function enableSpatialNav() {
  window.addEventListener('keydown', (e) => {
    if (!['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(e.key)) return
    if ((e.target as HTMLElement).matches('input, textarea')) return
    e.preventDefault()
    move(e.key as Dir)
  })
}
