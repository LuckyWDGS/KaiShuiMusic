// Package library 保存用户的本地音乐库：喜欢的歌曲、最近播放、自建歌单、收藏的歌单。
// 不需要登录，数据保存在本机 data/library.json；手机 / 电视端连接同一个服务时可共享。
package library

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/guohuiyuan/music-lib/model"
)

const maxHistory = 300

type Data struct {
	Likes     []model.Song      `json:"likes"`
	History   []model.Song      `json:"history"`
	Playlists []SavedPlaylist   `json:"playlists"`
	Created   []CreatedPlaylist `json:"created"`
	UpdatedAt int64             `json:"updated_at"`
}

// CreatedPlaylist 用户自建歌单
type CreatedPlaylist struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Songs     []model.Song `json:"songs"`
	CreatedAt int64        `json:"created_at"`
}

// SavedPlaylist 收藏的歌单 / 专辑 / 榜单
type SavedPlaylist struct {
	model.Playlist
	Kind string `json:"kind"` // playlist | album | toplist
}

type Store struct {
	mu   sync.Mutex
	path string
	data Data
}

func Open(path string) (*Store, error) {
	s := &Store{path: path}
	b, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(b, &s.data)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	s.normalize()
	return s, nil
}

func (s *Store) normalize() {
	if s.data.Likes == nil {
		s.data.Likes = []model.Song{}
	}
	if s.data.History == nil {
		s.data.History = []model.Song{}
	}
	if s.data.Playlists == nil {
		s.data.Playlists = []SavedPlaylist{}
	}
	if s.data.Created == nil {
		s.data.Created = []CreatedPlaylist{}
	}
	for i := range s.data.Created {
		if s.data.Created[i].Songs == nil {
			s.data.Created[i].Songs = []model.Song{}
		}
	}
}

var ErrNotFound = errors.New("歌单不存在")

func (s *Store) CreatePlaylist(name string) (Data, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return s.Snapshot(), errors.New("歌单名不能为空")
	}
	now := time.Now()
	return s.update(func(d *Data) {
		d.Created = append(d.Created, CreatedPlaylist{
			ID:        strconv.FormatInt(now.UnixNano(), 36),
			Name:      name,
			Songs:     []model.Song{},
			CreatedAt: now.Unix(),
		})
	})
}

func (s *Store) DeletePlaylist(id string) (Data, error) {
	return s.update(func(d *Data) {
		out := d.Created[:0]
		for _, p := range d.Created {
			if p.ID != id {
				out = append(out, p)
			}
		}
		d.Created = out
	})
}

// SetPlaylistSong 向自建歌单添加（加到最前）或移除歌曲
func (s *Store) SetPlaylistSong(id string, song model.Song, add bool) (Data, error) {
	found := false
	data, err := s.update(func(d *Data) {
		for i := range d.Created {
			if d.Created[i].ID != id {
				continue
			}
			found = true
			songs := removeSong(d.Created[i].Songs, songKey(song))
			if add {
				songs = append([]model.Song{slim(song)}, songs...)
			}
			d.Created[i].Songs = songs
		}
	})
	if err == nil && !found {
		err = ErrNotFound
	}
	return data, err
}

func (s *Store) Snapshot() Data {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func songKey(song model.Song) string { return song.Source + ":" + song.ID }

func (s *Store) SetLike(song model.Song, liked bool) (Data, error) {
	return s.update(func(d *Data) {
		d.Likes = removeSong(d.Likes, songKey(song))
		if liked {
			d.Likes = append([]model.Song{slim(song)}, d.Likes...)
		}
	})
}

func (s *Store) AddHistory(song model.Song) (Data, error) {
	return s.update(func(d *Data) {
		d.History = append([]model.Song{slim(song)}, removeSong(d.History, songKey(song))...)
		if len(d.History) > maxHistory {
			d.History = d.History[:maxHistory]
		}
	})
}

func (s *Store) ClearHistory() (Data, error) {
	return s.update(func(d *Data) { d.History = []model.Song{} })
}

func (s *Store) SetPlaylist(pl SavedPlaylist, saved bool) (Data, error) {
	return s.update(func(d *Data) {
		key := pl.Kind + ":" + pl.Source + ":" + pl.ID
		out := d.Playlists[:0]
		for _, p := range d.Playlists {
			if p.Kind+":"+p.Source+":"+p.ID != key {
				out = append(out, p)
			}
		}
		d.Playlists = out
		if saved {
			d.Playlists = append([]SavedPlaylist{pl}, d.Playlists...)
		}
	})
}

func (s *Store) update(fn func(*Data)) (Data, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.data)
	s.normalize()
	s.data.UpdatedAt = time.Now().Unix()
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return s.data, err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return s.data, err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return s.data, err
	}
	return s.data, os.Rename(tmp, s.path)
}

func removeSong(list []model.Song, key string) []model.Song {
	out := make([]model.Song, 0, len(list))
	for _, song := range list {
		if songKey(song) != key {
			out = append(out, song)
		}
	}
	return out
}

// slim 去掉一次性的播放地址，只保留识别歌曲所需的信息
func slim(song model.Song) model.Song {
	song.URL = ""
	return song
}
