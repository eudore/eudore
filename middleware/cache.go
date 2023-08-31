package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

type cache struct {
	sync.Mutex
	dura       time.Duration
	context    context.Context
	getKeyFunc func(eudore.Context) string
	waits      map[string]*sync.WaitGroup
	CacheStore
}

type CacheStore interface {
	Load(string) *CacheData
	Store(string, *CacheData)
}

/*
NewCacheFunc 函数创建一个缓存中间件，对Get请求具有缓存和SingleFlight双重效果，
无法获得中间件之前的响应header数据。

options:

	context.Context               =>    控制默认cacheMap清理过期数据的生命周期。
	time.Duration                 =>    请求数据缓存时间，默认秒。
	func(eudore.Context) string   =>    自定义缓存key，为空则跳过缓存。
	CacheStore			          =>    缓存存储对象。
*/
func NewCacheFunc(args ...any) eudore.HandlerFunc {
	c := &cache{
		dura:    DefaultCacheSaveTime,
		context: context.Background(),
		getKeyFunc: func(ctx eudore.Context) string {
			if ctx.Method() != eudore.MethodGet || ctx.GetHeader(eudore.HeaderUpgrade) != "" {
				return ""
			}
			return ctx.Request().URL.RequestURI()
		},
		waits: make(map[string]*sync.WaitGroup),
	}
	for _, i := range args {
		switch val := i.(type) {
		case time.Duration:
			c.dura = val
		case context.Context:
			c.context = val
		case func(eudore.Context) string:
			c.getKeyFunc = val
		case CacheStore:
			c.CacheStore = val
		}
	}
	if c.CacheStore == nil {
		c.CacheStore = newCacheMap(c.context, c.dura)
	}
	return c.Handle
}

func (cache *cache) Handle(ctx eudore.Context) {
	key := cache.getKeyFunc(ctx)
	if key == "" || ctx.GetHeader(eudore.HeaderConnection) == "Upgrade" {
		return
	}

	fullkey := fmt.Sprintf("%s:%s:%s", key,
		ctx.GetHeader(eudore.HeaderAccept),
		ctx.GetHeader(eudore.HeaderAcceptEncoding),
	)

	var wait *sync.WaitGroup
	var ok bool
	for {
		// load cache
		data := cache.Load(fullkey)
		if data != nil {
			data.writeData(ctx)
			ctx.SetParam("cache", key)
			ctx.End()
			return
		}

		// cas
		cache.Lock()
		wait, ok = cache.waits[fullkey]
		if !ok {
			wait = new(sync.WaitGroup)
			wait.Add(1)
			cache.waits[fullkey] = wait
			cache.Unlock()
			break
		}
		cache.Unlock()
		wait.Wait()
	}

	now := time.Now()
	w := &responseWriterCache{
		ResponseWriter: ctx.Response(),
		CacheHeader: http.Header{
			eudore.HeaderLastModified: {now.UTC().Format(http.TimeFormat)},
		},
	}
	ctx.SetResponse(w)
	ctx.Next()
	ctx.SetResponse(w.ResponseWriter)
	cache.Store(fullkey, &CacheData{
		Expired:      now.Add(cache.dura),
		ModifiedTime: w.CacheHeader.Get(eudore.HeaderLastModified),
		Status:       w.Status(),
		Header:       w.CacheHeader,
		Body:         w.CacheData.Bytes(),
	})

	cache.Lock()
	delete(cache.waits, fullkey)
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
			m.Map.Range(func(key, value any) bool {
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

// responseWriterCache 对象记录返回的响应数据
//
// Upgrade请求不会进入cache处理，push处理仅push主请求，缓存请求不push无明显影响。
type responseWriterCache struct {
	eudore.ResponseWriter
	CacheData    bytes.Buffer
	CacheHeader  http.Header
	ModifiedTime string
}

// Write 方法实现ResponseWriter中的Write方法。
func (w *responseWriterCache) Write(data []byte) (int, error) {
	if w.Size() == 0 {
		h := w.ResponseWriter.Header()
		h.Add(eudore.HeaderLastModified, w.ModifiedTime)
		for k, v := range w.CacheHeader {
			h[k] = v
		}
	}
	w.CacheData.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriterCache) WriteString(data string) (int, error) {
	if w.Size() == 0 {
		h := w.ResponseWriter.Header()
		h.Add(eudore.HeaderLastModified, w.ModifiedTime)
		for k, v := range w.CacheHeader {
			h[k] = v
		}
	}
	w.CacheData.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}

// Header 方法返回响应设置的header。
func (w *responseWriterCache) Header() http.Header {
	return w.CacheHeader
}

// CacheData 定义缓存的数据类型。
type CacheData struct {
	Expired      time.Time
	ModifiedTime string
	Status       int
	Header       http.Header
	Body         []byte
}

// writeData 方法将cache响应数据写入到请求响应。
func (w *CacheData) writeData(ctx eudore.Context) {
	ctx.SetHeader(eudore.HeaderXEudoreCache, "true")
	if ctx.GetHeader(eudore.HeaderIfModifiedSince) == w.ModifiedTime {
		ctx.SetHeader(eudore.HeaderLastModified, w.ModifiedTime)
		ctx.WriteHeader(eudore.StatusNotModified)
		return
	}
	h := ctx.Response().Header()
	for k, v := range w.Header {
		h[k] = v
	}
	ctx.WriteHeader(w.Status)
	ctx.Write(w.Body)
}
