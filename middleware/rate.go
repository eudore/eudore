package middleware

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// NewRateRequestFunc 返回一个限流处理函数。
//
// 每周期(默认秒)增加speed个令牌，最多拥有max个。
//
// options:
//
// context.Context               =>    控制cleanupVisitors退出的生命周期
//
// time.Duration                 =>    基础时间周期单位，默认秒
//
// func(eudore.Context) string   =>    限流获取key的函数，默认Context.ReadIP
func NewRateRequestFunc(speed, max int64, options ...interface{}) eudore.HandlerFunc {
	return newRate(speed, max, options...).HandlerRequest
}

// NewRateSpeedFunc 函数创建一个限速处理函数，不区分上下行流量。
//
// speed为速度(byte),max为默认初始化流量值，参数参考NewRateRequestFunc。
//
// speed速度不要小于通常Reader的缓冲区大小(最好大于4kB 4096)，否则无法请求到住够的令牌导致阻塞。
//
// Read时先请求缓冲区大小数量的令牌，然后返还未使用的令牌数量；Write时请求写入数据长度数量的令牌。
func NewRateSpeedFunc(speed, max int64, options ...interface{}) eudore.HandlerFunc {
	return newRate(speed, max, options...).HandlerSpeed
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

// HandlerRequest 方法实现eudore请求上下文处理函数。
func (r *rate) HandlerRequest(ctx eudore.Context) {
	key := r.GetKeyFunc(ctx)
	if !r.GetVisitor(key).WaitWithDeadline(ctx.GetContext(), 1) {
		ctx.WriteHeader(eudore.StatusTooManyRequests)
		ctx.Fatal("deny request of rate request: " + key)
		ctx.End()
	}
}

func (r *rate) HandlerSpeed(ctx eudore.Context) {
	rate := r.GetVisitor(r.GetKeyFunc(ctx))
	httpctx := ctx.GetContext()
	ctx.Request().Body = &rateRequqest{
		ReadCloser: ctx.Request().Body,
		Context:    httpctx,
		rateBucket: rate,
	}
	ctx.SetResponse(&rateResponse{
		ResponseWriter: ctx.Response(),
		Context:        httpctx,
		rateBucket:     rate,
	})
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
			r.mu.Lock()
			for key, v := range r.visitors {
				v.Lock()
				if v.last < dead {
					delete(r.visitors, key)
				}
				v.Unlock()
			}
			r.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

var errRateReadWaitLong = errors.New("If the github.com/eudore/eudore/middleware speed limit waiting time is too long, it will time out")
var errRateWriteWaitLong = errors.New("If the github.com/eudore/eudore/middleware speed limit waits for write time is too long, it will wait for timeout")

type rateRequqest struct {
	io.ReadCloser
	context.Context
	*rateBucket
}

type rateResponse struct {
	eudore.ResponseWriter
	context.Context
	*rateBucket
}

func (r *rateRequqest) Read(body []byte) (int, error) {
	length := len(body)
	if r.Wait(r.Context, int64(length)) {
		n, err := r.ReadCloser.Read(body)
		if length != n {
			r.Put(int64(length - n))
		}
		return n, err
	}
	err := r.Err()
	if err == nil {
		err = errRateReadWaitLong
	}
	return 0, err
}

func (r *rateResponse) Write(body []byte) (int, error) {
	if r.Wait(r.Context, int64(len(body))) {
		return r.ResponseWriter.Write(body)
	}
	err := r.Err()
	if err == nil {
		err = errRateWriteWaitLong
	}
	return 0, err
}

// rate 定义限流器
type rate struct {
	mu         sync.RWMutex
	visitors   map[string]*rateBucket
	GetKeyFunc func(eudore.Context) string
	speed      int64
	max        int64
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

func (r *rateBucket) Put(n int64) {
	r.Lock()
	r.last = r.last - n*r.speed
	r.Unlock()
}

func (r *rateBucket) Allow(n int64) bool {
	r.Lock()
	defer r.Unlock()
	now := time.Now().UnixNano()
	n = r.last + n*r.speed
	if n < now {
		r.last = n
		now = now - r.max
		if r.last < now {
			r.last = now
		}
		return true
	}
	return false
}

func (r *rateBucket) Wait(ctx context.Context, n int64) bool {
	r.Lock()
	now := time.Now().UnixNano()
	n = r.last + n*r.speed
	if n < now {
		r.last = n
		now = now - r.max
		if r.last < now {
			r.last = now
		}
		r.Unlock()
		return true
	}

	dead, ok := ctx.Deadline()
	if ok && dead.UnixNano() < n {
		r.Unlock()
		return false
	}

	// 预支令牌 等待可用
	ticker := time.NewTicker(time.Duration(n - now))
	defer ticker.Stop()
	r.last = n
	r.Unlock()
	select {
	case <-ticker.C:
		return true
	case <-ctx.Done():
		// 取消上下文不退还令牌
		return false
	}
}

func (r *rateBucket) WaitWithDeadline(ctx context.Context, n int64) bool {
	if _, ok := ctx.Deadline(); ok {
		return r.Wait(ctx, n)
	}
	return r.Allow(n)
}
