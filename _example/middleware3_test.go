package eudore_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func TestMiddlewareNethttpBasicAuth(*testing.T) {
	data := map[string]string{"user": "pw"}

	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPBasicAuthFunc(mux, data))

	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/2", http.Header{"Authorization": {"Basic dXNlcjpwdw=="}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpBodyLimit(*testing.T) {
	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPBodyLimitFunc(mux, 32))

	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/2", strings.NewReader("body"))
	app.NewRequest(nil, "GET", "/3", strings.NewReader("1234567890abcdefghijklmnopqrstuvwxyz"))
	app.NewRequest(nil, "GET", "/4", eudore.NewClientBodyForm(url.Values{"name": {"eudore"}}))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpBlack(*testing.T) {
	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPBlackFunc(mux, map[string]bool{
		"127.0.0.1/8":    true,
		"192.168.0.0/16": true,
		"10.0.0.0/8":     false,
	}))

	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "GET", "/eudore/debug/black/data")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")

	app.NewRequest(nil, "GET", "/eudore")
	app.NewRequest(nil, "GET", "/eudore", http.Header{eudore.HeaderXForwardedFor: {"192.168.1.4 192.168.1.1"}})
	app.NewRequest(nil, "GET", "/eudore", http.Header{eudore.HeaderXRealIP: {"127.0.0.1:29398"}})
	app.NewRequest(nil, "GET", "/eudore", http.Header{eudore.HeaderXRealIP: {"192.168.75.1:8298"}})
	app.NewRequest(nil, "GET", "/eudore", http.Header{eudore.HeaderXRealIP: {"10.1.1.1:2334"}})
	app.NewRequest(nil, "GET", "/eudore", http.Header{eudore.HeaderXRealIP: {"172.17.1.1:2334"}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpRateRequest(*testing.T) {
	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPRateRequestFunc(mux, 1, 3, func(req *http.Request) string {
		// 自定义限流key
		return req.UserAgent()
	}))

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpRewrite(*testing.T) {
	rewritedata := map[string]string{
		"/js/*":                    "/public/js/$0",
		"/api/v1/users/*/orders/*": "/api/v3/user/$0/order/$1",
		"/d/*":                     "/d/$0-$0",
		"/api/v1/*":                "/api/v3/$0",
		"/api/v2/*":                "/api/v3/$0",
		"/help/history*":           "/api/v3/history",
		"/help/history":            "/api/v3/history",
		"/help/*":                  "$0",
	}

	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPRewriteFunc(mux, rewritedata))

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/js/")
	app.NewRequest(nil, "GET", "/js/index.js")
	app.NewRequest(nil, "GET", "/api/v1/user")
	app.NewRequest(nil, "GET", "/api/v1/user/new")
	app.NewRequest(nil, "GET", "/api/v1/users/v3/orders/8920")
	app.NewRequest(nil, "GET", "/api/v1/users/orders")
	app.NewRequest(nil, "GET", "/api/v2")
	app.NewRequest(nil, "GET", "/api/v2/user")
	app.NewRequest(nil, "GET", "/d/3")
	app.NewRequest(nil, "GET", "/help/history")
	app.NewRequest(nil, "GET", "/help/historyv2")

	app.CancelFunc()
	app.Run()
}
