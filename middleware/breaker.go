package middleware

import (
	"runtime"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// 获取文件定义位置，静态ui文件在同目录。
func init() {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		StaticHTML = file[:len(file)-2] + "html"
	}
}

// 定义熔断器状态。
const (
	CircuitBreakerStatueClosed BreakState = iota
	CircuitBreakerStatueHalfOpen
	CircuitBreakerStatueOpen
)

// 半开状态时最大连续失败和最大连续成功次数。
var (
	MaxConsecutiveSuccesses uint32 = 10
	MaxConsecutiveFailures  uint32 = 10
	StaticHTML                     = ""
	// CircuitBreakerStatues 定义熔断状态字符串
	CircuitBreakerStatues = []string{"closed", "half-open", "open"}
)

type (
	// BreakState 是熔断器状态。
	BreakState int8
	// CircuitBreaker 定义熔断器。
	CircuitBreaker struct {
		mu            sync.RWMutex
		num           int
		Mapping       map[int]string                                       `json:"mapping"`
		Routes        map[string]*breakRoute                               `json:"routes"`
		OnStateChange func(eudore.Context, string, BreakState, BreakState) `json:"-"`
	}
	// breakRoute 定义单词路由的熔断数据。
	breakRoute struct {
		mu                   sync.Mutex
		ID                   int
		Name                 string
		BreakState           BreakState `json:"State"`
		LastTime             time.Time
		TotalSuccesses       uint64
		TotalFailures        uint64
		ConsecutiveSuccesses uint32
		ConsecutiveFailures  uint32
		OnStateChange        func(eudore.Context, string, BreakState, BreakState) `json:"-"`
	}
)

// NewCircuitBreaker 函数创建一个熔断器
func NewCircuitBreaker(r eudore.Router) *CircuitBreaker {
	cb := &CircuitBreaker{
		Mapping: make(map[int]string),
		Routes:  make(map[string]*breakRoute),
		OnStateChange: func(ctx eudore.Context, name string, from BreakState, to BreakState) {
			ctx.Infof("CircuitBreaker route %s change state from %s to %s", name, from, to)
		},
	}
	if r != nil {
		cb.RoutesInject(r)
	}
	return cb
}

// NewBreakFunc 方法定义熔断器处理eudore请求上下文函数。
func (cb *CircuitBreaker) NewBreakFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		name := ctx.GetParam("route")
		cb.mu.RLock()
		route, ok := cb.Routes[name]
		cb.mu.RUnlock()
		if !ok {
			cb.mu.Lock()
			route = &breakRoute{
				ID:            cb.num,
				Name:          name,
				LastTime:      time.Now(),
				OnStateChange: cb.OnStateChange,
			}
			cb.Mapping[cb.num] = name
			cb.Routes[name] = route
			cb.num++
			cb.mu.Unlock()
		}

		route.Handle(ctx)
	}
}

// RoutesInject 方法给给路由器注入熔断器的路由。
func (cb *CircuitBreaker) RoutesInject(r eudore.Router) {
	r.GetFunc("/ui", func(ctx eudore.Context) {
		if StaticHTML != "" {
			ctx.WriteFile(StaticHTML)
		} else {
			ctx.WriteString("breaker not set ui file path.")
		}
	})
	r.GetFunc("/list", func(ctx eudore.Context) {
		ctx.Render(cb.Routes)
	})
	r.GetFunc("/:id", func(ctx eudore.Context) {
		id := eudore.GetStringDefaultInt(ctx.GetParam("id"), -1)
		if id < 0 || id >= cb.num {
			ctx.Fatal("id is invalid")
			return
		}
		cb.mu.RLock()
		route := cb.Routes[cb.Mapping[id]]
		cb.mu.RUnlock()
		ctx.Render(route)

	})
	r.PutFunc("/:id/state/:state", func(ctx eudore.Context) {
		id := eudore.GetStringDefaultInt(ctx.GetParam("id"), -1)
		state := eudore.GetStringDefaultInt(ctx.GetParam("state"), -1)
		if id < 0 || id >= cb.num {
			ctx.Fatal("id is invalid")
			return
		}
		if state < -1 || state > 2 {
			ctx.Fatal("state is invalid")
			return
		}
		cb.mu.RLock()
		route := cb.Routes[cb.Mapping[id]]
		cb.mu.RUnlock()
		route.OnStateChange(ctx, route.Name, route.BreakState, BreakState(state))
		route.BreakState = BreakState(state)
		route.ConsecutiveSuccesses = 0
		route.ConsecutiveFailures = 0
	})
}

// Handle 方法实现路由条目处理熔断。
func (c *breakRoute) Handle(ctx eudore.Context) {
	if c.IsDeny() {
		ctx.WriteHeader(503)
		ctx.Fatal("CircuitBreaker")
		return
	}
	ctx.Next()
	if ctx.Response().Status() < 500 {
		c.onSuccess()
	} else {
		c.onFailure()
	}
}

// IsDeny 方法实现熔断器条目是否可以通过。
func (c *breakRoute) IsDeny() (b bool) {
	if c.BreakState == CircuitBreakerStatueHalfOpen {
		b = time.Now().Before(c.LastTime.Add(400 * time.Millisecond))
		if b {
			c.LastTime = time.Now()
		}
		return b
	}
	return c.BreakState == CircuitBreakerStatueOpen
}

// onSuccess 方法处理熔断器条目成功的情况。
func (c *breakRoute) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
	if c.BreakState != CircuitBreakerStatueClosed && c.ConsecutiveSuccesses > MaxConsecutiveSuccesses {
		c.ConsecutiveSuccesses = 0
		c.BreakState--
		c.LastTime = time.Now()
	}
}

// onFailure 方法处理熔断器条目失败的情况。
func (c *breakRoute) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
	if c.BreakState != CircuitBreakerStatueOpen && c.ConsecutiveFailures > MaxConsecutiveFailures {
		c.ConsecutiveFailures = 0
		c.BreakState++
		c.LastTime = time.Now()
	}
}

// String 方法实现string接口
func (state BreakState) String() string {
	return CircuitBreakerStatues[state]
}

// MarshalText 方法实现encoding.TextMarshaler接口。
func (state BreakState) MarshalText() (text []byte, err error) {
	text = []byte(state.String())
	return
}
