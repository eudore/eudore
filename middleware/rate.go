package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// NewRateFunc 返回一个限流处理函数。
//
// 每周期(默认秒)增加speed个令牌，最多拥有max个。
//
// options:
// context.Context               =>    控制cleanupVisitors退出的生命周期
// time.Duration                 =>    基础时间周期单位，默认秒
// func(eudore.Context) string   =>    限流获取key的函数，默认Context.ReadIP
func NewRateFunc(speed, max int64, options ...interface{}) eudore.HandlerFunc {
	return newRate(speed, max, options...).HandleHTTP
}

func newRate(speed, max int64, options ...interface{}) *rate {
	r := &rate{
		visitors: make(map[string]*rateBucket),
		GetKeyFunc: func(ctx eudore.Context) string {
			return ctx.RealIP()
		},
		speed: int64(time.Second) / speed,
		max:   int64(time.Second) / speed * max,
	}
	ctx := context.Background()
	for _, i := range options {
		switch val := i.(type) {
		case context.Context:
			ctx = val
		case time.Duration:
			r.speed = int64(val) / speed
			r.max = int64(val) / speed * max
		case func(eudore.Context) string:
			r.GetKeyFunc = val
		}
	}
	go r.cleanupVisitors(ctx)
	return r
}

// rate 定义限流器
type rate struct {
	mu         sync.RWMutex
	visitors   map[string]*rateBucket
	GetKeyFunc func(eudore.Context) string
	speed      int64
	max        int64
}

// HandleHTTP 方法实现eudore请求上下文处理函数。
func (r *rate) HandleHTTP(ctx eudore.Context) {
	key := r.GetKeyFunc(ctx)
	if !r.GetVisitor(key).WaitWithDeadline(ctx.GetContext()) {
		ctx.WriteHeader(eudore.StatusTooManyRequests)
		ctx.Fatal("deny request of rate: " + key)
		ctx.End()
	}
}

// GetVisitor 方法通过ip获得rateBucket。
func (r *rate) GetVisitor(key string) *rateBucket {
	r.mu.RLock()
	v, exists := r.visitors[key]
	r.mu.RUnlock()
	if !exists {
		limiter := newBucket(r.speed, r.max)
		r.mu.Lock()
		r.visitors[key] = limiter
		r.mu.Unlock()
		return limiter
	}
	return v
}

func (r *rate) cleanupVisitors(ctx context.Context) {
	dura := time.Duration(r.max) * 10
	if time.Millisecond < dura && dura < time.Minute {
		dura = time.Minute
	}
	for {
		select {
		case now := <-time.After(dura):
			dead := now.UnixNano() - int64(dura)
			for key, v := range r.visitors {
				if v.last < dead {
					r.mu.Lock()
					delete(r.visitors, key)
					r.mu.Unlock()
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

type rateBucket struct {
	sync.Mutex
	speed int64
	max   int64
	last  int64
}

func newBucket(speed, max int64) *rateBucket {
	return &rateBucket{
		speed: speed,
		max:   max,
		last:  time.Now().UnixNano() - max,
	}
}

func (r *rateBucket) Allow() bool {
	r.Lock()
	defer r.Unlock()
	now := time.Now().UnixNano()
	if r.last < now {
		r.last += r.speed
		now = now - r.max
		if r.last < now {
			r.last = now
		}
		return true
	}
	return false
}

func (r *rateBucket) Wait(ctx context.Context) bool {
	r.Lock()
	defer r.Unlock()
	now := time.Now().UnixNano()
	if r.last < now {
		r.last += r.speed
		now = now - r.max
		if r.last < now {
			r.last = now
		}
		return true
	}

	ticker := time.NewTicker(time.Duration(r.last - now))
	defer ticker.Stop()
	select {
	case <-ticker.C:
		r.last += r.speed
		return true
	case <-ctx.Done():
		return false
	}

}

func (r *rateBucket) WaitWithDeadline(ctx context.Context) bool {
	if _, ok := ctx.Deadline(); ok {
		return r.Wait(ctx)
	}
	return r.Allow()
}
