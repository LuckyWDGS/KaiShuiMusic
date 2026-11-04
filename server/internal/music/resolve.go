package music

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/guohuiyuan/go-music-dl/core"
	"github.com/guohuiyuan/music-lib/model"
	"github.com/guohuiyuan/music-lib/soda"
)

// 换源时搜索的平台，按优先级排列（汽水/5sing 不作为换源目标：前者无登录只有试听）
var switchSources = []string{"netease", "qq", "kugou", "kuwo", "migu", "qianqian"}

const (
	switchCandidatesPerSource = 8
	switchHighConfidence      = 0.98
	switchValidateLimit       = 10
	// 已有试听片段时换源的总时限。找不到时各平台要搜两轮再逐个校验，可能耗时近 20 秒，到时限就先用试听
	switchTimeout = 8 * time.Second
)

type Resolved struct {
	Song model.Song `json:"song"`
	// Preview 表示最终只能播放试听片段
	Preview bool `json:"preview"`
	// Switched 表示已自动换到其它平台的同名歌曲
	Switched bool `json:"switched"`
}

// Resolve 判断歌曲能否完整播放；汽水会员曲目（无登录只有试听）或失效歌曲会自动匹配其它平台的同一首歌。
// 这样用户无需登录也能听完整歌曲。
func (s *Service) Resolve(song model.Song) (*Resolved, error) {
	if song.ID == "" || song.Source == "" {
		return nil, errors.New("缺少歌曲信息")
	}
	key := "resolve:" + song.Source + ":" + song.ID
	return cached(s.cache, key, time.Hour, func() (*Resolved, error) {
		preview, playable := s.checkPlayable(&song)
		if playable && !preview {
			return &Resolved{Song: song}, nil
		}
		if !playable {
			// 没有试听片段可以兜底，换源多久都要等
			if alt := findSwitchSong(song); alt != nil {
				return &Resolved{Song: *alt, Switched: true}, nil
			}
			return nil, errors.New("暂无可播放的音源")
		}
		alt, _ := withTimeout(switchTimeout, func() (*model.Song, error) { return findSwitchSong(song), nil })
		if alt != nil {
			return &Resolved{Song: *alt, Switched: true}, nil
		}
		return &Resolved{Song: song, Preview: true}, nil
	})
}

type sodaInfoEntry struct {
	info    *soda.DownloadInfo
	expires time.Time
}

// SodaInfo 获取汽水播放地址并短暂缓存。解析可播放性和取流都要用到，
// 浏览器拖动进度也会多次请求，缓存后可省去每次 2~3 秒的取地址耗时。
func (s *Service) SodaInfo(song *model.Song) (*soda.DownloadInfo, error) {
	if v, ok := s.sodaInfos.Load(song.ID); ok {
		if e := v.(sodaInfoEntry); time.Now().Before(e.expires) {
			return e.info, nil
		}
	}
	info, err := core.NewSodaClient(core.CM.Get("soda")).GetDownloadInfo(song)
	if err != nil {
		return nil, err
	}
	if info == nil || info.URL == "" {
		return nil, errors.New("汽水未返回播放地址")
	}
	s.sodaInfos.Store(song.ID, sodaInfoEntry{info: info, expires: time.Now().Add(5 * time.Minute)})
	return info, nil
}

func (s *Service) checkPlayable(song *model.Song) (preview bool, playable bool) {
	if song.Source == "soda" {
		info, err := s.SodaInfo(song)
		if err != nil {
			return false, false
		}
		return info.Preview, true
	}
	_, err := s.PlayURL(song)
	return false, err == nil
}

// PlayURL 获取汽水以外平台的播放地址并短暂缓存。解析可播放性、取流、拖动进度都要用到，
// 上游每次要 1~3 秒，缓存后歌曲解析完就能直接开始取流。
func (s *Service) PlayURL(song *model.Song) (string, error) {
	return cached(s.cache, "url:"+song.Source+":"+song.ID, 5*time.Minute, func() (string, error) {
		fn := core.GetDownloadFunc(song.Source)
		if fn == nil {
			return "", errors.New("不支持的音源")
		}
		u, err := withTimeout(searchTimeout, func() (string, error) { return fn(song) })
		if err != nil {
			// 上游偶发失败时再试一次，否则播放器会直接跳到下一首
			u, err = withTimeout(searchTimeout, func() (string, error) { return fn(song) })
		}
		if err != nil {
			return "", err
		}
		if !strings.HasPrefix(u, "http") {
			return "", errors.New("未返回播放地址")
		}
		return u, nil
	})
}

