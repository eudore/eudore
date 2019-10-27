package host

import (
	"github.com/eudore/eudore"
	"path"
	"strings"
)

// RouterHost 使用Host匹配进行路由。
type RouterHost struct {
	eudore.RouterMethod
	Default eudore.Router
	Hosts   []string
	Routers []eudore.Router
}

var _ eudore.Router = (*RouterHost)(nil)

// InitAddHost 定义添加host参数的全局中间件函数，需要使用Eudore，Core无法使用该功能。
func InitAddHost(ctx eudore.Context) {
	ctx.AddParam("host", ctx.Host())
}

// NewRouterHost 函数创建一个默认的hst路由。
func NewRouterHost() *RouterHost {
	r := &RouterHost{
		Default: eudore.NewRouterRadix(),
	}
	r.RouterMethod = &eudore.RouterMethodStd{RouterCore: r}
	return r
}

func (r *RouterHost) getRouter(path string) eudore.Router {
	args := strings.Split(path, " ")
	for _, arg := range args[1:] {
		if arg[:5] == "host=" {
			r := r.matchRouter(arg[1:])
			if r != nil {
				return r
			}
		}
	}
	return r.Default
}

func (r *RouterHost) matchRouter(host string) eudore.Router {
	for i, h := range r.Hosts {
		if b, _ := path.Match(h, host); b {
			return r.Routers[i]
		}
	}
	return r.Default
}

// RegisterHost 给Host路由器注册域名的子路由器。
//
// 如果host为空字符串，设置为默认子路由器。
func (r *RouterHost) RegisterHost(host string, router eudore.Router) {
	if host == "" {
		r.Default = router
		return
	}
	for i, h := range r.Hosts {
		if h == host {
			r.Routers[i] = router
			return
		}
	}
	r.Hosts = append(r.Hosts, host)
	r.Routers = append(r.Routers, router)
}

// RegisterMiddleware 方法根据host选择路由器然后注册中间件。
func (r *RouterHost) RegisterMiddleware(path string, handler eudore.HandlerFuncs) {
	r.getRouter(path).RegisterMiddleware(path, handler)
}

// RegisterHandler 方法根据host选择路由器然后注册路由。
func (r *RouterHost) RegisterHandler(method string, path string, handler eudore.HandlerFuncs) {
	r.getRouter(path).RegisterHandler(method, path, handler)
}

// Match 方法根据host选择路由器然后匹配请求。
func (r *RouterHost) Match(method, path string, params eudore.Params) eudore.HandlerFuncs {
	return r.matchRouter(params.Get("host")).Match(method, path, params)
}

// Set 方法传递路由器的Set行为。
func (r *RouterHost) Set(key string, i interface{}) (err error) {
	_, err = eudore.Set(r.Default, key, i)
	for _, r2 := range r.Routers {
		eudore.Set(r2, key, i)
	}
	return
}
