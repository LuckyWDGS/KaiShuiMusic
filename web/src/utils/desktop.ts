// 桌面壳（WebView2 无边框窗口）注入的窗口控制接口；在浏览器中运行时不存在
/** tip:<0|1>:<歌名 - 歌手> 更新托盘图标的悬停提示和播放状态 */
type WinAction = 'drag' | 'min' | 'max' | 'close' | 'mini:on' | 'mini:off' | `tip:${string}`

const win = () => (window as unknown as { kaishuiWin?: (a: WinAction) => Promise<void> }).kaishuiWin

export const isDesktopShell = () => typeof win() === 'function'

export function winAction(a: WinAction) {
  win()?.(a)
}

/** 绑定到标题栏区域：按住空白处拖动窗口，双击最大化 / 还原 */
export function onTitleMouseDown(e: MouseEvent) {
  if (e.button !== 0 || !isDesktopShell()) return
  if ((e.target as HTMLElement).closest('button, a, input, .search, [data-no-drag]')) return
  if (e.detail === 2) winAction('max')
  else winAction('drag')
}
