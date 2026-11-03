<script setup lang="ts">
// 推荐：单曲推荐流（参照汽水音乐）。滚轮 / 方向键切换歌曲。
import { computed, onActivated, onDeactivated, onMounted, ref, watch } from 'vue'
import { Loader2, AudioLines, X } from 'lucide-vue-next'
import SongStage from '@/components/SongStage.vue'
import { api } from '@/api'
import type { Song } from '@/types'
import { usePlayer } from '@/stores/player'
import { updateThemeFromCover } from '@/stores/theme'

const player = usePlayer()

const error = ref('')
const active = ref(false)
const inFeed = computed(() => player.context === 'feed' && !!player.current)
// 还没开始收听推荐流时展示推荐流第一首（预览），播放条等控制会从这里开始播放
const song = computed<Song | null>(() => (inFeed.value ? player.current : player.feedPreview[0] ?? null))

async function loadFeed() {
  error.value = ''
  try {
    player.feedPreview = await api.feed(1)
  } catch (e) {
    error.value = (e as Error).message
  }
}

async function exitStation() {
  player.leaveStation()
  if (!player.feedPreview.length) await loadFeed()
  player.startPreview(0)
}

// 滚轮 / 方向键：预览状态下由播放器直接开始推荐流
const go = (delta: 1 | -1) => (delta > 0 ? player.next() : player.prev())

let wheelLock = 0
function onWheel(e: WheelEvent) {
  if (Math.abs(e.deltaY) < 30 || Date.now() - wheelLock < 800) return
  if ((e.target as HTMLElement).closest('.lyric')) return
  wheelLock = Date.now()
  go(e.deltaY > 0 ? 1 : -1)
}

function onKey(e: KeyboardEvent) {
  if ((e.target as HTMLElement).matches('input, textarea') || player.showNowPlaying) return
  if (e.key === 'ArrowDown') { e.preventDefault(); go(1) }
  if (e.key === 'ArrowUp') { e.preventDefault(); go(-1) }
}

// 推荐页展示的歌曲还没播放时：主题色跟随它，并在后台提前解析播放地址，点播放即可开始
watch(song, (s) => {
  if (!s || inFeed.value) return
  player.prepare(s, true)
  if (active.value) updateThemeFromCover(s.cover, s.source)
}, { immediate: true })

onMounted(() => {
  if (!player.feedPreview.length) loadFeed()
})
onActivated(() => {
  active.value = true
  player.previewActive = true
  window.addEventListener('keydown', onKey)
  if (song.value && !inFeed.value) updateThemeFromCover(song.value.cover, song.value.source)
})
onDeactivated(() => {
  active.value = false
  player.previewActive = false
  window.removeEventListener('keydown', onKey)
  if (!inFeed.value) updateThemeFromCover(player.current?.cover, player.current?.source)
})
</script>

<template>
  <div class="for-you" @wheel="onWheel">
    <div v-if="error" class="state">
      <p>{{ error }}</p>
      <button class="btn primary" @click="loadFeed">重新加载</button>
    </div>
    <div v-else-if="!song" class="state"><Loader2 :size="28" class="spin" /></div>
    <Transition v-else name="swap" mode="out-in">
      <SongStage :key="song.source + song.id" :song="song" />
    </Transition>
    <div v-if="inFeed && player.station" class="station">
      <AudioLines :size="14" />{{ player.station.name }}
      <button title="退出模式，回到推荐" @click="exitStation"><X :size="14" /></button>
    </div>
  </div>
</template>

<style scoped>
.for-you { height: 100%; position: relative; overflow: hidden; }
.station {
  position: absolute; top: 8px; left: 32px; display: flex; align-items: center; gap: 6px; height: 30px; padding: 0 6px 0 12px;
  border-radius: 15px; background: rgba(255, 255, 255, .14); font-size: 13px; font-weight: 600;
}
.station button { display: flex; padding: 4px; border-radius: 50%; color: var(--text-2); }
.station button:hover { color: var(--text); background: rgba(255, 255, 255, .12); }
.state { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.swap-enter-active, .swap-leave-active { transition: opacity .25s, transform .25s; }
.swap-enter-from { opacity: 0; transform: translateY(24px); }
.swap-leave-to { opacity: 0; transform: translateY(-24px); }
</style>
