package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"golang.org/x/time/rate"
)

// Rate 定义限流器
type Rate struct {
	visitors map[string]*visitor
	mtx      sync.RWMutex
	Rate     rate.Limit
	Burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRate 创建一个限流器。
//
// 周期内增加r2个令牌，最多拥有burst个。
func NewRate(ctx context.Context, r2, burst int) *Rate {
	r := &Rate{
		visitors: make(map[string]*visitor),
		Rate:     rate.Limit(r2),
		Burst:    burst,
	}
	go r.cleanupVisitors(ctx)
	return r
}

// NewRateFunc 返回一个限流处理函数。
func NewRateFunc(ctx context.Context, r2, burst int) eudore.HandlerFunc {
	return NewRate(ctx, r2, burst).HandleHTTP
}

// ServeHTTP 方法实现http.Handler接口。
func (r *Rate) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := getRealClientIP(req)
	limiter := r.GetVisitor(key)
	if !limiter.Allow() {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		req.Method = "Deny"
		return
	}
}

// HandleHTTP 方法实现eudore请求上下文处理函数。
func (r *Rate) HandleHTTP(ctx eudore.Context) {
	key := ctx.RealIP()
	limiter := r.GetVisitor(key)
	if !limiter.Allow() {
		ctx.WriteHeader(http.StatusTooManyRequests)
		ctx.Fatal("rate: " + key)
		ctx.End()
	}
}

// GetVisitor 方法通过ip获得*rate.Limiter。
func (r *Rate) GetVisitor(key string) *rate.Limiter {
	r.mtx.RLock()
	v, exists := r.visitors[key]
	if !exists {
		r.mtx.RUnlock()
		return r.AddVisitor(key)
	}
	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	r.mtx.RUnlock()
	return v.limiter
}

// AddVisitor Change the the map to hold values of the type visitor.
func (r *Rate) AddVisitor(key string) *rate.Limiter {
	limiter := rate.NewLimiter(r.Rate, r.Burst)
	r.mtx.Lock()
	r.visitors[key] = &visitor{limiter, time.Now()}
	r.mtx.Unlock()
	return limiter
}

func (r *Rate) cleanupVisitors(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(time.Minute)
			for key, v := range r.visitors {
				if time.Now().Sub(v.lastSeen) > time.Minute {
					r.mtx.Lock()
					delete(r.visitors, key)
					r.mtx.Unlock()
				}
			}
		}
	}
}

// getRealClientIP 函数获取http请求的真实ip
func getRealClientIP(r *http.Request) string {
	xforward := r.Header.Get("X-Forwarded-For")
	if "" == xforward {
		return strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}

	return strings.SplitN(string(xforward), ",", 2)[0]
}
