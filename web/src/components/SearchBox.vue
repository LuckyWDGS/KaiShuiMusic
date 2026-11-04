<script setup lang="ts">
// 顶栏搜索框（参照汽水音乐）：聚焦时下拉显示历史搜索
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search, X, Trash2 } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const HISTORY_KEY = 'kaishui.searchHistory'

const keyword = ref(typeof route.query.q === 'string' ? route.query.q : '')
const focused = ref(false)
const history = ref<string[]>(load())

function load(): string[] {
  try {
    return JSON.parse(localStorage.getItem(HISTORY_KEY) || '[]')
  } catch {
    return []
  }
}
function save() {
  try { localStorage.setItem(HISTORY_KEY, JSON.stringify(history.value)) } catch { /* ignore */ }
}

function submit(q = keyword.value) {
  q = q.trim()
  if (!q) return
  keyword.value = q
  history.value = [q, ...history.value.filter((h) => h !== q)].slice(0, 12)
  save()
  focused.value = false
  ;(document.activeElement as HTMLElement)?.blur()
  router.push({ name: 'search', query: { q, type: route.name === 'search' ? route.query.type : undefined } })
}

function onBlur() {
  setTimeout(() => (focused.value = false), 150)
}
function clearHistory() {
  history.value = []
  save()
}
</script>

<template>
  <div class="search" :class="{ focused }">
    <Search :size="16" class="ico" />
    <input
      v-model="keyword"
      placeholder="歌手、歌曲或专辑名"
      @focus="focused = true"
      @blur="onBlur"
      @keydown.enter="submit()"
      @keydown.esc="($event.target as HTMLInputElement).blur()"
    />
    <button v-if="keyword && focused" class="clear" @mousedown.prevent="keyword = ''"><X :size="15" /></button>

    <div v-if="focused && history.length" class="dropdown" @mousedown.prevent>
      <div class="block-head">
        <span>历史搜索</span>
        <button class="trash" title="清空历史搜索" @click="clearHistory"><Trash2 :size="17" /></button>
      </div>
      <div class="chips">
        <button v-for="h in history" :key="h" class="chip" @click="submit(h)">{{ h }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search {
  position: relative; display: flex; align-items: center; width: 240px; height: 32px; padding: 0 12px;
  border-radius: 10px; background: rgba(255, 255, 255, .12); border: 1px solid transparent; transition: border-color .15s;
}
.search.focused { border-color: rgba(255, 255, 255, .7); }
.ico { color: var(--text-2); flex-shrink: 0; }
input { flex: 1; min-width: 0; height: 100%; border: 0; outline: 0; background: none; padding: 0 6px; font-size: 13px; font-weight: 600; user-select: text; }
input::placeholder { color: var(--text-3); }
.clear { color: var(--text-3); display: flex; }
.clear:hover { color: var(--text); }
.dropdown {
  position: absolute; top: 40px; left: -40px; width: 320px; z-index: 50; padding: 16px; border-radius: 10px;
  background: var(--surface); box-shadow: 0 16px 40px rgba(0, 0, 0, .45); max-height: 460px; overflow-y: auto;
}
@media (max-width: 760px) {
  .dropdown { left: 0; right: 0; width: auto; }
}
.block-head { display: flex; align-items: center; justify-content: space-between; font-size: 14px; font-weight: 700; margin-bottom: 12px; }
.trash { display: flex; color: var(--text-2); }
.trash:hover { color: var(--text); }
.chips { display: flex; flex-wrap: wrap; gap: 10px 8px; }
.chip { height: 32px; padding: 0 16px; border-radius: 16px; background: rgba(255, 255, 255, .08); font-size: 13px; font-weight: 600; color: var(--text-2); }
.chip:hover { color: var(--text); background: rgba(255, 255, 255, .14); }
</style>
