package breaker

import (
	"sync"
	"github.com/eudore/eudore"
)

const (
	CircuitBreakerStatueClosed int8	= iota
	CircuitBreakerStatueHalfOpen
	CircuitBreakerStatueOpen 
)

type CircuitBreaker struct {
	statue		int8
	mu			sync.Mutex
	success		int64
	failure		int64
}

func (cb *CircuitBreaker) Handler(ctx eudore.Context) {
	switch cb.statue {
	case CircuitBreakerStatueClosed:
		// Close状态,正常处理并统计结果，失败过多进入半开状态
		ctx.Next()
		if ctx.Response().Status() < 500 {
			atomic.AddUint64(cb.success, 1)	
		}else {
			atomic.AddUint64(cb.failure, 1)
			if cb.failure > cb.success/10 {
				cb.mu.Lock()
				cb.statue = CircuitBreakerStatueHalfOpen
				cb.mu.Unock()
			}
		}
	case CircuitBreakerStatueHalfOpen:
		// HalfOpen状态，允许部分服务尝试，并统计情况。
		// TODO：未添加限流和状态变更前的统计
		ctx.Next()
		if ctx.Response().Status() < 500 {
			cb.mu.Lock()
			cb.statue = CircuitBreakerStatueClosed
			cb.mu.Unock()
		}else {
			cb.mu.Lock()
			cb.statue = CircuitBreakerStatueOpen
			cb.mu.Unock()
		}
	case CircuitBreakerStatueOpen:
		// open状态拒绝服务
		ctx.End()
		return
	}	
}
