// 歌曲的 ⋯ 菜单与"收藏到歌单"菜单（参照汽水音乐的菜单项）
import { useRouter } from 'vue-router'
import { Plus, Heart, ListMusic } from 'lucide-vue-next'
import { LOCAL_FILE } from '@/api'
import type { Song } from '@/types'
import { useLibrary } from '@/stores/library'
import { usePlayer } from '@/stores/player'
import { openMenu, reopenMenu, type MenuItem } from '@/stores/menu'
import { toast } from '@/stores/toast'
import { downloadSong } from './actions'

export function useSongActions() {
  const router = useRouter()
  const player = usePlayer()
  const lib = useLibrary()

  function collectItems(song: Song): MenuItem[] {
    return [
      { label: '创建新歌单', icon: Plus, action: () => { lib.collecting = song; lib.creating = true } },
      {
        label: '我喜欢的音乐',
        icon: Heart,
        action: () => (lib.isLiked(song) ? toast('已经在我喜欢的音乐里了') : lib.toggleLike(song)),
      },
      ...lib.data.created.map((p) => ({ label: p.name, icon: ListMusic, action: () => lib.addToPlaylist(p.id, song) })),
    ]
  }

  /** 收藏到歌单：从触发按钮旁弹出歌单菜单 */
  function openCollectMenu(e: MouseEvent, song: Song) {
    openMenu(e, collectItems(song))
  }

  /** 歌曲 ⋯ 菜单；onRemove 存在时显示"删除" */
  function openSongMenu(e: MouseEvent, song: Song, onRemove?: (s: Song) => void) {
    const local = song.source === LOCAL_FILE
    const items: MenuItem[] = [{ label: '下一首播放', action: () => player.insertNext(song) }]
    if (!local) {
      items.push({ label: '添加到歌单', action: () => reopenMenu(collectItems(song)) })
      items.push({ label: '查看歌手', action: () => router.push({ name: 'search', query: { q: song.artist } }) })
      if (song.album) {
        items.push({
          label: '查看专辑',
          action: () => song.album_id
            ? router.push({ name: 'detail', params: { kind: 'album', source: song.source, id: song.album_id }, query: { name: song.album, cover: song.cover } })
            : router.push({ name: 'search', query: { q: `${song.album} ${song.artist}`, type: 'album' } }),
        })
      }
    }
    items.push({
      label: '复制曲名',
      action: () => navigator.clipboard.writeText(`${song.name} - ${song.artist}`).then(() => toast('已复制'), () => toast('复制失败')),
    })
    if (!local) items.push({ label: '下载', action: () => downloadSong(song) })
    if (onRemove) items.push({ label: '删除', danger: true, action: () => onRemove(song) })
    openMenu(e, items)
  }

  return { openSongMenu, openCollectMenu }
}
