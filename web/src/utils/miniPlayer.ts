// 任务栏播放器（主窗口一侧）：开关状态，以及通过 BroadcastChannel 把播放状态同步给任务栏里的小播放器
import { ref, watch } from 'vue'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import { isDesktopShell, winAction } from './desktop'

export const MINI_CHANNEL = 'kaishui-mini'
const KEY = 'kaishui.miniPlayer'

export interface MiniState {
  song: { name: string; artist: string; cover: string; source: string } | null
  playing: boolean
  lyric: string
  liked: boolean
}
export type MiniMessage =
  | { type: 'state'; state: MiniState }
  | { type: 'hello' }
  | { type: 'cmd'; cmd: 'toggle' | 'next' | 'prev' | 'like' }

function loadPref() {
  try {
    return localStorage.getItem(KEY) !== '0' // 默认开启
  } catch {
    return true
  }
}

export const miniOn = ref(loadPref())

export function toggleMini() {
  miniOn.value = !miniOn.value
  try { localStorage.setItem(KEY, miniOn.value ? '1' : '0') } catch { /* ignore */ }
  winAction(miniOn.value ? 'mini:on' : 'mini:off')
}

/** 系统托盘：右键菜单的播放控制，以及把正在播放的歌名同步到托盘图标的悬停提示 */
function initTraySync() {
  const player = usePlayer()
  ;(window as unknown as { __kaishuiTray: (cmd: string) => void }).__kaishuiTray = (cmd) => {
    if (cmd === 'toggle') player.toggle()
    else if (cmd === 'next') player.next()
    else if (cmd === 'prev') player.prev()
  }
  watch(() => [player.current, player.playing] as const, ([song, playing]) => {
    winAction(`tip:${playing ? 1 : 0}:${song ? `${song.name} - ${song.artist}` : ''}`)
  }, { immediate: true })
}

/** 主窗口启动时调用：托盘同步、恢复任务栏播放器开关并开始同步 */
export function initMiniSync() {
  if (!isDesktopShell()) return
  initTraySync()
  if (typeof BroadcastChannel === 'undefined') return
  const player = usePlayer()
  const lib = useLibrary()
  const ch = new BroadcastChannel(MINI_CHANNEL)

  const snapshot = (): MiniState => {
    const s = player.current
    const line = player.lyrics[player.lyricIndex]?.text
    return {
      song: s ? { name: s.name, artist: s.artist, cover: s.cover, source: s.source } : null,
      playing: player.playing,
      lyric: line || (player.lyrics.length ? '' : '暂无歌词，请欣赏'),
      liked: lib.isLiked(s),
    }
  }
  const send = () => ch.postMessage({ type: 'state', state: snapshot() } satisfies MiniMessage)

  watch(() => [player.current, player.playing, player.lyricIndex, player.lyrics, lib.data.likes], send)
  ch.onmessage = (e: MessageEvent<MiniMessage>) => {
    const msg = e.data
    if (msg.type === 'hello') send()
    else if (msg.type === 'cmd') {
      if (msg.cmd === 'toggle') player.toggle()
      else if (msg.cmd === 'next') player.next()
      else if (msg.cmd === 'prev') player.prev()
      else if (msg.cmd === 'like' && player.current) lib.toggleLike(player.current)
    }
  }
  if (miniOn.value) winAction('mini:on')
}
