export interface LyricLine {
  time: number
  text: string
}

const TIME_TAG = /\[(\d{1,3}):(\d{1,2})(?:[.:](\d{1,3}))?\]/g

/** 解析 LRC 歌词，兼容一行多个时间标签以及逐字歌词中的 <mm:ss.xx> 标签 */
export function parseLyric(raw: string): LyricLine[] {
  const lines: LyricLine[] = []
  for (const row of raw.split(/\r?\n/)) {
    const times: number[] = []
    let m: RegExpExecArray | null
    TIME_TAG.lastIndex = 0
    while ((m = TIME_TAG.exec(row))) {
      const frac = m[3] ? Number(m[3]) / 10 ** m[3].length : 0
      times.push(Number(m[1]) * 60 + Number(m[2]) + frac)
    }
    if (!times.length) continue
    const text = row
      .replace(TIME_TAG, '')
      .replace(/<\d{1,3}:\d{1,2}(?:[.:]\d{1,3})?>/g, '')
      .replace(/\(\d+,\d+(?:,\d+)?\)/g, '')
      .trim()
    for (const time of times) lines.push({ time, text })
  }
  lines.sort((a, b) => a.time - b.time)
  return lines.filter((l, i) => l.text || (i > 0 && lines[i - 1].text))
}

export function currentLineIndex(lines: LyricLine[], time: number) {
  let lo = 0
  let hi = lines.length - 1
  let ans = -1
  while (lo <= hi) {
    const mid = (lo + hi) >> 1
    if (lines[mid].time <= time + 0.15) {
      ans = mid
      lo = mid + 1
    } else hi = mid - 1
  }
  return ans
}
