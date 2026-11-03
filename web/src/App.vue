<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { updateThemeFromCover } from '@/stores/theme'
import { initMiniSync } from '@/utils/miniPlayer'
import PcLayout from '@/layouts/PcLayout.vue'
import MobileLayout from '@/layouts/MobileLayout.vue'
import LibraryDialogs from '@/components/LibraryDialogs.vue'
import ContextMenu from '@/components/ContextMenu.vue'
import { platform } from '@/platform'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import { toastState } from '@/stores/toast'

const player = usePlayer()
const lib = useLibrary()

// 全局快捷键：空格 播放/暂停，Ctrl+←/→ 切歌
function onKey(e: KeyboardEvent) {
  if ((e.target as HTMLElement).matches('input, textarea')) return
  if (e.code === 'Space') {
    e.preventDefault()
    player.toggle()
  } else if (e.ctrlKey && e.key === 'ArrowRight') player.next()
  else if (e.ctrlKey && e.key === 'ArrowLeft') player.prev()
}

// 推荐页展示的歌曲（尚未开始播放时）也会改变主题色，见 ForYou.vue
watch(() => player.current?.cover, () => updateThemeFromCover(player.current?.cover, player.current?.source), { immediate: true })

onMounted(() => {
  lib.load().then(initMiniSync)
  window.addEventListener('keydown', onKey)
})
</script>

<template>
  <div :class="`platform-${platform}`" style="height: 100%">
    <PcLayout v-if="platform === 'pc'" />
    <MobileLayout v-else />
    <LibraryDialogs />
    <ContextMenu />
    <Transition name="toast">
      <div v-if="toastState.visible" class="toast">{{ toastState.text }}</div>
    </Transition>
  </div>
</template>

<style>
.toast {
  position: fixed; left: 50%; top: 72px; transform: translateX(-50%); z-index: 100; max-width: 80vw;
  padding: 10px 20px; border-radius: 22px; background: var(--surface);
  font-size: 13px; box-shadow: 0 8px 30px rgba(0, 0, 0, .4); pointer-events: none;
}
.toast-enter-active, .toast-leave-active { transition: opacity .2s, transform .2s; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translate(-50%, -8px); }
</style>
