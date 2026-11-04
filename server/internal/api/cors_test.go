package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalHost(t *testing.T) {
	for host, want := range map[string]bool{
		"127.0.0.1:37618":    true,
		"localhost:5173":     true,
		"192.168.1.10:8686":  true,
		"[::1]:8686":         true,
		"my-pc:8686":         true,
		"my-pc.local:8686":   true,
		"evil.example.com":   false,
		"8.8.8.8:80":         false,
		"rebind.attacker.cn": false,
	} {
		if got := localHost(host); got != want {
			t.Errorf("localHost(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestWebsitesCannotCallLocalAPI(t *testing.T) {
	h := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	send := func(host, origin string) int {
		r := httptest.NewRequest(http.MethodPost, "/api/settings", nil)
		r.Host = host
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	if c := send("127.0.0.1:37618", ""); c != http.StatusOK {
		t.Errorf("同源请求被拒绝: %d", c)
	}
	if c := send("127.0.0.1:37618", "http://127.0.0.1:37618"); c != http.StatusOK {
		t.Errorf("本机来源被拒绝: %d", c)
	}
	if c := send("127.0.0.1:37618", "https://evil.example.com"); c != http.StatusForbidden {
		t.Errorf("外部网页未被拒绝: %d", c)
	}
	if c := send("rebind.example.com:37618", "http://rebind.example.com:37618"); c != http.StatusForbidden {
		t.Errorf("DNS 重绑定未被拒绝: %d", c)
	}
}
