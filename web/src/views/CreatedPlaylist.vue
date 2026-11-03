<script setup lang="ts">
// 自建歌单
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import PlaylistHead from '@/components/PlaylistHead.vue'
import SongList from '@/components/SongList.vue'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import type { Song } from '@/types'
import type { MenuItem } from '@/stores/menu'

const route = useRoute()
const router = useRouter()
const player = usePlayer()
const lib = useLibrary()

const id = route.params.id as string
const playlist = computed(() => lib.data.created.find((p) => p.id === id))

async function remove() {
  if (!playlist.value || !confirm(`确定删除歌单“${playlist.value.name}”吗？`)) return
  if (await lib.deletePlaylist(id)) router.replace('/')
}
const menu: MenuItem[] = [{ label: '删除歌单', danger: true, action: remove }]
const removeSong = (s: Song) => lib.removeFromPlaylist(id, s)
</script>

<template>
  <div v-if="playlist" class="page">
    <PlaylistHead
      :title="playlist.name"
      :cover="playlist.songs[0]?.cover"
      :source="playlist.songs[0]?.source"
      sub="我创建的歌单"
      :count="playlist.songs.length"
      :menu="menu"
      @play="player.playList(playlist.songs, 0, 'list', playlist.name)"
    />
    <div v-if="!playlist.songs.length" class="empty">歌单还是空的，播放时点击播放条上的“收藏”即可添加</div>
    <SongList v-else :songs="playlist.songs" :source="playlist.name" show-source @remove="removeSong" />
  </div>
  <div v-else class="empty">歌单不存在</div>
</template>
