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
		BreakerStaticHTML = file[:len(file)-2] + "html"
	}
}

// 定义熔断器状态。
const (
	BreakerStatueClosed BreakerState = iota
	BreakerStatueHalfOpen
	BreakerStatueOpen
)

// 半开状态时最大连续失败和最大连续成功次数。
var (
	MaxConsecutiveSuccesses uint32 = 10
	MaxConsecutiveFailures  uint32 = 10
	BreakerStaticHTML              = ""
	// BreakerStatues 定义熔断状态字符串
	BreakerStatues = []string{"closed", "half-open", "open"}
)

// BreakerState 是熔断器状态。
type BreakerState int8

// Breaker 定义熔断器。
type Breaker struct {
	mu            sync.RWMutex
	num           int
	Mapping       map[int]string                                           `json:"mapping"`
	Routes        map[string]*breakRoute                                   `json:"routes"`
	OnStateChange func(eudore.Context, string, BreakerState, BreakerState) `json:"-"`
}

// breakRoute 定义单词路由的熔断数据。
type breakRoute struct {
	mu                   sync.Mutex
	ID                   int
	Name                 string
	BreakerState         BreakerState `json:"State"`
	LastTime             time.Time
	TotalSuccesses       uint64
	TotalFailures        uint64
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
	OnStateChange        func(eudore.Context, string, BreakerState, BreakerState) `json:"-"`
}

// NewBreaker 函数创建一个熔断器
func NewBreaker() *Breaker {
	return &Breaker{
		Mapping: make(map[int]string),
		Routes:  make(map[string]*breakRoute),
		OnStateChange: func(ctx eudore.Context, name string, from BreakerState, to BreakerState) {
			ctx.Infof("Breaker route %s change state from %s to %s", name, from, to)
		},
	}
}

// NewBreakFunc 方法定义熔断器处理eudore请求上下文函数。
func (b *Breaker) NewBreakFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		name := ctx.GetParam("route")
		b.mu.RLock()
		route, ok := b.Routes[name]
		b.mu.RUnlock()
		if !ok {
			b.mu.Lock()
			route = &breakRoute{
				ID:            b.num,
				Name:          name,
				LastTime:      time.Now(),
				OnStateChange: b.OnStateChange,
			}
			b.Mapping[b.num] = name
			b.Routes[name] = route
			b.num++
			b.mu.Unlock()
		}

		route.Handle(ctx)
	}
}

// InjectRoutes 方法给给路由器注入熔断器的路由。
func (b *Breaker) InjectRoutes(r eudore.Router) *Breaker {
	r.GetFunc("/breaker/ui", b.ui)
	r.GetFunc("/breaker/data", b.data)
	r.GetFunc("/breaker/:id", b.getRoute)
	r.PutFunc("/breaker/:id/state/:state", b.putRouteState)
	return b
}

func (b *Breaker) ui(ctx eudore.Context) {
	if BreakerStaticHTML != "" {
		ctx.WriteFile(BreakerStaticHTML)
	} else {
		ctx.WriteString("breaker not set ui file path.")
	}
}

func (b *Breaker) data(ctx eudore.Context) {
	ctx.Render(b.Routes)
}

func (b *Breaker) getRoute(ctx eudore.Context) {
	id := eudore.GetStringDefaultInt(ctx.GetParam("id"), -1)
	if id < 0 || id >= b.num {
		ctx.Fatal("id is invalid")
		return
	}
	b.mu.RLock()
	route := b.Routes[b.Mapping[id]]
	b.mu.RUnlock()
	ctx.Render(route)
}

func (b *Breaker) putRouteState(ctx eudore.Context) {
	id := eudore.GetStringDefaultInt(ctx.GetParam("id"), -1)
	state := eudore.GetStringDefaultInt(ctx.GetParam("state"), -1)
	if id < 0 || id >= b.num {
		ctx.Fatal("id is invalid")
		return
	}
	if state < -1 || state > 2 {
		ctx.Fatal("state is invalid")
		return
	}
	b.mu.RLock()
	route := b.Routes[b.Mapping[id]]
	b.mu.RUnlock()
	route.OnStateChange(ctx, route.Name, route.BreakerState, BreakerState(state))
	route.BreakerState = BreakerState(state)
	route.ConsecutiveSuccesses = 0
	route.ConsecutiveFailures = 0
}

// Handle 方法实现路由条目处理熔断。
func (c *breakRoute) Handle(ctx eudore.Context) {
	if c.IsDeny() {
		ctx.WriteHeader(503)
		ctx.Fatal("Breaker")
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
	if c.BreakerState == BreakerStatueHalfOpen {
		b = time.Now().Before(c.LastTime.Add(400 * time.Millisecond))
		if b {
			c.LastTime = time.Now()
		}
		return b
	}
	return c.BreakerState == BreakerStatueOpen
}

// onSuccess 方法处理熔断器条目成功的情况。
func (c *breakRoute) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
	if c.BreakerState != BreakerStatueClosed && c.ConsecutiveSuccesses > MaxConsecutiveSuccesses {
		c.ConsecutiveSuccesses = 0
		c.BreakerState--
		c.LastTime = time.Now()
	}
}

// onFailure 方法处理熔断器条目失败的情况。
func (c *breakRoute) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
	if c.BreakerState != BreakerStatueOpen && c.ConsecutiveFailures > MaxConsecutiveFailures {
		c.ConsecutiveFailures = 0
		c.BreakerState++
		c.LastTime = time.Now()
	}
}

// String 方法实现string接口
func (state BreakerState) String() string {
	return BreakerStatues[state]
}

// MarshalText 方法实现encoding.TextMarshaler接口。
func (state BreakerState) MarshalText() (text []byte, err error) {
	text = []byte(state.String())
	return
}
