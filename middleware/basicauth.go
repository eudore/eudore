package middleware

import (
	"encoding/base64"
	"fmt"
	"github.com/eudore/eudore"
)

// NewBasicAuthFunc 创建一个Basic auth认证中间件。
//
// 需要realm的值和保存用户密码的map。
func NewBasicAuthFunc(realm string, names map[string]string) eudore.HandlerFunc {
	if realm == "" {
		realm = "Basic"
	} else {
		realm = fmt.Sprintf("Basic realm=\"%s\"", realm)
	}
	checks := make(map[string]struct{}, len(names))
	for name, pass := range names {
		checks[base64.StdEncoding.EncodeToString([]byte(name+":"+pass))] = struct{}{}
	}
	return func(ctx eudore.Context) {
		auth := ctx.GetHeader("Authorization")
		if len(auth) > 5 && auth[:6] == "Basic " {
			_, ok := checks[auth[6:]]
			if ok {
				return
			}
		}
		ctx.SetHeader("WWW-Authenticate", realm)
		ctx.WriteHeader(401)
		ctx.End()
	}
}
