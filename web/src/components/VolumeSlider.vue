<script setup lang="ts">
// 竖向音量滑块（参照汽水音乐）：圆角竖条，从下往上填充，可拖动 / 点击 / 滚轮调节
import { computed, ref } from 'vue'

const props = defineProps<{ value: number }>()
const emit = defineEmits<{ change: [v: number] }>()

const track = ref<HTMLElement>()
const dragging = ref(false)
const percent = computed(() => Math.round(Math.max(0, Math.min(1, props.value)) * 100))

function valueAt(e: PointerEvent) {
  const r = track.value!.getBoundingClientRect()
  return Math.max(0, Math.min(1, (r.bottom - e.clientY) / r.height))
}
function down(e: PointerEvent) {
  dragging.value = true
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
  emit('change', valueAt(e))
}
function move(e: PointerEvent) {
  if (dragging.value) emit('change', valueAt(e))
}
function up() {
  dragging.value = false
}
</script>

<template>
  <div class="volume-slider" :class="{ dragging }" :title="`音量 ${percent}%`" @pointerdown="down" @pointermove="move" @pointerup="up" @pointercancel="up">
    <div ref="track" class="track">
      <div class="fill" :style="{ height: percent + '%' }"><span class="knob"></span></div>
    </div>
  </div>
</template>

<style scoped>
.volume-slider {
  width: 36px; height: 150px; padding: 16px 0; border-radius: 18px; display: flex; justify-content: center;
  background: rgba(205, 215, 215, .55); backdrop-filter: blur(16px); box-shadow: 0 8px 24px rgba(0, 0, 0, .25);
  cursor: pointer; touch-action: none;
}
.track { position: relative; width: 4px; height: 100%; border-radius: 2px; background: rgba(255, 255, 255, .45); }
.fill { position: absolute; left: 0; right: 0; bottom: 0; border-radius: 2px; background: #fff; }
.knob {
  position: absolute; left: 50%; top: 0; width: 14px; height: 14px; margin: -7px 0 0 -7px; border-radius: 50%;
  background: #fff; box-shadow: 0 1px 4px rgba(0, 0, 0, .3); transition: transform .1s;
}
.volume-slider:hover .knob, .dragging .knob { transform: scale(1.15); }
</style>
