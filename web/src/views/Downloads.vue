<script setup lang="ts">
import { computed, onActivated, ref } from 'vue'
import PlaylistHead from '@/components/PlaylistHead.vue'
import SongList from '@/components/SongList.vue'
import { api, LOCAL_FILE } from '@/api'
import type { Song } from '@/types'
import { usePlayer } from '@/stores/player'
import type { MenuItem } from '@/stores/menu'
import { platform } from '@/platform'

const player = usePlayer()
const menu = computed<MenuItem[]>(() => [
  ...(platform.value === 'pc' ? [{ label: '打开文件夹', action: () => api.openDownloads() }] : []),
  { label: '刷新', action: load },
])
const dir = ref('')
const files = ref<{ name: string; size: number; mtime: number }[]>([])
const loaded = ref(false)

// 下载文件名模板为 "歌手 - 歌名"，据此还原歌曲信息
const songs = computed<Song[]>(() =>
  files.value.map((f) => {
    const base = f.name.split('/').pop()!.replace(/\.[^.]+$/, '')
    const [artist, ...rest] = base.split(' - ')
    return {
      id: f.name,
      source: LOCAL_FILE,
      name: rest.length ? rest.join(' - ') : base,
      artist: rest.length ? artist : '本地音乐',
      album: '',
      duration: 0,
      cover: '',
    }
  }),
)

async function load() {
  try {
    const r = await api.downloads()
    dir.value = r.dir
    files.value = r.files
  } finally {
    loaded.value = true
  }
}
onActivated(load)
</script>

<template>
  <div class="page">
    <PlaylistHead title="本地下载" :sub="dir" :count="songs.length" :menu="menu" @play="player.playList(songs, 0, 'list', '本地下载')" />
    <div v-if="loaded && !songs.length" class="empty">还没有下载的歌曲，在歌曲的 ⋯ 菜单里点击“下载”即可保存到本地</div>
    <SongList v-else :songs="songs" :show-album="false" source="本地下载" />
  </div>
</template>
