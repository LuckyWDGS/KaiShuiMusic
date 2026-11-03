<script setup lang="ts">
// 平台歌单 / 专辑 / 榜单详情
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Loader2 } from 'lucide-vue-next'
import PlaylistHead from '@/components/PlaylistHead.vue'
import SongList from '@/components/SongList.vue'
import { api } from '@/api'
import type { DetailKind, Playlist, Song } from '@/types'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import type { MenuItem } from '@/stores/menu'
import { sourceName } from '@/utils/format'

const route = useRoute()
const player = usePlayer()
const lib = useLibrary()

const kind = route.params.kind as DetailKind
const source = route.params.source as string
const id = route.params.id as string
const q = (k: string) => (typeof route.query[k] === 'string' ? (route.query[k] as string) : '')

const songs = ref<Song[]>([])
const loading = ref(true)
const error = ref('')

const info = computed<Playlist>(() => ({
  id,
  source,
  name: q('name') || '歌单',
  cover: q('cover') || songs.value[0]?.cover || '',
  description: q('desc'),
  track_count: songs.value.length,
  play_count: 0,
}))
const kindLabel = { playlist: '歌单', album: '专辑', toplist: '榜单' }[kind] || '歌单'
const menu = computed<MenuItem[]>(() => [
  { label: lib.isSaved(kind, info.value) ? '取消收藏' : '收藏歌单', action: () => lib.toggleSave(kind, info.value) },
])

async function load() {
  loading.value = true
  error.value = ''
  try {
    songs.value = await api.detail(kind, source, id)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>

<template>
  <div class="page">
    <PlaylistHead
      :title="info.name"
      :cover="info.cover"
      :source="source"
      :sub="`${kindLabel} · ${sourceName(source)}`"
      :desc="info.description"
      :count="songs.length"
      :menu="menu"
      @play="player.playList(songs, 0, 'list', info.name)"
    />
    <div v-if="loading" class="empty"><Loader2 :size="26" class="spin" /></div>
    <div v-else-if="error" class="empty">
      <p>{{ error }}</p>
      <button class="btn" @click="load">重试</button>
    </div>
    <div v-else-if="!songs.length" class="empty">这里还没有歌曲</div>
    <SongList v-else :songs="songs" :source="info.name" />
  </div>
</template>

<style scoped>
.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
