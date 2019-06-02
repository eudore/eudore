package debug

import (
	"github.com/eudore/eudore"
)

type (
	RouterDebug struct {
		eudore.RouterMethod		`json:"-" xml:"-" set:"-"`
		eudore.RouterCore		`json:"-" xml:"-" set:"-"`
		router		eudore.Router	`json:"-" xml:"-" set:"-"`
		Methods		[]string	`json:"methods"`
		Paths		[]string	`json:"paths"`
	}
	RouterDebugCore struct {
		eudore.RouterCore		`json:"-" set:"-"`
	}
)

func init() {
	eudore.RegisterComponent(eudore.ComponentRouterDebugName, func(arg interface{}) (eudore.Component, error) {
		return NewRouterDebug(arg)
	})
}

func NewRouterDebug(i interface{}) (eudore.Router, error) {
	r2, err := eudore.NewRouterFull(i)
	r := &RouterDebug{
		RouterCore: r2,
		router:		r2,
	}
	r.RouterMethod = &eudore.RouterMethodStd{
		RouterCore:			r,
		ControllerParseFunc:	eudore.ControllerBaseParseFunc,
	}
	r.GetFunc("/eudore/debug/router/data", r.getData)
	r.GetFunc("/eudore/debug/router/ui", func(ctx eudore.Context) {
		if UIpath != "" {
			ctx.WriteFile(UIpath)
		}else {
			ctx.WriteString(UIString)
		}
	})
	return r, err
}

func (r *RouterDebug) RegisterHandler(method, path string, hs eudore.HandlerFuncs) {
	r.Methods = append(r.Methods, method)
	r.Paths = append(r.Paths, path)
	r.RouterCore.RegisterHandler(method, path, hs)
}

func (r *RouterDebug) getData(ctx eudore.Context) {
	ctx.WriteRender(r)
}

func (r *RouterDebug) Set(key string, i interface{}) error {
	return eudore.ComponentSet(r.router, key, i)
}

func (*RouterDebug) GetName() string {
	return eudore.ComponentRouterDebugName
}

func (*RouterDebug) Version() string {
	return eudore.ComponentRouterDebugVersion
}

