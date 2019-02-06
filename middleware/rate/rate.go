package rate

import (
	"sync"
	"time"
	"strings"
	"net/http"
	"eudore"
	"golang.org/x/time/rate"
)

var rates []*Rate

func init() {
	go cleanupVisitors()
}

type Rate struct{
	visitors	map[string]*visitor
	mtx			sync.Mutex
	Rate	rate.Limit
	Burst	int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}


func NewRate(r2, burst int) *Rate {
	r := &Rate{
		visitors:	make(map[string]*visitor),
		Rate:	rate.Limit(r2),
		Burst:	burst,
	}
	rates = append(rates, r)
	return r
}

func (r *Rate) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ip := GetRealClientIP(req)
	limiter := r.GetVisitor(ip)
	if !limiter.Allow() {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		req.Method = "Deny"
		return
	}
}

func (r *Rate) Handle(ctx eudore.Context) {
	ip := ctx.RemoteAddr()
	limiter := r.GetVisitor(ip)
	if !limiter.Allow() {
		ctx.Info("rate: " + ctx.RemoteAddr())
		ctx.WriteHeader(http.StatusTooManyRequests)
		ctx.End()
	}
}


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


// Change the the map to hold values of the type visitor.
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
		time.Sleep(time.Minute)
		for _,i := range rates {
			i.mtx.Lock()
			for ip, v := range i.visitors {
				if time.Now().Sub(v.lastSeen) > 3 * time.Minute {
					delete(i.visitors, ip)
				}
			}
			i.mtx.Unlock()
		}
	}
}


func GetRealClientIP(r *http.Request ) string {
	xforward := r.Header.Get("X-Forwarded-For")
	if "" == xforward {
		return strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}

	return strings.SplitN(string(xforward), ",", 2)[0]
}
