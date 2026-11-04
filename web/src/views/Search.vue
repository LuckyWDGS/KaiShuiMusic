<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Play, Loader2 } from 'lucide-vue-next'
import SongList from '@/components/SongList.vue'
import PlaylistCard from '@/components/PlaylistCard.vue'
import { api } from '@/api'
import type { Playlist, SearchType, Song } from '@/types'
import { usePlayer } from '@/stores/player'

const route = useRoute()
const router = useRouter()
const player = usePlayer()

const keyword = typeof route.query.q === 'string' ? route.query.q : ''
const type = (['song', 'playlist', 'album'].includes(route.query.type as string) ? route.query.type : 'song') as SearchType
const TYPES: { id: SearchType; label: string }[] = [
  { id: 'song', label: '单曲' },
  { id: 'playlist', label: '歌单' },
  { id: 'album', label: '专辑' },
]

const songs = ref<Song[]>([])
const playlists = ref<Playlist[]>([])
const loading = ref(false)
const error = ref('')
const aborter = new AbortController()

// 同一首歌只保留最先返回的平台的结果
const seen = new Set<string>()
const fresh = (s: Song) => {
  const key = `${s.name}|${s.artist}`.toLowerCase().replace(/\s+/g, '')
  if (seen.has(key)) return false
  seen.add(key)
  return true
}

async function search() {
  if (!keyword) return
  loading.value = true
  error.value = ''
  try {
    // 各平台同时搜索，先返回的先显示，后到的结果追加在后面，已显示的内容位置不变
    await api.search(keyword, type, (c) => {
      songs.value.push(...(c.songs || []).filter(fresh))
      playlists.value.push(...(c.playlists || []))
    }, aborter.signal)
  } catch (e) {
    if (!aborter.signal.aborted) error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function switchType(t: SearchType) {
  router.replace({ name: 'search', query: { q: keyword, type: t } })
}
onMounted(search)
onBeforeUnmount(() => aborter.abort())
</script>

<template>
  <div class="page">
    <h1 class="title">搜索 <span class="kw">{{ keyword }}</span></h1>
    <div class="bar">
      <div class="text-tabs">
        <button v-for="t in TYPES" :key="t.id" :class="{ active: type === t.id }" @click="switchType(t.id)">{{ t.label }}</button>
      </div>
      <button v-if="type === 'song' && songs.length" class="play-all" @click="player.playList(songs, 0, 'list', `搜索：${keyword}`)">
        <Play :size="14" fill="currentColor" />播放全部
      </button>
    </div>

    <div v-if="loading && !songs.length && !playlists.length" class="empty"><Loader2 :size="26" class="spin" /><p>正在全网搜索…</p></div>
    <div v-else-if="error && !songs.length && !playlists.length" class="empty">{{ error }}</div>
    <template v-else-if="type === 'song'">
      <div v-if="!songs.length" class="empty">没有找到相关歌曲</div>
      <div v-else class="card">
        <SongList :songs="songs" :source="`搜索：${keyword}`" />
      </div>
    </template>
    <template v-else>
      <div v-if="!playlists.length" class="empty">没有找到相关{{ type === 'album' ? '专辑' : '歌单' }}</div>
      <div v-else class="grid-cards">
        <PlaylistCard v-for="p in playlists" :key="p.source + p.id" :item="p" :kind="type === 'album' ? 'album' : 'playlist'" />
      </div>
    </template>
    <div v-if="loading && (songs.length || playlists.length)" class="more"><Loader2 :size="16" class="spin" />正在加载更多结果…</div>
  </div>
</template>

<style scoped>
.title { font-size: 22px; font-weight: 700; margin: 4px 0 18px; }
.kw { color: #20c25e; }
.bar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.text-tabs { display: flex; gap: 32px; }
.text-tabs button { position: relative; font-size: 18px; font-weight: 700; color: var(--text-2); padding-bottom: 8px; }
.text-tabs button:hover { color: var(--text); }
.text-tabs button.active { color: var(--text); }
.text-tabs button.active::after {
  content: ''; position: absolute; left: 50%; bottom: 0; width: 12px; height: 3px; margin-left: -6px; border-radius: 2px; background: var(--text);
}
.play-all {
  display: inline-flex; align-items: center; gap: 6px; height: 30px; padding: 0 12px; border-radius: 6px;
  background: #e9e9e9; color: #121212; font-size: 13px; font-weight: 700;
}
.card { background: #1b1b1b; border-radius: 10px; padding: 12px 0 8px; }
.card :deep(.song-list) { margin: 0; }
.card :deep(.row) { padding: 0 16px 0 24px; }
.more { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 16px 0; font-size: 13px; color: var(--text-3); }
.spin { animation: spin 1s linear infinite; color: var(--accent); }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
