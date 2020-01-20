package pprof

import (
	"github.com/eudore/eudore"
	"net/http/pprof"
)

// RoutesInject 函数实现注入pprof路由。
func RoutesInject(r eudore.Router) {
	r = r.Group("/pprof")
	// fixpprof 如果X-Content-Type-Options=nosniff直接返回html文本。
	r.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
	})
	r.AnyFunc("/", pprof.Index)
	r.AnyFunc("/*", fixpath, pprof.Index)
	r.AnyFunc("/cmdline", pprof.Cmdline)
	r.AnyFunc("/profile", pprof.Profile)
	r.AnyFunc("/symbol", pprof.Symbol)
	r.AnyFunc("/trace", pprof.Trace)
}

// 修复pprof前缀要求
func fixpath(ctx eudore.Context) {
	ctx.Request().URL.Path = "/debug/pprof/" + ctx.GetParam("*")
}
