<script setup lang="ts">
// 唱盘机：圆形封面作为唱片 + 唱臂。播放时唱臂落到唱片上、封面缓慢旋转；暂停时唱臂抬起、封面停在当前角度
import CoverImg from './CoverImg.vue'

defineProps<{ cover?: string; source?: string; playing: boolean }>()
</script>

<template>
  <div class="turntable" :class="{ playing }">
    <div class="disc">
      <CoverImg :src="cover" :size="600" :source="source" radius="50%" class="art" />
      <span class="shine"></span>
    </div>
    <svg class="arm" viewBox="0 0 100 240" aria-hidden="true">
      <circle cx="60" cy="24" r="17" fill="rgba(255,255,255,.95)" />
      <circle cx="60" cy="24" r="7" fill="rgba(0,0,0,.18)" />
      <path d="M60 24 L60 150 L40 192" stroke="rgba(255,255,255,.95)" stroke-width="5" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <rect x="26" y="187" width="22" height="34" rx="5" fill="#fff" transform="rotate(26 37 204)" />
    </svg>
  </div>
</template>

<style scoped>
/* 以正方形为画布：封面圆盘偏左放，右上角留给唱臂 */
.turntable { position: relative; width: 100%; aspect-ratio: 1; }
.disc {
  position: absolute; left: 2%; top: 7%; width: 88%; aspect-ratio: 1; border-radius: 50%;
  box-shadow: 0 22px 50px rgba(0, 0, 0, .32), 0 0 0 1px rgba(255, 255, 255, .14);
  animation: spin 30s linear infinite; animation-play-state: paused;
}
.playing .disc { animation-play-state: running; }
.disc .art { position: absolute; inset: 0; width: 100%; height: 100%; }
/* 细腻的光泽与内描边，随封面一起转 */
.shine {
  position: absolute; inset: 0; border-radius: 50%; pointer-events: none;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, .12), inset 0 0 40px rgba(0, 0, 0, .18);
  background: conic-gradient(from 35deg, rgba(255, 255, 255, .1), transparent 10%, transparent 48%, rgba(255, 255, 255, .06) 55%, transparent 64%);
}
.arm {
  position: absolute; left: 69%; top: -1%; width: 25%; height: 60%; overflow: visible;
  transform-origin: 60% 10%; transform: rotate(-26deg); transition: transform .6s cubic-bezier(.3, .8, .3, 1);
  filter: drop-shadow(0 6px 10px rgba(0, 0, 0, .28));
}
.playing .arm { transform: rotate(0deg); }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
