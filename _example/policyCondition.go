package main

/*
pbac 通过策略限制访问权限，每个策略拥有多条描述，按照顺序依次匹配，命中则执行effect。

pbac条件允许多种多样的属性限制方法，额外条件可以使用ram.RegisterCondition函数注册条件。
*/

import (
	"fmt"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/policy"
)

func main() {
	policy.RegisterCondition("browser", func(interface{}) policy.Condition { return nil })
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	policys := policy.NewPolicys()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "Policy", "Resource", "Userid"))
	app.AddMiddleware(policys.HandleHTTP)
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), eudore.HandlerEmpty)
	}
	app.AnyFunc("/* action=hello resource-prefix=/", eudore.HandlerEmpty)

	policys.AddPolicy(&policy.Policy{
		PolicyID:    1,
		Description: "or",
		Statement: []byte(`[
			{"conditions":{"or":{"time":{"before":"2019-9-27"}}}},
			{"effect":true,"conditions":{"or":{"sourceip":["127.0.0.1"],"method":["GET"]}}}
		]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    2,
		Description: "sourceip",
		Statement:   []byte(`[{"effect":true,"action":["1"],"conditions":{"or":{"method":["PUT"],"sourceip":["127.0.0.1","192.0.2.1"]}}}]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    3,
		Description: "params",
		Statement:   []byte(`[{"effect":false,"action":["2"],"conditions":{"params":{"route":["/1","/2"]}}}]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    4,
		Description: "error",
		Statement:   []byte(`[{"effect":true,"action":["2"],"conditions":{"and":"/1","params":[]}}]`),
	})
	policys.AddMember(&policy.Member{
		UserID:   1,
		PolicyID: 1,
	})
	policys.AddMember(&policy.Member{
		UserID:   1,
		PolicyID: 2,
		Index:    10,
	})
	policys.AddMember(&policy.Member{
		UserID:   1,
		PolicyID: 3,
		Index:    10,
	})
	policys.AddMember(&policy.Member{
		UserID:     1,
		PolicyID:   4,
		Index:      11,
		Expiration: time.Now().Add(-100 * time.Hour),
	})

	client := httptest.NewClient(app)
	client.AddHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationJSONUtf8)
	client.AddHeaderValue(eudore.HeaderAuthorization, policys.NewBearer(1, "", "", time.Now().Add(87600*time.Hour).Unix()))
	client.NewRequest("GET", "/policys").Do()
	client.NewRequest("GET", "/1").Do()
	client.NewRequest("GET", "/1").WithRemoteAddr("192.168.1.99").Do()
	client.NewRequest("GET", "/2").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
