<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, getServerBase, setServerBase } from '@/api'
import { platform } from '@/platform'
import { isAppShell } from '@/utils/appShell'
import { toast } from '@/stores/toast'

const downloadDir = ref('')
const dataDir = ref('')
const version = ref('')
const serverBase = ref(getServerBase())
// APP 内置服务，下载位置固定，也不需要连接电脑
const app = isAppShell()

onMounted(async () => {
  try {
    const s = await api.settings()
    downloadDir.value = s.downloadDir
    dataDir.value = s.dataDir
    version.value = s.version
  } catch (e) {
    toast((e as Error).message)
  }
})

async function saveDir() {
  try {
    const s = await api.saveSettings({ downloadDir: downloadDir.value })
    downloadDir.value = s.downloadDir
    toast('已保存')
  } catch (e) {
    toast((e as Error).message)
  }
}

function saveServer() {
  setServerBase(serverBase.value.trim())
  location.reload()
}
</script>

<template>
  <div class="page settings">
    <h1 class="page-title">设置</h1>

    <section class="card">
      <h3>下载</h3>
      <label>下载保存位置</label>
      <p v-if="app" class="hint">{{ downloadDir }}</p>
      <div v-else class="row">
        <input v-model="downloadDir" placeholder="默认保存在数据目录的 data/downloads" />
        <button class="btn primary" @click="saveDir">保存</button>
      </div>
      <p class="hint">下载时会自动选择可完整播放的音源{{ app ? '' : '，并尽量写入封面和歌词（需要本机安装 ffmpeg）' }}。</p>
    </section>

    <section v-if="!app && (platform !== 'pc' || serverBase)" class="card">
      <h3>连接电脑</h3>
      <label>服务地址（手机 / 电视连接局域网内运行开水音乐的电脑）</label>
      <div class="row">
        <input v-model="serverBase" placeholder="例如 http://192.168.1.10:8686，留空表示本机" />
        <button class="btn primary" @click="saveServer">连接</button>
      </div>
    </section>

    <section class="card">
      <h3>关于</h3>
      <p>开水音乐 v{{ version }} · 免登录，打开即听</p>
      <p v-if="!app" class="hint">数据目录：{{ dataDir }}</p>
      <p class="hint">
        音乐源能力基于开源项目
        <a href="https://github.com/guohuiyuan/go-music-dl" target="_blank">go-music-dl</a>
        （AGPL-3.0）。音乐版权归各平台及权利人所有，仅供个人学习交流使用。
      </p>
    </section>
  </div>
</template>

<style scoped>
.settings { max-width: 760px; }
.card { background: var(--bg-elevated); border-radius: var(--radius); padding: 20px 24px; margin-bottom: 18px; }
h3 { margin: 0 0 14px; font-size: 16px; }
label { display: block; font-size: 13px; color: var(--text-2); margin-bottom: 8px; }
.row { display: flex; gap: 10px; }
input {
  flex: 1; height: 36px; padding: 0 12px; border-radius: 8px; border: 1px solid var(--border);
  background: rgba(255, 255, 255, .04); outline: none; user-select: text;
}
input:focus { border-color: var(--accent); }
p { margin: 6px 0; font-size: 14px; }
.hint { font-size: 12px; color: var(--text-3); line-height: 1.7; }
.hint a { color: var(--accent); }
.hint { word-break: break-all; }
@media (max-width: 760px) {
  .card { padding: 16px; }
}
</style>
