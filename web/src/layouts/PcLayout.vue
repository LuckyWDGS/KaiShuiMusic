<script setup lang="ts">
// PC 布局（参照汽水音乐 PC 版）：左侧导航 + 顶部居中搜索 + 底部通栏播放条，整窗跟随封面取色
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Play, Compass, Heart, Clock3, Download, ListMusic, Plus, Settings, ChevronLeft, ChevronRight, Star, Minus, Square, X,
  PictureInPicture2, Crown,
} from 'lucide-vue-next'
import { toast } from '@/stores/toast'
import { isDesktopShell, onTitleMouseDown, winAction } from '@/utils/desktop'
import { miniOn, toggleMini } from '@/utils/miniPlayer'
import SearchBox from '@/components/SearchBox.vue'
import PlayerBar from '@/components/PlayerBar.vue'
import QueuePanel from '@/components/QueuePanel.vue'
import NowPlaying from '@/components/NowPlaying.vue'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'

const router = useRouter()
const route = useRoute()
const player = usePlayer()
const lib = useLibrary()
const desktop = isDesktopShell()
// 推荐页与听歌模式整窗使用封面色，其它页面深色底
const immersive = computed(() => route.name === 'foryou' || player.showNowPlaying)

const mine = [
  { to: '/mine/likes', label: '我喜欢的音乐', icon: Heart },
  { to: '/mine/history', label: '历史播放', icon: Clock3 },
  { to: '/mine/downloads', label: '本地下载', icon: Download },
]
</script>

<template>
  <div class="pc" :class="{ immersive }">
    <aside class="sidebar">
      <div class="brand" @mousedown="onTitleMouseDown">
        <img src="/wordmark.png" alt="开水音乐" draggable="false" />
      </div>
      <nav>
        <RouterLink to="/" class="nav" exact-active-class="active" @click="player.showNowPlaying = false">
          <Play :size="16" fill="currentColor" />推荐
        </RouterLink>
        <RouterLink to="/modes" class="nav" active-class="active" @click="player.showNowPlaying = false">
          <Compass :size="17" />听歌模式
        </RouterLink>

        <div class="group">我的音乐</div>
        <RouterLink v-for="n in mine" :key="n.to" :to="n.to" class="nav" active-class="active" @click="player.showNowPlaying = false">
          <component :is="n.icon" :size="17" />{{ n.label }}
        </RouterLink>

        <div class="group">
          创建的歌单
          <button class="add" title="新建歌单" @click="lib.creating = true"><Plus :size="16" /></button>
        </div>
        <RouterLink
          v-for="p in lib.data.created"
          :key="p.id"
          :to="`/mine/playlist/${p.id}`"
          class="nav"
          active-class="active"
          @click="player.showNowPlaying = false"
        >
          <ListMusic :size="17" /><span class="ellipsis">{{ p.name }}</span>
        </RouterLink>

        <template v-if="lib.data.playlists.length">
          <div class="group">收藏的歌单</div>
          <RouterLink
            v-for="p in lib.data.playlists"
            :key="p.kind + p.source + p.id"
            :to="{ name: 'detail', params: { kind: p.kind, source: p.source, id: p.id }, query: { name: p.name, cover: p.cover, desc: p.description } }"
            class="nav"
            active-class="active"
            @click="player.showNowPlaying = false"
          >
            <Star :size="17" /><span class="ellipsis">{{ p.name }}</span>
          </RouterLink>
        </template>
      </nav>
    </aside>

    <section class="main">
      <header class="topbar" @mousedown="onTitleMouseDown">
        <div class="history-nav">
          <button class="icon-btn" title="后退" @click="router.back()"><ChevronLeft :size="20" /></button>
          <button class="icon-btn" title="前进" @click="router.forward()"><ChevronRight :size="20" /></button>
        </div>
        <SearchBox class="search" />
        <div class="top-right">
          <button class="svip-pill" title="开水音乐全部歌曲免费畅听" @click="toast('开水音乐全部歌曲免费畅听，无需开通会员')">
            <Crown :size="13" />SVIP 畅听中
          </button>
          <RouterLink to="/settings" class="icon-btn" title="设置" @click="player.showNowPlaying = false"><Settings :size="18" /></RouterLink>
          <template v-if="desktop">
            <button class="icon-btn mini-toggle" :class="{ on: miniOn }" :title="miniOn ? '关闭任务栏播放器' : '开启任务栏播放器'" @click="toggleMini">
              <PictureInPicture2 :size="18" />
            </button>
            <span class="sep"></span>
            <button class="icon-btn win" title="最小化" @click="winAction('min')"><Minus :size="18" /></button>
            <button class="icon-btn win" title="最大化" @click="winAction('max')"><Square :size="14" /></button>
            <button class="icon-btn win close" title="关闭" @click="winAction('close')"><X :size="19" /></button>
          </template>
        </div>
      </header>
      <!-- 听歌模式占据内容区，页面本身保持挂载以保留滚动位置 -->
      <main v-show="!player.showNowPlaying" class="view">
        <RouterView v-slot="{ Component, route }">
          <KeepAlive :max="8">
            <component :is="Component" :key="route.fullPath" />
          </KeepAlive>
        </RouterView>
      </main>
      <Transition name="rise">
        <NowPlaying v-if="player.showNowPlaying" class="view" />
      </Transition>
    </section>

    <QueuePanel />
    <PlayerBar class="bar" />
  </div>
