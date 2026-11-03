<script setup lang="ts">
// 底部播放条（参照汽水音乐）：顶边通栏进度线，中间胶囊播放键，右侧音源 / 词 / 模式 / 音量 / 列表。
// 推荐页（整窗封面色）左侧只放操作按钮和时间；其它页面左侧显示封面、歌名 – 歌手和一排操作按钮。
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  Play, Pause, SkipBack, SkipForward, Heart, Repeat, Repeat1, Shuffle, Volume2, Volume1, VolumeX,
  ListMusic, ListPlus, ThumbsDown, MoreVertical, Loader2,
} from 'lucide-vue-next'
import CoverImg from './CoverImg.vue'
import ProgressBar from './ProgressBar.vue'
import VolumeSlider from './VolumeSlider.vue'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import { formatTime, sourceName } from '@/utils/format'
import { useSongActions } from '@/utils/songActions'
import { LOCAL_FILE } from '@/api'

const route = useRoute()
const player = usePlayer()
const lib = useLibrary()
const { openSongMenu, openCollectMenu } = useSongActions()

const immersive = computed(() => route.name === 'foryou' || player.showNowPlaying)
const modeIcon = computed(() => ({ list: Repeat, single: Repeat1, shuffle: Shuffle })[player.mode])
const modeTitle = computed(() => ({ list: '列表循环', single: '单曲循环', shuffle: '随机播放' })[player.mode])
const volIcon = computed(() => (player.muted || player.volume === 0 ? VolumeX : player.volume < 0.5 ? Volume1 : Volume2))
const sourceLabel = computed(() => {
  const r = player.resolved
  if (!r) return ''
  if (r.song.source === LOCAL_FILE) return '本地'
  return sourceName(r.song.source) + (r.preview ? '·试听' : '')
})
const timeText = computed(() => `${formatTime(player.currentTime)} / ${formatTime(player.duration)}`)

function dislike() {
  // 不喜欢：从播放列表移除并播放下一首
  if (player.index >= 0) player.removeAt(player.index)
}
function onWheelVolume(e: WheelEvent) {
  player.setVolume(player.volume + (e.deltaY < 0 ? 0.05 : -0.05))
}
</script>

<template>
  <footer class="player-bar" :class="{ immersive }">
    <ProgressBar class="progress" :value="player.currentTime" :max="player.duration" :title="timeText" @change="player.seek" />

    <div class="left">
      <template v-if="!immersive && player.current">
        <button class="thumb" title="打开推荐页查看歌词" @click="player.showNowPlaying = true">
          <CoverImg :src="player.current.cover" :size="120" :source="player.current.source" radius="4px" />
        </button>
        <div class="song">
          <div class="line ellipsis"><b>{{ player.current.name }}</b><span class="artist"> – {{ player.current.artist }}</span></div>
          <div class="ops">
            <button class="op" :class="{ liked: lib.isLiked(player.current) }" title="喜欢" @click="lib.toggleLike(player.current)">
              <Heart :size="19" :fill="lib.isLiked(player.current) ? 'currentColor' : 'none'" />
            </button>
            <button v-if="player.current.source !== LOCAL_FILE" class="op" title="收藏" @click="openCollectMenu($event, player.current)"><ListPlus :size="19" /></button>
            <button class="op" title="不喜欢" @click="dislike"><ThumbsDown :size="18" /></button>
            <button class="op" title="更多" @click="openSongMenu($event, player.current)"><MoreVertical :size="18" /></button>
          </div>
        </div>
      </template>
      <template v-else-if="player.current">
        <button v-if="player.current.source !== LOCAL_FILE" class="op" title="收藏" @click="openCollectMenu($event, player.current)"><ListPlus :size="21" /></button>
        <button class="op" title="不喜欢" @click="dislike"><ThumbsDown :size="20" /></button>
        <button class="op" title="更多" @click="openSongMenu($event, player.current)"><MoreVertical :size="20" /></button>
        <span class="time">{{ timeText }}</span>
      </template>
    </div>

    <div class="center">
      <button class="icon-btn ctrl" title="上一首" @click="player.prev()"><SkipBack :size="22" fill="currentColor" /></button>
      <button class="play-btn" :title="player.playing ? '暂停' : '播放'" @click="player.toggle()">
        <Loader2 v-if="player.loading && player.playing" :size="24" class="spin" />
        <Pause v-else-if="player.playing" :size="24" fill="currentColor" />
        <Play v-else :size="24" fill="currentColor" class="play-icon" />
      </button>
      <button class="icon-btn ctrl" title="下一首" @click="player.next()"><SkipForward :size="22" fill="currentColor" /></button>
    </div>

    <div class="right">
      <span v-if="sourceLabel" class="pill" :title="player.resolved?.switched ? '已自动切换到可完整播放的音源' : '当前音源'">{{ sourceLabel }}</span>
      <button class="icon-btn text" :class="{ active: player.showNowPlaying }" title="歌词" @click="player.showNowPlaying = !player.showNowPlaying">词</button>
      <button class="icon-btn" :title="modeTitle" @click="player.cycleMode()"><component :is="modeIcon" :size="20" /></button>
      <div class="volume" @wheel.prevent="onWheelVolume">
        <button class="icon-btn" title="静音" @click="player.muted = !player.muted"><component :is="volIcon" :size="21" /></button>
        <div class="vol-pop">
          <VolumeSlider :value="player.muted ? 0 : player.volume" @change="player.setVolume" />
        </div>
      </div>
      <button class="icon-btn" :class="{ active: player.showQueue }" title="播放列表" @click="player.showQueue = !player.showQueue">
        <ListMusic :size="21" />
      </button>
    </div>
  </footer>
