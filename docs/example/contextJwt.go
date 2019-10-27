package main

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/eudore/eudore"
)

type contextJwt struct {
	eudore.Context
}

func init() {
	eudore.RegisterHandlerFunc(func(fn func(contextJwt)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(contextJwt{ctx})
		}
	})
}

var hmacSampleSecret = []byte("secret")

func (ctx contextJwt) parseJwt() map[string]interface{} {
	tokenString := ctx.GetHeader(eudore.HeaderAuthorization)
	token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
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

func main() {
	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx contextJwt) {
		fmt.Println(ctx.parseJwt())
	})
}
