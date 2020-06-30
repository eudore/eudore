package host

import (
	"strings"

	"github.com/eudore/eudore"
)

// RouterHost 使用Host匹配执行路由行为，仅对RouterCore部分方法有效。
type RouterHost struct {
	eudore.Router
	core *RouterCoreHost
}

// RouterCoreHost 实现根据host值执行RouterCore行为。
type RouterCoreHost struct {
	Routers map[string]eudore.Router
	Hosts   []string
	Matchs  []eudore.Router
}

var _ eudore.Router = (*RouterHost)(nil)
var _ eudore.RouterCore = (*RouterCoreHost)(nil)

// AddHostHandler 函数将host值加入到Params中
func AddHostHandler(ctx eudore.Context) {
	ctx.AddParam("host", ctx.Host())
}

// NewRouterHost 函数创建一个默认的hst路由。
func NewRouterHost() *RouterHost {
	hostcore := &RouterCoreHost{
		Routers: map[string]eudore.Router{
			"": eudore.NewRouterRadix(),
		},
	}
	return &RouterHost{
		Router: eudore.NewRouterStd(hostcore),
		core:   hostcore,
	}
}

// GetRouter 方法获取一个域名对应的路由器。
func (r *RouterHost) GetRouter(path string) eudore.Router {
	return r.core.getRouter(path)
}

// SetRouter 方法设置一个域名使用的路由器
func (r *RouterHost) SetRouter(host string, router eudore.Router) {
	r.core.setRouter(host, router)
}

func (r *RouterCoreHost) getRouter(host string) eudore.Router {
	router, ok := r.Routers[host]
	if ok {
		return router
	}
	for i, h := range r.Hosts {
		if matchStar(h, host) {
			return r.Matchs[i]
		}
	}
	return r.Routers[""]
}

func (r *RouterCoreHost) getRouters(host string) (rs []eudore.Router) {
	// 获取常量host
	router, ok := r.Routers[host]
	if ok {
		return []eudore.Router{router}
	}
	// 获取模式host
	for i, h := range r.Hosts {
		if matchStar(h, host) {
			rs = append(rs, r.Matchs[i])
			return
		}
	}
	// 返回默认host
	if len(rs) == 0 {
		rs = []eudore.Router{r.Routers[""]}
	}
	return
}

// setRouter 给Host路由器注册域名的子路由器。
//
// 如果host为空字符串，设置为默认子路由器。
func (r *RouterCoreHost) setRouter(host string, router eudore.Router) {
	pos := strings.IndexByte(host, '*')
	if pos == -1 {
		r.Routers[host] = router
		return
	}
	for i, h := range r.Hosts {
		if h == host {
			r.Matchs[i] = router
			return
		}
	}
	r.Hosts = append(r.Hosts, host)
	r.Matchs = append(r.Matchs, router)
}

// HandleFunc 方法根据host选择路由器然后注册路由。
func (r *RouterCoreHost) HandleFunc(method string, path string, handler eudore.HandlerFuncs) {
	for _, i := range r.getRouters(getHostOfPath(path)) {
		i.HandleFunc(method, path, handler)
	}
}

// Match 方法根据host选择路由器然后匹配请求。
func (r *RouterCoreHost) Match(method, path string, params *eudore.Params) eudore.HandlerFuncs {
	return r.getRouter(params.Get("host")).Match(method, path, params)
}

// Set 方法传递路由器的Set行为。
func (r *RouterCoreHost) Set(key string, i interface{}) (err error) {
	for _, r2 := range r.Routers {
		eudore.Set(r2, key, i)
	}
	return
}

// getHostOfPath 函数从path中提取到host的值。
func getHostOfPath(path string) string {
	args := strings.Split(path, " ")
	for _, arg := range args[1:] {
		if arg[:5] == "host=" {
			return arg[6:]
		}
	}
	return ""
}

// matchStar 模式匹配对象，允许使用带'*'的模式。
func matchStar(obj, patten string) bool {
	ps := strings.Split(patten, "*")
	if len(ps) < 2 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, ps[0]) {
		return false
	}
	for _, i := range ps {
		if i == "" {
			continue
		}
		pos := strings.Index(obj, i)
		if pos == -1 {
			return false
		}
		obj = obj[pos+len(i):]
	}
	return true
}
