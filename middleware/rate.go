package middleware

import (
	"github.com/eudore/eudore"
	"golang.org/x/time/rate"
	"net/http"
	"strings"
	"sync"
	"time"
)

var rates []*Rate
var ratemu sync.Mutex

func init() {
	go cleanupVisitors()
}

// Rate 定义限流器
type Rate struct {
	visitors map[string]*visitor
	mtx      sync.Mutex
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
func NewRate(r2, burst int) *Rate {
	r := &Rate{
		visitors: make(map[string]*visitor),
		Rate:     rate.Limit(r2),
		Burst:    burst,
	}
	ratemu.Lock()
	rates = append(rates, r)
	ratemu.Unlock()
	return r
}

// NewRateFunc 返回一个限流处理函数。
func NewRateFunc(r2, burst int) eudore.HandlerFunc {
	return NewRate(r2, burst).Handle
}

// ServeHTTP 方法实现http.Handler接口。
func (r *Rate) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ip := GetRealClientIP(req)
	limiter := r.GetVisitor(ip)
	if !limiter.Allow() {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		req.Method = "Deny"
		return
	}
}

// Handle 方法实现eudore请求上下文处理函数。
func (r *Rate) Handle(ctx eudore.Context) {
	ip := ctx.RealIP()
	limiter := r.GetVisitor(ip)
	if !limiter.Allow() {
		ctx.Info("rate: " + ip)
		ctx.WriteHeader(http.StatusTooManyRequests)
		ctx.End()
	}
}

// GetVisitor 方法通过ip获得*rate.Limiter。
func (r *Rate) GetVisitor(ip string) *rate.Limiter {
	r.mtx.Lock()
	v, exists := r.visitors[ip]
	if !exists {
		r.mtx.Unlock()
		return r.AddVisitor(ip)
	}
	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	r.mtx.Unlock()
	return v.limiter
}

// AddVisitor Change the the map to hold values of the type visitor.
func (r *Rate) AddVisitor(ip string) *rate.Limiter {
	limiter := rate.NewLimiter(r.Rate, r.Burst)
	r.mtx.Lock()
	// Include the current time when creating a new visitor.
	r.visitors[ip] = &visitor{limiter, time.Now()}
	r.mtx.Unlock()
	return limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute / 100)
		ratemu.Lock()
		for _, i := range rates {
			i.mtx.Lock()
			for ip, v := range i.visitors {
				if time.Now().Sub(v.lastSeen) > 3*time.Minute {
					delete(i.visitors, ip)
				}
			}
			i.mtx.Unlock()
		}
		ratemu.Unlock()
	}
}

// GetRealClientIP 函数获取http请求的真实ip
func GetRealClientIP(r *http.Request) string {
	xforward := r.Header.Get("X-Forwarded-For")
	if "" == xforward {
		return strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}

	return strings.SplitN(string(xforward), ",", 2)[0]
}
