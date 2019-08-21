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
		StaticHtml = file[:len(file)-2] + "html"
	}
}

// 定义熔断器状态。
const (
	CircuitBreakerStatueClosed State = iota
	CircuitBreakerStatueHalfOpen
	CircuitBreakerStatueOpen
)

// 半开状态时最大连续失败和最大连续成功次数。
var (
	MaxConsecutiveSuccesses uint32 = 10
	MaxConsecutiveFailures  uint32 = 10
	StaticHtml                     = ""
	// CircuitBreakerStatues 定义熔断状态字符串
	CircuitBreakerStatues = []string{"closed", "half-open", "open"}
)

type (
	// State 是熔断器状态。
	State int8
	// CircuitBreaker 定义熔断器。
	CircuitBreaker struct {
		mu            sync.RWMutex
		num           int
		Mapping       map[int]string                             `json:"mapping"`
		Routes        map[string]*Route                          `json:"routes"`
		OnStateChange func(eudore.Context, string, State, State) `json:"-"`
	}
	// Route 定义单词路由的熔断数据。
	Route struct {
		mu                   sync.Mutex `json:"-"`
		Id                   int
		Name                 string
		State                State
		LastTime             time.Time
		TotalSuccesses       uint64
		TotalFailures        uint64
		ConsecutiveSuccesses uint32
		ConsecutiveFailures  uint32
		OnStateChange        func(eudore.Context, string, State, State) `json:"-"`
	}
)

// NewCircuitBreaker 函数创建一个熔断器
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		Mapping: make(map[int]string),
		Routes:  make(map[string]*Route),
		OnStateChange: func(ctx eudore.Context, name string, from State, to State) {
			ctx.Infof("CircuitBreaker route %s change state from %s to %s", name, from, to)
		},
	}
}

// Handle 方法定义熔断器处理eudore请求上下文函数。
func (cb *CircuitBreaker) Handle(ctx eudore.Context) {
	name := ctx.GetParam("route")
	cb.mu.RLock()
	route, ok := cb.Routes[name]
	cb.mu.RUnlock()
	if !ok {
		cb.mu.Lock()
		route = &Route{
			Id:            cb.num,
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

// InjectRoutes 方法给给路由器注入熔断器的路由。
func (cb *CircuitBreaker) InjectRoutes(r eudore.RouterMethod) {
	r.GetFunc("/ui", func(ctx eudore.Context) {
		if StaticHtml != "" {
			ctx.WriteFile(StaticHtml)
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
		route.OnStateChange(ctx, route.Name, route.State, State(state))
		route.State = State(state)
		route.ConsecutiveSuccesses = 0
		route.ConsecutiveFailures = 0
	})
}

// Handle 方法实现路由条目处理熔断。
func (c *Route) Handle(ctx eudore.Context) {
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
func (c *Route) IsDeny() (b bool) {
	if c.State == CircuitBreakerStatueHalfOpen {
		b = time.Now().Before(c.LastTime.Add(400 * time.Millisecond))
		if b {
			c.LastTime = time.Now()
		}
		return b
	}
	return c.State == CircuitBreakerStatueOpen
}

// onSuccess 方法处理熔断器条目成功的情况。
func (c *Route) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
	if c.State != CircuitBreakerStatueClosed && c.ConsecutiveSuccesses > MaxConsecutiveSuccesses {
		c.ConsecutiveSuccesses = 0
		c.State--
		c.LastTime = time.Now()
	}
}

// onFailure 方法处理熔断器条目失败的情况。
func (c *Route) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
	if c.State != CircuitBreakerStatueOpen && c.ConsecutiveFailures > MaxConsecutiveFailures {
		c.ConsecutiveFailures = 0
		c.State++
		c.LastTime = time.Now()
	}
}

// String 方法实现string接口
func (state State) String() string {
	return CircuitBreakerStatues[state]
}

// MarshalText 方法实现encoding.TextMarshaler接口。
func (state State) MarshalText() (text []byte, err error) {
	text = []byte(state.String())
	return
}
