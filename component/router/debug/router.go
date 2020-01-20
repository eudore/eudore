package debug

import (
	"github.com/eudore/eudore"
	"runtime"
)

type (
	// RouterCoreDebug 定义debug路由器。
	RouterCoreDebug struct {
		eudore.RouterCore `json:"-" xml:"-" set:"-"`
		Methods           []string `json:"methods"`
		Paths             []string `json:"paths"`
	}
	// RouterDebugCore 定义debug路由器核心。
	RouterDebugCore struct {
		eudore.RouterCore `json:"-" set:"-"`
	}
)

var _ eudore.RouterCore = (*RouterCoreDebug)(nil)

// StaticHTML 定义ui文件的位置，默认是该文件同名称后缀为html
var StaticHTML = ""

// 获取文件定义位置，静态ui文件在同目录。
func init() {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		StaticHTML = file[:len(file)-2] + "html"
	}
}

// NewRouterDebug 函数创建一个debug路由器，默认使用RouterCoreFull为核心。
func NewRouterDebug() eudore.Router {
	return NewRouterDebugWithCore(eudore.NewRouterFull())
}

// NewRouterDebugWithCore 函数指定路由核心创建一个debug路由器。
func NewRouterDebugWithCore(core eudore.RouterCore) eudore.Router {
	r := &RouterCoreDebug{
		RouterCore: core,
	}
	r.RegisterHandler("GET", "/eudore/debug/router/data", eudore.HandlerFuncs{r.getData})
	r.RegisterHandler("GET", "/eudore/debug/router/ui", eudore.HandlerFuncs{r.showUI})
	return eudore.NewRouterStd(r)
}

// RegisterHandler 实现eudore.RouterCore接口，记录全部路由路径。
func (r *RouterCoreDebug) RegisterHandler(method, path string, hs eudore.HandlerFuncs) {
	r.Methods = append(r.Methods, method)
	r.Paths = append(r.Paths, path)
	r.RouterCore.RegisterHandler(method, path, hs)
}

// getData 方法返回debug路由信息数据。
func (r *RouterCoreDebug) getData(ctx eudore.Context) {
	ctx.Render(r)
}

// showUI 方法返回debug路由器的html ui。
func (r *RouterCoreDebug) showUI(ctx eudore.Context) {
	if StaticHTML != "" {
		ctx.WriteFile(StaticHTML)
	} else {
		ctx.WriteString("breaker not set ui file path.")
	}
}
