// "我的"音乐库：喜欢、最近播放、收藏歌单。数据保存在后端本地文件，无需登录。
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '@/api'
import type { DetailKind, LibraryData, Playlist, Song } from '@/types'
import { songKey } from '@/utils/format'
import { toast } from './toast'

export const useLibrary = defineStore('library', () => {
  const data = ref<LibraryData>({ likes: [], history: [], playlists: [], created: [] })
  /** 正在"收藏到歌单"的歌曲，非空时弹出歌单选择框 */
  const collecting = ref<Song | null>(null)
  /** 是否显示"新建歌单"对话框 */
  const creating = ref(false)
  const likedKeys = computed(() => new Set(data.value.likes.map(songKey)))
  const savedKeys = computed(() => new Set(data.value.playlists.map((p) => `${p.kind}:${p.source}:${p.id}`)))

  async function load() {
    try {
      data.value = { ...data.value, ...(await api.library()) }
    } catch { /* 后端未启动时保持空 */ }
  }

  const isLiked = (song?: Song | null) => !!song && likedKeys.value.has(songKey(song))

  async function toggleLike(song: Song) {
    const liked = !isLiked(song)
    try {
      data.value = await api.like(song, liked)
      toast(liked ? '已添加到我喜欢' : '已取消喜欢')
    } catch (e) {
      toast((e as Error).message)
    }
  }

  async function addHistory(song: Song) {
    try {
      data.value = await api.addHistory(song)
    } catch { /* ignore */ }
  }

  async function clearHistory() {
    data.value = await api.clearHistory()
  }

  const isSaved = (kind: DetailKind, p: { source: string; id: string }) => savedKeys.value.has(`${kind}:${p.source}:${p.id}`)

  async function toggleSave(kind: DetailKind, playlist: Playlist) {
    const saved = !isSaved(kind, playlist)
    try {
      data.value = await api.savePlaylist({ ...playlist, kind }, saved)
      toast(saved ? '已收藏歌单' : '已取消收藏')
    } catch (e) {
      toast((e as Error).message)
    }
  }

  async function run(fn: () => Promise<LibraryData>, ok?: string) {
    try {
      data.value = await fn()
      if (ok) toast(ok)
      return true
    } catch (e) {
      toast((e as Error).message)
      return false
    }
  }

  const createPlaylist = (name: string) => run(() => api.createPlaylist(name), '歌单已创建')
  const deletePlaylist = (id: string) => run(() => api.deletePlaylist(id), '歌单已删除')
  const addToPlaylist = (id: string, song: Song) => run(() => api.setPlaylistSong(id, song, true), '已收藏到歌单')
  const removeFromPlaylist = (id: string, song: Song) => run(() => api.setPlaylistSong(id, song, false), '已从歌单移除')

  return {
    data, collecting, creating, load, isLiked, toggleLike, addHistory, clearHistory, isSaved, toggleSave,
    createPlaylist, deletePlaylist, addToPlaylist, removeFromPlaylist,
  }
})
