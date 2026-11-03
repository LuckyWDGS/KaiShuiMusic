import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import MiniApp from './MiniApp.vue'
import { router } from './router'
import { platform } from './platform'
import { enableSpatialNav } from './utils/spatialNav'
import './styles/global.css'

if (platform.value === 'mini') {
  // 任务栏播放器：独立的小页面，不加载播放器和路由
  createApp(MiniApp).mount('#app')
} else {
  if (platform.value === 'tv') enableSpatialNav()
  createApp(App).use(createPinia()).use(router).mount('#app')
}
