package main

/*
先设置bind调用validate app.Binder = eudore.NewValidateBinder(app.Binder)，然后bind时会调用Validate验证数据。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type userRequest struct {
	Username string `validate:"regexp:^[a-zA-Z]*$"`
	Name     string `validate:"nozero"`
	Age      int    `validate:"min:21,max:40"`
	Password string `validate:"len:>7"`
}

func main() {
	app := eudore.NewApp()
	app.Binder = eudore.NewBinderValidate(app.Binder)

	// 上传文件信息
	app.PutFunc("/file/data/:path", func(ctx eudore.Context) {
		var user userRequest
		ctx.Bind(&user)
		ctx.RenderWith(&user, eudore.RenderIndentJSON)
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/file/data/2").WithHeaderValue("Content-Type", "application/json").WithBodyString(`{"username":"abc","name":"eudore","age":21,"password":"12345678"}`).Do().CheckStatus(200).Out()
	client.NewRequest("PUT", "/file/data/2").WithHeaderValue("Content-Type", "application/json").WithBodyString(`{"username":"abc","name":"eudore","age":21,"password":"12345"}`).Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
