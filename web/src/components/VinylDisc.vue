<script setup lang="ts">
// 唱片封面：一张圆形黑胶唱片直接浮在背景上，封面印在唱片中央；播放时缓慢旋转，暂停时停在当前角度
import CoverImg from './CoverImg.vue'

defineProps<{ cover?: string; source?: string; spinning: boolean }>()
</script>

<template>
  <div class="vinyl">
    <div class="disc" :class="{ spinning }">
      <CoverImg :src="cover" :size="600" :source="source" radius="50%" class="art" />
      <span class="shine"></span>
    </div>
  </div>
</template>

<style scoped>
.vinyl { width: 100%; aspect-ratio: 1; display: flex; align-items: center; justify-content: center; }
/* 外圈细黑胶边带纹路，中间是封面 */
.disc {
  position: relative; width: 100%; aspect-ratio: 1; border-radius: 50%;
  background: repeating-radial-gradient(circle, #141414 0 2px, #1d1d1d 2px 3px);
  box-shadow: 0 16px 40px rgba(0, 0, 0, .28);
  animation: spin 24s linear infinite; animation-play-state: paused;
}
.disc.spinning { animation-play-state: running; }
.disc .art { position: absolute; inset: 4%; width: 92%; height: 92%; }
/* 唱片高光，随唱片一起转 */
.shine {
  position: absolute; inset: 0; border-radius: 50%; pointer-events: none;
  background: conic-gradient(from 30deg, rgba(255, 255, 255, .1), transparent 15%, transparent 45%, rgba(255, 255, 255, .07) 55%, transparent 70%);
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
