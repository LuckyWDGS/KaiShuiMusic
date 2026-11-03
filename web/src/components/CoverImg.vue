<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getServerBase } from '@/api'
import { coverUrl } from '@/utils/format'

const props = defineProps<{ src?: string; size?: number; source?: string; radius?: string }>()

// 加载失败时依次尝试：https → 原地址 → 后端代理
const attempt = ref(0)
watch(() => props.src, () => (attempt.value = 0))

const url = computed(() => {
  if (!props.src) return ''
  if (attempt.value === 0) return coverUrl(props.src, props.size)
  if (attempt.value === 1) return props.src
  if (attempt.value === 2) {
    const q = new URLSearchParams({ url: props.src, source: props.source || '' })
    return `${getServerBase()}/api/cover?${q}`
  }
  return ''
})
</script>

<template>
  <div class="cover" :style="{ borderRadius: radius ?? 'var(--radius-sm)' }">
    <img v-if="url" :src="url" loading="lazy" referrerpolicy="no-referrer" alt="" @error="attempt++" />
    <svg v-else class="placeholder" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
      <path d="M9 18V5l12-2v13" /><circle cx="6" cy="18" r="3" /><circle cx="18" cy="16" r="3" />
    </svg>
  </div>
</template>

<style scoped>
.cover { position: relative; overflow: hidden; background: #26262b; aspect-ratio: 1; flex-shrink: 0; }
.cover img { width: 100%; height: 100%; object-fit: cover; }
.placeholder { position: absolute; inset: 28%; width: 44%; height: 44%; color: var(--text-3); }
</style>
