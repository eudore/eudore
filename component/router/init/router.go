package init

import (
	"github.com/eudore/eudore"
)

// 将处理函数转换成路由器。
// 设计目标是处理初始化时的请求，所有注册代理给目标路由器，初始化完成后被目标路由器替换。
type RouterInit struct {
	eudore.Router
	hs eudore.HandlerFuncs
}

func init() {
	eudore.RegisterComponent(eudore.ComponentRouterInitName, func(arg interface{}) (eudore.Component, error) {
		return NewRouterInit(arg)
	})
}

func NewRouterInit(arg interface{}) (eudore.Router, error) {
	h, ok := arg.(eudore.HandlerFunc)
	if !ok {
		h, ok = arg.(func(eudore.Context))
	}
	if !ok {
		return nil, eudore.ErrRouterSetNoSupportType
	}
	return &RouterInit{hs: eudore.HandlerFuncs{h}}, nil
}

func (r *RouterInit) Match(string, string, eudore.Params) eudore.HandlerFuncs {
	// Do nothing because empty router does not process entries.
	return r.hs
}

func (*RouterInit) GetName() string {
	return eudore.ComponentRouterInitName
}

func (*RouterInit) Version() string {
	return eudore.ComponentRouterInitVersion
}
