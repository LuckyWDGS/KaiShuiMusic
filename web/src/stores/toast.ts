import { reactive } from 'vue'

export const toastState = reactive({ text: '', visible: false })
let timer: number | undefined

export function toast(text: string, ms = 2200) {
  toastState.text = text
  toastState.visible = true
  clearTimeout(timer)
  timer = window.setTimeout(() => (toastState.visible = false), ms)
}
