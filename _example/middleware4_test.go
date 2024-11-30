package eudore_test

import (
	"strconv"
	"testing"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
)

func TestMiddlewareCORS(t *testing.T) {
	corsPattern := []string{"localhost"}
	corsHeaders := map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
		"access-control-Max-Age":           "1000",
	}
	NewCORSFunc(nil, corsHeaders)

	app := NewApp()
	app.AddMiddleware("global", NewCORSFunc(corsPattern, corsHeaders))
	app.AnyFunc("/*", HandlerEmpty)

	reqs := []struct {
		method string
		origin string
		status int
	}{
		{MethodOptions, DefaultClientInternalHost, 405},
		{MethodGet, DefaultClientInternalHost, 200},
		{MethodOptions, DefaultClientHost, 204},
		{MethodGet, DefaultClientHost, 200},
		{MethodGet, "127.0.0.1", 403},
	}
	for i, r := range reqs {
		err := app.NewRequest(r.method, "/"+strconv.Itoa(i),
			NewClientHeader(HeaderOrigin, "http://"+r.origin),
			NewClientCheckStatus(r.status),
		)
		if err != nil {
			t.Error(err)
		}
	}
	app.GetRequest("/", NewClientHeader(HeaderOrigin, "a"))
	app.GetRequest("/", NewClientHeader(HeaderOrigin, "https://a"))
	app.CancelFunc()
	app.Run()
}

func TestMiddlewareReferer(t *testing.T) {
	app := NewApp()
	app.AddMiddleware(NewRefererCheckFunc(map[string]bool{
		"*":                      false,
		"":                       true,
		"origin":                 true,
		"eudore.cn/*":            true,
		"www.eudore.cn/*":        true,
		"*.eudore.cn/*":          true,
		"*.eudore.cn/*/public/*": true,
	}))
	app.AnyFunc("/*", HandlerEmpty)

	reqs := []struct {
		referer string
		status  int
	}{
		{"http://" + DefaultClientInternalHost + "/", 200},
		{"http://eudore.cn/", 200},
		{"http://www.eudore.cn/", 200},
		{"http://godoc.eudore.cn/", 200},
		{"http://godoc.eudore.cn/pkg/public/app.js", 200},
		{"http://www.example.com/", 403},
		{"", 200},
		{"origin", 403},
	}
	for i, r := range reqs {
		err := app.GetRequest("/"+strconv.Itoa(i),
			NewClientHeader(HeaderReferer, r.referer),
			NewClientCheckStatus(r.status),
		)
		if err != nil {
			t.Error(err)
		}
	}

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRewrite(t *testing.T) {
	rewritedata := map[string]string{
		"/public/*":    "/static/$0",
		"/api/v1/*":    "/api/v3/$0",
		"/api/v2/*":    "/api/v3/$0",
		"/user/*":      "/users/$0",
		"/user/*/icon": "/users/icon/$0",
		"/split1/*":    "$0",
		"/split2/*":    "/$0-$1-$",
		"/u1":          "",
		"/u2":          "",
		"/u3":          "",
		"/u4":          "",
	}

	app := NewApp()
	app.AddMiddleware("global", NewRewriteFunc(rewritedata))
	app.AnyFunc("*", func(ctx Context) {
		ctx.WriteString(ctx.Path())
	})

	reqs := []struct {
		path   string
		target string
	}{
		{"/api/v1/name", "/api/v3/name"},
		{"/api/v2/name", "/api/v3/name"},
		{"/user/eudore", "/users/eudore"},
		{"/user/eudore/icon", "/users/icon/eudore"},
		{"/public/app.js", "/static/app.js"},
		{"/metrics", "/metrics"},
		{"/split1/name", "name"},
		{"/split2/name", "/name-$1-$"},
	}
	for _, r := range reqs {
		err := app.GetRequest(r.path,
			NewClientCheckStatus(200),
			NewClientCheckBody(r.target),
		)
		if err != nil {
			t.Error(err)
		}
	}

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRouter(*testing.T) {
	routerdata := map[string]interface{}{
		"/api/:v/*": func(ctx Context) {
			ctx.Request().URL.Path = "/api/v3/" + ctx.GetParam("*")
		},
		"GET /api/:v/*": func(ctx Context) {
			ctx.WriteHeader(403)
			ctx.End()
		},
	}

	app := NewApp()
	app.AddMiddleware(NewRoutesFunc(routerdata))
	app.AnyFunc("/*", HandlerEmpty)

	app.GetRequest("/api/v1/user")
	app.PutRequest("/api/v1/user")
	app.PutRequest("/api/v2/user")

	app.CancelFunc()
	app.Run()
}
