// 与后端（music-lib model）一致的数据结构

export interface Song {
  id: string
  name: string
  artist: string
  album: string
  album_id?: string
  duration: number
  source: string
  cover: string
  link?: string
  extra?: Record<string, string>
  is_vip?: boolean
}

/** 一个平台返回的一批搜索结果 */
export interface SearchChunk {
  source?: string
  songs?: Song[]
  playlists?: Playlist[]
  parsed?: boolean
}

export interface Playlist {
  id: string
  name: string
  cover: string
  track_count: number
  play_count: number
  creator?: string
  description?: string
  source: string
  link?: string
  extra?: Record<string, string>
}

export interface TopList {
  id: string
  name: string
  cover: string
  description?: string
  update_frequency?: string
  track_count: number
  source: string
  extra?: Record<string, string>
}

export interface PlaylistCategory {
  id: string
  name: string
  group: string
  source: string
  hot?: boolean
}

export type DetailKind = 'playlist' | 'album' | 'toplist'

export interface SavedPlaylist extends Playlist {
  kind: DetailKind
}

export interface CreatedPlaylist {
  id: string
  name: string
  songs: Song[]
  created_at: number
}

export interface LibraryData {
  likes: Song[]
  history: Song[]
  playlists: SavedPlaylist[]
  created: CreatedPlaylist[]
}

export interface Home {
  playlists: Playlist[]
  toplists: TopList[]
  categories: PlaylistCategory[]
}

export interface Resolved {
  song: Song
  preview: boolean
  switched: boolean
}

export interface SourceInfo {
  id: string
  name: string
}

export type SearchType = 'song' | 'playlist' | 'album'

/** 听歌模式频道 */
export interface Station {
  id: string
  kind: 'toplist' | 'category' | 'local'
  name: string
  desc?: string
  icon?: string
  cover?: string
}
