package ram

import (
	"github.com/eudore/eudore"
)


func DefaultHandle(ctx eudore.Context, r RamHandler) {
	action := GetActionDefault(ctx)
	if len(action) > 0 && !HandleDefaultRam(GetIdDefault(ctx), action, ctx, r.RamHandle) {
		ForbiddenDefault(ctx)	
	}
}

// 处理默认id为0的验证方法
func HandleDefaultRam(id int, action string, ctx eudore.Context,fn RamHandleFunc) bool {
	// 验证权限

	if isgrant, ok := fn(id, action, ctx); ok {
		return isgrant
	}
	// 处理默认权限
	if id != 0 {
		if isgrant, ok := fn(0, action, ctx); ok {
			return isgrant
		}
	}
	return false
}


func GetActionDefault(ctx eudore.Context) string {
	return ctx.GetParam(eudore.ParamAction)
}

func ForbiddenDefault(ctx eudore.Context) {
	ctx.WriteHeader(403)
	ctx.WriteRender(map[string]interface{}{
		eudore.ParamRam:		ctx.GetParam(eudore.ParamRam),
		eudore.ParamAction:		ctx.GetParam(eudore.ParamAction),
	})
	ctx.End()
}

func GetIdDefault(ctx eudore.Context) (id int) {
	return eudore.GetInt(ctx.GetParam(UserIdString))
}