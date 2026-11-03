// 后端接口封装。PC 版与后端同源；手机 / 电视版可在设置中填写局域网内 PC 的服务地址。
import type {
  DetailKind, Home, LibraryData, Playlist, Resolved, SavedPlaylist, SearchType, Song, SourceInfo, Station,
} from '@/types'

const BASE_KEY = 'kaishui.serverBase'

export function getServerBase(): string {
  try {
    return localStorage.getItem(BASE_KEY) || ''
  } catch {
    return ''
  }
}

export function setServerBase(base: string) {
  try {
    if (base) localStorage.setItem(BASE_KEY, base.replace(/\/+$/, ''))
    else localStorage.removeItem(BASE_KEY)
  } catch { /* 忽略存储不可用 */ }
}

function url(path: string, params?: Record<string, string | number | undefined>) {
  const qs = new URLSearchParams()
  for (const [k, v] of Object.entries(params || {})) {
    if (v !== undefined && v !== '') qs.set(k, String(v))
  }
  const q = qs.toString()
  return `${getServerBase()}/api${path}${q ? `?${q}` : ''}`
}

async function request<T>(path: string, init?: RequestInit & { params?: Record<string, string | number | undefined> }): Promise<T> {
  const res = await fetch(url(path, init?.params), {
    ...init,
    headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || `请求失败 (${res.status})`)
  return data as T
}

const post = <T>(path: string, body: unknown) => request<T>(path, { method: 'POST', body: JSON.stringify(body) })

export const api = {
  sources: () => request<Record<string, SourceInfo[]>>('/sources'),
  home: () => request<Home>('/home'),
  feed: (page: number) => request<Song[]>('/feed', { params: { page } }),
  stations: () => request<{ common: Station[]; more: Station[] }>('/stations'),
  station: (kind: string, id: string) => request<Song[]>('/station', { params: { kind, id } }),
  detail: (kind: DetailKind, source: string, id: string) =>
    request<Song[]>('/detail', { params: { kind, source, id } }),
  search: (q: string, type: SearchType, sources?: string[]) =>
    request<{ songs: Song[] | null; playlists: Playlist[] | null; parsed: boolean }>('/search', {
      params: { q, type, sources: sources?.join(',') },
    }),

  resolve: (song: Song) => post<Resolved>('/resolve', song),
  lyric: (song: Song) => post<{ lyric: string }>('/lyric', song),

  library: () => request<LibraryData>('/library'),
  like: (song: Song, liked: boolean) => post<LibraryData>('/library/like', { song, liked }),
  addHistory: (song: Song) => post<LibraryData>('/library/history', song),
  clearHistory: () => request<LibraryData>('/library/history', { method: 'DELETE' }),
  savePlaylist: (playlist: SavedPlaylist, saved: boolean) => post<LibraryData>('/library/playlist', { playlist, saved }),
  createPlaylist: (name: string) => post<LibraryData>('/library/created', { name }),
  deletePlaylist: (id: string) => request<LibraryData>(`/library/created/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  setPlaylistSong: (id: string, song: Song, add: boolean) =>
    post<LibraryData>(`/library/created/${encodeURIComponent(id)}/songs`, { song, add }),

  download: (song: Song) => post<{ filename: string; skipped: boolean; warning?: string }>('/download', song),
  downloads: () => request<{ dir: string; files: { name: string; size: number; mtime: number }[] }>('/downloads'),
  openDownloads: () => post<{ dir: string }>('/downloads/open', {}),

  settings: () => request<{ downloadDir: string; dataDir: string; version: string }>('/settings'),
  saveSettings: (s: { downloadDir: string }) => post<{ downloadDir: string }>('/settings', s),
}

/** 本地下载文件的虚拟来源，id 为下载目录内的相对路径 */
export const LOCAL_FILE = 'localfile'

/** 音频流地址，供 <audio> 直接播放 */
export function streamUrl(song: Song) {
  if (song.source === LOCAL_FILE) return localFileUrl(song.id)
  return url('/stream', {
    id: song.id,
    source: song.source,
    name: song.name,
    artist: song.artist,
    album: song.album,
    extra: song.extra ? JSON.stringify(song.extra) : undefined,
  })
}

export function localFileUrl(name: string) {
  return url('/downloads/file', { name })
}
