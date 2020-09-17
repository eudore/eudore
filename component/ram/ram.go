// Package ram 定义供acl、rbac、pbac公共使用的方法和对象。
package ram

import (
	"github.com/eudore/eudore"
)

type (
	// Handler 定义Ram处理接口
	Handler interface {
		Name() string
		Match(int, string, eudore.Context) (bool, bool)
		// return1 验证结果 return2 是否验证
	}
	// Allow 定义全部允许处理
	Allow struct{}
	// Deny 定义全部拒绝处理
	Deny struct{}
)

var (
	_ Handler = (*Acl)(nil)
	_ Handler = (*Rbac)(nil)
	_ Handler = (*Pbac)(nil)
	_ Handler = (*Allow)(nil)
	_ Handler = (*Deny)(nil)
)

// Name 方法返回deny name。
func (Deny) Name() string {
	return "deny"
}

// Match 方法实现Handler接口。
func (Deny) Match(int, string, eudore.Context) (bool, bool) {
	return false, true
}

// Name 方法返回allow name。
func (Allow) Name() string {
	return "allow"
}

// Match 方法实现andler接口。
func (Allow) Match(int, string, eudore.Context) (bool, bool) {
	return true, true
}

// NewMiddleware 函数使用多个ram.Handler创建一个eudore中间件处理函数。
func NewMiddleware(rams ...Handler) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		// 如果请求用户资源是用户本身的直接通过，UID、UNAME由用户信息中间件加载，userid、username由路由参数加载。
		if ctx.GetParam("userid") == ctx.GetParam("UID") && ctx.GetParam("userid") != "" {
			return
		}
		if ctx.GetParam("username") == ctx.GetParam("UNAME") && ctx.GetParam("username") != "" {
			return
		}

		// 执行ram鉴权逻辑
		action := ctx.GetParam("action")
		if action == "" {
			return
		}

		name, result := MatchAction(rams, ctx, action)
		ctx.AddParam(eudore.ParamRAM, name)
		if !result {
			forbiddenFunc(ctx)
		}
	}
}

// MatchAction 方法使用[]ram.Handler判断是否匹配指定action。
func MatchAction(rams []Handler, ctx eudore.Context, action string) (string, bool) {
	uid := eudore.GetStringInt(ctx.GetParam("UID"))
	for {
		// 依次检查每种ram是否匹配
		for _, ram := range rams {
			result, ok := ram.Match(uid, action, ctx)
			if ok {
				return ram.Name(), result
			}
		}
		// 执行非0和0两种userid匹配,用户0相当于用户的默认的权限。
		if uid == 0 {
			break
		} else {
			uid = 0
		}
	}
	return "deny", false
}

func forbiddenFunc(ctx eudore.Context) {
	ctx.WriteHeader(403)
	ctx.Render(map[string]string{
		eudore.ParamRAM:    ctx.GetParam("ram"),
		eudore.ParamAction: ctx.GetParam("action"),
		"message":          "Forbidden",
	})
	ctx.End()
}
