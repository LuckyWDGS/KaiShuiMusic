<script setup lang="ts">
// 歌曲舞台（参照汽水音乐推荐页）：左侧封面 + 歌名歌手 + 喜欢，右侧歌词。推荐页与听歌模式共用。
import { computed } from 'vue'
import { Heart } from 'lucide-vue-next'
import VinylDisc from './VinylDisc.vue'
import LyricView from './LyricView.vue'
import type { Song } from '@/types'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import { songKey, sourceName } from '@/utils/format'

// 播放控制统一使用底部播放条，封面只做展示
const props = defineProps<{ song: Song; large?: boolean }>()
const player = usePlayer()
const lib = useLibrary()

const isCurrent = computed(() => !!player.current && songKey(props.song) === songKey(player.current))
const playingSource = computed(() => (isCurrent.value && player.resolved ? player.resolved.song.source : props.song.source))
</script>

<template>
  <div class="stage" :class="{ large }">
    <div class="row">
      <div class="left">
        <VinylDisc :cover="song.cover" :source="song.source" :spinning="isCurrent && player.playing" />
        <div class="info">
          <div class="meta">
            <div class="name ellipsis">{{ song.name }}</div>
            <div class="artist">
              <span class="ellipsis">{{ song.artist }}</span>
              <span class="tag" :title="isCurrent && player.resolved?.switched ? '已自动切换到可完整播放的音源' : '音源'">
                {{ sourceName(playingSource) }}{{ isCurrent && player.resolved?.preview ? ' · 试听' : '' }}
              </span>
            </div>
          </div>
          <button class="like" :class="{ liked: lib.isLiked(song) }" title="喜欢" @click="lib.toggleLike(song)">
            <Heart :size="26" fill="currentColor" />
          </button>
        </div>
      </div>
      <div class="right">
        <LyricView v-if="isCurrent" :anchor="large ? 0.3 : 0.08" :large="large" />
        <div v-else class="tip">点击下方播放键开始收听</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stage {
  --cover: clamp(200px, calc(50vh - 72px), 400px);
  height: 100%; display: flex; align-items: center; justify-content: center; padding: 0 48px 40px;
}
.stage.large { --cover: clamp(240px, calc(62vh - 90px), 520px); }
.row { display: flex; align-items: flex-start; gap: 56px; width: 100%; justify-content: center; }
.large .row { gap: 80px; align-items: center; }
.left { width: var(--cover); flex-shrink: 0; }
.info { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-top: 30px; }
.meta { min-width: 0; }
.name { font-size: 20px; font-weight: 700; line-height: 1.3; }
.artist { display: flex; align-items: center; gap: 8px; margin-top: 6px; font-size: 13px; color: var(--text-2); font-weight: 600; min-width: 0; }
.tag { flex-shrink: 0; font-size: 11px; padding: 1px 7px; border-radius: 5px; background: rgba(255, 255, 255, .16); color: var(--text); }
.like { color: rgba(255, 255, 255, .92); flex-shrink: 0; padding-top: 2px; transition: transform .15s; }
.like:hover { transform: scale(1.1); }
.like.liked { color: var(--danger); }
/* 歌词首行与封面中部对齐，向下延伸并渐隐 */
.right { width: min(380px, 42%); height: calc(var(--cover) * .9); margin-top: calc(var(--cover) * .42); }
.large .right { width: min(520px, 45%); height: calc(var(--cover) * 1.15); margin-top: 0; }
.tip { color: rgba(255, 255, 255, .55); font-size: 22px; font-weight: 700; padding-top: 8px; }

/* 窄屏（手机 / 小窗口）：封面在上，歌词在下 */
@media (max-width: 760px) {
  .stage { --cover: min(72vw, 340px); padding: 56px 24px 24px; align-items: flex-start; }
  .row { flex-direction: column; align-items: center; gap: 20px; }
  .right { width: 100%; height: 30vh; margin-top: 0; }
}
</style>
