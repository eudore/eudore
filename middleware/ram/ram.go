/*
定义供acl、rbac、pbac公共使用的方法和对象。
*/
package ram

import (
	"eudore"
)

const (
	UserIdString			=	"uid"
	ForbiddenString			=	"Forbidden"
)

type (
	GetActionFunc func(eudore.Context) string
	GetIdFunc func(eudore.Context) int
	ForbiddenFunc func(eudore.Context)
	RamHandleFunc func(int, string, eudore.Context) (bool, bool)
	RamHandler interface {
		RamHandle(int, string, eudore.Context) (bool, bool)
		// return1 验证结果 return2 是否验证
		// RamHandle(int, string, eudore.Context) (bool, bool)

	}
	RamHttp struct {
		RamHandler
		GetId		GetIdFunc
		GetAction	GetActionFunc
		Forbidden	ForbiddenFunc
	}
	RamAny struct {
		Rams		[]RamHandler	
	}
	RanAnd struct {
		Rams		[]RamHandler
	}
	Deny struct{}
	Allow struct{}
)

var (
	DenyHander = Deny{}
	AllowHanlder = Allow{}
)

func NewRamHttp(rams ...RamHandler) *RamHttp {
	r := &RamHttp{
		GetId:			GetIdDefault,
		GetAction:		GetActionDefault,
		Forbidden:		ForbiddenDefault,
	}
	switch len(rams) {
	case 0:
	case 1:
		r.RamHandler = rams[0]
	default:
		r.RamHandler = NewRamAny(rams...)
	}
	return r
}

func (r *RamHttp) Handle(ctx eudore.Context) {
	action := r.GetAction(ctx)
	if len(action) > 0 && !HandleDefaultRam(r.GetId(ctx), action, ctx, r.RamHandler.RamHandle) {
		r.Forbidden(ctx)	
	}
}

//
func (r *RamHttp) Set(f1 GetIdFunc, f2 GetActionFunc, f3 ForbiddenFunc) *RamHttp {
	if f1 != nil {
		r.GetId = f1
	}
	if f2 != nil {
		r.GetAction = f2
	}
	if f3 != nil {
		r.Forbidden = f3
	}
	return r
}


func NewRamAny(hs ...RamHandler) *RamAny {
	return &RamAny{
		Rams:		hs,
	}
}

func (r *RamAny) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	for _, h := range r.Rams {
		isgrant, ok := h.RamHandle(id, action, ctx)
		if ok {
			ctx.SetParam(eudore.ParamRam, "arm-" + ctx.GetParam(eudore.ParamRam))
			return isgrant, true
		}
	}
	ctx.SetParam(eudore.ParamRam, "arm-default")
	return false, false
}

func (r *RamAny) Handle(ctx eudore.Context) {
	DefaultHandle(ctx, r)
}



func (Deny) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "deny")
	return false, true
}

func (Allow) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRam, "allow")
	return true, true
}
