package middleware

import (
	"net/textproto"
	"strings"

	"github.com/eudore/eudore"
)

// NewCorsFunc 函数创建一个Cors处理函数。
//
// origins是允许的origin，headers是跨域验证成功的添加的headers，例如："Access-Control-Allow-Credentials"、"Access-Control-Allow-Headers"等。
//
// 如果origins为空，设置为*。
// 如果Access-Control-Allow-Methods header为空，设置为*。
//
// Cors中间件注册不是全局中间件时，需要最后注册一次Options /*或404方法，否则Options请求匹配了默认404没有经过Cors中间件处理。
func NewCorsFunc(origins []string, headers map[string]string) eudore.HandlerFunc {
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	for k, v := range headers {
		delete(headers, k)
		headers[textproto.CanonicalMIMEHeaderKey(k)] = v
	}
	if headers["Access-Control-Allow-Methods"] == "" {
		headers["Access-Control-Allow-Methods"] = "*"
	}
	return func(ctx eudore.Context) {
		origin := ctx.GetHeader("Origin")
		// 检查是否未同源请求,cors和upgrade时存在origin header。
		if origin == "" || ctx.GetHeader(eudore.HeaderUpgrade) != "" {
			return
		}
		origin = strings.TrimPrefix(strings.TrimPrefix(origin, "http://"), "https://")

		if !validateOrigin(origins, origin) {
			ctx.WriteHeader(403)
			ctx.End()
			return
		}

		h := ctx.Response().Header()
		h.Add("Access-Control-Allow-Origin", ctx.GetHeader("Origin"))
		if ctx.Method() == eudore.MethodOptions {
			for k, v := range headers {
				h[k] = append(h[k], v)
			}
			ctx.WriteHeader(204)
			ctx.End()
		}
	}
}

// validateOrigin 方法检查origin是否合法。
func validateOrigin(origins []string, origin string) bool {
	for _, i := range origins {
		if matchStar(origin, i) {
			return true
		}
	}
	return false
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
