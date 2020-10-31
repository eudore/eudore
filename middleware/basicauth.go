package middleware

import (
	"encoding/base64"
	"github.com/eudore/eudore"
)

// NewBasicAuthFunc 创建一个Basic auth认证中间件。
//
// names为保存用户密码的map。
func NewBasicAuthFunc(names map[string]string) eudore.HandlerFunc {
	checks := make(map[string]string, len(names))
	for name, pass := range names {
		checks[base64.StdEncoding.EncodeToString([]byte(name+":"+pass))] = name
	}
	return func(ctx eudore.Context) {
		auth := ctx.GetHeader(eudore.HeaderAuthorization)
		if len(auth) > 5 && auth[:6] == "Basic " {
			name, ok := checks[auth[6:]]
			if ok {
				ctx.SetParam("basicauth", name)
				return
			}
		}
		ctx.SetHeader(eudore.HeaderWWWAuthenticate, "Basic")
		ctx.WriteHeader(401)
		ctx.End()
	}
}
