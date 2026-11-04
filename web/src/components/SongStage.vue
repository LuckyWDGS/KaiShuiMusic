<script setup lang="ts">
// 歌曲舞台：左侧唱盘机 + 歌名歌手，右侧歌词。推荐页与歌词页共用。
// 播放控制统一使用底部播放条；还没开始播放的歌也会先加载并显示歌词。
import { computed, ref, watch } from 'vue'
import { Heart } from 'lucide-vue-next'
import Turntable from './Turntable.vue'
import LyricView from './LyricView.vue'
import { api } from '@/api'
import type { Song } from '@/types'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import { songKey } from '@/utils/format'
import { parseLyric, type LyricLine } from '@/utils/lyric'

const props = defineProps<{ song: Song; large?: boolean }>()
const player = usePlayer()
const lib = useLibrary()

const isCurrent = computed(() => !!player.current && songKey(props.song) === songKey(player.current))

// 未播放歌曲的歌词（预览）
const previewLines = ref<LyricLine[]>([])
const previewLoading = ref(false)
watch(() => [props.song, isCurrent.value] as const, async ([song, current]) => {
  if (current) return
  previewLoading.value = true
  previewLines.value = []
  try {
    const { lyric } = await api.lyric(song)
    if (songKey(song) === songKey(props.song)) previewLines.value = lyric ? parseLyric(lyric) : []
  } catch { /* 没有歌词 */ } finally {
    previewLoading.value = false
  }
}, { immediate: true })
</script>

<template>
  <div class="stage" :class="{ large }">
    <div class="row">
      <div class="left">
        <Turntable :cover="song.cover" :source="song.source" :playing="isCurrent && player.playing" />
        <div class="info">
          <div class="title">
            <span class="name ellipsis">{{ song.name }}</span>
            <button class="like" :class="{ liked: lib.isLiked(song) }" :title="lib.isLiked(song) ? '取消喜欢' : '喜欢'" @click="lib.toggleLike(song)">
              <Heart :size="22" :fill="lib.isLiked(song) ? 'currentColor' : 'none'" />
            </button>
          </div>
          <div class="artist">
            <span class="ellipsis">{{ song.artist }}</span>
            <span v-if="isCurrent && player.resolved?.preview" class="tag">试听</span>
          </div>
        </div>
      </div>
      <div class="right">
        <LyricView v-if="isCurrent" :anchor="0.38" :large="large" />
        <LyricView v-else :lines="previewLines" :loading="previewLoading" :anchor="0.38" :large="large" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.stage {
  --cover: clamp(240px, calc(62vh - 70px), 480px);
  height: 100%; display: flex; align-items: center; justify-content: center; padding: 0 48px 24px;
}
.stage.large { --cover: clamp(260px, calc(64vh - 90px), 520px); }
.row { display: flex; align-items: center; gap: 64px; width: 100%; max-width: 1100px; justify-content: center; }
.left { width: var(--cover); flex-shrink: 0; display: flex; flex-direction: column; align-items: center; }
/* 唱片偏左放置，信息与唱片中心对齐 */
.info { width: 88%; margin: 4px 0 0 -8%; text-align: center; min-width: 0; }
.title { display: flex; align-items: center; justify-content: center; gap: 8px; min-width: 0; }
.name { font-size: 20px; font-weight: 700; line-height: 1.3; }
.like { display: flex; flex-shrink: 0; color: rgba(255, 255, 255, .85); transition: transform .15s, color .15s; }
.like:hover { transform: scale(1.12); color: #fff; }
.like.liked { color: #fe2c55; }
.artist {
  display: flex; align-items: center; justify-content: center; gap: 8px; margin-top: 6px; min-width: 0;
  font-size: 13px; font-weight: 600; color: var(--text-2);
}
.tag { flex-shrink: 0; font-size: 11px; padding: 1px 7px; border-radius: 5px; background: rgba(255, 255, 255, .16); color: var(--text); }
/* 歌词与唱片垂直居中对齐 */
.right { width: min(420px, 44%); height: calc(var(--cover) * 1.1); }
.large .right { width: min(520px, 46%); }

/* 手机：唱片在上，歌词占满剩余高度 */
@media (max-width: 760px) {
  .stage { --cover: min(72vw, 36vh); padding: 4px 24px 0; align-items: stretch; }
  .row { flex-direction: column; gap: 10px; height: 100%; justify-content: flex-start; }
  .info { width: 100%; margin: 2px 0 0; }
  .right { width: 100%; height: auto; flex: 1; min-height: 0; }
  .right :deep(.lyric) { text-align: center; }
  .right :deep(.pad-top) { height: 8%; }
  .right :deep(.tip) { padding-top: 8%; }
  .right :deep(.line) { font-size: 16px; line-height: 24px; padding: 6px 0; }
  .right :deep(.tip) { font-size: 16px; line-height: 24px; }
}
</style>
