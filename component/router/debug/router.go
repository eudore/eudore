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

// StaticHtml 定义ui文件的位置，默认是该文件同名称后缀为html
var StaticHtml = ""

// 获取文件定义位置，静态ui文件在同目录。
func init() {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		StaticHtml = file[:len(file)-2] + "html"
	}
}

// NewRouterDebug 函数创建一个debug路由器，默认封装RouterFull。
func NewRouterDebug() eudore.Router {
	r2 := eudore.NewRouterFull()
	r := &RouterDebug{
		RouterCore: r2,
		router:     r2,
	}
	r.RouterMethod = &eudore.RouterMethodStd{
		RouterCore: r,
	}
	r.GetFunc("/eudore/debug/router/data", r.getData)
	r.GetFunc("/eudore/debug/router/ui", func(ctx eudore.Context) {
		if StaticHtml != "" {
			ctx.WriteFile(StaticHtml)
		} else {
			ctx.WriteString("breaker not set ui file path.")
		}
	})
	return r
}

// RegisterHandler 实现eudore.RouterCore接口，记录全部路由路径。
func (r *RouterDebug) RegisterHandler(method, path string, hs eudore.HandlerFuncs) {
	r.Methods = append(r.Methods, method)
	r.Paths = append(r.Paths, path)
	r.RouterCore.RegisterHandler(method, path, hs)
}

func (r *RouterDebug) getData(ctx eudore.Context) {
	ctx.Render(r)
}

// Set 方法传递路由器的Set行为。
func (r *RouterDebug) Set(key string, i interface{}) (err error) {
	_, err = eudore.Set(r.router, key, i)
	return
}
