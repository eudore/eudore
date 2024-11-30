package middleware

import (
	"context"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// rate defines the rate limiter.
type rate struct {
	mu         sync.RWMutex
	visitors   map[string]*rateBucket
	GetKeyFunc func(eudore.Context) string
	speed      int64
	total      int64
}

type rateBucket struct {
	sync.Mutex
	speed int64
	max   int64
	last  int64
}

// NewRateRequestFunc function creates middleware to implement request limit,
//
// Use memory-based token bucket rate limiting.
//
// Speed is tokens per second, with a maximum of total tokens.
//
// This middleware does not support cluster mode.
//
// options: [NewOptionKeyFunc] [NewOptionRateCleanup].
func NewRateRequestFunc(speed, total int64, options ...Option) Middleware {
	r := newRate(speed, total, options)
	return func(ctx eudore.Context) {
		key := r.GetKeyFunc(ctx)
		if key == "" {
			return
		}
		now, at, ok := r.GetVisitor(key).Allow()
		state := now - at
		ctx.SetHeader(eudore.HeaderXRateLimit, fi64(r.total/r.speed))
		ctx.SetHeader(eudore.HeaderXRateReset, fi64((at+r.total)/1000000000))
		if ok {
			ctx.SetHeader(eudore.HeaderXRateRemaining, fi64(state/r.speed))
			return
		}

		retry := int((r.speed - state) / 1000000000)
		if retry < DefaultRateRetryMin {
			retry = DefaultRateRetryMin
		}
		ctx.SetHeader(eudore.HeaderRetryAfter, strconv.Itoa(retry))
		writePage(ctx, eudore.StatusTooManyRequests, DefaultPageRate, key)
		ctx.End()
	}
}

func fi64(i int64) string {
	return strconv.FormatInt(i, 10)
}

// The NewRateSpeedFunc function creates middleware to implement rate limiting,
// without distinguishing between upstream and downstream traffic.
//
// speed is the speed (byte), total is the default initialization flow value,
// refer: [NewRateRequestFunc].
//
// The speed should not be less than the [io.Reader] buffer size
// (preferably greater than 4kB 4096),
// otherwise the unable to get token will cause blocking.
//
// When reading, first request the tokens of the buffer size (512),
// and then return the number of unused tokens;
// when writing, request the tokens of the write data length.
func NewRateSpeedFunc(speed, total int64, options ...Option) Middleware {
	r := newRate(speed, total, options)
	return func(ctx eudore.Context) {
		key := r.GetKeyFunc(ctx)
		if key == "" {
			return
		}
		rate := r.GetVisitor(key)
		httpctx := ctx.Context()
		req := ctx.Request()
		if req.ContentLength != 0 {
			req.Body = &requqestReaderRate{
				ReadCloser: req.Body,
				Context:    httpctx,
				rateBucket: rate,
			}
		}
		ctx.SetResponse(&responseWriterRate{
			ResponseWriter: ctx.Response(),
			Context:        httpctx,
			rateBucket:     rate,
		})
	}
}

func newRate(speed, total int64, options []Option) *rate {
	r := &rate{
		visitors: make(map[string]*rateBucket),
		GetKeyFunc: func(ctx eudore.Context) string {
			return ctx.RealIP()
		},
		speed: int64(time.Second) / speed,
		total: int64(time.Second) / speed * total,
	}
	applyOption(r, options)

	return r
}

// The GetVisitor method gets the rateBucket through the key.
func (r *rate) GetVisitor(key string) *rateBucket {
	r.mu.RLock()
	v, exists := r.visitors[key]
	r.mu.RUnlock()
	if !exists {
		limiter := &rateBucket{
			speed: r.speed,
			max:   r.total,
			last:  time.Now().UnixNano() - r.total,
		}
		r.mu.Lock()
		r.visitors[key] = limiter
		r.mu.Unlock()
		return limiter
	}
	return v
}

// The cleanupVisitors method periodically clears unactive rates.
func (r *rate) cleanupVisitors(ctx context.Context, t time.Duration, less int) {
	last := time.Duration(r.total) * 10
	if time.Millisecond < t && t < last {
		t = last
	}
	for {
		select {
		case now := <-time.After(t):
			r.mu.RLock()
			if len(r.visitors) < less {
				r.mu.RUnlock()
				break
			}
			r.mu.RUnlock()

			dead := now.UnixNano() - int64(last)
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

func (r *rateBucket) Put(n int64) {
	r.Lock()
	r.last -= n * r.speed
	r.Unlock()
}

func (r *rateBucket) Allow() (int64, int64, bool) {
	r.Lock()
	defer r.Unlock()
	now := time.Now().UnixNano()
	next := r.last + r.speed
	if next < now {
		r.last = next
		if r.last < now-r.max {
			r.last = now - r.max
		}
		return now, r.last, true
	}
	return now, next, false
}

func (r *rateBucket) WaitN(ctx context.Context, n int64) bool {
	r.Lock()
	now := time.Now().UnixNano()
	n = r.last + n*r.speed
	if n < now {
		r.last = n
		now -= r.max
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

	// prepay token and wait for it to become available
	ticker := time.NewTicker(time.Duration(n - now))
	defer ticker.Stop()
	r.last = n
	r.Unlock()
	select {
	case <-ticker.C:
		return true
	case <-ctx.Done():
		// cancelling the context does not return the token
		return false
	}
}

type requqestReaderRate struct {
	io.ReadCloser
	context.Context
	*rateBucket
}

type responseWriterRate struct {
	eudore.ResponseWriter
	context.Context
	*rateBucket
}

func (r *requqestReaderRate) Read(body []byte) (int, error) {
	length := len(body)
	if r.WaitN(r.Context, int64(length)) {
		n, err := r.ReadCloser.Read(body)
		if length != n {
			r.Put(int64(length - n))
		}
		return n, err
	}
	return 0, r.Err()
}

func (r *responseWriterRate) Write(data []byte) (int, error) {
	if r.WaitN(r.Context, int64(len(data))) {
		return r.ResponseWriter.Write(data)
	}
	return 0, r.Err()
}

func (r *responseWriterRate) WriteString(data string) (int, error) {
	if r.WaitN(r.Context, int64(len(data))) {
		return r.ResponseWriter.WriteString(data)
	}
	return 0, r.Err()
}
