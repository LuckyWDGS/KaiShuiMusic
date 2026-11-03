<script setup lang="ts">
// "新建歌单"弹窗；从"收藏到歌单"菜单进入时，建好后把歌曲直接放进新歌单
import { nextTick, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { useLibrary } from '@/stores/library'

const lib = useLibrary()
const name = ref('')
const input = ref<HTMLInputElement>()

watch(() => lib.creating, (v) => {
  if (v) {
    name.value = ''
    nextTick(() => input.value?.focus())
  } else {
    lib.collecting = null
  }
})

async function create() {
  if (!name.value.trim()) return
  const song = lib.collecting
  if (await lib.createPlaylist(name.value.trim())) {
    const created = lib.data.created[lib.data.created.length - 1]
    if (song && created) await lib.addToPlaylist(created.id, song)
    lib.creating = false
  }
}
</script>

<template>
  <Transition name="fade">
    <div v-if="lib.creating" class="mask" @click.self="lib.creating = false">
      <div class="dialog">
        <header>
          <span>新建歌单</span>
          <button class="icon-btn" @click="lib.creating = false"><X :size="18" /></button>
        </header>
        <input ref="input" v-model="name" maxlength="40" placeholder="输入歌单名称" @keydown.enter="create" @keydown.esc="lib.creating = false" />
        <div class="actions">
          <button class="btn" @click="lib.creating = false">取消</button>
          <button class="btn primary" :disabled="!name.trim()" @click="create">创建</button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.mask { position: fixed; inset: 0; z-index: 90; background: rgba(0, 0, 0, .45); display: flex; align-items: center; justify-content: center; }
.dialog {
  width: 380px; border-radius: 12px; padding: 8px 20px 20px;
  background: var(--surface); box-shadow: 0 20px 60px rgba(0, 0, 0, .45);
}
header { display: flex; align-items: center; justify-content: space-between; padding: 6px 0 14px; font-size: 16px; font-weight: 700; }
input {
  width: 100%; height: 40px; padding: 0 12px; border-radius: 8px; border: 1px solid var(--border);
  background: rgba(255, 255, 255, .06); outline: none; user-select: text;
}
input:focus { border-color: rgba(255, 255, 255, .5); }
.actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 18px; }
.fade-enter-active, .fade-leave-active { transition: opacity .15s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
