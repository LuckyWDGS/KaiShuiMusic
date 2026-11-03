package music

import (
	"sync"
	"time"
)

// ttlCache 是一个极简的内存缓存，用于缓存首页、榜单、歌单等变化不频繁的数据，
// 避免每次打开页面都去请求上游平台。
type ttlCache struct {
	mu    sync.Mutex
	items map[string]cacheItem
	max   int
}

type cacheItem struct {
	value   any
	expires time.Time
}

func newTTLCache(max int) *ttlCache {
	return &ttlCache{items: make(map[string]cacheItem), max: max}
}

func (c *ttlCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(it.expires) {
		delete(c.items, key)
		return nil, false
	}
	return it.value, true
}

func (c *ttlCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.max > 0 && len(c.items) >= c.max {
		now := time.Now()
		for k, it := range c.items {
			if now.After(it.expires) {
				delete(c.items, k)
			}
		}
		// 仍然超限时随机淘汰一项（map 遍历顺序随机）
		if len(c.items) >= c.max {
			for k := range c.items {
				delete(c.items, k)
				break
			}
		}
	}
	c.items[key] = cacheItem{value: value, expires: time.Now().Add(ttl)}
}

// cached 读取缓存，未命中时调用 load 并在成功后写入缓存。
func cached[T any](c *ttlCache, key string, ttl time.Duration, load func() (T, error)) (T, error) {
	if v, ok := c.Get(key); ok {
		if t, ok := v.(T); ok {
			return t, nil
		}
	}
	v, err := load()
	if err != nil {
		return v, err
	}
	c.Set(key, v, ttl)
	return v, nil
}
