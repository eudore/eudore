package main

/*
pbac 通过策略限制访问权限，每个策略拥有多条描述，按照顺序依次匹配，命中则执行effect。

pbac条件允许多种多样的属性限制方法，额外条件可以使用ram.RegisterCondition函数注册条件。
*/

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/middleware/ram"
)

func main() {
	pbac := ram.NewPbac()
	pbac.AddPolicyStringJson(1, `{"version":"1","description":"AdministratorAccess","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}`)
	pbac.AddPolicyStringJson(2, `{"version":"1","description":"Get method allow","statement":[{"effect":true,"action":["*"],"resource":["*"],"conditions":{"method":["GET"]}}]}`)
	pbac.AddPolicyStringJson(3, `{"version":"1","description":"Get method allow","statement":[{"effect":true,"action":["3"],"resource":["*"]}]}`)
	// pbac.BindPolicy(1,1)
	pbac.BindPolicy(1, 2)
	pbac.BindPolicy(1, 3)

	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "action", "ram", "route", "resource", "browser"))
	// 测试给予参数 UID=2  即用户id为2，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, "1")
	})
	app.AddMiddleware(ram.NewMiddleware(pbac))
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), fmt.Sprintf("hello %d", i))
	}
	app.AnyFunc("/* action=hello", "hello")

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/hello").Do().CheckStatus(200)
	client.NewRequest("GET", "/1").Do().CheckStatus(200)
	client.NewRequest("GET", "/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4").Do().CheckStatus(200)
	client.NewRequest("PUT", "/5").Do().CheckStatus(200)
	client.NewRequest("PUT", "/6").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
