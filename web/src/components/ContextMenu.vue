<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { closeMenu, menuState, type MenuItem } from '@/stores/menu'

const el = ref<HTMLElement>()
const left = ref(0)

// 防止菜单超出窗口右侧
watch(() => [menuState.visible, menuState.x], async () => {
  if (!menuState.visible) return
  left.value = menuState.x
  await nextTick()
  const w = el.value?.offsetWidth || 200
  if (left.value + w > window.innerWidth - 8) left.value = window.innerWidth - w - 8
})

function run(item: MenuItem) {
  closeMenu()
  item.action()
}

const onDown = (e: MouseEvent) => {
  if (menuState.visible && !el.value?.contains(e.target as Node)) closeMenu()
}
const onKey = (e: KeyboardEvent) => e.key === 'Escape' && closeMenu()
onMounted(() => {
  window.addEventListener('mousedown', onDown, true)
  window.addEventListener('keydown', onKey)
  window.addEventListener('blur', closeMenu)
  window.addEventListener('resize', closeMenu)
})
onBeforeUnmount(() => {
  window.removeEventListener('mousedown', onDown, true)
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('blur', closeMenu)
  window.removeEventListener('resize', closeMenu)
})
</script>

<template>
  <div
    v-if="menuState.visible"
    ref="el"
    class="menu"
    :style="{ left: left + 'px', [menuState.up ? 'bottom' : 'top']: (menuState.up ? `calc(100% - ${menuState.y}px)` : menuState.y + 'px') }"
    @contextmenu.prevent
  >
    <button v-for="(it, i) in menuState.items" :key="i" class="item" :class="{ danger: it.danger }" @click="run(it)">
      <component :is="it.icon" v-if="it.icon" :size="18" />
      <span class="ellipsis">{{ it.label }}</span>
    </button>
  </div>
</template>

<style scoped>
.menu {
  position: fixed; z-index: 95; min-width: 200px; max-width: 280px; max-height: 60vh; overflow-y: auto; padding: 6px;
  border-radius: 8px; background: var(--surface); box-shadow: 0 12px 32px rgba(0, 0, 0, .45);
}
.item {
  display: flex; align-items: center; gap: 10px; width: 100%; height: 40px; padding: 0 12px; border-radius: 6px;
  font-size: 14px; font-weight: 600; text-align: left;
}
.item:hover { background: rgba(255, 255, 255, .08); }
.item.danger { color: #ff6b81; }
</style>
