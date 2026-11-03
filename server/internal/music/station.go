package music

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/guohuiyuan/music-lib/model"
)

// Station 听歌模式里的一个频道：来源可以是汽水的榜单 / 推荐歌单（toplist），或歌单分类（category）。
// kind 为 local 的频道（如"熟悉模式"）由前端用本地的喜欢和历史生成。
type Station struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"` // toplist | category | local
	Name  string `json:"name"`
	Desc  string `json:"desc,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Cover string `json:"cover,omitempty"`
}

type Stations struct {
	Common []Station `json:"common"`
	More   []Station `json:"more"`
}

// 常用模式：固定频道。汽水的推荐歌单免登录时常常只返回几首，所以优先用分类和官方榜单。
var commonStations = []Station{
	{ID: "familiar", Kind: "local", Name: "熟悉模式", Icon: "heart"},
	{ID: "7060812597884869927", Kind: "toplist", Name: "新鲜模式", Icon: "sparkles"},
	{ID: "69", Kind: "category", Name: "Chill 放松", Icon: "leaf"},
	{ID: "7036274230471712007", Kind: "toplist", Name: "热歌漫游", Icon: "flame"},
	{ID: "45", Kind: "category", Name: "图书馆", Icon: "library"},
	{ID: "40", Kind: "category", Name: "专注模式", Icon: "book"},
	{ID: "7415959718721494311", Kind: "toplist", Name: "摸鱼", Icon: "fish"},
	{ID: "7031096410419283998", Kind: "toplist", Name: "深夜 EMO", Icon: "moon"},
}

// 探索更多新模式：按曲风 / 场景分类
var moreStations = []Station{
	{ID: "14", Kind: "category", Name: "流行", Desc: "当下最火的流行金曲"},
	{ID: "20", Kind: "category", Name: "国风", Desc: "大气旋律，惊艳世界"},
	{ID: "8", Kind: "category", Name: "华语", Desc: "耳熟能详的华语好歌"},
	{ID: "9", Kind: "category", Name: "欧美", Desc: "原汁原味的欧美节奏"},
	{ID: "18", Kind: "category", Name: "民谣", Desc: "一把吉他，一段故事"},
	{ID: "38", Kind: "category", Name: "说唱", Desc: "节奏拉满，态度在线"},
	{ID: "16", Kind: "category", Name: "电子", Desc: "律动不停，开车必备"},
	{ID: "15", Kind: "category", Name: "摇滚", Desc: "热血沸腾的现场感"},
	{ID: "19", Kind: "category", Name: "R&B", Desc: "慵懒随性的转音"},
	{ID: "7061475546400005410", Kind: "toplist", Name: "欧美榜", Desc: "欧美热度最高的单曲"},
}

const stationsDiskFile = "stations.json"

// Stations 返回听歌模式的频道列表；"更多"频道的封面取该分类第一个歌单的封面。
// 结果同时保存到磁盘，程序重启后无需再等待上游即可显示。
func (s *Service) Stations() (*Stations, error) {
	return cached(s.cache, "stations", 6*time.Hour, func() (*Stations, error) {
		var disk Stations
		if s.loadDisk(stationsDiskFile, 3*24*time.Hour, &disk) && len(disk.More) == len(moreStations) {
			// 磁盘缓存只用于秒开，后台顺便刷新一份新的
			go func() {
				if st, err := s.fetchStations(); err == nil {
					s.cache.Set("stations", st, 6*time.Hour)
				}
			}()
			disk.Common = commonStations
			return &disk, nil
		}
		return s.fetchStations()
	})
}

func (s *Service) fetchStations() (*Stations, error) {
	more := make([]Station, len(moreStations))
	copy(more, moreStations)
	lists, _ := s.TopLists(FeedSource)
	var wg sync.WaitGroup
	for i := range more {
		st := &more[i]
		if st.Kind == "toplist" {
			for _, tl := range lists {
				if tl.ID == st.ID {
					st.Cover = tl.Cover
				}
			}
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if pls, err := s.CategoryPlaylists(FeedSource, st.ID, 1); err == nil && len(pls) > 0 {
				st.Cover = pls[0].Cover
			}
		}()
	}
	wg.Wait()
	st := &Stations{Common: commonStations, More: more}
	s.saveDisk(stationsDiskFile, st)
	return st, nil
}

// StationSongs 返回频道的歌曲（已随机打散，第一首优先选无需换源的非会员歌曲，点开即播）。
func (s *Service) StationSongs(kind, id string) ([]model.Song, error) {
	songs, err := cached(s.cache, "station:"+kind+":"+id, 30*time.Minute, func() ([]model.Song, error) {
		return s.fetchStationSongs(kind, id)
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Song, len(songs))
	copy(out, songs)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	s.putFastFirst(out, "station:"+kind+":"+id)
	return out, nil
}

// putFastFirst 把预热时确认可直接播放的歌曲放到第一首
func (s *Service) putFastFirst(songs []model.Song, key string) {
	v, ok := s.fastStart.Load(key)
	if !ok {
		return
	}
	for i, song := range songs {
		if song.ID == v.(string) {
			songs[0], songs[i] = songs[i], songs[0]
			return
		}
	}
}

// findFastStart 在前几首里找一首汽水能直接播放完整版的歌（不用换源，点开即播），记录下来。
// 只查汽水播放地址（约 1 秒，结果缓存），不触发较慢的换源搜索。
func (s *Service) findFastStart(songs []model.Song, key string) {
	if _, ok := s.fastStart.Load(key); ok {
		return
	}
	for i := 0; i < len(songs) && i < 8; i++ {
		song := songs[i]
		if song.Source != "soda" {
			continue
		}
		if info, err := s.SodaInfo(&song); err == nil && !info.Preview {
			s.fastStart.Store(key, song.ID)
			return
		}
	}
}

// WarmStations 后台预热：频道列表和所有频道的歌曲，打开听歌模式、点选频道时无需等待
func (s *Service) WarmStations() {
	st, err := s.Stations()
	if err != nil {
		return
	}
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for _, list := range [][]Station{st.Common, st.More} {
		for _, station := range list {
			if station.Kind == "local" {
				continue
			}
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer func() { <-sem; wg.Done() }()
				if songs, err := s.StationSongs(station.Kind, station.ID); err == nil {
					s.findFastStart(songs, "station:"+station.Kind+":"+station.ID)
				}
			}()
		}
	}
	wg.Wait()
}

func (s *Service) fetchStationSongs(kind, id string) ([]model.Song, error) {
	var songs []model.Song
	switch kind {
	case "toplist":
		list, err := s.Detail("toplist", FeedSource, id)
		if err != nil {
			return nil, err
		}
		songs = append(songs, list...)
	case "category":
		pls, err := s.CategoryPlaylists(FeedSource, id, 1)
		if err != nil {
			return nil, err
		}
		if len(pls) > 4 {
			pls = pls[:4]
		}
		groups := make([][]model.Song, len(pls))
		var wg sync.WaitGroup
		for i, pl := range pls {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if list, err := s.Detail("playlist", FeedSource, pl.ID); err == nil {
					groups[i] = list
				}
			}()
		}
		wg.Wait()
		seen := map[string]bool{}
		for _, song := range interleave(groups, 0) {
			key := strings.ToLower(song.Name + "|" + song.Artist)
			if !seen[key] {
				seen[key] = true
				songs = append(songs, song)
			}
		}
	default:
		return nil, fmt.Errorf("未知的频道类型: %s", kind)
	}
	if len(songs) == 0 {
		return nil, fmt.Errorf("这个频道暂时没有歌曲")
	}
	return songs, nil
}
