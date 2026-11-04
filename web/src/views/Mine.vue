<script setup lang="ts">
// 我的（手机版）：PC 版放在侧边栏的内容集中到这一页
import { Heart, Clock3, Download, ListMusic, Star, Plus, Settings, ChevronRight } from 'lucide-vue-next'
import SvipBanner from '@/components/SvipBanner.vue'
import { useLibrary } from '@/stores/library'

const lib = useLibrary()
const entries = [
  { to: '/mine/likes', label: '我喜欢的音乐', icon: Heart },
  { to: '/mine/history', label: '历史播放', icon: Clock3 },
  { to: '/mine/downloads', label: '本地下载', icon: Download },
  { to: '/settings', label: '设置', icon: Settings },
]
</script>

<template>
  <div class="page mine">
    <div class="brand"><img src="/wordmark.png" alt="开水音乐" /></div>
    <SvipBanner />

    <div class="group">
      <RouterLink v-for="e in entries" :key="e.to" :to="e.to" class="item">
        <component :is="e.icon" :size="20" /><span class="ellipsis">{{ e.label }}</span><ChevronRight :size="18" class="go" />
      </RouterLink>
    </div>

    <div class="title">
      创建的歌单
      <button aria-label="新建歌单" @click="lib.creating = true"><Plus :size="20" /></button>
    </div>
    <div v-if="lib.data.created.length" class="group">
      <RouterLink v-for="p in lib.data.created" :key="p.id" :to="`/mine/playlist/${p.id}`" class="item">
        <ListMusic :size="20" /><span class="ellipsis">{{ p.name }}</span><ChevronRight :size="18" class="go" />
      </RouterLink>
    </div>

    <template v-if="lib.data.playlists.length">
      <div class="title">收藏的歌单</div>
      <div class="group">
        <RouterLink
          v-for="p in lib.data.playlists"
          :key="p.kind + p.source + p.id"
          :to="{ name: 'detail', params: { kind: p.kind, source: p.source, id: p.id }, query: { name: p.name, cover: p.cover, desc: p.description } }"
          class="item"
        >
          <Star :size="20" /><span class="ellipsis">{{ p.name }}</span><ChevronRight :size="18" class="go" />
        </RouterLink>
      </div>
    </template>
  </div>
</template>

<style scoped>
.brand { padding: 6px 2px 14px; }
.brand img { height: 24px; width: auto; }
.group { border-radius: var(--radius); background: var(--bg-elevated); overflow: hidden; }
.item { display: flex; align-items: center; gap: 12px; height: 52px; padding: 0 14px; font-size: 15px; font-weight: 600; }
.item:active { background: var(--bg-hover); }
.item span { flex: 1; min-width: 0; }
.go { color: var(--text-3); flex-shrink: 0; }
.title { display: flex; align-items: center; justify-content: space-between; margin: 22px 2px 10px; font-size: 14px; font-weight: 700; color: var(--text-2); }
.title button { display: flex; color: var(--text-2); padding: 4px; }
</style>
