package music

import (
	"context"
	"testing"
	"time"

	"github.com/guohuiyuan/music-lib/model"
)

func TestFirstComeEmitsFastestFirst(t *testing.T) {
	delays := map[string]time.Duration{"slow": 300 * time.Millisecond, "fast": 0, "mid": 100 * time.Millisecond, "empty": 0}
	var got []string
	start := time.Now()
	var firstAt time.Duration
	err := firstCome(context.Background(), []string{"slow", "mid", "fast", "empty"}, func(src string) SearchChunk {
		time.Sleep(delays[src])
		if src == "empty" {
			return SearchChunk{Source: src}
		}
		return SearchChunk{Source: src, Songs: []model.Song{{ID: src}}}
	}, func(c SearchChunk) {
		if got == nil {
			firstAt = time.Since(start)
		}
		got = append(got, c.Source)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != "fast" || got[1] != "mid" || got[2] != "slow" {
		t.Fatalf("顺序不对: %v", got)
	}
	if firstAt > 100*time.Millisecond {
		t.Fatalf("第一批结果要等 %v，被慢平台拖住了", firstAt)
	}
}

func TestFirstComeStopsWhenCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := firstCome(ctx, []string{"a"}, func(string) SearchChunk {
		time.Sleep(time.Second)
		return SearchChunk{}
	}, func(SearchChunk) { t.Error("取消后不应再有结果") })
	if err == nil {
		t.Fatal("应返回取消错误")
	}
}

func TestCachedSharesConcurrentLoad(t *testing.T) {
	c := newTTLCache(8)
	calls := make(chan struct{}, 8)
	done := make(chan string, 4)
	for range 4 {
		go func() {
			v, _ := cached(c, "k", time.Minute, func() (string, error) {
				calls <- struct{}{}
				time.Sleep(50 * time.Millisecond)
				return "v", nil
			})
			done <- v
		}()
	}
	for range 4 {
		if v := <-done; v != "v" {
			t.Fatalf("结果不对: %q", v)
		}
	}
	if len(calls) != 1 {
		t.Fatalf("上游被调用了 %d 次", len(calls))
	}
}
