<script setup lang="ts">
// 手机版播放控制：操作按钮 + 进度条 + 播放键。推荐页和全屏播放页共用
import { computed } from 'vue'
import {
  Play, Pause, SkipBack, SkipForward, Repeat, Repeat1, Shuffle, ListMusic, ListPlus, ThumbsDown, MoreHorizontal, Loader2,
} from 'lucide-vue-next'
import ProgressBar from './ProgressBar.vue'
import { usePlayer } from '@/stores/player'
import { formatTime } from '@/utils/format'
import { useSongActions } from '@/utils/songActions'
import { LOCAL_FILE } from '@/api'

const player = usePlayer()
const { openSongMenu, openCollectMenu } = useSongActions()
const modeIcon = computed(() => ({ list: Repeat, single: Repeat1, shuffle: Shuffle })[player.mode])

function dislike() {
  if (player.index >= 0) player.removeAt(player.index)
}
</script>

<template>
  <div class="controls">
    <div v-if="player.current" class="ops">
      <button v-if="player.current.source !== LOCAL_FILE" aria-label="收藏" @click="openCollectMenu($event, player.current)"><ListPlus :size="22" /></button>
      <button aria-label="不喜欢" @click="dislike"><ThumbsDown :size="21" /></button>
      <button aria-label="更多" @click="openSongMenu($event, player.current)"><MoreHorizontal :size="22" /></button>
    </div>
    <div class="progress">
      <span>{{ formatTime(player.currentTime) }}</span>
      <ProgressBar :value="player.currentTime" :max="player.duration" @change="player.seek" />
      <span>{{ formatTime(player.duration) }}</span>
    </div>
    <div class="buttons">
      <button aria-label="播放模式" @click="player.cycleMode()"><component :is="modeIcon" :size="22" /></button>
      <button aria-label="上一首" @click="player.prev()"><SkipBack :size="26" fill="currentColor" /></button>
      <button class="play" :aria-label="player.playing ? '暂停' : '播放'" @click="player.toggle()">
        <Loader2 v-if="player.loading" :size="28" class="spin" />
        <Pause v-else-if="player.playing" :size="28" fill="currentColor" />
        <Play v-else :size="28" fill="currentColor" class="play-icon" />
      </button>
      <button aria-label="下一首" @click="player.next()"><SkipForward :size="26" fill="currentColor" /></button>
      <button aria-label="播放列表" @click="player.showQueue = true"><ListMusic :size="22" /></button>
    </div>
  </div>
</template>

<style scoped>
.controls { padding: 0 24px 10px; flex-shrink: 0; }
.ops { display: flex; justify-content: flex-end; gap: 8px; margin: 0 -8px 2px; }
.ops button, .buttons button { display: flex; align-items: center; justify-content: center; width: 44px; height: 44px; }
.progress { display: flex; align-items: center; gap: 10px; font-size: 11px; font-weight: 600; color: var(--text-2); font-variant-numeric: tabular-nums; }
.progress :deep(.bar) { flex: 1; height: 24px; }
.progress :deep(.rail) { background: rgba(255, 255, 255, .22); }
.progress :deep(.knob) { opacity: 1; }
.buttons { display: flex; align-items: center; justify-content: space-between; margin-top: 2px; }
.buttons .play { width: 64px; height: 64px; border-radius: 50%; background: rgba(255, 255, 255, .16); }
.play-icon { margin-left: 3px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
