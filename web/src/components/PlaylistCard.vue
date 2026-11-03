<script setup lang="ts">
import { useRouter } from 'vue-router'
import { Play, Headphones } from 'lucide-vue-next'
import CoverImg from './CoverImg.vue'
import type { DetailKind, Playlist, TopList } from '@/types'
import { formatCount, sourceName } from '@/utils/format'

const props = defineProps<{ item: Playlist | TopList; kind?: DetailKind; showSource?: boolean }>()
const router = useRouter()

function open() {
  const it = props.item
  router.push({
    name: 'detail',
    params: { kind: props.kind || 'playlist', source: it.source, id: it.id },
    query: { name: it.name, cover: it.cover, desc: it.description || '' },
  })
}
const plays = () => formatCount((props.item as Playlist).play_count)
</script>

<template>
  <button class="card" @click="open">
    <div class="cover-wrap">
      <CoverImg :src="item.cover" :size="300" :source="item.source" radius="var(--radius)" />
      <span v-if="plays()" class="plays"><Headphones :size="12" />{{ plays() }}</span>
      <span class="play"><Play :size="20" fill="currentColor" /></span>
    </div>
    <div class="name">{{ item.name }}</div>
    <div v-if="showSource || (kind === 'toplist' && (item as TopList).update_frequency)" class="sub ellipsis">
      <template v-if="kind === 'toplist' && (item as TopList).update_frequency">{{ (item as TopList).update_frequency }}</template>
      <template v-else>{{ sourceName(item.source) }}</template>
    </div>
  </button>
</template>

<style scoped>
.card { text-align: left; width: 100%; min-width: 0; }
.cover-wrap { position: relative; border-radius: var(--radius); overflow: hidden; }
.cover-wrap :deep(img) { transition: transform .35s ease; }
.card:hover .cover-wrap :deep(img) { transform: scale(1.05); }
.plays {
  position: absolute; top: 8px; right: 8px; display: inline-flex; align-items: center; gap: 3px;
  font-size: 11px; padding: 2px 7px; border-radius: 10px; background: rgba(0, 0, 0, .45); backdrop-filter: blur(6px);
}
.play {
  position: absolute; right: 10px; bottom: 10px; width: 40px; height: 40px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; color: #121212;
  background: var(--accent); opacity: 0; transform: translateY(6px); transition: all .2s ease;
  box-shadow: 0 6px 18px rgba(0, 0, 0, .4);
}
.card:hover .play { opacity: 1; transform: none; }
.name {
  margin-top: 10px; font-size: 14px; line-height: 1.45; display: -webkit-box; -webkit-line-clamp: 2;
  line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.sub { margin-top: 4px; font-size: 12px; color: var(--text-3); }
</style>