</template>

<style scoped>
.pc {
  position: relative; height: 100%; display: grid; grid-template-columns: var(--sidebar-w) 1fr;
  grid-template-rows: 1fr var(--playerbar-h); overflow: hidden; background: var(--bg);
}
.pc.immersive { background: linear-gradient(180deg, var(--theme-top), var(--theme-bottom)); }
.sidebar { grid-row: 1; background: var(--bg-sidebar); display: flex; flex-direction: column; padding: 0 16px 12px; min-height: 0; }
.brand { display: flex; align-items: center; height: 72px; padding: 16px 10px 0; flex-shrink: 0; }
.brand img { height: 23px; width: auto; pointer-events: none; }
nav { flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: 4px; padding-top: 22px; scrollbar-width: none; }
nav::-webkit-scrollbar { display: none; }
.group {
  display: flex; align-items: center; justify-content: space-between; font-size: 13px; color: var(--text-3);
  padding: 22px 10px 8px; flex-shrink: 0;
}
.add { color: var(--text-3); display: flex; }
.add:hover { color: var(--text); }
.nav {
  display: flex; align-items: center; gap: 10px; height: 40px; padding: 0 10px; border-radius: 8px; width: 100%;
  color: var(--text-2); font-size: 14px; font-weight: 600; transition: background .15s, color .15s; flex-shrink: 0; text-align: left;
}
.nav:hover { color: var(--text); background: var(--bg-hover); }
.nav.active { color: var(--text); background: var(--bg-active); }
.nav :deep(svg) { flex-shrink: 0; }
.main { grid-row: 1; display: flex; flex-direction: column; min-width: 0; min-height: 0; }
.topbar { height: var(--topbar-h); display: grid; grid-template-columns: 1fr auto 1fr; align-items: center; padding: 0 24px; flex-shrink: 0; }
.history-nav { display: flex; gap: 2px; }
.history-nav .icon-btn { color: var(--text-3); }
.top-right { display: flex; align-items: center; justify-content: flex-end; gap: 4px; margin-right: -12px; }
.top-right .icon-btn { color: var(--text-2); width: 34px; height: 34px; border-radius: 8px; }
.top-right .icon-btn:hover { color: var(--text); }
.win.close:hover { background: #e81123; color: #fff; }
.svip-pill {
  display: inline-flex; align-items: center; gap: 4px; height: 26px; padding: 0 10px; margin-right: 6px; border-radius: 13px;
  font-size: 12px; font-weight: 700; color: #3a2a10; background: linear-gradient(90deg, #f6dfae, #e2b86b);
  white-space: nowrap; flex-shrink: 0;
}
.svip-pill:hover { filter: brightness(1.06); }
.mini-toggle.on { color: #121212 !important; background: rgba(255, 255, 255, .85); }
.sep { width: 1px; height: 18px; background: rgba(255, 255, 255, .2); margin: 0 6px; }
.view { flex: 1; min-height: 0; overflow-y: auto; position: relative; }
.bar { grid-column: 1 / -1; grid-row: 2; }
.rise-enter-active, .rise-leave-active { transition: transform .35s cubic-bezier(.2, .8, .2, 1), opacity .35s; }
.rise-enter-from, .rise-leave-to { transform: translateY(40px); opacity: 0; }
</style>
