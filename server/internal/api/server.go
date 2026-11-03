// Package api 提供给前端（PC / 手机 / 电视）的 JSON 接口。
// 所有端共用同一套接口，前端只负责界面差异。
package api

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/library"
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/music"
	"github.com/guohuiyuan/go-music-dl/core"
	"github.com/guohuiyuan/music-lib/model"
)

const Version = "0.1.0"

type Server struct {
	music   *music.Service
	lib     *library.Store
	static  fs.FS
	dataDir string
	audio   *audioCache
}

func New(svc *music.Service, lib *library.Store, static fs.FS, dataDir string) *Server {
	return &Server{music: svc, lib: lib, static: static, dataDir: dataDir, audio: newAudioCache(6)}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "version": Version})
	})
	mux.HandleFunc("GET /api/sources", s.handleSources)

	// 推荐 / 音乐馆
	mux.HandleFunc("GET /api/home", s.handleHome)
	mux.HandleFunc("GET /api/feed", s.handleFeed)
	mux.HandleFunc("GET /api/stations", s.handleStations)
	mux.HandleFunc("GET /api/station", s.handleStation)
	mux.HandleFunc("GET /api/categories", s.handleCategories)
	mux.HandleFunc("GET /api/category_playlists", s.handleCategoryPlaylists)
	mux.HandleFunc("GET /api/toplists", s.handleTopLists)
	mux.HandleFunc("GET /api/detail", s.handleDetail)
	mux.HandleFunc("GET /api/search", s.handleSearch)

	// 播放
	mux.HandleFunc("POST /api/resolve", s.handleResolve)
	mux.HandleFunc("POST /api/lyric", s.handleLyric)
	mux.HandleFunc("GET /api/stream", s.handleStream)
	mux.HandleFunc("GET /api/cover", s.handleCover)

	// 我的（本地音乐库，无需登录）
	mux.HandleFunc("GET /api/library", s.handleLibrary)
	mux.HandleFunc("POST /api/library/like", s.handleLike)
	mux.HandleFunc("POST /api/library/history", s.handleHistory)
	mux.HandleFunc("DELETE /api/library/history", s.handleClearHistory)
	mux.HandleFunc("POST /api/library/playlist", s.handleSavePlaylist)
	mux.HandleFunc("POST /api/library/created", s.handleCreatePlaylist)
	mux.HandleFunc("DELETE /api/library/created/{id}", s.handleDeletePlaylist)
	mux.HandleFunc("POST /api/library/created/{id}/songs", s.handlePlaylistSong)

	// 下载 / 本地文件
	mux.HandleFunc("POST /api/download", s.handleDownload)
	mux.HandleFunc("GET /api/downloads", s.handleDownloads)
	mux.HandleFunc("GET /api/downloads/file", s.handleLocalFile)
	mux.HandleFunc("POST /api/downloads/open", s.handleOpenFolder)

	mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	mux.HandleFunc("POST /api/settings", s.handleSaveSettings)

	mux.Handle("/", s.spaHandler())
	return withCORS(mux)
}

// ---------- 推荐 / 音乐馆 ----------

func (s *Server) handleSources(w http.ResponseWriter, r *http.Request) {
	type src struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	list := func(ids []string) []src {
		out := make([]src, 0, len(ids))
		for _, id := range ids {
			out = append(out, src{id, core.GetSourceDescription(id)})
		}
		return out
	}
	writeJSON(w, map[string]any{
		"song":     list(music.DefaultSongSources),
		"playlist": list(music.DefaultPlaylistSources),
		"album":    list(music.DefaultAlbumSources),
		"toplist":  list(music.TopListSources),
		"all":      list(core.GetAllSourceNames()),
	})
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	home, err := s.music.Home()
	respond(w, home, err)
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	songs, err := s.music.Feed(page)
	respond(w, songs, err)
}

func (s *Server) handleStations(w http.ResponseWriter, r *http.Request) {
	st, err := s.music.Stations()
	respond(w, st, err)
}

func (s *Server) handleStation(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	songs, err := s.music.StationSongs(q.Get("kind"), q.Get("id"))
	respond(w, songs, err)
}

func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := s.music.Categories(queryOr(r, "source", music.FeedSource))
	respond(w, cats, err)
}

func (s *Server) handleCategoryPlaylists(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pls, err := s.music.CategoryPlaylists(queryOr(r, "source", music.FeedSource), q.Get("id"), page)
	respond(w, pls, err)
}

func (s *Server) handleTopLists(w http.ResponseWriter, r *http.Request) {
	lists, err := s.music.TopLists(queryOr(r, "source", music.FeedSource))
	respond(w, lists, err)
}

func (s *Server) handleDetail(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("source") == "" || q.Get("id") == "" {
		writeError(w, http.StatusBadRequest, "缺少参数")
		return
	}
	songs, err := s.music.Detail(queryOr(r, "kind", "playlist"), q.Get("source"), q.Get("id"))
	respond(w, songs, err)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var sources []string
	if v := strings.TrimSpace(q.Get("sources")); v != "" {
		sources = strings.Split(v, ",")
	}
	res, err := s.music.Search(q.Get("q"), queryOr(r, "type", "song"), sources)
	respond(w, res, err)
}

// ---------- 播放 ----------

func (s *Server) handleResolve(w http.ResponseWriter, r *http.Request) {
	var song model.Song
	if !readJSON(w, r, &song) {
		return
	}
	res, err := s.music.Resolve(song)
	respond(w, res, err)
}

