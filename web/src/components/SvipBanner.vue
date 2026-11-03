<script setup lang="ts">
// SVIP 横幅（仿汽水音乐样式，纯装饰）：开水音乐不需要账号，所有歌曲免费听，点击只给出提示
import { toast } from '@/stores/toast'

const tip = () => toast('开水音乐全部歌曲免费畅听，无需开通会员')
</script>

<template>
  <div class="svip" @click="tip">
    <span class="text">尊享 SVIP 专属音质 · 开水音乐免费畅听</span>
    <span class="badge">已开通 <b>SVIP</b></span>
  </div>
</template>

<style scoped>
.svip {
  display: flex; align-items: center; justify-content: space-between; height: 52px; padding: 0 10px 0 14px; margin-bottom: 14px;
  border-radius: 10px; background: linear-gradient(90deg, #2c271f, #26231f 60%, #2a2620); cursor: pointer;
}
.text {
  font-size: 14px; font-weight: 700; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  background: linear-gradient(90deg, #f6dfae, #e2b86b); -webkit-background-clip: text; background-clip: text; color: transparent;
}
.badge {
  display: inline-flex; align-items: center; gap: 4px; height: 32px; padding: 0 14px; border-radius: 16px;
  background: rgba(246, 223, 174, .12); color: #f3d9a4; font-size: 13px; font-weight: 700;
  white-space: nowrap; flex-shrink: 0; margin-left: 12px;
}
.badge b { font-style: italic; letter-spacing: .5px; }
.svip:hover .badge { background: rgba(246, 223, 174, .2); }
</style>
