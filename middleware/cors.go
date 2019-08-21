package middleware

import (
	"github.com/eudore/eudore"
)

type (
	// Cors 定义Cors对象。
	Cors struct {
		origins []string
		headers map[string]string
	}
)

// NewCors 函数创建应该Cors对象。
//
// 如果origins为空，设置为*。
// 如果Access-Control-Allow-Methods header为空，设置为*。
func NewCors(origins []string, headers map[string]string) *Cors {
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	if headers["Access-Control-Allow-Methods"] == "" {
		headers["Access-Control-Allow-Methods"] = "*"
	}
	return &Cors{
		origins: origins,
		headers: headers,
	}
}

// NewCorsFunc 函数创建应该CORES中间件。
func NewCorsFunc(origins []string, headers map[string]string) eudore.HandlerFunc {
	return NewCors(origins, headers).Handle
}

// Handle 方法实现eudore上下文请求函数。
func (cors *Cors) Handle(ctx eudore.Context) {
	origin := ctx.GetHeader("Origin")
	if origin == "" {
		return
	}

	// 检查是否未同源请求。
	host := ctx.Host()
	if origin == "http://"+host || origin == "https://"+host {
		return
	}

	if !cors.validateOrigin(origin) {
		ctx.WriteHeader(403)
		ctx.End()
		return
	}

	h := ctx.Response().Header()
	if ctx.Method() == eudore.MethodOptions {
		for k, v := range cors.headers {
			h.Add(k, v)
		}
		ctx.WriteHeader(204)
		ctx.End()
	}
	h.Add("Access-Control-Allow-Origin", origin)
}

// validateOrigin 方法检查origin是否合法。
func (cors *Cors) validateOrigin(origin string) bool {
	for _, i := range cors.origins {
		if eudore.MatchStar(origin, i) {
			return true
		}
	}
	return false
}
