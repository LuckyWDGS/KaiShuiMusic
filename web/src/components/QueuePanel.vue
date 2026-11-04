<script setup lang="ts">
// 播放列表面板（参照汽水音乐）：右侧整高抽屉，标题下方显示来源；当前歌曲封面叠加跳动图标，悬停出现删除按钮
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ListEnd, Trash2, AudioLines } from 'lucide-vue-next'
import CoverImg from './CoverImg.vue'
import { usePlayer } from '@/stores/player'
import { songKey } from '@/utils/format'
import { platform } from '@/platform'

const player = usePlayer()
const panel = ref<HTMLElement>()
const list = ref<HTMLElement>()

watch(() => player.showQueue, (v) => {
  if (v) nextTick(() => list.value?.querySelector('.current')?.scrollIntoView({ block: 'center' }))
})

// 点击面板外部收起
// PC 双击播放，手机点一下播放
const tapPlay = (i: number) => platform.value === 'mobile' && player.load(i)

const onDown = (e: MouseEvent) => {
  const t = e.target as HTMLElement
  if (!player.showQueue || panel.value?.contains(t) || t.closest('.menu, [title="播放列表"]')) return
  player.showQueue = false
}
onMounted(() => window.addEventListener('mousedown', onDown))
onBeforeUnmount(() => window.removeEventListener('mousedown', onDown))
</script>

<template>
  <Transition name="slide">
    <aside v-if="player.showQueue" ref="panel" class="queue" @keydown.esc="player.showQueue = false">
      <div class="top">
        <button class="icon-btn" title="收起" @click="player.showQueue = false"><ListEnd :size="20" /></button>
        <button v-if="player.queue.length" class="clear" title="清空播放列表" @click="player.clearQueue()"><Trash2 :size="16" /></button>
      </div>
      <header>
        <div class="title">播放列表</div>
        <div class="from ellipsis">{{ player.queueSource || `共 ${player.queue.length} 首` }}</div>
      </header>
      <div ref="list" class="list">
        <div
          v-for="(s, i) in player.queue"
          :key="songKey(s) + i"
          class="item"
          :class="{ current: i === player.index }"
          @click="tapPlay(i)"
          @dblclick="player.load(i)"
        >
          <div class="cover">
            <CoverImg :src="s.cover" :size="100" :source="s.source" radius="4px" />
            <span v-if="i === player.index" class="eq"><AudioLines :size="18" /></span>
          </div>
          <div class="meta">
            <div class="name ellipsis">{{ s.name }}</div>
            <div class="artist ellipsis">{{ s.artist }}</div>
          </div>
          <button class="rm" title="从播放列表删除" @click.stop="player.removeAt(i)"><Trash2 :size="16" /></button>
        </div>
        <div v-if="!player.queue.length" class="empty">播放列表是空的</div>
      </div>
    </aside>
  </Transition>
</template>

<style scoped>
.queue {
  position: absolute; right: 0; top: 0; bottom: 0; width: 420px; z-index: 40;
  background: var(--surface); display: flex; flex-direction: column; box-shadow: -12px 0 40px rgba(0, 0, 0, .35);
}
.top { display: flex; align-items: center; justify-content: space-between; padding: 12px 16px 0 12px; }
.clear { color: var(--text-3); display: flex; padding: 6px; border-radius: 6px; }
.clear:hover { color: var(--text); background: rgba(255, 255, 255, .08); }
header { padding: 14px 20px 10px; }
.title { font-size: 18px; font-weight: 700; }
.from { font-size: 13px; font-weight: 600; color: var(--text-2); margin-top: 4px; }
.list { flex: 1; overflow-y: auto; padding: 0 0 12px; }
.item { display: flex; align-items: center; gap: 16px; height: 64px; padding: 0 12px 0 20px; }
.item:hover { background: rgba(255, 255, 255, .06); }
.cover { position: relative; width: 44px; flex-shrink: 0; }
.eq { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(0, 0, 0, .45); border-radius: 4px; }
.meta { flex: 1; min-width: 0; }
.name { font-size: 15px; font-weight: 700; }
.artist { font-size: 13px; font-weight: 600; color: var(--text-2); margin-top: 3px; }
.rm { display: none; width: 32px; height: 32px; border-radius: 6px; align-items: center; justify-content: center; background: rgba(255, 255, 255, .08); color: var(--text-2); }
.rm:hover { color: var(--text); }
.item:hover .rm { display: flex; }
.slide-enter-active, .slide-leave-active { transition: transform .25s ease; }
.slide-enter-from, .slide-leave-to { transform: translateX(100%); }
/* 手机：从底部弹出，删除按钮常驻 */
@media (max-width: 760px) {
  .queue { top: 28%; width: 100%; z-index: 50; border-radius: 16px 16px 0 0; padding-bottom: var(--safe-bottom); box-shadow: 0 -12px 40px rgba(0, 0, 0, .4); }
  .rm { display: flex; background: none; }
  .slide-enter-from, .slide-leave-to { transform: translateY(100%); }
}
</style>
