package middleware

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

type cache struct {
	sync.Mutex
	waits      map[string]*sync.WaitGroup
	storage    cacheData
	GetKeyFunc func(ctx eudore.Context) string
}

type cacheData interface {
	LoadData(key string) *cacheResponse
	SaveData(key string, val *cacheResponse)
}

// cacheResponse defines the cached data type.
type cacheResponse struct {
	Expired time.Time
	Status  int
	Header  http.Header
	Body    []byte
}

// The NewCacheFunc function creates middleware to implement
// cache response content,
// which has the dual effects of Cache and SingleFlight, except panic.
//
// This middleware caches data in memory and is used to cache API data.
// Caching files is not recommended.
//
// Skip non-Get methods, Websocket and SSE by default.
//
// Cannot get response headers before this middleware.
//
// This middleware does not support cluster mode.
// Cache requests are idempotent and do not rely on the cluster.
//
// options: [NewOptionKeyFunc] [NewOptionCacheClear].
func NewCacheFunc(dura time.Duration, options ...Option) Middleware {
	c := newCache(options)
	return func(ctx eudore.Context) {
		key := c.GetKeyFunc(ctx)
		if key == "" {
			return
		}
		fullkey := fmt.Sprintf("%s:%s:%s", key,
			formatAccept(ctx.GetHeader(eudore.HeaderAccept)),
			ctx.GetHeader(eudore.HeaderAcceptEncoding),
		)

		wait := c.load(ctx, fullkey)
		if wait == nil {
			return
		}
		defer func() {
			// waits for done and cleanup
			c.Lock()
			delete(c.waits, fullkey)
			c.Unlock()
			wait.Done()
		}()

		now := time.Now()
		w := &responseWriterCache{
			ResponseWriter: ctx.Response(),
			h: http.Header{
				eudore.HeaderLastModified: {now.UTC().Format(http.TimeFormat)},
			},
		}
		ctx.SetResponse(w)
		defer ctx.SetResponse(w.ResponseWriter)
		ctx.Next()
		c.storage.SaveData(fullkey, &cacheResponse{
			Expired: now.Add(dura),
			Status:  w.Status(),
			Header:  w.h,
			Body:    w.w.Bytes(),
		})
	}
}

func newCache(options []Option) *cache {
	c := &cache{
		waits:   make(map[string]*sync.WaitGroup),
		storage: new(cacheMap),
		GetKeyFunc: func(ctx eudore.Context) string {
			if ctx.Method() != eudore.MethodGet ||
				ctx.GetHeader(eudore.HeaderConnection) ==
					eudore.HeaderValueUpgrade ||
				ctx.GetHeader(eudore.HeaderAccept) ==
					eudore.MimeTextEventStream {
				return ""
			}
			return ctx.Request().URL.RequestURI()
		},
	}
	applyOption(c, options)
	return c
}

func (c *cache) load(ctx eudore.Context, key string) *sync.WaitGroup {
	for {
		// load cache
		data := c.storage.LoadData(key)
		if data != nil {
			// write cache data
			headerCopy(ctx.Response().Header(), data.Header)
			ctx.WriteHeader(data.Status)
			if len(data.Body) != 0 {
				_, _ = ctx.Write(data.Body)
			}
			ctx.End()
			return nil
		}

		c.Lock()
		wait, ok := c.waits[key]
		if !ok {
			// Create the first SingleFlight request
			wait = new(sync.WaitGroup)
			wait.Add(1)
			c.waits[key] = wait
			c.Unlock()
			return wait
		}
		c.Unlock()
		// Waiting for other requests to done
		wait.Wait()
	}
}

// The formatAccept function filters invalid Accept.
func formatAccept(accept string) string {
	var accepts []string
	for _, accept := range strings.Split(accept, ",") {
		k, _, _ := strings.Cut(accept, ";")
		_, ok := DefaultCacheAllowAccept[k]
		if ok {
			accepts = append(accepts, k)
		}
	}
	return strings.Join(accepts, ",")
}

// responseWriterCache defines cached response data.
type responseWriterCache struct {
	eudore.ResponseWriter
	w bytes.Buffer
	h http.Header
}

func (w *responseWriterCache) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriterCache) Write(data []byte) (int, error) {
	if w.Size() == 0 {
		headerCopy(w.ResponseWriter.Header(), w.h)
	}
	w.w.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriterCache) WriteString(data string) (int, error) {
	if w.Size() == 0 {
		headerCopy(w.ResponseWriter.Header(), w.h)
	}
	w.w.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}

// The WriteHeader method adds the Cache-Header when writing for the first time.
func (w *responseWriterCache) WriteHeader(code int) {
	if w.Size() == 0 {
		headerCopy(w.ResponseWriter.Header(), w.h)
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterCache) Header() http.Header {
	return w.h
}

// refer: [responseWriterTimeout.Body].
func (w *responseWriterCache) Body() []byte {
	return w.w.Bytes()
}

// The Flush method is not supported.
func (w *responseWriterCache) Flush() {}

// The Hijack method is not supported.
func (w *responseWriterCache) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, eudore.ErrContextNotHijacker
}

type cacheMap struct {
	sync.Map
}

func (c *cacheMap) LoadData(key string) *cacheResponse {
	data, ok := c.Map.Load(key)
	if !ok {
		return nil
	}
	item := data.(*cacheResponse)
	if time.Now().After(item.Expired) {
		c.Map.Delete(key)
		return nil
	}
	return item
}

func (c *cacheMap) SaveData(key string, val *cacheResponse) {
	c.Map.Store(key, val)
}

func (c *cacheMap) cleanupExpired(ctx context.Context, t time.Duration) {
	for {
		select {
		case now := <-time.After(t):
			c.Map.Range(func(key, value any) bool {
				item := value.(*cacheResponse)
				if now.After(item.Expired) {
					c.Map.Delete(key)
				}
				return true
			})
		case <-ctx.Done():
			return
		}
	}
}
