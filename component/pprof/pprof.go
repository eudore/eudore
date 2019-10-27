package pprof

import (
	"net/http/pprof"

	"github.com/eudore/eudore"
)

// RoutesInject 函数实现注入pprof路由。
func RoutesInject(r eudore.RouterMethod) {
	r = r.Group("/pprof")
	r.AnyFunc("/", pprof.Index)
	r.AnyFunc("/*", pprof.Index)
	r.AnyFunc("/cmdline", pprof.Cmdline)
	r.AnyFunc("/profile", pprof.Profile)
	r.AnyFunc("/symbol", pprof.Symbol)
	r.AnyFunc("/trace", pprof.Trace)
}
