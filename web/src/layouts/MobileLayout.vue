<script setup lang="ts">
// 手机 / 电视布局（预留，基础可用）：顶部搜索 + 内容 + 迷你播放条 + 底部导航。
// 页面、播放器、音乐库与 PC 版完全共用，后续在此基础上细化触屏手势与遥控器交互。
import { Play as PlayTab, Heart, Play, Pause, SkipForward, ChevronDown } from 'lucide-vue-next'
import SearchBox from '@/components/SearchBox.vue'
import CoverImg from '@/components/CoverImg.vue'
import SongStage from '@/components/SongStage.vue'
import QueuePanel from '@/components/QueuePanel.vue'
import { usePlayer } from '@/stores/player'

const player = usePlayer()
const tabs = [
  { to: '/', label: '推荐', icon: PlayTab },
  { to: '/mine/likes', label: '我的', icon: Heart },
]
</script>

<template>
  <div class="mobile">
    <header class="top"><SearchBox class="search" /></header>
    <main class="view">
      <RouterView v-slot="{ Component, route }">
        <KeepAlive :max="6"><component :is="Component" :key="route.fullPath" /></KeepAlive>
      </RouterView>
    </main>

    <div v-if="player.current" class="mini" @click="player.showNowPlaying = true">
      <CoverImg :src="player.current.cover" :size="100" :source="player.current.source" />
      <div class="meta">
        <div class="ellipsis">{{ player.current.name }}</div>
        <div class="ellipsis muted">{{ player.current.artist }}</div>
      </div>
      <button class="icon-btn" @click.stop="player.toggle()">
        <Pause v-if="player.playing" :size="22" fill="currentColor" /><Play v-else :size="22" fill="currentColor" />
      </button>
      <button class="icon-btn" @click.stop="player.next()"><SkipForward :size="20" fill="currentColor" /></button>
      <div class="mini-progress" :style="{ width: (player.duration ? (player.currentTime / player.duration) * 100 : 0) + '%' }"></div>
    </div>

    <nav class="tabbar">
      <RouterLink v-for="t in tabs" :key="t.to" :to="t.to" class="tab-item" exact-active-class="active">
        <component :is="t.icon" :size="22" /><span>{{ t.label }}</span>
      </RouterLink>
    </nav>

    <Transition name="rise">
      <div v-if="player.showNowPlaying && player.current" class="full">
        <button class="icon-btn close" @click="player.showNowPlaying = false"><ChevronDown :size="28" /></button>
        <SongStage :song="player.current" />
      </div>
    </Transition>
    <QueuePanel />
  </div>
</template>

<style scoped>
.mobile { height: 100%; display: flex; flex-direction: column; position: relative; --playerbar-h: 0px; }
.top { padding: 10px 14px; padding-top: max(10px, env(safe-area-inset-top)); display: flex; justify-content: center; }
.search { width: 100% !important; }
.view { flex: 1; min-height: 0; overflow-y: auto; }
.view :deep(.page) { padding: 4px 14px 24px; }
.view :deep(.grid-cards) { grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 16px 12px; }
.mini {
  position: relative; display: flex; align-items: center; gap: 10px; margin: 0 8px 6px; padding: 8px 10px;
  border-radius: 12px; background: rgba(255, 255, 255, .12); overflow: hidden;
}
.mini > .cover { width: 42px; }
.meta { flex: 1; min-width: 0; font-size: 14px; }
.meta .muted { font-size: 12px; margin-top: 2px; }
.mini-progress { position: absolute; left: 0; bottom: 0; height: 2px; background: #fff; }
.tabbar { display: flex; padding-bottom: env(safe-area-inset-bottom); background: var(--bg-sidebar); }
.tab-item { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 2px; padding: 8px 0; font-size: 11px; color: var(--text-3); }
.tab-item.active { color: var(--text); }
.full { position: absolute; inset: 0; z-index: 40; background: linear-gradient(180deg, var(--theme-top), var(--theme-bottom)); }
.close { position: absolute; top: 12px; left: 12px; z-index: 2; color: #fff; }
.rise-enter-active, .rise-leave-active { transition: transform .3s ease; }
.rise-enter-from, .rise-leave-to { transform: translateY(100%); }
</style>
