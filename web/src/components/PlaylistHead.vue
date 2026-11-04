<script setup lang="ts">
// 歌单页头部（参照汽水音乐）：左侧封面，右侧标题、副标题，底部"播放全部 N"与 ⋯ 菜单
import { Play, MoreHorizontal } from 'lucide-vue-next'
import CoverImg from './CoverImg.vue'
import { openMenu, type MenuItem } from '@/stores/menu'

const props = defineProps<{ title: string; cover?: string; source?: string; sub?: string; desc?: string; count: number; menu?: MenuItem[] }>()
const emit = defineEmits<{ play: [] }>()
</script>

<template>
  <header class="head">
    <div class="cover">
      <CoverImg v-if="cover" :src="cover" :size="400" :source="source" radius="8px" />
      <div v-else class="placeholder"><img src="/logo.png" alt="" /></div>
    </div>
    <div class="info">
      <h1 class="ellipsis">{{ title }}</h1>
      <div v-if="sub" class="sub ellipsis">{{ sub }}</div>
      <p v-if="desc" class="desc">{{ desc }}</p>
      <div class="actions">
        <button class="play" :disabled="!count" @click="emit('play')">
          <Play :size="16" fill="currentColor" />播放全部<template v-if="count"> {{ count }}</template>
        </button>
        <button v-if="props.menu?.length" class="more" title="更多" @click="openMenu($event, props.menu!)"><MoreHorizontal :size="20" /></button>
        <slot />
      </div>
    </div>
  </header>
</template>

<style scoped>
.head { display: flex; gap: 20px; margin: 4px 0 20px; }
.cover { width: 200px; flex-shrink: 0; }
.placeholder {
  aspect-ratio: 1; border-radius: 8px; background: #262626; display: flex; align-items: center; justify-content: center;
}
.placeholder img { width: 64px; opacity: .35; filter: grayscale(1); }
.info { flex: 1; min-width: 0; display: flex; flex-direction: column; padding-top: 2px; }
h1 { font-size: 20px; font-weight: 700; margin: 0 0 10px; }
.sub { font-size: 13px; font-weight: 600; color: var(--text-2); }
.desc {
  margin: 8px 0 0; font-size: 13px; line-height: 1.6; color: var(--text-3); max-width: 640px;
  display: -webkit-box; -webkit-line-clamp: 2; line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.actions { display: flex; align-items: center; gap: 16px; margin-top: auto; }
.play {
  display: inline-flex; align-items: center; gap: 6px; height: 32px; padding: 0 14px; border-radius: 6px;
  background: #e9e9e9; color: #121212; font-size: 14px; font-weight: 700; white-space: nowrap; flex-shrink: 0;
}
.play:hover { background: #fff; }
.play:disabled { background: #2c2c2c; color: var(--text-3); cursor: default; }
.more { display: flex; color: var(--text-2); padding: 4px; border-radius: 6px; }
.more:hover { color: var(--text); background: rgba(255, 255, 255, .08); }
@media (max-width: 760px) {
  .head { gap: 14px; }
  .cover { width: 120px; }
  h1 { font-size: 17px; margin-bottom: 6px; }
}
</style>
