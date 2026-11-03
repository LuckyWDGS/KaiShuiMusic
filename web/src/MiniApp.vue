<script setup lang="ts">
// 任务栏播放器（参照汽水音乐）：封面、歌名、当前歌词、上一首 / 播放 / 下一首、喜欢。
// 不自己播放，只通过 BroadcastChannel 显示主窗口的状态并发送控制指令。
import { onMounted, ref } from 'vue'
import { Play, Pause, SkipBack, SkipForward, Heart } from 'lucide-vue-next'
import { coverUrl } from '@/utils/format'
import { MINI_CHANNEL, type MiniMessage, type MiniState } from '@/utils/miniPlayer'

const state = ref<MiniState>({ song: null, playing: false, lyric: '', liked: false })
const ch = new BroadcastChannel(MINI_CHANNEL)
ch.onmessage = (e: MessageEvent<MiniMessage>) => {
  if (e.data.type === 'state') state.value = e.data.state
}
const cmd = (c: 'toggle' | 'next' | 'prev' | 'like') => ch.postMessage({ type: 'cmd', cmd: c } satisfies MiniMessage)

type MiniBridge = (a: 'show' | 'colors') => Promise<string[] | null>
const bridge = () => (window as unknown as { kaishuiMini?: MiniBridge }).kaishuiMini

// 桌面壳采样任务栏颜色后调用，让播放器和任务栏融为一体
function setColors(bg: string, fg: string) {
  const s = document.documentElement.style
  s.setProperty('--tb-bg', bg)
  s.setProperty('--tb-fg', fg)
}
;(window as unknown as { __kaishuiTaskbar: typeof setColors }).__kaishuiTaskbar = setColors

onMounted(async () => {
  ch.postMessage({ type: 'hello' } satisfies MiniMessage)
  const c = await bridge()?.('colors')
  if (c?.length === 2) setColors(c[0], c[1])
})
</script>

<template>
  <div class="mini" @contextmenu.prevent>
    <template v-if="state.song">
      <button class="info" title="打开开水音乐" @click="bridge()?.('show')">
        <img v-if="state.song.cover" class="cover" :src="coverUrl(state.song.cover, 100)" alt="" referrerpolicy="no-referrer" />
        <span class="text">
          <span class="name">{{ state.song.name }}<span class="artist"> - {{ state.song.artist }}</span></span>
          <span class="lyric">{{ state.lyric || '…' }}</span>
        </span>
      </button>
      <div class="ctrls">
        <button title="上一首" @click="cmd('prev')"><SkipBack :size="16" fill="currentColor" /></button>
        <button class="play" :title="state.playing ? '暂停' : '播放'" @click="cmd('toggle')">
          <Pause v-if="state.playing" :size="18" fill="currentColor" />
          <Play v-else :size="18" fill="currentColor" style="margin-left: 2px" />
        </button>
        <button title="下一首" @click="cmd('next')"><SkipForward :size="16" fill="currentColor" /></button>
        <button :class="{ liked: state.liked }" title="喜欢" @click="cmd('like')">
          <Heart :size="16" :fill="state.liked ? 'currentColor' : 'none'" />
        </button>
      </div>
    </template>
    <button v-else class="info idle" @click="bridge()?.('show')">
      <img class="cover" src="/logo.png" alt="" />
      <span class="text"><span class="name">开水音乐</span><span class="lyric">点击打开，开始听歌</span></span>
    </button>
  </div>
</template>

<style>
/* 背景透明，透出任务栏；--tb-bg 只用来判断文字深浅 */
:root { --tb-bg: #1c1c1c; --tb-fg: #ffffff; }
html, html body { background: transparent; color: var(--tb-fg); }
</style>

<style scoped>
.mini { height: 100%; display: flex; align-items: center; gap: 6px; padding: 0 6px 0 4px; }
.info { flex: 1; min-width: 0; display: flex; align-items: center; gap: 8px; height: 80%; padding: 0 4px; border-radius: 6px; text-align: left; }
.info:hover { background: color-mix(in srgb, var(--tb-fg) 8%, transparent); }
.cover { width: 30px; height: 30px; border-radius: 4px; object-fit: cover; flex-shrink: 0; }
.text { min-width: 0; display: flex; flex-direction: column; gap: 1px; }
.name, .lyric { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.name { font-size: 12px; font-weight: 600; }
.artist { opacity: .7; font-weight: 400; }
.lyric { font-size: 11px; opacity: .65; }
.ctrls { display: flex; align-items: center; gap: 2px; flex-shrink: 0; }
.ctrls button { width: 28px; height: 28px; display: flex; align-items: center; justify-content: center; border-radius: 6px; opacity: .9; }
.ctrls button:hover { opacity: 1; background: color-mix(in srgb, var(--tb-fg) 10%, transparent); }
.ctrls .play { width: 32px; height: 32px; }
.ctrls .liked { color: #fe2c55; opacity: 1; }
</style>
