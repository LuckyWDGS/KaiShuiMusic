<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps<{ value: number; max: number }>()
const emit = defineEmits<{ change: [v: number] }>()

const track = ref<HTMLElement>()
const dragging = ref(false)
const dragValue = ref(0)

const percent = computed(() => {
  const v = dragging.value ? dragValue.value : props.value
  return props.max > 0 ? Math.min(100, (v / props.max) * 100) : 0
})

function valueAt(e: PointerEvent) {
  const r = track.value!.getBoundingClientRect()
  return Math.max(0, Math.min(1, (e.clientX - r.left) / r.width)) * props.max
}
function down(e: PointerEvent) {
  if (!props.max) return
  dragging.value = true
  dragValue.value = valueAt(e)
  ;(e.target as HTMLElement).setPointerCapture(e.pointerId)
}
function move(e: PointerEvent) {
  if (dragging.value) dragValue.value = valueAt(e)
}
function up() {
  if (!dragging.value) return
  dragging.value = false
  emit('change', dragValue.value)
}
</script>

<template>
  <div ref="track" class="bar" :class="{ dragging }" @pointerdown="down" @pointermove="move" @pointerup="up" @pointercancel="up">
    <div class="rail"><div class="fill" :style="{ width: percent + '%' }"><span class="knob"></span></div></div>
  </div>
</template>

<style scoped>
.bar { height: 14px; display: flex; align-items: center; cursor: pointer; touch-action: none; }
.rail { position: relative; width: 100%; height: 3px; border-radius: 2px; background: rgba(255, 255, 255, .14); transition: height .15s; }
.bar:hover .rail, .dragging .rail { height: 5px; }
.fill { position: relative; height: 100%; border-radius: inherit; background: var(--text); }
.bar:hover .fill, .dragging .fill { background: var(--accent); }
.knob {
  position: absolute; right: -6px; top: 50%; width: 12px; height: 12px; margin-top: -6px; border-radius: 50%;
  background: #fff; opacity: 0; transition: opacity .15s; box-shadow: 0 1px 4px rgba(0, 0, 0, .4);
}
.bar:hover .knob, .dragging .knob { opacity: 1; }
</style>
