// 全局弹出菜单（歌曲 ⋯ 菜单、收藏到歌单等），由 components/ContextMenu.vue 渲染
import { markRaw, reactive, type Component } from 'vue'

export interface MenuItem {
  label: string
  icon?: Component
  danger?: boolean
  action: () => void
}

export const menuState = reactive({
  visible: false,
  x: 0,
  y: 0,
  /** 菜单向上展开（锚点在窗口下半部分时，如播放条上的按钮） */
  up: false,
  items: [] as MenuItem[],
})

/** 在触发元素旁边打开菜单 */
export function openMenu(e: MouseEvent, items: MenuItem[]) {
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  const up = rect.top > window.innerHeight * 0.55
  menuState.items = items.map((it) => ({ ...it, icon: it.icon && markRaw(it.icon) }))
  menuState.up = up
  menuState.x = rect.left
  menuState.y = up ? rect.top - 6 : rect.bottom + 6
  menuState.visible = true
}

/** 在上一个菜单的位置换一组菜单项（二级菜单，如"添加到歌单"） */
export function reopenMenu(items: MenuItem[]) {
  menuState.items = items.map((it) => ({ ...it, icon: it.icon && markRaw(it.icon) }))
  menuState.visible = true
}

export function closeMenu() {
  menuState.visible = false
}
