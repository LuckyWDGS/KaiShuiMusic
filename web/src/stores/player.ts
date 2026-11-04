// 播放器核心：播放队列、播放模式、歌词、自动换源。PC / 手机 / 电视三端共用。
import { defineStore } from 'pinia'
import { computed, ref, shallowRef, watch } from 'vue'
import { api, LOCAL_FILE, streamUrl } from '@/api'
import type { Resolved, Song, Station } from '@/types'
import { coverUrl, songKey } from '@/utils/format'
import { currentLineIndex, parseLyric, type LyricLine } from '@/utils/lyric'
import { useLibrary } from './library'
import { toast } from './toast'

export type PlayMode = 'list' | 'single' | 'shuffle'
/** feed：为你推荐歌曲流（播完自动加载更多）；list：普通歌单 / 搜索结果 */
export type QueueContext = 'feed' | 'list'

const STORE_KEY = 'kaishui.player'

export const usePlayer = defineStore('player', () => {
  // 两个音频元素轮换：audio 正在播放，standby 在后台提前缓冲下一首（或推荐页待播的歌），
  // 切歌时直接换成 standby，省去重新连接音源、缓冲开头的几秒等待。
  const audios = [new Audio(), new Audio()]
  for (const a of audios) a.preload = 'auto'
  let audio = audios[0]
  let standby = audios[1]
  let standbyUrl = ''

  const queue = ref<Song[]>([])
  const index = ref(-1)
  const context = ref<QueueContext>('list')
  /** 播放列表的来源名称，显示在播放列表面板标题下方（如"我喜欢的音乐"） */
  const queueSource = ref('')
  /** 正在收听的听歌模式频道；为空表示默认的推荐流 */
  const station = ref<Station | null>(null)
  const mode = ref<PlayMode>('list')
  const playing = ref(false)
  const loading = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const volume = ref(0.8)
  const muted = ref(false)
  const resolved = shallowRef<Resolved | null>(null)
  const lyrics = shallowRef<LyricLine[]>([])
  const lyricLoading = ref(false)

  // 界面状态
  const showNowPlaying = ref(false)
  const showQueue = ref(false)
  /** 本次启动已经开始过播放（恢复的播放列表还没播放时为 false） */
  const started = ref(false)

  const current = computed<Song | null>(() => queue.value[index.value] ?? null)
  const lyricIndex = computed(() => currentLineIndex(lyrics.value, currentTime.value))

  const resolveCache = new Map<string, Promise<Resolved>>()
  let loadToken = 0
  let errorStreak = 0
  let feedPage = 1
  let feedLoading = false

  function resolveSong(song: Song) {
    // 本地下载的文件直接播放，无需解析
    if (song.source === LOCAL_FILE) return Promise.resolve<Resolved>({ song, preview: false, switched: false })
    const key = songKey(song)
    let p = resolveCache.get(key)
    if (!p) {
      p = api.resolve(song)
      p.catch(() => resolveCache.delete(key))
      resolveCache.set(key, p)
    }
    return p
  }

  async function load(i: number, autoplay = true) {
    const song = queue.value[i]
    if (!song) return
    const token = ++loadToken
    started.value = true
    index.value = i
    resolved.value = null
    lyrics.value = []
    currentTime.value = 0
    duration.value = song.duration || 0
    loading.value = true
    audio.pause()
    updateMediaSession(song)
    loadLyric(song, token)
    persist()

    try {
      const res = await resolveSong(song)
      if (token !== loadToken) return
      resolved.value = res
      useUrl(streamUrl(res.song))
      if (res.preview) toast('该歌曲暂无完整音源，仅可试听片段')
      if (autoplay) {
        await audio.play().catch((e) => {
          if (e?.name !== 'AbortError') playing.value = false
        })
        useLibrary().addHistory(song)
      }
      errorStreak = 0
      prefetchNext()
    } catch (e) {
      if (token !== loadToken) return
      loading.value = false
      onFailed((e as Error).message)
    }
    if (context.value === 'feed' && i >= queue.value.length - 3) loadMoreFeed()
  }

  async function loadLyric(song: Song, token: number) {
    lyricLoading.value = true
    try {
      let { lyric } = await api.lyric(song)
      // 原平台没有歌词时，用换源后的歌曲再取一次
      if (!lyric && token === loadToken) {
        const res = await resolveSong(song).catch(() => null)
        if (res?.switched) lyric = (await api.lyric(res.song)).lyric
      }
      if (token === loadToken) lyrics.value = lyric ? parseLyric(lyric) : []
    } catch {
      if (token === loadToken) lyrics.value = []
    } finally {
      if (token === loadToken) lyricLoading.value = false
    }
  }

  /** 切到指定地址：standby 已经缓冲好这首就直接换过去，否则在当前元素上加载 */
  function useUrl(url: string) {
    if (standbyUrl === url && !standby.error) {
      const old = audio
      audio = standby
      standby = old
      standbyUrl = ''
      old.pause()
      old.removeAttribute('src')
      old.load()
      audio.currentTime = 0
      if (isFinite(audio.duration) && audio.duration > 0) duration.value = audio.duration
      return
    }
    audio.src = url
  }

  /** 让 standby 预先缓冲某首歌 */
  function buffer(song: Song) {
    resolveSong(song).then((res) => {
      if (current.value && songKey(current.value) === songKey(song) && !audio.paused) return
      const url = streamUrl(res.song)
      if (standbyUrl === url || audio.getAttribute('src') === url) return
      standbyUrl = url
      standby.src = url
      standby.load()
    }, () => {})
  }

  /** 提前解析后面两首的播放地址（换源可能要几秒），并缓冲下一首的音频，切歌时即可秒播 */
  function prefetchNext() {
    const len = queue.value.length
    const first = nextIndex()
    if (first < 0) return
    const nextSong = queue.value[first]
    if (nextSong && nextSong !== current.value) buffer(nextSong)
    const after = queue.value[(first + 1) % len]
    if (after && after !== current.value) prepare(after)
  }

  /** 后台预先解析某首歌的播放地址（结果缓存）；withAudio 时同时缓冲音频（推荐页待播的歌） */
  function prepare(song: Song, withAudio = false) {
    if (withAudio) buffer(song)
    else resolveSong(song).catch(() => {})
  }

  function onFailed(msg?: string) {
    errorStreak++
    if (errorStreak >= 3 || queue.value.length <= 1) {
      playing.value = false
      toast(msg ? `播放失败：${msg}` : '播放失败，请稍后重试')
      return
    }
    toast('当前歌曲无法播放，已为你切到下一首')
    advance()
  }

  function nextIndex() {
    const len = queue.value.length
    if (!len) return -1
    if (mode.value === 'shuffle' && context.value !== 'feed' && len > 1) {
      let r = index.value
      while (r === index.value) r = Math.floor(Math.random() * len)
      return r
    }
    return (index.value + 1) % len
  }

  // ---------- 对外操作 ----------

  /** 播放一个列表，从 start 开始；source 为列表来源名称 */
  function playList(songs: Song[], start = 0, ctx: QueueContext = 'list', source = '') {
    if (!songs.length) return
    queue.value = [...songs]
    context.value = ctx
    queueSource.value = source || (ctx === 'feed' ? '推荐' : '')
    if (ctx === 'feed') feedPage = 1
    else station.value = null
    load(Math.max(0, Math.min(start, songs.length - 1)))
  }

  /** 开始收听听歌模式里的频道（以推荐流方式连续播放） */
  async function playStation(st: Station) {
    const songs = st.kind === 'local' ? familiarSongs() : await api.station(st.kind, st.id)
    if (!songs.length) throw new Error('这个模式暂时没有歌曲，先去听几首喜欢的歌吧')
    station.value = st
    playList(songs, 0, 'feed', st.name)
  }

  /** 熟悉模式：喜欢的歌 + 最近播放，随机打散 */
  function familiarSongs() {
    const lib = useLibrary()
    const seen = new Set<string>()
    const list = [...lib.data.likes, ...lib.data.history].filter((s) => {
      const k = songKey(s)
      if (seen.has(k) || s.source === LOCAL_FILE) return false
      seen.add(k)
      return true
    })
    for (let i = list.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[list[i], list[j]] = [list[j], list[i]]
    }
    return list
  }

  /** 退出频道，回到默认推荐流 */
  function leaveStation() {
    station.value = null
  }

  /** 播放单曲：已在队列中则跳过去，否则插入到当前歌曲之后 */
  function play(song: Song) {
    const i = queue.value.findIndex((s) => songKey(s) === songKey(song))
    if (i >= 0) return load(i)
    insertNext(song, false)
    load(index.value + 1)
  }

  function insertNext(song: Song, notify = true) {
    const list = queue.value.filter((s) => songKey(s) !== songKey(song))
    const cur = current.value
    const at = cur ? list.findIndex((s) => songKey(s) === songKey(cur)) + 1 : 0
    list.splice(at, 0, song)
    queue.value = list
    if (cur) index.value = at - 1
    if (context.value === 'feed') context.value = 'list'
    if (notify) toast('已添加到下一首播放')
    persist()
  }

  function removeAt(i: number) {
    if (i === index.value) {
      queue.value.splice(i, 1)
      if (queue.value.length) load(Math.min(i, queue.value.length - 1), playing.value)
      else clearQueue()
      return
    }
    queue.value.splice(i, 1)
    if (i < index.value) index.value--
    persist()
  }

  function clearQueue() {
    loadToken++
    for (const a of audios) {
      a.pause()
      a.removeAttribute('src')
    }
    standbyUrl = ''
    started.value = false
    queue.value = []
    index.value = -1
    resolved.value = null
    lyrics.value = []
    playing.value = false
    persist()
  }

  // ---------- 推荐页预览 ----------
  // 推荐页在还没开始收听推荐流时会展示推荐流的第一首（预览）。此时播放条、空格键、媒体键、
  // 任务栏播放器的播放 / 切歌都应作用于页面上展示的推荐流，而不是旧的播放列表（或空列表）。

  /** 推荐页准备好的推荐流歌曲 */
  const feedPreview = shallowRef<Song[]>([])
  /** 推荐页当前是否可见 */
  const previewActive = ref(false)
  const inPreview = () => previewActive.value && !started.value && context.value !== 'feed' && feedPreview.value.length > 0

  /** 从推荐流第 i 首开始播放 */
  function startPreview(i = 0) {
    const list = feedPreview.value
    if (!list.length) return
    station.value = null
    playList(list, (i + list.length) % list.length, 'feed')
  }

  function toggle() {
    if (inPreview()) return startPreview(0)
    if (!current.value) return
    if (!audio.src || !resolved.value) return load(index.value)
    if (audio.paused) audio.play().catch(() => {})
    else audio.pause()
  }

  function next() {
    if (inPreview()) return startPreview(1)
    advance()
  }

  /** 播完 / 出错时自动切到下一首（不受推荐页预览影响） */
  function advance() {
    const i = nextIndex()
    if (i >= 0) load(i)
  }

  function prev() {
    if (inPreview()) return startPreview(-1)
    if (audio.currentTime > 5) {
      audio.currentTime = 0
      return
    }
    const len = queue.value.length
    if (len) load((index.value - 1 + len) % len)
  }

  function seek(t: number) {
    if (!isFinite(t)) return
    audio.currentTime = t
    currentTime.value = t
  }

  function setVolume(v: number) {
    volume.value = Math.max(0, Math.min(1, v))
    muted.value = volume.value === 0
  }

  function cycleMode() {
    const order: PlayMode[] = ['list', 'single', 'shuffle']
    mode.value = order[(order.indexOf(mode.value) + 1) % order.length]
    toast({ list: '列表循环', single: '单曲循环', shuffle: '随机播放' }[mode.value])
  }

  async function loadMoreFeed() {
    if (feedLoading) return
    feedLoading = true
    try {
      const st = station.value
      const more = !st ? await api.feed(++feedPage) : st.kind === 'local' ? familiarSongs() : await api.station(st.kind, st.id)
      const seen = new Set(queue.value.map(songKey))
      queue.value.push(...more.filter((s) => !seen.has(songKey(s))))
    } catch { /* 下次再试 */ } finally {
      feedLoading = false
    }
  }

  // ---------- audio 事件 ----------

  /** 两个元素都监听，但只处理当前正在播放（active）的那个，后台缓冲的元素的事件忽略 */
  function onAudio(type: string, fn: () => void) {
    for (const a of audios) a.addEventListener(type, () => a === audio && fn())
  }
  onAudio('play', () => (playing.value = true))
  onAudio('pause', () => (playing.value = false))
  onAudio('playing', () => (loading.value = false))
  onAudio('waiting', () => (loading.value = true))
  onAudio('canplay', () => (loading.value = false))
  onAudio('timeupdate', () => (currentTime.value = audio.currentTime))
  onAudio('durationchange', () => {
    if (isFinite(audio.duration) && audio.duration > 0) duration.value = audio.duration
  })
  onAudio('ended', () => {
    if (mode.value === 'single') {
      audio.currentTime = 0
      audio.play().catch(() => {})
    } else advance()
  })
  onAudio('error', () => {
    if (!audio.getAttribute('src') || !resolved.value) return
    loading.value = false
    onFailed()
  })

  // 必须在下面的 watch 之前恢复，否则 immediate watch 会先把空状态写回存储
  restore()

  watch([volume, muted], () => {
    for (const a of audios) {
      a.volume = volume.value
      a.muted = muted.value
    }
    persist()
  }, { immediate: true })
  watch(mode, persist)

  // ---------- 系统媒体控制（媒体键、锁屏、蓝牙耳机） ----------

  function updateMediaSession(song: Song) {
    if (!('mediaSession' in navigator)) return
    navigator.mediaSession.metadata = new MediaMetadata({
      title: song.name,
      artist: song.artist,
      album: song.album,
      artwork: song.cover ? [{ src: coverUrl(song.cover), sizes: '512x512' }] : [],
    })
    document.title = `${song.name} - ${song.artist} | 开水音乐`
  }
  if ('mediaSession' in navigator) {
    const ms = navigator.mediaSession
    ms.setActionHandler('play', () => toggle())
    ms.setActionHandler('pause', () => audio.pause())
    ms.setActionHandler('previoustrack', () => prev())
    ms.setActionHandler('nexttrack', () => next())
  }

  // ---------- 持久化（刷新后恢复队列，不自动播放） ----------

  function persist() {
    try {
      localStorage.setItem(STORE_KEY, JSON.stringify({
        queue: queue.value.slice(0, 500),
        index: index.value,
        context: context.value,
        queueSource: queueSource.value,
        station: station.value,
        mode: mode.value,
        volume: volume.value,
        muted: muted.value,
      }))
    } catch { /* ignore */ }
  }

  function restore() {
    try {
      const s = JSON.parse(localStorage.getItem(STORE_KEY) || 'null')
      if (!s) return
      queue.value = s.queue || []
      index.value = Math.min(s.index ?? -1, queue.value.length - 1)
      context.value = s.context || 'list'
      queueSource.value = s.queueSource || ''
      station.value = s.station || null
      mode.value = s.mode || 'list'
      volume.value = s.volume ?? 0.8
      muted.value = !!s.muted
      if (current.value) {
        duration.value = current.value.duration || 0
        updateMediaSession(current.value)
        loadLyric(current.value, loadToken)
      }
    } catch { /* ignore */ }
  }

  return {
    queue, index, context, mode, playing, loading, currentTime, duration, volume, muted,
    resolved, lyrics, lyricIndex, lyricLoading, showNowPlaying, showQueue, current,
    playList, play, insertNext, removeAt, clearQueue, toggle, next, prev, seek, setVolume, cycleMode,
    load, loadMoreFeed, queueSource, station, playStation, leaveStation, feedPreview, previewActive, startPreview, prepare, started,
  }
})
