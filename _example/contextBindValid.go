package main

/*

Validater是Bind
*/

import (
	"strings"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

type userRequest struct {
	Username string `validate:"regexp:^[a-zA-Z]*$"`
	Name     string `validate:"nozero"`
	Age      int    `validate:"min:21,max:40"`
	Password string `validate:"len>7"`
}

func main() {
	app := eudore.NewApp()
	// 设置Bind，在Bind成功后执行Validater。
	app.SetValue(eudore.ContextKeyBind, eudore.NewHandlerDataFuncs(
		eudore.NewHandlerDataBinds(nil),
		eudore.NewHandlerDataValidateStruct(app),
		// 自定义测试Hook
		func(ctx eudore.Context, data any) error {
			app.Debugf("bind %s data: %#v", ctx.Path(), data)
			return nil
		},
	))
	// 使用Context配置生效
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AddMiddleware(middleware.NewLoggerFunc(app.Logger))
	app.PutFunc("/user/:name", func(ctx eudore.Context) error {
		var user userRequest
		return ctx.Bind(&user)
	})

	app.NewRequest("PUT", "/user/1",
		eudore.NewClientHeader("Content-Type", "application/json"),
		strings.NewReader(`{"username":"abc","name":""}`),
	)
	app.NewRequest("PUT", "/user/2",
		eudore.NewClientHeader("Content-Type", "application/json"),
		strings.NewReader(`{"username":"abc","name":"eudore","age":20}`),
	)
	app.NewRequest("PUT", "/user/3",
		eudore.NewClientHeader("Content-Type", "application/json"),
		strings.NewReader(`{"username":"abc","name":"eudore","age":22,"password":"12345"}`),
	)
	app.NewRequest("PUT", "/user/4",
		eudore.NewClientHeader("Content-Type", "application/json"),
		strings.NewReader(`{"username":"abc","name":"eudore","age":22,"password":"12345678"}`),
	)

	app.Listen(":8088")
	app.Run()
}
