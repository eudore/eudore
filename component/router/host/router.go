package host

import (
	"github.com/eudore/eudore"
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

func init() {
	eudore.RegisterComponent(eudore.ComponentRouterHostName, func(arg interface{}) (eudore.Component, error) {
		return NewRouterHost(arg)
	})
}

func InitAddHost(ctx eudore.Context) {
	ctx.AddParam("host", ctx.Host())
}

func NewRouterHost(interface{}) (eudore.Router, error) {
	r := &RouterHost{}
	r.RouterMethod = &eudore.RouterMethodStd{RouterCore: r}
	return r, nil
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
		if eudore.MatchStar(host, h) {
			return r.Routers[i]
		}
	}
	return r.Default
}

func (r *RouterHost) RegisterHost(host string, router eudore.Router) {
	r.Hosts = append(r.Hosts, host)
	r.Routers = append(r.Routers, router)
}

func (r *RouterHost) RegisterMiddleware(method string, path string, handler eudore.HandlerFuncs) {
	r.getRouter(path).RegisterMiddleware(method, path, handler)
}

func (r *RouterHost) RegisterHandler(method string, path string, handler eudore.HandlerFuncs) {
	r.getRouter(path).RegisterHandler(method, path, handler)
}

func (r *RouterHost) Match(method, path string, params eudore.Params) eudore.HandlerFuncs {
	return r.getRouter(params.GetParam("host")).Match(method, path, params)
}

func (r *RouterHost) GetName() string {
	return eudore.ComponentRouterHostName
}

func (r *RouterHost) Version() string {
	return eudore.ComponentRouterHostVersion
}
