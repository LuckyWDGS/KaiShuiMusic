<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { usePlayer } from '@/stores/player'

/** anchor：当前句停在可视区域的位置（0 顶部，0.5 居中） */
const props = withDefaults(defineProps<{ align?: 'left' | 'center'; anchor?: number; large?: boolean }>(), {
  align: 'left',
  anchor: 0.4,
  large: false,
})
const player = usePlayer()
const box = ref<HTMLElement>()
let userScrolling = 0

function scrollToActive(smooth = true) {
  const el = box.value
  if (!el || Date.now() - userScrolling < 3000) return
  const line = el.querySelector<HTMLElement>('.line.active')
  if (!line) return el.scrollTo({ top: 0, behavior: smooth ? 'smooth' : 'auto' })
  el.scrollTo({ top: line.offsetTop - el.clientHeight * props.anchor, behavior: smooth ? 'smooth' : 'auto' })
}

watch(() => player.lyricIndex, () => nextTick(() => scrollToActive()))
watch(() => player.lyrics, () => nextTick(() => scrollToActive(false)))
</script>

<template>
  <div
    ref="box"
    class="lyric"
    :class="[`align-${props.align}`, { large }]"
    :style="{ '--anchor': props.anchor }"
    @wheel.passive="userScrolling = Date.now()"
  >
    <div v-if="player.lyricLoading && !player.lyrics.length" class="tip">歌词加载中…</div>
    <div v-else-if="!player.lyrics.length" class="tip">{{ player.current ? '纯音乐，请欣赏' : '' }}</div>
    <template v-else>
      <div class="pad-top"></div>
      <p
        v-for="(l, i) in player.lyrics"
        :key="i"
        class="line"
        :class="{ active: i === player.lyricIndex }"
        @click="player.seek(l.time)"
      >{{ l.text || '♪' }}</p>
      <div class="pad-bottom"></div>
    </template>
  </div>
</template>

<style scoped>
.lyric {
  height: 100%; overflow-y: auto; scrollbar-width: none; position: relative;
  mask-image: linear-gradient(#000 0, #000 55%, transparent 100%);
}
.lyric::-webkit-scrollbar { display: none; }
.pad-top { height: calc(var(--anchor) * 100%); }
.pad-bottom { height: 70%; }
.line {
  margin: 0; padding: 10px 0; font-size: 22px; line-height: 34px; font-weight: 700; color: rgba(255, 255, 255, .42);
  transition: color .35s; cursor: pointer;
}
.line:hover { color: rgba(255, 255, 255, .7); }
.line.active { color: #fff; }
.large .line { font-size: 28px; line-height: 42px; padding: 11px 0; }
.align-center { text-align: center; }
.tip { padding-top: calc(var(--anchor) * 100%); color: rgba(255, 255, 255, .6); font-size: 22px; font-weight: 700; line-height: 54px; }
</style>
