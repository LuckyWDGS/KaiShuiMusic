<script setup lang="ts">
// 听歌模式（参照汽水音乐）：常用模式按钮 + 探索更多新模式卡片，点击后按频道连续播放并回到推荐页
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Heart, Sparkles, Leaf, Flame, Library, BookOpen, Fish, Moon, Radio, Loader2, AudioLines } from 'lucide-vue-next'
import CoverImg from '@/components/CoverImg.vue'
import { api } from '@/api'
import type { Station } from '@/types'
import { usePlayer } from '@/stores/player'
import { toast } from '@/stores/toast'

const router = useRouter()
const player = usePlayer()

const ICONS = { heart: Heart, sparkles: Sparkles, leaf: Leaf, flame: Flame, library: Library, book: BookOpen, fish: Fish, moon: Moon }
const common = ref<Station[]>([])
const more = ref<Station[]>([])
const error = ref('')
const starting = ref('')

async function load() {
  error.value = ''
  try {
    const r = await api.stations()
    common.value = r.common
    more.value = r.more
  } catch (e) {
    error.value = (e as Error).message
  }
}

const isActive = (st: Station) => player.context === 'feed' && player.station?.id === st.id

async function start(st: Station) {
  if (starting.value) return
  starting.value = st.id
  try {
    await player.playStation(st)
    router.push('/')
  } catch (e) {
    toast((e as Error).message)
  } finally {
    starting.value = ''
  }
}
onMounted(load)
</script>

<template>
  <div class="page modes">
    <h1 class="title">听歌模式</h1>
    <div v-if="error" class="empty">
      <p>{{ error }}</p>
      <button class="btn primary" @click="load">重新加载</button>
    </div>
    <template v-else>
      <div class="group">常用模式</div>
      <div class="common">
        <button v-for="st in common" :key="st.id" class="mode" :class="{ active: isActive(st) }" @click="start(st)">
          <Loader2 v-if="starting === st.id" :size="22" class="spin" />
          <AudioLines v-else-if="isActive(st)" :size="22" />
          <component :is="ICONS[st.icon as keyof typeof ICONS] || Radio" v-else :size="22" />
          <span>{{ st.name }}</span>
        </button>
        <template v-if="!common.length">
          <div v-for="i in 8" :key="i" class="mode skeleton"></div>
        </template>
      </div>

      <div class="group">探索更多新模式</div>
      <div class="more">
        <button v-for="st in more" :key="st.id" class="card" :class="{ active: isActive(st) }" @click="start(st)">
          <CoverImg :src="st.cover" :size="360" source="soda" radius="0" class="bg" />
          <span class="shade"></span>
          <span class="text">
            <span class="name">
              <Loader2 v-if="starting === st.id" :size="16" class="spin" />
              <AudioLines v-else-if="isActive(st)" :size="16" />
              {{ st.name }}
            </span>
            <span class="desc ellipsis">{{ st.desc }}</span>
          </span>
        </button>
        <template v-if="!more.length">
          <div v-for="i in 8" :key="i" class="card skeleton"></div>
        </template>
      </div>
    </template>
  </div>
</template>

<style scoped>
.modes { padding-top: 0; }
.title { font-size: 22px; font-weight: 700; margin: 4px 0 22px; }
.group { font-size: 13px; font-weight: 600; color: var(--text-2); margin: 0 0 12px; }
.common { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 28px; }
.mode {
  display: flex; align-items: center; gap: 10px; height: 48px; padding: 0 14px; border-radius: 10px;
  background: #222; font-size: 14px; font-weight: 600; transition: background .15s; text-align: left;
}
.mode:hover { background: #2c2c2c; }
.mode.active { background: #333; }
.more { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.card { position: relative; aspect-ratio: 197 / 132; border-radius: 10px; overflow: hidden; text-align: left; background: #222; }
.card .bg { position: absolute; inset: 0; width: 100%; height: 100%; aspect-ratio: auto; }
.card :deep(img) { transition: transform .4s ease; }
.card:hover :deep(img) { transform: scale(1.06); }
.shade { position: absolute; inset: 0; background: linear-gradient(180deg, rgba(0, 0, 0, .05) 30%, rgba(0, 0, 0, .55)); }
.text { position: absolute; left: 12px; right: 12px; bottom: 12px; display: flex; flex-direction: column; gap: 3px; }
.name { display: flex; align-items: center; gap: 6px; font-size: 16px; font-weight: 700; }
.desc { font-size: 12px; color: rgba(255, 255, 255, .8); }
.card.active { outline: 2px solid rgba(255, 255, 255, .8); outline-offset: -2px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 760px) {
  .common, .more { grid-template-columns: repeat(2, 1fr); }
}
</style>
