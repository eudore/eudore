package middleware

import (
	"net/textproto"
	"strings"

	"github.com/eudore/eudore"
)

// NewCorsFunc 函数创建一个Cors处理函数。
//
// pattens是允许的origin，headers是跨域验证成功的添加的headers，例如："Access-Control-Allow-Credentials"、"Access-Control-Allow-Headers"等。
//
// 如果pattens为空，允许任意origin。
// 如果Access-Control-Allow-Methods header为空，设置为*。
//
// Cors中间件注册不是全局中间件时，需要最后注册一次Options /*或404方法，否则Options请求匹配了默认404没有经过Cors中间件处理。
func NewCorsFunc(pattens []string, headers map[string]string) eudore.HandlerFunc {
	corsHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		corsHeaders[textproto.CanonicalMIMEHeaderKey(k)] = v
	}
	if corsHeaders[eudore.HeaderAccessControlAllowMethods] == "" {
		corsHeaders[eudore.HeaderAccessControlAllowMethods] = "*"
	}
	return func(ctx eudore.Context) {
		origin := ctx.GetHeader(eudore.HeaderOrigin)
		pos := strings.Index(origin, "://")
		if pos != -1 {
			origin = origin[pos+3:]
		}
		// 检查是否未同源请求,cors和upgrade时存在origin header。
		if origin == "" || origin == ctx.Host() {
			return
		}

		if !validateOrigin(pattens, origin) {
			ctx.WriteHeader(eudore.StatusForbidden)
			ctx.End()
			return
		}

		h := ctx.Response().Header()
		h.Add(eudore.HeaderAccessControlAllowOrigin, ctx.GetHeader(eudore.HeaderOrigin))
		if ctx.Method() == eudore.MethodOptions {
			for k, v := range corsHeaders {
				h[k] = append(h[k], v)
			}
			ctx.WriteHeader(eudore.StatusNoContent)
			ctx.End()
		}
	}
}

// validateOrigin 方法检查origin是否合法。
func validateOrigin(pattens []string, origin string) bool {
	for _, patten := range pattens {
		if matchStar(patten, origin) {
			return true
		}
	}
	return pattens == nil
}

// matchStar 模式匹配对象，允许使用带'*'的模式。
func matchStar(patten, obj string) bool {
	parts := strings.Split(patten, "*")
	if len(parts) < 2 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, parts[0]) {
		return false
	}
	for _, i := range parts {
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
