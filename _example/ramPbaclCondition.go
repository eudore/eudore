package main

/*
使用RegisterCondition扩展一个pbac条件，例如要求https请求。
ConditionBrowser扩展: https://github.com/eudore/website/blob/master/framework/rambrowser.go
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/ram"
	"github.com/eudore/eudore/middleware"
)

func main() {
	ram.RegisterCondition("https", newHttpsCondistion)
	pbac := ram.NewPbac()
	pbac.AddPolicyStringJSON(1, `{"version":"1","description":"get resource /2","statement":[{"effect":true,"action":["*"],"resource":["/2"],"conditions":{"https":true}}]}`)
	pbac.BindPolicy(1, 0, 1)

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	// 测试给予id等于请求参数id，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, ctx.GetQuery("id"))
	})
	app.AddMiddleware(ram.NewMiddleware(pbac))
	app.AnyFunc("/* action=hello resource-prefix=/", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/hello?id=1").WithTLS().Do().CheckStatus(200)
	client.NewRequest("PUT", "/hello?id=1").Do().CheckStatus(403)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type httpsCondition struct {
	is bool
}

func newHttpsCondistion(i interface{}) ram.Condition {
	b, ok := i.(bool)
	if !ok {
		return nil
	}
	return httpsCondition{b}
}

// Name 方法返回条件名称。
func (cond httpsCondition) Name() string {
	return "https"
}

// Match 方法匹配or条件。
func (cond httpsCondition) Match(ctx eudore.Context) bool {
	return ctx.Istls() == cond.is
}
