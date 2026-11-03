<script setup lang="ts">
// 听歌模式：大封面 + 大字歌词，专注听歌
import { onBeforeUnmount, onMounted } from 'vue'
import SongStage from './SongStage.vue'
import { usePlayer } from '@/stores/player'

const player = usePlayer()

const onKey = (e: KeyboardEvent) => e.key === 'Escape' && (player.showNowPlaying = false)
onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div class="now-playing">
    <SongStage v-if="player.current" :song="player.current" large />
    <div v-else class="empty">还没有正在播放的歌曲</div>
  </div>
</template>

<style scoped>
.now-playing { height: 100%; }
</style>
