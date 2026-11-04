<script setup lang="ts">
// 手机布局（电视暂时共用）：推荐页整屏唱盘 + 上下滑切歌，其它页面底部迷你播放条，全屏播放页从迷你条打开。
// 页面、播放器、音乐库与 PC 版完全共用。
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Play as PlayTab, Compass, User, Play, Pause, SkipForward, ChevronDown, ChevronLeft } from 'lucide-vue-next'
import SearchBox from '@/components/SearchBox.vue'
import CoverImg from '@/components/CoverImg.vue'
import SongStage from '@/components/SongStage.vue'
import QueuePanel from '@/components/QueuePanel.vue'
import MobileControls from '@/components/MobileControls.vue'
import { usePlayer } from '@/stores/player'
import { TAB_ROUTES } from '@/utils/appShell'

const route = useRoute()
const router = useRouter()
const player = usePlayer()

const tabs = [
  { to: '/', label: '推荐', icon: PlayTab },
  { to: '/modes', label: '听歌模式', icon: Compass },
  { to: '/mine', label: '我的', icon: User },
]
const onTab = computed(() => TAB_ROUTES.includes(String(route.name)))
const onStage = computed(() => route.name === 'foryou')
</script>

<template>
  <div class="mobile" :class="{ immersive: onStage }">
    <header class="top">
      <button v-if="!onTab" class="icon-btn back" aria-label="返回" @click="router.back()"><ChevronLeft :size="24" /></button>
      <SearchBox class="search" />
    </header>
    <main class="view" :class="{ stage: onStage }">
      <RouterView v-slot="{ Component, route: r }">
        <KeepAlive :max="6"><component :is="Component" :key="r.fullPath" /></KeepAlive>
      </RouterView>
    </main>

    <MobileControls v-if="onStage" />
    <div v-else-if="player.current" class="mini" @click="player.showNowPlaying = true">
      <CoverImg :src="player.current.cover" :size="100" :source="player.current.source" radius="6px" />
      <div class="meta">
        <div class="ellipsis">{{ player.current.name }}</div>
        <div class="ellipsis muted">{{ player.current.artist }}</div>
      </div>
      <button class="icon-btn" aria-label="播放或暂停" @click.stop="player.toggle()">
        <Pause v-if="player.playing" :size="22" fill="currentColor" /><Play v-else :size="22" fill="currentColor" />
      </button>
      <button class="icon-btn" aria-label="下一首" @click.stop="player.next()"><SkipForward :size="20" fill="currentColor" /></button>
      <div class="mini-progress" :style="{ width: (player.duration ? (player.currentTime / player.duration) * 100 : 0) + '%' }"></div>
    </div>

    <nav v-if="onTab" class="tabbar">
      <RouterLink v-for="t in tabs" :key="t.to" :to="t.to" class="tab-item" exact-active-class="active">
        <component :is="t.icon" :size="22" /><span>{{ t.label }}</span>
      </RouterLink>
    </nav>
    <div v-else class="safe-bottom"></div>

    <Transition name="rise">
      <div v-if="player.showNowPlaying && player.current" class="full immersive">
        <button class="icon-btn close" aria-label="收起" @click="player.showNowPlaying = false"><ChevronDown :size="28" /></button>
        <SongStage class="stage-body" :song="player.current" />
        <MobileControls />
        <div class="safe-bottom"></div>
      </div>
    </Transition>
    <QueuePanel />
  </div>
</template>

<style scoped>
.mobile { height: 100%; display: flex; flex-direction: column; position: relative; overflow: hidden; --playerbar-h: 0px; background: var(--bg); }
.mobile.immersive { background: linear-gradient(180deg, var(--theme-top), var(--theme-bottom)); }
.top { display: flex; align-items: center; gap: 4px; padding: calc(var(--safe-top) + 8px) 14px 8px; flex-shrink: 0; }
.back { margin-left: -8px; flex-shrink: 0; color: var(--text); }
.search { flex: 1; width: auto !important; height: 38px; }
.view { flex: 1; min-height: 0; overflow-y: auto; }
.view.stage { overflow: hidden; }
.view :deep(.page) { padding: 4px 14px 24px; }
.view :deep(.grid-cards) { grid-template-columns: repeat(auto-fill, minmax(140px, 1fr)); gap: 16px 12px; }
.mini {
  position: relative; display: flex; align-items: center; gap: 10px; margin: 0 8px 6px; padding: 8px 10px;
  border-radius: 12px; background: var(--surface); overflow: hidden; flex-shrink: 0;
}
.mini > .cover { width: 42px; flex-shrink: 0; }
.meta { flex: 1; min-width: 0; font-size: 14px; font-weight: 600; }
.meta .muted { font-size: 12px; margin-top: 2px; color: var(--text-2); }
.mini-progress { position: absolute; left: 0; bottom: 0; height: 2px; background: #fff; }
.tabbar { display: flex; padding-bottom: var(--safe-bottom); flex-shrink: 0; }
.mobile:not(.immersive) .tabbar { background: var(--bg-sidebar); }
.tab-item { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 2px; padding: 8px 0 6px; font-size: 11px; font-weight: 600; color: var(--text-3); }
.tab-item.active { color: var(--text); }
.safe-bottom { height: var(--safe-bottom); flex-shrink: 0; }
.full {
  position: absolute; inset: 0; z-index: 40; display: flex; flex-direction: column; padding-top: calc(var(--safe-top) + 48px);
  background: linear-gradient(180deg, var(--theme-top), var(--theme-bottom));
}
.close { position: absolute; top: calc(var(--safe-top) + 6px); left: 12px; z-index: 2; color: #fff; }
.stage-body { flex: 1; min-height: 0; }
.rise-enter-active, .rise-leave-active { transition: transform .3s ease; }
.rise-enter-from, .rise-leave-to { transform: translateY(100%); }
</style>
