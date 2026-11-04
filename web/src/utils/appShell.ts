// 安卓 APP 壳（WebView）注入的 KaiShuiApp 接口；在浏览器和桌面版中不存在
import { watch } from 'vue'
import type { Router } from 'vue-router'
import { getServerBase } from '@/api'
import { closeMenu, menuState } from '@/stores/menu'
import { usePlayer } from '@/stores/player'
import { coverUrl } from './format'

interface AppBridge {
  playback(title: string, artist: string, cover: string, playing: boolean, duration: number, position: number): void
}
type ShellWindow = Window & {
  KaiShuiApp?: AppBridge
  __kaishuiMedia?: (cmd: string, value: number) => void
  __kaishuiBack?: () => boolean
}

const shell = window as ShellWindow

export const isAppShell = () => !!shell.KaiShuiApp

/** 主标签页：返回键到这里后退到后台，而不是继续后退 */
export const TAB_ROUTES = ['foryou', 'modes', 'mine']

/** 通知栏封面经内置服务获取（防盗链由服务处理），系统无法访问相对地址，需要完整地址 */
function notificationCover(cover?: string, source?: string) {
  if (!cover) return ''
  const qs = new URLSearchParams({ url: coverUrl(cover, 480), source: source || '' })
  return `${getServerBase() || location.origin}/api/cover?${qs}`
}

export function initAppShell(router: Router) {
  const app = shell.KaiShuiApp
  if (!app) return
  const player = usePlayer()

  const sync = () => {
    const s = player.current
    if (!s) return
    app.playback(s.name, s.artist, notificationCover(s.cover, s.source), player.playing, player.duration, player.currentTime)
  }
  watch(() => [player.current, player.playing, Math.round(player.duration)], sync)
  // 拖动进度后同步给锁屏进度条；正常播放时系统自己推算，不用每秒同步
  watch(() => player.currentTime, (t, old) => {
    if (Math.abs(t - old) > 2) sync()
  })

  shell.__kaishuiMedia = (cmd, value) => {
    if (cmd === 'toggle') player.toggle()
    else if (cmd === 'play' && !player.playing) player.toggle()
    else if (cmd === 'pause' && player.playing) player.toggle()
    else if (cmd === 'next') player.next()
    else if (cmd === 'prev') player.prev()
    else if (cmd === 'seek') player.seek(value / 1000)
  }

  shell.__kaishuiBack = () => {
    if (menuState.visible) {
      closeMenu()
      return true
    }
    if (player.showQueue) {
      player.showQueue = false
      return true
    }
    if (player.showNowPlaying) {
      player.showNowPlaying = false
      return true
    }
    if (TAB_ROUTES.includes(String(router.currentRoute.value.name))) return false
    router.back()
    return true
  }
}
