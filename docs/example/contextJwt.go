package main

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/eudore/eudore"
)

type ContextJwt struct {
	eudore.Context
}

func init() {
	eudore.RegisterHandlerFunc(func(fn func(ContextJwt)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(ContextJwt{ctx})
		}
	})
}

var hmacSampleSecret = []byte("secret")

func (ctx ContextJwt) ParseJwt() map[string]interface{} {
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
	app.GetFunc("/*", func(ctx ContextJwt) {
		fmt.Println(ctx.ParseJwt())
	})
}
