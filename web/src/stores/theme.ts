// 跟随歌曲封面变色：取当前封面的主色，压低饱和度和亮度后作为整个窗口的背景色。
import { getServerBase } from '@/api'

const DEFAULT = { h: 190, s: 14 }
let token = 0

function apply(h: number, s: number) {
  const root = document.documentElement.style
  root.setProperty('--theme-h', String(Math.round(h)))
  root.setProperty('--theme-s', `${Math.round(s)}%`)
}

function rgbToHsl(r: number, g: number, b: number) {
  r /= 255; g /= 255; b /= 255
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const l = (max + min) / 2
  if (max === min) return { h: 0, s: 0, l }
  const d = max - min
  const s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
  let h = max === r ? (g - b) / d + (g < b ? 6 : 0) : max === g ? (b - r) / d + 2 : (r - g) / d + 4
  return { h: h * 60, s, l }
}

/** 从封面图提取主色；图片经后端代理加载，保证同源可读取像素 */
export function updateThemeFromCover(cover?: string, source?: string) {
  const my = ++token
  if (!cover) return apply(DEFAULT.h, DEFAULT.s)
  const img = new Image()
  img.crossOrigin = 'anonymous'
  img.onload = () => {
    if (my !== token) return
    try {
      const size = 24
      const canvas = document.createElement('canvas')
      canvas.width = canvas.height = size
      const ctx = canvas.getContext('2d', { willReadFrequently: true })!
      ctx.drawImage(img, 0, 0, size, size)
      const data = ctx.getImageData(0, 0, size, size).data
      // 按饱和度加权求色相，忽略接近黑白的像素
      let x = 0, y = 0, wsum = 0, ssum = 0, n = 0
      for (let i = 0; i < data.length; i += 4) {
        const { h, s, l } = rgbToHsl(data[i], data[i + 1], data[i + 2])
        if (l < 0.08 || l > 0.94) continue
        const w = s * (1 - Math.abs(l - 0.5))
        const rad = (h * Math.PI) / 180
        x += Math.cos(rad) * w
        y += Math.sin(rad) * w
        wsum += w
        ssum += s
        n++
      }
      if (!n || wsum < 0.5) return apply(DEFAULT.h, 6)
      const hue = ((Math.atan2(y, x) * 180) / Math.PI + 360) % 360
      const sat = Math.min(30, Math.max(8, (ssum / n) * 45))
      apply(hue, sat)
    } catch {
      apply(DEFAULT.h, DEFAULT.s)
    }
  }
  img.onerror = () => my === token && apply(DEFAULT.h, DEFAULT.s)
  const q = new URLSearchParams({ url: cover, source: source || '' })
  img.src = `${getServerBase()}/api/cover?${q}`
}

apply(DEFAULT.h, DEFAULT.s)
