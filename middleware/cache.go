package middleware

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

type cache struct {
	sync.Mutex
	dura    time.Duration
	context context.Context
	waits   map[string]*sync.WaitGroup
	cacheStore
}

type cacheStore interface {
	Load(string) *CacheData
	Store(string, *CacheData)
}

// NewCacheFunc 函数创建一个缓存中间件，对Get请求具有缓存和SingleFlight双重效果，无法获得中间件之前的响应header数据。
//
// options:
//
// context.Context               =>    控制默认cacheMap清理过期数据的生命周期
//
// time.Duration                 =>    请求数据缓存时间，默认秒
//
// cacheStore			         =>    缓存存储对象
func NewCacheFunc(args ...interface{}) eudore.HandlerFunc {
	c := &cache{
		dura:    time.Second,
		context: context.Background(),
		waits:   make(map[string]*sync.WaitGroup),
	}
	for _, i := range args {
		switch val := i.(type) {
		case time.Duration:
			c.dura = val
		case context.Context:
			c.context = val
		case cacheStore:
			c.cacheStore = val
		}
	}
	if c.cacheStore == nil {
		c.cacheStore = newCacheMap(c.context, c.dura)
	}
	return c.Handle
}

func (cache *cache) Handle(ctx eudore.Context) {
	if ctx.Method() != eudore.MethodGet || ctx.GetHeader(eudore.HeaderUpgrade) != "" {
		return
	}

	key := ctx.Request().URL.RequestURI()
	var wait *sync.WaitGroup
	var ok bool
	for {
		// load cache
		data := cache.Load(key)
		if data != nil {
			data.writeData(ctx.Response())
			ctx.End()
			return
		}

		// cas
		cache.Lock()
		wait, ok = cache.waits[key]
		if !ok {
			wait = new(sync.WaitGroup)
			wait.Add(1)
			cache.waits[key] = wait
			cache.Unlock()
			break
		}
		cache.Unlock()
		wait.Wait()
	}

	w := ctx.Response()
	resp := &cacheResponset{
		ResponseWriter: w,
		header:         make(http.Header),
	}
	ctx.SetResponse(resp)
	ctx.Next()
	ctx.SetResponse(w)
	cache.Store(key, &CacheData{
		Expired: time.Now().Add(cache.dura),
		Status:  w.Status(),
		Header:  w.Header(),
		Body:    resp.Buffer.Bytes(),
	})

	cache.Lock()
	delete(cache.waits, key)
	wait.Done()
	cache.Unlock()
}

type cacheMap struct {
	sync.Map
}

func newCacheMap(ctx context.Context, t time.Duration) *cacheMap {
	c := new(cacheMap)
	go c.Run(ctx, t*2)
	return c
}

func (m *cacheMap) Load(key string) *CacheData {
	data, ok := m.Map.Load(key)
	if !ok {
		return nil
	}
	item := data.(*CacheData)
	if time.Now().After(item.Expired) {
		m.Map.Delete(key)
		return nil
	}
	return item
}

func (m *cacheMap) Store(key string, val *CacheData) {
	m.Map.Store(key, val)
}

func (m *cacheMap) Run(ctx context.Context, t time.Duration) {
	for {
		select {
		case now := <-time.After(t):
			m.Map.Range(func(key, value interface{}) bool {
				item := value.(*CacheData)
				if now.After(item.Expired) {
					m.Map.Delete(key)
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}

// cacheResponset 对象记录返回的响应数据
//
// Upgrade请求不会进入cache处理，push处理仅push主请求，缓存请求不push无明显影响。
type cacheResponset struct {
	eudore.ResponseWriter
	header http.Header
	bytes.Buffer
}

// Write 方法实现ResponseWriter中的Write方法。
func (w *cacheResponset) Write(data []byte) (int, error) {
	if w.Size() == 0 {
		h := w.ResponseWriter.Header()
		for k, v := range w.header {
			h[k] = v
		}
	}
	w.Buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

// Header 方法返回响应设置的header。
func (w *cacheResponset) Header() http.Header {
	return w.header
}

// CacheData 定义缓存的数据类型。
type CacheData struct {
	Expired time.Time
	Status  int
	Header  http.Header
	Body    []byte
}

// writeData 方法将cache响应数据写入到请求响应。
func (w *CacheData) writeData(resp eudore.ResponseWriter) {
	resp.WriteHeader(w.Status)
	h := resp.Header()
	for k, v := range w.Header {
		h[k] = v
	}
	resp.Write(w.Body)
}
