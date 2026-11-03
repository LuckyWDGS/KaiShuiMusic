<script setup lang="ts">
// 我喜欢的音乐 / 历史播放（共用一个页面，按路由 meta 区分）
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import PlaylistHead from '@/components/PlaylistHead.vue'
import SongList from '@/components/SongList.vue'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import type { Song } from '@/types'
import type { MenuItem } from '@/stores/menu'

const route = useRoute()
const player = usePlayer()
const lib = useLibrary()

const isLikes = route.meta.list === 'likes'
const title = isLikes ? '我喜欢的音乐' : '历史播放'
const songs = computed(() => (isLikes ? lib.data.likes : lib.data.history))
const menu = computed<MenuItem[]>(() => (isLikes || !songs.value.length ? [] : [{ label: '清空历史播放', danger: true, action: () => lib.clearHistory() }]))

function remove(s: Song) {
  if (isLikes) lib.toggleLike(s)
}
</script>

<template>
  <div class="page">
    <PlaylistHead :title="title" :cover="songs[0]?.cover" :source="songs[0]?.source" :count="songs.length" :menu="menu" @play="player.playList(songs, 0, 'list', title)" />
    <div v-if="!songs.length" class="empty">
      {{ isLikes ? '还没有喜欢的歌曲，听到喜欢的歌点一下 ♥ 吧' : '还没有播放记录' }}
    </div>
    <SongList v-else-if="isLikes" :songs="songs" :source="title" show-source @remove="remove" />
    <SongList v-else :songs="songs" :source="title" show-source />
  </div>
</template>
