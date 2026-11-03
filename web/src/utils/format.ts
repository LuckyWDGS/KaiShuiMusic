export const SOURCE_NAMES: Record<string, string> = {
  netease: '网易云',
  qq: 'QQ音乐',
  kugou: '酷狗',
  kuwo: '酷我',
  migu: '咪咕',
  soda: '汽水',
  qianqian: '千千',
  bilibili: 'B站',
  apple: 'Apple',
  joox: 'JOOX',
  jamendo: 'Jamendo',
  fivesing: '5sing',
  local: '本地',
}

export const sourceName = (s: string) => SOURCE_NAMES[s] || s

export function formatTime(sec: number) {
  if (!isFinite(sec) || sec < 0) sec = 0
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

export function formatCount(n: number) {
  if (!n) return ''
  if (n >= 1e8) return `${(n / 1e8).toFixed(1)}亿`
  if (n >= 1e4) return `${(n / 1e4).toFixed(1)}万`
  return String(n)
}

export function formatSize(b: number) {
  if (b >= 1 << 20) return `${(b / (1 << 20)).toFixed(1)} MB`
  return `${Math.round(b / 1024)} KB`
}

export const songKey = (s: { source: string; id: string }) => `${s.source}:${s.id}`

/** 有的平台封面是 http，统一转 https 以免混合内容警告（汽水、网易云均支持） */
export function coverUrl(url?: string, size?: number) {
  if (!url) return ''
  let u = url.replace(/^http:\/\//, 'https://')
  if (size && /music\.126\.net/.test(u) && !u.includes('param=')) {
    u += `${u.includes('?') ? '&' : '?'}param=${size}y${size}`
  }
  // 汽水封面默认 960px，列表里请求小图更省流量（加载失败时 CoverImg 会回退原图）
  if (size && size <= 480) u = u.replace(/resize:960:960/, 'resize:360:360')
  return u
}
