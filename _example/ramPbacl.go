package main

/*
pbac 通过策略限制访问权限，每个策略拥有多条描述，按照顺序依次匹配，命中则执行effect。

pbac条件允许多种多样的属性限制方法，额外条件可以使用ram.RegisterCondition函数注册条件。
*/

import (
	"encoding/json"
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/ram"
	"github.com/eudore/eudore/middleware"
)

func main() {
	pbac := ram.NewPbac()
	pbac.AddPolicyStringJson(1, `{"version":"1","description":"access action 1","statement":[{"effect":true,"action":["1"],"resource":["*"]}]}`)
	pbac.AddPolicyStringJson(2, `{"version":"1","description":"get resource /2","statement":[{"effect":true,"action":["*"],"resource":["/2"],"conditions":{"method":["GET"]}}]}`)
	pbac.AddPolicyStringJson(3, `{"version":"1","description":"test error","statement":[{"effect":true,"action":["*"],"resource":["/2"],"conditions":{"browser":["GET"],"or":"or","and":"and"}},{"effect":"error test"}]}`)
	pbac.AddPolicyStringJson(4, `{"version":"1","description":"test or、get","statement":[{"effect":true,"action":["*"],"resource":["*"],"conditions":{"or":{"method":["GET"],"method":["POST"]}}}]}`)
	pbac.AddPolicyStringJson(5, `{"version":"1","description":"test time","statement":[{"effect":true,"action":["*"],"resource":["*"],"conditions":{"or":{"method":["GET"],"time":{"after":"2020-01-01"}}}}]}`)
	pbac.AddPolicyStringJson(6, `{"version":"1","description":"local api","statement":[{"effect":true,"action":["*"],"resource":["api/*"],"conditions":{"and":{"method":["GET"],"sourceip":["127.0.0.1","192.168.0.0/24"]}}}]}`)
	pbac.AddPolicyStringJson(7, `{"version":"1","description":"AdministratorAccess","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}`)
	pbac.AddPolicyStringJson(8, `{"version":"1","description":"AdministratorAccess","statement":[{"effect":true,"action":["*"],"resource":["api/*/user/*"]}]}`)

	// id=1 测试stmt
	pbac.BindPolicy(1, 0, 1)
	pbac.BindPolicy(1, 100, 2)
	// id=2 or time method
	pbac.BindPolicy(2, 1, 4)
	pbac.BindPolicy(2, 1, 5)
	// id=3 and sourceip
	pbac.BindPolicyString(3, 0, "6,7,xxx")
	pbac.DeletePolicy(7)
	// id=4 matchStar
	pbac.BindPolicy(4, 1, 8)

	body, err := json.Marshal(pbac.Policys)
	fmt.Println(string(body), err)

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	// 测试给予id等于请求参数id，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, ctx.GetQuery("id"))
	})
	app.AddMiddleware(ram.NewMiddleware(pbac))
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), eudore.HandlerEmpty)
	}
	app.AnyFunc("/* action=hello resource-prefix=/", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1?id=1").Do().CheckStatus(200)
	client.NewRequest("GET", "/2?id=1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/hello?id=1").Do().CheckStatus(403)

	client.NewRequest("GET", "/1?id=2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/1?id=2").Do().CheckStatus(200)

	client.NewRequest("GET", "/api/v1?id=3").Do().CheckStatus(403)
	client.NewRequest("GET", "/api/v1?id=3").WithRemoteAddr("127.0.0.1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/api/v1?id=3").Do().CheckStatus(403)

	client.NewRequest("PUT", "/api/v1?id=4").Do().CheckStatus(403)
	client.NewRequest("PUT", "/api1/v1?id=4").Do().CheckStatus(403)
	client.NewRequest("PUT", "/api/v1/auth/2?id=4").Do().CheckStatus(403)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
