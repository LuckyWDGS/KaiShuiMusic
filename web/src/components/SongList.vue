<script setup lang="ts">
// 歌曲表格（参照汽水音乐）：# / 喜欢 / 歌曲 / 歌手 / 专辑 / 时长 / ⋯
import { getCurrentInstance } from 'vue'
import { Heart, Play, AudioLines, MoreHorizontal } from 'lucide-vue-next'
import type { Song } from '@/types'
import { usePlayer } from '@/stores/player'
import { useLibrary } from '@/stores/library'
import { formatTime, songKey, sourceName } from '@/utils/format'
import { useSongActions } from '@/utils/songActions'

const props = defineProps<{ songs: Song[]; showSource?: boolean; showAlbum?: boolean; source?: string }>()
const emit = defineEmits<{ remove: [song: Song] }>()
// 父组件监听了 remove 时，⋯ 菜单里才有"删除"
const removable = !!getCurrentInstance()?.vnode.props?.onRemove
const player = usePlayer()
const lib = useLibrary()
const { openSongMenu } = useSongActions()

const isCurrent = (s: Song) => !!player.current && songKey(player.current) === songKey(s)
const playAt = (i: number) => player.playList(props.songs, i, 'list', props.source)
const more = (e: MouseEvent, s: Song) => openSongMenu(e, s, removable ? (song) => emit('remove', song) : undefined)
</script>

<template>
  <div class="song-list" :class="{ 'no-album': showAlbum === false }">
    <div class="row head">
      <span class="idx">#</span>
      <span></span>
      <span>歌曲</span>
      <span>歌手</span>
      <span v-if="showAlbum !== false">专辑</span>
      <span class="dur">时长</span>
      <span></span>
    </div>
    <div
      v-for="(s, i) in songs"
      :key="songKey(s) + i"
      class="row"
      :class="{ current: isCurrent(s) }"
      tabindex="0"
      @dblclick="playAt(i)"
      @keydown.enter="playAt(i)"
      @contextmenu.prevent="more($event, s)"
    >
      <span class="idx">
        <AudioLines v-if="isCurrent(s) && player.playing" :size="16" class="eq" />
        <span v-else class="num">{{ i + 1 }}</span>
        <button class="hover-play" title="播放" @click="playAt(i)"><Play :size="15" fill="currentColor" /></button>
      </span>
      <button class="like" :class="{ liked: lib.isLiked(s) }" :title="lib.isLiked(s) ? '取消喜欢' : '喜欢'" @click.stop="lib.toggleLike(s)">
        <Heart :size="19" :fill="lib.isLiked(s) ? 'currentColor' : 'none'" />
      </button>
      <span class="title">
        <span class="name ellipsis">{{ s.name }}</span>
        <span v-if="showSource" class="tag">{{ sourceName(s.source) }}</span>
      </span>
      <span class="artist ellipsis">{{ s.artist }}</span>
      <span v-if="showAlbum !== false" class="album ellipsis">{{ s.album }}</span>
      <span class="dur">{{ s.duration ? formatTime(s.duration) : '' }}</span>
      <button class="more" title="更多" @click.stop="more($event, s)"><MoreHorizontal :size="18" /></button>
    </div>
  </div>
</template>

<style scoped>
.song-list { margin: 0 -32px; }
.row {
  display: grid; grid-template-columns: 44px 36px minmax(0, 3fr) minmax(0, 2fr) minmax(0, 2fr) 56px 40px;
  align-items: center; column-gap: 12px; height: 48px; padding: 0 28px 0 32px; cursor: default;
}
.no-album .row { grid-template-columns: 44px 36px minmax(0, 3fr) minmax(0, 2fr) 56px 40px; }
.row:not(.head):hover, .row:focus-visible { background: rgba(255, 255, 255, .05); outline: none; }
.head { height: 40px; color: var(--text-3); font-size: 13px; }
.idx { position: relative; color: var(--text-2); font-size: 14px; font-variant-numeric: tabular-nums; display: flex; align-items: center; }
.hover-play { display: none; color: var(--text); }
.row:not(.head):hover .num, .row:not(.head):hover .eq { display: none; }
.row:not(.head):hover .hover-play { display: inline-flex; }
.eq { color: var(--text); }
.like { display: flex; color: var(--text-2); }
.like:hover { color: var(--text); }
.like.liked { color: #fe2c55; }
.title { display: flex; align-items: center; gap: 6px; min-width: 0; }
.name { font-size: 15px; font-weight: 700; }
.tag {
  flex-shrink: 0; font-size: 10px; line-height: 14px; padding: 0 4px; border-radius: 3px; font-weight: 600;
  color: #20c25e; background: rgba(32, 194, 94, .14);
}
.artist, .album { font-size: 14px; color: var(--text-2); }
.dur { text-align: right; font-size: 14px; color: var(--text-2); font-variant-numeric: tabular-nums; }
.head .dur { font-size: 13px; color: var(--text-3); }
.more { display: flex; justify-content: center; color: var(--text-2); padding: 6px; border-radius: 6px; }
.more:hover { color: var(--text); background: rgba(255, 255, 255, .08); }
@media (max-width: 760px) {
  .song-list { margin: 0 -14px; }
  .row, .no-album .row { grid-template-columns: 32px 28px minmax(0, 1fr) 40px !important; padding: 0 14px; }
  .row > :nth-child(4), .row > :nth-child(5), .row > .dur { display: none; }
}
</style>
