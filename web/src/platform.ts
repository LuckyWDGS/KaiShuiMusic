// 运行平台识别：PC / 手机 / 电视。界面布局按平台切换，业务逻辑（播放器、接口、音乐库）三端共用。
// 可通过 URL 参数 ?platform=pc|mobile|tv 强制指定（桌面壳、电视壳启动时传入），并会被记住。
import { ref } from 'vue'

/** mini：桌面版嵌在任务栏里的小播放器 */
export type Platform = 'pc' | 'mobile' | 'tv' | 'mini'

const KEY = 'kaishui.platform'

function detect(): Platform {
  const forced = new URLSearchParams(location.search).get('platform')
  if (forced === 'mini') return 'mini'
  if (forced === 'pc' || forced === 'mobile' || forced === 'tv') {
    try { sessionStorage.setItem(KEY, forced) } catch { /* ignore */ }
    return forced
  }
  try {
    const saved = sessionStorage.getItem(KEY)
    if (saved === 'pc' || saved === 'mobile' || saved === 'tv') return saved
  } catch { /* ignore */ }

  const ua = navigator.userAgent
  if (/\b(TV|SmartTV|AFT|BRAVIA|Android TV|GoogleTV|HiTV|MiTV)\b/i.test(ua)) return 'tv'
  if (/Android|iPhone|iPad|Mobile/i.test(ua) || window.innerWidth < 768) return 'mobile'
  return 'pc'
}

export const platform = ref<Platform>(detect())
export const isPC = () => platform.value === 'pc'