func (s *Server) handleLyric(w http.ResponseWriter, r *http.Request) {
	var song model.Song
	if !readJSON(w, r, &song) {
		return
	}
	writeJSON(w, map[string]string{"lyric": s.music.Lyric(&song)})
}

func (s *Server) handleCover(w http.ResponseWriter, r *http.Request) {
	u := strings.TrimSpace(r.URL.Query().Get("url"))
	if u == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data, contentType, err := core.FetchBytesWithMime(u, r.URL.Query().Get("source"))
	if err != nil || len(data) == 0 {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if contentType == "" {
		contentType = "image/jpeg"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

// ---------- 我的 ----------

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.lib.Snapshot())
}

func (s *Server) handleLike(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Song  model.Song `json:"song"`
		Liked bool       `json:"liked"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	data, err := s.lib.SetLike(body.Song, body.Liked)
	respond(w, data, err)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	var song model.Song
	if !readJSON(w, r, &song) {
		return
	}
	data, err := s.lib.AddHistory(song)
	respond(w, data, err)
}

func (s *Server) handleClearHistory(w http.ResponseWriter, r *http.Request) {
	data, err := s.lib.ClearHistory()
	respond(w, data, err)
}

func (s *Server) handleSavePlaylist(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Playlist library.SavedPlaylist `json:"playlist"`
		Saved    bool                  `json:"saved"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	data, err := s.lib.SetPlaylist(body.Playlist, body.Saved)
	respond(w, data, err)
}

func (s *Server) handleCreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	data, err := s.lib.CreatePlaylist(body.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, data)
}

func (s *Server) handleDeletePlaylist(w http.ResponseWriter, r *http.Request) {
	data, err := s.lib.DeletePlaylist(r.PathValue("id"))
	respond(w, data, err)
}

func (s *Server) handlePlaylistSong(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Song model.Song `json:"song"`
		Add  bool       `json:"add"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	data, err := s.lib.SetPlaylistSong(r.PathValue("id"), body.Song, body.Add)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, data)
}

// ---------- 下载 ----------

var audioExts = map[string]bool{".mp3": true, ".flac": true, ".m4a": true, ".ogg": true, ".wav": true, ".aac": true, ".wma": true}

func (s *Server) downloadDir() string {
	return core.DownloadDirOrDefault(core.GetWebSettings().DownloadDir)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	var song model.Song
	if !readJSON(w, r, &song) {
		return
	}
	// 先解析可播放音源，保证下载到的是完整歌曲而不是试听片段
	if res, err := s.music.Resolve(song); err == nil {
		song = res.Song
	}
	settings := core.GetWebSettings()
	dedup, _ := core.LoadDownloadDedupSet()
	result, err := core.DownloadWithDedupCheckWithTemplate(&song, s.downloadDir(), true, true, settings.DownloadFilenameTemplate, dedup)
	if err != nil {
		writeError(w, http.StatusBadGateway, "下载失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]any{"filename": result.Filename, "skipped": result.Skipped, "warning": result.Warning})
}

type localFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"mtime"`
}

func (s *Server) handleDownloads(w http.ResponseWriter, r *http.Request) {
	dir := s.downloadDir()
	files := []localFile{}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !audioExts[strings.ToLower(filepath.Ext(p))] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		files = append(files, localFile{Name: filepath.ToSlash(rel), Size: info.Size(), ModTime: info.ModTime().Unix()})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].ModTime > files[j].ModTime })
	writeJSON(w, map[string]any{"dir": dir, "files": files})
}

func (s *Server) handleLocalFile(w http.ResponseWriter, r *http.Request) {
	dir, _ := filepath.Abs(s.downloadDir())
	p, _ := filepath.Abs(filepath.Join(dir, filepath.FromSlash(r.URL.Query().Get("name"))))
	if !strings.HasPrefix(p, dir+string(filepath.Separator)) || !audioExts[strings.ToLower(filepath.Ext(p))] {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, p)
}

func (s *Server) handleOpenFolder(w http.ResponseWriter, r *http.Request) {
	dir, _ := filepath.Abs(s.downloadDir())
	_ = os.MkdirAll(dir, 0o755)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	_ = cmd.Start()
	writeJSON(w, map[string]string{"dir": dir})
}

// ---------- 设置 ----------

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	dir, _ := filepath.Abs(s.downloadDir())
	writeJSON(w, map[string]any{"downloadDir": dir, "dataDir": s.dataDir, "version": Version})
}

func (s *Server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DownloadDir string `json:"downloadDir"`
	}
	if !readJSON(w, r, &body) {
		return
	}
	settings := core.GetWebSettings()
	settings.DownloadDir = strings.TrimSpace(body.DownloadDir)
	if err := core.SaveWebSettings(settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.handleGetSettings(w, r)
}

// ---------- 静态文件 ----------

// spaHandler 提供前端页面；找不到的路径回退到 index.html，交给前端路由处理。
func (s *Server) spaHandler() http.Handler {
	fileServer := http.FileServerFS(s.static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(s.static, name); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
			name = "index.html"
		}
		if name == "index.html" {
			w.Header().Set("Cache-Control", "no-cache")
		} else if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		fileServer.ServeHTTP(w, r)
	})
}

// ---------- 工具 ----------

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func queryOr(r *http.Request, key, def string) string {
	if v := strings.TrimSpace(r.URL.Query().Get(key)); v != "" {
		return v
	}
	return def
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return false
	}
	return true
}

func respond(w http.ResponseWriter, v any, err error) {
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, v)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
