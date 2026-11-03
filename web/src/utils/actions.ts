import { api } from '@/api'
import { toast } from '@/stores/toast'
import type { Song } from '@/types'

export async function downloadSong(song: Song) {
  toast(`开始下载：${song.name}`)
  try {
    const r = await api.download(song)
    toast(r.skipped ? '已下载过，跳过' : `下载完成：${r.filename}`)
  } catch (e) {
    toast((e as Error).message)
  }
}
