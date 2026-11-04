package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/guohuiyuan/go-music-dl/core"
	"github.com/guohuiyuan/music-lib/model"
	"github.com/guohuiyuan/music-lib/soda"
)

// handleStream 代理音频流供 <audio> 播放，支持 Range 拖动。
// 逻辑移植自 go-music-dl 的 /download?stream=1。
func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	song := &model.Song{
		ID:     q.Get("id"),
		Source: q.Get("source"),
		Name:   q.Get("name"),
		Artist: q.Get("artist"),
		Album:  q.Get("album"),
		Cover:  q.Get("cover"),
	}
	if raw := q.Get("extra"); raw != "" {
		_ = json.Unmarshal([]byte(raw), &song.Extra)
	}
	if song.ID == "" || song.Source == "" {
		http.Error(w, "missing params", http.StatusBadRequest)
		return
	}

	var downloadURL string
	if song.Source == "soda" {
		info, err := s.music.SodaInfo(song)
		if err != nil {
			http.Error(w, "soda info error", http.StatusBadGateway)
			return
		}
		if strings.TrimSpace(info.PlayAuth) != "" {
			// 加密音频需要整段下载后解密，缓存最近几首，避免拖动进度时重复下载
			data, err := s.audio.load(song.Source+":"+song.ID, func() ([]byte, error) { return fetchSodaAudio(info) })
			if err != nil {
				http.Error(w, "soda stream error", http.StatusBadGateway)
				return
			}
			w.Header().Set("Content-Type", core.AudioMimeByExt(core.DetectAudioExt(data)))
			http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
			return
		}
		// 明文音频直接流式转发，起播更快
		downloadURL = info.URL
	} else {
		u, err := s.music.PlayURL(song)
		if err != nil {
			http.Error(w, "failed to get url", http.StatusNotFound)
			return
		}
		downloadURL = u
	}

	rangeHeader := r.Header.Get("Range")
	// 汽水 CDN 原生支持 Range，直接转发；其它平台先探测，处理不支持 Range 的情况
	if song.Source != "soda" {
		rf, handled, err := core.NewSourceRangeFetch(downloadURL, song.Source, rangeHeader)
		if err != nil {
			http.Error(w, "upstream range error", http.StatusBadGateway)
			return
		}
		if handled {
			ext := strings.TrimPrefix(strings.ToLower(rf.Ext), ".")
			if ext == "" {
				ext = "mp3"
			}
			w.Header().Set("Content-Type", core.AudioMimeByExt(ext))
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", strconv.FormatInt(rf.ContentLength, 10))
			if rf.ContentRange != "" {
				w.Header().Set("Content-Range", rf.ContentRange)
			}
			w.WriteHeader(rf.StatusCode)
			_ = rf.WriteTo(w)
			return
		}
	}

	req, err := core.BuildSourceRequest(http.MethodGet, downloadURL, song.Source, rangeHeader)
	if err != nil {
		http.Error(w, "upstream request error", http.StatusBadGateway)
		return
	}
	req = req.WithContext(r.Context())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "upstream stream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		switch k {
		case "Transfer-Encoding", "Date", "Access-Control-Allow-Origin", "Set-Cookie", "Content-Disposition":
		default:
			w.Header()[k] = v
		}
	}
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if ct == "" || strings.HasPrefix(ct, "application/octet-stream") || strings.HasPrefix(ct, "text/") {
		ext := core.DetectAudioExtByContentType(ct)
		if ext == "" {
			if u, err := url.Parse(downloadURL); err == nil {
				ext = strings.TrimPrefix(strings.ToLower(path.Ext(u.Path)), ".")
			}
		}
		if ext == "" {
			ext = "mp3"
		}
		w.Header().Set("Content-Type", core.AudioMimeByExt(ext))
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func fetchSodaAudio(info *soda.DownloadInfo) ([]byte, error) {
	req, err := core.BuildSourceRequest(http.MethodGet, info.URL, "soda", "")
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 2 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// 公开试听路径返回明文音频（无 play_auth），会员整曲需要解密
	if strings.TrimSpace(info.PlayAuth) == "" {
		return data, nil
	}
	return soda.DecryptAudio(data, info.PlayAuth)
}

// audioCache 缓存最近播放的需要整段下载的音频（如汽水），按插入顺序淘汰。
type audioCache struct {
	mu      sync.Mutex
	max     int
	order   []string
	items   map[string][]byte
	loading map[string]*sync.WaitGroup
}

func newAudioCache(max int) *audioCache {
	return &audioCache{max: max, items: map[string][]byte{}, loading: map[string]*sync.WaitGroup{}}
}

func (c *audioCache) load(key string, fetch func() ([]byte, error)) ([]byte, error) {
	c.mu.Lock()
	if data, ok := c.items[key]; ok {
		c.mu.Unlock()
		return data, nil
	}
	// 同一首歌并发请求（浏览器常会同时发多个 Range 请求）只下载一次
	if wg, ok := c.loading[key]; ok {
		c.mu.Unlock()
		wg.Wait()
		c.mu.Lock()
		data, ok := c.items[key]
		c.mu.Unlock()
		if ok {
			return data, nil
		}
		return fetch()
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	c.loading[key] = wg
	c.mu.Unlock()

	data, err := fetch()

	c.mu.Lock()
	delete(c.loading, key)
	if err == nil {
		c.items[key] = data
		c.order = append(c.order, key)
		for len(c.order) > c.max {
			delete(c.items, c.order[0])
			c.order = c.order[1:]
		}
	}
	c.mu.Unlock()
	wg.Done()
	return data, err
}
