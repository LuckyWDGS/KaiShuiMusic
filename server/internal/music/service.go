// Package music 在 go-music-dl 的 core 包之上提供面向 App 的聚合能力：
// 首页推荐、为你推荐歌曲流、聚合搜索、歌单/专辑/榜单详情、歌词以及可播放性解析（自动换源）。
// 具体的各平台接口全部复用 go-music-dl / music-lib，这里只做编排与缓存。
package music

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/guohuiyuan/go-music-dl/core"
	"github.com/guohuiyuan/music-lib/model"
)

const (
	// FeedSource 是"为你推荐"歌曲流和首页内容的主要来源
	FeedSource = "soda"

	searchTimeout = 8 * time.Second
	feedPageSize  = 20
)

var (
	DefaultSongSources     = []string{"netease", "qq", "kugou", "kuwo", "migu", "soda"}
	DefaultPlaylistSources = []string{"netease", "qq", "kugou", "soda"}
	DefaultAlbumSources    = []string{"netease", "qq", "kugou", "soda"}
	// TopListSources 是"排行榜"页面可切换的平台
	TopListSources = []string{"soda", "netease", "qq", "kugou", "kuwo"}
)

type Service struct {
	cache     *ttlCache
	sodaInfos sync.Map // song id -> sodaInfoEntry
}

func NewService() *Service {
	return &Service{cache: newTTLCache(512)}
}

// ---------- 首页 ----------

type Home struct {
	Playlists  []model.Playlist         `json:"playlists"`
	TopLists   []model.TopList          `json:"toplists"`
	Categories []model.PlaylistCategory `json:"categories"`
}

func (s *Service) Home() (*Home, error) {
	return cached(s.cache, "home", 30*time.Minute, func() (*Home, error) {
		home := &Home{}
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			home.Playlists = s.recommendPlaylists()
		}()
		go func() {
			defer wg.Done()
			if tl, err := s.TopLists(FeedSource); err == nil {
				home.TopLists = tl
			}
		}()
		go func() {
			defer wg.Done()
			if cats, err := s.Categories(FeedSource); err == nil {
				home.Categories = cats
			}
		}()
		wg.Wait()
		if len(home.Playlists) == 0 && len(home.TopLists) == 0 {
			return nil, errors.New("首页内容加载失败，请检查网络")
		}
		return home, nil
	})
}

// recommendPlaylists 汇总无需登录即可获取的推荐歌单：汽水"每日推荐"分类 + 网易云推荐歌单。
func (s *Service) recommendPlaylists() []model.Playlist {
	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		groups = make([][]model.Playlist, 2)
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		if pls, err := s.CategoryPlaylists(FeedSource, "0", 1); err == nil {
			mu.Lock()
			groups[0] = pls
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		if fn := core.GetRecommendFunc("netease"); fn != nil {
			if pls, err := withTimeout(searchTimeout, fn); err == nil {
				for i := range pls {
					pls[i].Source = "netease"
				}
				mu.Lock()
				groups[1] = pls
				mu.Unlock()
			}
		}
	}()
	wg.Wait()
	return interleave(groups, 30)
}

// Feed 返回"为你推荐"的歌曲流：合并汽水官方榜单歌曲，去重后随机打散，按页返回。
func (s *Service) Feed(page int) ([]model.Song, error) {
	pool, err := cached(s.cache, "feed-pool", 30*time.Minute, func() ([]model.Song, error) {
		lists, err := s.TopLists(FeedSource)
		if err != nil {
			return nil, err
		}
		if len(lists) > 5 {
			lists = lists[:5]
		}
		groups := make([][]model.Song, len(lists))
		var wg sync.WaitGroup
		for i, tl := range lists {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if songs, err := s.Detail("toplist", FeedSource, tl.ID); err == nil {
					groups[i] = songs
				}
			}()
		}
		wg.Wait()
		seen := make(map[string]bool)
		var pool []model.Song
		for _, g := range groups {
			for _, song := range g {
				key := strings.ToLower(song.Name + "|" + song.Artist)
				if seen[key] {
					continue
				}
				seen[key] = true
				pool = append(pool, song)
			}
		}
		if len(pool) == 0 {
			return nil, errors.New("暂时没有推荐歌曲")
		}
		rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
		return pool, nil
	})
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	start := (page - 1) * feedPageSize
	if start >= len(pool) {
		// 歌曲流循环播放：超出范围后从头开始
		start %= len(pool)
	}
	end := min(start+feedPageSize, len(pool))
	return pool[start:end], nil
}

// ---------- 音乐馆 ----------

func (s *Service) Categories(source string) ([]model.PlaylistCategory, error) {
	return cached(s.cache, "cats:"+source, 6*time.Hour, func() ([]model.PlaylistCategory, error) {
		fn := core.GetPlaylistCategoriesFunc(source)
		if fn == nil {
			return nil, fmt.Errorf("%s 不支持歌单分类", core.GetSourceDescription(source))
		}
		return withTimeout(searchTimeout, fn)
	})
}

func (s *Service) CategoryPlaylists(source, categoryID string, page int) ([]model.Playlist, error) {
	if page < 1 {
		page = 1
	}
	key := fmt.Sprintf("catpl:%s:%s:%d", source, categoryID, page)
	return cached(s.cache, key, 30*time.Minute, func() ([]model.Playlist, error) {
		fn := core.GetCategoryPlaylistsFunc(source)
		if fn == nil {
			return nil, fmt.Errorf("%s 不支持歌单分类", core.GetSourceDescription(source))
		}
		pls, err := withTimeout(searchTimeout, func() ([]model.Playlist, error) { return fn(categoryID, page, 30) })
		for i := range pls {
			pls[i].Source = source
		}
		return pls, err
	})
}

