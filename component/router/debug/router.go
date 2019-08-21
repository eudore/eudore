package debug

import (
	"github.com/eudore/eudore"
	"runtime"
)

type (
	// RouterDebug 定义debug路由器。
	RouterDebug struct {
		eudore.RouterMethod `json:"-" xml:"-" set:"-"`
		eudore.RouterCore   `json:"-" xml:"-" set:"-"`
		router              eudore.Router
		Methods             []string `json:"methods"`
		Paths               []string `json:"paths"`
	}
	// RouterDebugCore 定义debug路由器核心。
	RouterDebugCore struct {
		eudore.RouterCore `json:"-" set:"-"`
	}
)

var _ eudore.Router = (*RouterDebug)(nil)
var StaticHtml = ""

// 获取文件定义位置，静态ui文件在同目录。
func init() {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		StaticHtml = file[:len(file)-2] + "html"
	}
}

// NewRouterDebug 函数创建一个debug路由器，默认封装RouterFull。
func NewRouterDebug(i interface{}) (eudore.Router, error) {
	r2, err := eudore.NewRouterFull(i)
	r := &RouterDebug{
		RouterCore: r2,
		router:     r2,
	}
	r.RouterMethod = &eudore.RouterMethodStd{
		RouterCore:          r,
		ControllerParseFunc: eudore.ControllerBaseParseFunc,
	}
	r.GetFunc("/eudore/debug/router/data", r.getData)
	r.GetFunc("/eudore/debug/router/ui", func(ctx eudore.Context) {
		if StaticHtml != "" {
			ctx.WriteFile(StaticHtml)
		} else {
			ctx.WriteString("breaker not set ui file path.")
		}
	})
	return r, err
}

// RegisterHandler 实现eudore.RouterCore接口，记录全部路由路径。
func (r *RouterDebug) RegisterHandler(method, path string, hs eudore.HandlerFuncs) {
	r.Methods = append(r.Methods, method)
	r.Paths = append(r.Paths, path)
	r.RouterCore.RegisterHandler(method, path, hs)
}

func (r *RouterDebug) getData(ctx eudore.Context) {
	ctx.WriteRender(r)
}

// Set 方法传递路由器的Set行为。
func (r *RouterDebug) Set(key string, i interface{}) error {
	return eudore.ComponentSet(r.router, key, i)
}
