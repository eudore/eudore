package main

/*
定义一种解析jwt的扩展请求上下文实现解析jwt数据。
*/

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type contextJwt struct {
	eudore.Context
}

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AddHandlerExtend(func(fn func(contextJwt)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(contextJwt{ctx})
		}
	})
	app.GetFunc("/*", func(ctx contextJwt) {
		fmt.Println(ctx.parseJwt())
	})
}

var hmacSampleSecret = []byte("secret")

func (ctx contextJwt) parseJwt() map[string]interface{} {
	tokenString := ctx.GetHeader(eudore.HeaderAuthorization)
	token, err := jwt.Parse(tokenString[7:], func(*jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if err != nil {
		ctx.Error(err)
		return nil
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims
	}
	return nil
}
