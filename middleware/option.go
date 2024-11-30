package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

// Option defines some middleware optional and modifies default configuration.
type Option func(any)

// NewOptionKeyFunc function creates option to modify GetKeyFunc.
//
// If GetKeyFunc returns an empty string,
// the corresponding middleware will be skipped.
//
// middleware: [NewCSRFFunc] [NewCircuitBreakerFunc] [NewCacheFunc]
// [NewRateRequestFunc] [NewRateSpeedFunc].
func NewOptionKeyFunc(fn func(eudore.Context) string) Option {
	return func(data any) {
		switch v := data.(type) {
		case *breaker:
			v.GetKeyFunc = fn
		case *rate:
			v.GetKeyFunc = fn
		case *cache:
			v.GetKeyFunc = fn
		case *csrf:
			v.GetKeyFunc = fn
		}
	}
}

// NewOptionRouter function creates options for registering middleware API.
//
// NewBlackFunc middleware will add [sync.RWMutex].
//
// middleware: [NewCircuitBreakerFunc] [NewBlackListFunc].
func NewOptionRouter(router eudore.Router) Option {
	return func(data any) {
		switch v := data.(type) {
		case *breaker:
			router.GetFunc("/breaker/data", v.data)
			router.GetFunc("/breaker/:id", v.get)
			router.PutFunc("/breaker/:id/state/:state", v.putState)
		case *black:
			v.White4 = &subnetListMutex{subnetList: v.White4}
			v.Black4 = &subnetListMutex{subnetList: v.Black4}
			v.White6 = &subnetListMutex{subnetList: v.White6}
			v.Black6 = &subnetListMutex{subnetList: v.Black6}
			router.GetFunc("/black/data", v.data)
			router.PutFunc("/black/allow/:ip list=white", v.putIP)
			router.PutFunc("/black/deny/:ip list=black", v.putIP)
			router.DeleteFunc("/black/allow/:ip list=white", v.deleteIP)
			router.DeleteFunc("/black/deny/:ip list=black", v.deleteIP)
		}
	}
}

// NewOptionCSRFCookie function creates a CSRF option setting read-write cookie.
func NewOptionCSRFCookie(cookie http.Cookie) Option {
	return func(data any) {
		v, ok := data.(*csrf)
		if ok {
			v.Cookie = cookie
		}
	}
}

// NewOptionRateCleanup function creates Cache option to clean up expired data.
func NewOptionCacheCleanup(ctx context.Context, t time.Duration) Option {
	return func(data any) {
		v, ok := data.(*cache)
		if ok {
			m, ok := v.storage.(*cacheMap)
			if ok {
				go m.cleanupExpired(ctx, t)
			}
		}
	}
}

// NewOptionCircuitBreakerConfig function creates options to modify Breaker
// config.
//
// Maybe add GetBreakrEntryFunc to implement different Breaker strategies.
func NewOptionCircuitBreakerConfig(maxSuccesses, maxFailures int,
	dura, wait time.Duration,
) Option {
	return func(data any) {
		v, ok := data.(*breaker)
		if ok {
			v.GetBreakrEntryFunc = newBreakerEntryfunc(
				maxSuccesses, maxFailures, dura, wait,
			)
		}
	}
}

// NewOptionRateCleanup function creates Rate option to cleanup expired buckets.
// And sets the less number of buckets to use when cleaning.
func NewOptionRateCleanup(ctx context.Context, t time.Duration, less int,
) Option {
	return func(data any) {
		v, ok := data.(*rate)
		if ok {
			go v.cleanupVisitors(ctx, t, less)
		}
	}
}

func applyOption(data any, options []Option) {
	for i := range options {
		options[i](data)
	}
}

func writePage(ctx eudore.Context, code int, msg, value string) {
	if msg != "" {
		ctx.WriteStatus(code)
		_ = ctx.Render(strings.Replace(msg, "{{value}}", value, 1))
	} else {
		ctx.WriteHeader(code)
	}
}

func headerCopy(dst, src map[string][]string) {
	for key, vals := range src {
		dst[key] = append(dst[key], vals...)
	}
}

func headerVary(h http.Header, vary string) {
	v := h[eudore.HeaderVary]
	if v == nil {
		h[eudore.HeaderVary] = []string{vary}
	} else {
		v[0] = v[0] + ", " + vary
	}
}