</template>

<style scoped>
.player-bar {
  position: relative; z-index: 30; height: var(--playerbar-h); display: grid;
  grid-template-columns: 1fr auto 1fr; align-items: center; padding: 6px 20px 0;
  background: var(--theme-bar); transition: background .6s;
}
.player-bar.immersive { background: transparent; }
.progress { position: absolute; left: 0; right: 0; top: -7px; }
.progress :deep(.rail) { height: 3px; background: rgba(255, 255, 255, .15); border-radius: 0; }
.progress :deep(.fill) { background: rgba(255, 255, 255, .55) !important; }
.immersive .progress :deep(.rail) { height: 2px; background: rgba(255, 255, 255, .22); }
.immersive .progress :deep(.fill) { background: rgba(255, 255, 255, .85) !important; }
.progress :deep(.knob) { opacity: 1; width: 9px; height: 9px; right: -4px; margin-top: -4.5px; }
.progress:hover :deep(.knob) { width: 12px; height: 12px; right: -6px; margin-top: -6px; }
.left { display: flex; align-items: center; gap: 4px; min-width: 0; }
.thumb { width: 56px; flex-shrink: 0; border-radius: 4px; overflow: hidden; }
.song { min-width: 0; margin-left: 16px; }
.line { font-size: 15px; }
.line b { font-weight: 700; }
.artist { color: var(--text-2); font-weight: 600; }
.ops { display: flex; gap: 14px; margin-top: 6px; }
.op { display: flex; align-items: center; justify-content: center; color: var(--text); opacity: .85; padding: 2px; }
.op:hover { opacity: 1; }
.op.liked { color: #fe2c55; opacity: 1; }
.immersive .left { gap: 10px; }
.immersive .op { width: 34px; height: 34px; }
.time { margin-left: 10px; font-size: 13px; font-weight: 600; color: var(--text-2); font-variant-numeric: tabular-nums; white-space: nowrap; }
.center { display: flex; align-items: center; gap: 30px; }
.ctrl { color: var(--text) !important; }
.play-btn {
  width: 84px; height: 48px; border-radius: 24px; background: rgba(255, 255, 255, .1);
  display: flex; align-items: center; justify-content: center; transition: background .15s;
}
.play-btn:hover { background: rgba(255, 255, 255, .18); }
.play-icon { margin-left: 3px; }
.right { display: flex; align-items: center; justify-content: flex-end; gap: 8px; }
.right .icon-btn { color: var(--text); }
.pill {
  height: 26px; line-height: 26px; padding: 0 9px; border-radius: 7px; background: rgba(255, 255, 255, .12);
  font-size: 12px; font-weight: 700; white-space: nowrap; margin-right: 6px; cursor: default;
}
.text { font-size: 17px; font-weight: 700; }
.icon-btn.active { color: var(--text); background: var(--bg-active); }
.volume { position: relative; }
.vol-pop { position: absolute; bottom: 42px; left: 50%; transform: translateX(-50%); display: none; padding-bottom: 6px; }
.volume:hover .vol-pop { display: block; }
/* 透明的连接区域，鼠标从图标移到滑块时不会断开悬停 */
.vol-pop::after { content: ''; position: absolute; left: 0; right: 0; bottom: -10px; height: 16px; }
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
