// Package ram 定义供acl、rbac、pbac公共使用的方法和对象。
package ram

import (
	"github.com/eudore/eudore"
)

const (
	// UserIdString 定义获取用户id的参数名称。
	UserIdString = "UID"
)

type (
	// GetActionFunc 定义Ram获得Action的函数。
	GetActionFunc func(eudore.Context) string
	// GetIdFunc 定义Ram获得id的函数。
	GetIdFunc func(eudore.Context) int
	// ForbiddenFunc 定义Ram执行403的处理函数。
	ForbiddenFunc func(eudore.Context)
	// RamHandleFunc 定义Ram处理一个认证请求的函数。
	RamHandleFunc func(int, string, eudore.Context) (bool, bool)
	// RamHandler 定义Ram处理接口
	RamHandler interface {
		RamHandle(int, string, eudore.Context) (bool, bool)
		// return1 验证结果 return2 是否验证

	}
	// RamHttp 定义Ram处理一个eudore http请求上下文。
	RamHttp struct {
		RamHandler
		GetId     GetIdFunc
		GetAction GetActionFunc
		Forbidden ForbiddenFunc
	}
	// RamAny 是一个Ram组合处理，多个Ram任意一个通过即可。
	RamAny struct {
		Rams []RamHandler
	}
	// RanAnd 是一个Ram组合处理，要求多个Ram全部通过。
	RanAnd struct {
		Rams []RamHandler
	}
	// Deny 定义默认全部拒绝处理
	Deny struct{}
	// Allow 定义默认全部通过处理
	Allow struct{}
)

var (
	// DenyHander 是拒绝处理者
	DenyHander = Deny{}
	// AllowHanlder 是允许处理者
	AllowHanlder = Allow{}
)

// NewRamHttp 创建一个eudore请求处理者，需要给予Ram处理者。
//
// 如果没有Ram处理者默认使用全部拒绝，如果有多个Ram处理者，使用或逻辑。多Ram任意一个允许即可。
func NewRamHttp(rams ...RamHandler) *RamHttp {
	r := &RamHttp{
		GetId:     GetIdDefault,
		GetAction: GetActionDefault,
		Forbidden: ForbiddenDefault,
	}
	switch len(rams) {
	case 0:
		r.RamHandler = AllowHanlder
	case 1:
		r.RamHandler = rams[0]
	default:
		r.RamHandler = NewRamAny(rams...)
	}
	return r
}

// Handle 方法实现eudore请求上下文处理函数。
func (r *RamHttp) NewRamFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		action := r.GetAction(ctx)
		if len(action) > 0 && !HandleDefaultRam(r.GetId(ctx), action, ctx, r.RamHandler.RamHandle) {
			r.Forbidden(ctx)
		}
	}
}

// NewRamAny 函数创建一个或逻辑的ram组合处理者。
func NewRamAny(hs ...RamHandler) *RamAny {
	return &RamAny{
		Rams: hs,
	}
}

// RamHandle 方法实现RamHandler接口。
func (r *RamAny) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	for _, h := range r.Rams {
		isgrant, ok := h.RamHandle(id, action, ctx)
		if ok {
			ctx.SetParam(eudore.ParamRAM, "ram-"+ctx.GetParam(eudore.ParamRAM))
			return isgrant, true
		}
	}
	ctx.SetParam(eudore.ParamRAM, "ram-default")
	return false, false
}

// Handle 方法实现eudore请求上下文处理函数。
func (r *RamAny) Handle(ctx eudore.Context) {
	DefaultHandle(ctx, r)
}

// RamHandle 方法实现RamHandler接口。
func (Deny) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRAM, "deny")
	return false, true
}

// RamHandle 方法实现RamHandler接口。
func (Allow) RamHandle(id int, action string, ctx eudore.Context) (bool, bool) {
	ctx.SetParam(eudore.ParamRAM, "allow")
	return true, true
}
