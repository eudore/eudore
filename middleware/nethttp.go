package middleware

/*
定义中间件在net/http库下的实现或者入口。
*/

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/eudore/eudore"
)

// NewNetHTTPBasicAuthFunc 函数创建一个net/http BasicAuth中间件处理函数，文档见NewBasicAuthFunc函数。
func NewNetHTTPBasicAuthFunc(next http.Handler, names map[string]string) http.HandlerFunc {
	checks := make(map[string]string, len(names))
	for name, pass := range names {
		checks[base64.StdEncoding.EncodeToString([]byte(name+":"+pass))] = name
	}
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if len(auth) > 5 && auth[:6] == "Basic " {
			name, ok := checks[auth[6:]]
			if ok {
				r.Header.Add("Basic-Name", name)
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", "Basic")
		w.WriteHeader(401)
	}
}

// NewNetHTTPBlackFunc 函数创建一个net/http黑名单中间件处理函数，文档见NewBlackFunc函数。
func NewNetHTTPBlackFunc(next http.Handler, data map[string]bool) http.HandlerFunc {
	b := newBlack()
	for k, v := range data {
		if v {
			b.InsertWhite(k)
		} else {
			b.InsertBlack(k)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ip := ip2int(getRealClientIP(r))
		if b.White.Look(ip) {
			next.ServeHTTP(w, r)
			return
		}
		if b.Black.Look(ip) {
			w.WriteHeader(403)
			w.Write([]byte("black list deny your ip " + getRealClientIP(r)))
		} else {
			next.ServeHTTP(w, r)
		}
	}
}

// NewNetHTTPRewriteFunc 函数创建一个net/http路径重写中间件处理函数，文档见NewRewriteFunc函数。
func NewNetHTTPRewriteFunc(next http.Handler, data map[string]string) http.HandlerFunc {
	node := new(rewriteNode)
	for k, v := range data {
		node.insert(k, v)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = node.rewrite(r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

// NewNetHTTPRateRequestFunc 函数创建一个net/http限流中间件处理函数，文档见NewRateFunc函数。
func NewNetHTTPRateRequestFunc(next http.Handler, speed, max int64, options ...interface{}) http.HandlerFunc {
	getKeyFunc := getRealClientIP
	for _, i := range options {
		if fn, ok := i.(func(*http.Request) string); ok {
			getKeyFunc = fn
			break
		}
	}
	r := newRate(speed, max, options...)
	return func(w http.ResponseWriter, req *http.Request) {
		key := getKeyFunc(req)
		if r.GetVisitor(key).WaitWithDeadline(req.Context(), 1) {
			next.ServeHTTP(w, req)
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("deny request of rate request: " + key))
	}
}

// getRealClientIP 函数获取http请求的真实ip
func getRealClientIP(r *http.Request) string {
	if realip := r.Header.Get(eudore.HeaderXRealIP); realip != "" {
		return realip
	}
	if xforward := r.Header.Get(eudore.HeaderXForwardedFor); xforward != "" {
		return strings.SplitN(string(xforward), ",", 2)[0]
	}
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}