func (s *Service) TopLists(source string) ([]model.TopList, error) {
	return cached(s.cache, "toplists:"+source, 2*time.Hour, func() ([]model.TopList, error) {
		fn := core.GetTopListsFunc(source)
		if fn == nil {
			return nil, fmt.Errorf("%s 不支持排行榜", core.GetSourceDescription(source))
		}
		lists, err := withTimeout(searchTimeout, fn)
		for i := range lists {
			lists[i].Source = source
		}
		return lists, err
	})
}

// Detail 返回歌单 / 专辑 / 榜单的歌曲列表。
func (s *Service) Detail(kind, source, id string) ([]model.Song, error) {
	key := fmt.Sprintf("detail:%s:%s:%s", kind, source, id)
	return cached(s.cache, key, 15*time.Minute, func() ([]model.Song, error) {
		var fn func(string) ([]model.Song, error)
		switch kind {
		case "album":
			fn = core.GetAlbumDetailFunc(source)
		case "toplist":
			fn = core.GetTopListSongsFunc(source)
		default:
			fn = core.GetPlaylistDetailFunc(source)
		}
		if fn == nil {
			return nil, fmt.Errorf("%s 暂不支持该内容", core.GetSourceDescription(source))
		}
		songs, err := withTimeout(20*time.Second, func() ([]model.Song, error) { return fn(id) })
		for i := range songs {
			songs[i].Source = source
		}
		return songs, err
	})
}

// ---------- 搜索 ----------

type SearchResult struct {
	Songs     []model.Song     `json:"songs"`
	Playlists []model.Playlist `json:"playlists"`
	// Parsed 表示关键字是一个分享链接，已直接解析
	Parsed bool `json:"parsed"`
}

func (s *Service) Search(keyword, searchType string, sources []string) (*SearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return &SearchResult{}, nil
	}
	if strings.HasPrefix(keyword, "http") {
		return s.parseLink(keyword)
	}
	if len(sources) == 0 {
		switch searchType {
		case "playlist":
			sources = DefaultPlaylistSources
		case "album":
			sources = DefaultAlbumSources
		default:
			sources = DefaultSongSources
		}
	}
	key := fmt.Sprintf("search:%s:%s:%s", searchType, strings.Join(sources, ","), keyword)
	return cached(s.cache, key, 10*time.Minute, func() (*SearchResult, error) {
		res := &SearchResult{}
		var wg sync.WaitGroup
		switch searchType {
		case "playlist", "album":
			groups := make([][]model.Playlist, len(sources))
			for i, src := range sources {
				fn := core.GetPlaylistSearchFunc(src)
				if searchType == "album" {
					fn = core.GetAlbumSearchFunc(src)
				}
				if fn == nil {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					pls, err := withTimeout(searchTimeout, func() ([]model.Playlist, error) { return fn(keyword) })
					if err != nil {
						return
					}
					for j := range pls {
						pls[j].Source = src
					}
					groups[i] = pls
				}()
			}
			wg.Wait()
			res.Playlists = interleave(groups, 0)
		default:
			groups := make([][]model.Song, len(sources))
			for i, src := range sources {
				fn := core.GetSearchFunc(src)
				if fn == nil {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					songs, err := withTimeout(searchTimeout, func() ([]model.Song, error) { return fn(keyword) })
					if err != nil {
						return
					}
					for j := range songs {
						songs[j].Source = src
					}
					groups[i] = songs
				}()
			}
			wg.Wait()
			res.Songs = interleave(groups, 0)
		}
		return res, nil
	})
}

func (s *Service) parseLink(link string) (*SearchResult, error) {
	src := core.DetectSource(link)
	if src == "" {
		return nil, errors.New("无法识别该链接")
	}
	if fn := core.GetParseFunc(src); fn != nil {
		if song, err := fn(link); err == nil && song != nil {
			song.Source = src
			return &SearchResult{Songs: []model.Song{*song}, Parsed: true}, nil
		}
	}
	for _, fn := range []func(string) (*model.Playlist, []model.Song, error){core.GetParsePlaylistFunc(src), core.GetParseAlbumFunc(src)} {
		if fn == nil {
			continue
		}
		if _, songs, err := fn(link); err == nil {
			for i := range songs {
				songs[i].Source = src
			}
			return &SearchResult{Songs: songs, Parsed: true}, nil
		}
	}
	return nil, fmt.Errorf("暂不支持解析该%s链接", core.GetSourceDescription(src))
}

// ---------- 歌词 ----------

func (s *Service) Lyric(song *model.Song) string {
	key := "lyric:" + song.Source + ":" + song.ID
	lrc, _ := cached(s.cache, key, 2*time.Hour, func() (string, error) {
		fn := core.GetLyricFunc(song.Source)
		if fn == nil {
			return "", errors.New("unsupported")
		}
		return withTimeout(searchTimeout, func() (string, error) { return fn(song) })
	})
	return lrc
}

// ---------- 工具 ----------

// withTimeout 给上游调用加超时，防止某个平台卡住拖慢整个页面。
func withTimeout[T any](d time.Duration, fn func() (T, error)) (T, error) {
	type result struct {
		v   T
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := fn()
		ch <- result{v, err}
	}()
	select {
	case r := <-ch:
		return r.v, r.err
	case <-time.After(d):
		var zero T
		return zero, errors.New("请求超时")
	}
}

// interleave 轮流从各个来源取结果，让不同平台的结果交错排列。limit <= 0 表示不限制。
func interleave[T any](groups [][]T, limit int) []T {
	var out []T
	for i := 0; ; i++ {
		added := false
		for _, g := range groups {
			if i < len(g) {
				out = append(out, g[i])
				added = true
				if limit > 0 && len(out) >= limit {
					return out
				}
			}
		}
		if !added {
			return out
		}
	}
}
