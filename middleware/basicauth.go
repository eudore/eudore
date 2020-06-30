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
		ctx.SetHeader(eudore.HeaderWWWAuthenticate, realm)
		ctx.WriteHeader(401)
		ctx.End()
	}
}