type switchCandidate struct {
	song    model.Song
	score   float64
	durDiff int
}

func findSwitchSong(orig model.Song) *model.Song {
	name := strings.TrimSpace(orig.Name)
	artist := strings.TrimSpace(orig.Artist)
	if name == "" {
		return nil
	}
	keyword := strings.TrimSpace(name + " " + artist)

	// 各平台并行搜索。某个平台的最佳结果高度匹配且确认可播放时立即返回，不再等待较慢的平台；
	// 否则等所有平台搜完，按匹配度排序后再逐个校验。
	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		candidates []switchCandidate
		fast       = make(chan model.Song, len(switchSources))
	)
	for _, src := range switchSources {
		if src == orig.Source {
			continue
		}
		fn := core.GetSearchFunc(src)
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			found := searchCandidates(fn, src, keyword, name, artist, orig.Duration)
			if len(found) == 0 {
				return
			}
			sortCandidates(found)
			mu.Lock()
			candidates = append(candidates, found...)
			mu.Unlock()
			best := found[0]
			if best.score >= switchHighConfidence && (orig.Duration == 0 || best.durDiff <= 3) && core.ValidatePlayable(&best.song) {
				fast <- best.song
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case song := <-fast:
		return withCover(song, orig)
	case <-done:
	}
	select {
	case song := <-fast: // 所有平台结束前最后一个高匹配结果
		return withCover(song, orig)
	default:
	}
	if len(candidates) == 0 {
		return nil
	}
	sortCandidates(candidates)
	if len(candidates) > switchValidateLimit {
		candidates = candidates[:switchValidateLimit]
	}

	// 并行校验可播放性，取排序最靠前的可用结果
	valid := make([]bool, len(candidates))
	for i := range candidates {
		wg.Add(1)
		go func() {
			defer wg.Done()
			valid[i] = core.ValidatePlayable(&candidates[i].song)
		}()
	}
	wg.Wait()
	for i, ok := range valid {
		if ok {
			return withCover(candidates[i].song, orig)
		}
	}
	return nil
}

func searchCandidates(fn core.SearchFunc, src, keyword, name, artist string, duration int) []switchCandidate {
	res, err := withTimeout(6*time.Second, func() ([]model.Song, error) { return fn(keyword) })
	if (err != nil || len(res) == 0) && artist != "" {
		res, _ = withTimeout(6*time.Second, func() ([]model.Song, error) { return fn(name) })
	}
	var out []switchCandidate
	for i := 0; i < len(res) && i < switchCandidatesPerSource; i++ {
		cand := res[i]
		cand.Source = src
		score := core.CalcSongSimilarity(name, artist, cand.Name, cand.Artist)
		if score <= 0 {
			continue
		}
		durDiff := 0
		if duration > 0 && cand.Duration > 0 {
			if !core.IsDurationClose(duration, cand.Duration) {
				continue
			}
			durDiff = core.IntAbs(duration - cand.Duration)
		}
		out = append(out, switchCandidate{song: cand, score: score, durDiff: durDiff})
	}
	return out
}

// sortCandidates 按匹配度、时长差排序；同分时按 switchSources 的平台优先级
func sortCandidates(c []switchCandidate) {
	rank := make(map[string]int, len(switchSources))
	for i, src := range switchSources {
		rank[src] = i
	}
	sort.SliceStable(c, func(i, j int) bool {
		a, b := c[i], c[j]
		if a.score != b.score {
			return a.score > b.score
		}
		if a.durDiff != b.durDiff {
			return a.durDiff < b.durDiff
		}
		return rank[a.song.Source] < rank[b.song.Source]
	})
}

// withCover 换源后的歌曲没有封面时沿用原歌曲的封面
func withCover(song, orig model.Song) *model.Song {
	if song.Cover == "" {
		song.Cover = orig.Cover
	}
	return &song
}
