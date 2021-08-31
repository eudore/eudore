package main

/*
pbac 通过策略限制访问权限，每个策略拥有多条描述，按照顺序依次匹配，命中则执行effect。

pbac条件允许多种多样的属性限制方法，额外条件可以使用ram.RegisterCondition函数注册条件。
*/

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/policy"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	os.Remove("./policy.db")
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	policys := policy.NewPolicys()
	db, err := sql.Open("sqlite3", "./policy.db")
	// db, err = sql.Open("postgres", "host=172.17.0.1 port=5432 user=jass password=TPG4ppk4rlncL3lO dbname=jass sslmode=disable")
	if err != nil {
		app.Options(err)
		return
	}
	defer os.Remove("./policy.db")

	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "Policy", "Resource", "Userid"))
	app.AddMiddleware(policys.HandleHTTP)
	app.AddController(policys.NewPolicysController("sqltie", db))
	policys.NewPolicysController("postgres", db)
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), eudore.HandlerEmpty)
	}
	app.AnyFunc("/* action=hello resource-prefix=/", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.AddHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationJSONUtf8)
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":1,"policy_name":"AdministratorAccess","description":"AdministratorAccess","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":2,"policy_name":"AccessApiUser","description":"AccessApiUser","statement":[{"effect":true,"action":["*"],"resource":["api/*/user/*"]}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":3,"policy_name":"LocalAccess","description":"local api","statement":[{"effect":true,"action":["*"],"resource":["api/*"],"conditions":{"and":{"method":["GET"],"sourceip":["127.0.0.1","192.168.0.0/24"]}}}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":4,"policy_name":"TimeBefore","description":"test time","statement":[{"effect":true,"action":["*"],"resource":["*"],"conditions":{"or":{"method":["GET"],"time":{"after":"2222-01-01"}}}}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":5,"policy_name":"access 1","description":"access action 1","statement":[{"effect":true,"action":["/"],"resource":["*"]}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":6,"policy_name":"access 2","description":"get resource /2","statement":[{"effect":true,"action":["*"],"resource":["/2"],"conditions":{"method":["GET"]}}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(`{"policy_id":7,"policy_name":"access put","statement":[{"effect":true,"action":["*"],"resource":["/2"]}]}`).Do()
	client.NewRequest("POST", "/policys").WithBody(``).Do()
	client.NewRequest("PUT", "/policys/7").WithBody(`{"policy_id":7,"policy_name":"access put","statement":[{"effect":true,"action":["*"],"resource":["/2"]}]}`).Do()
	client.NewRequest("PUT", "/policys/7").Do()
	client.NewRequest("DELETE", "/policys/7").Do()

	client.NewRequest("GET", "/policys/reload/policys").Do()
	client.NewRequest("GET", "/policys").Do()
	client.NewRequest("GET", "/policys/1").Do().Out()

	client.NewRequest("POST", "/policys/1/members").WithBody(`{"user_id":1,"index":1}`).Do()
	client.NewRequest("POST", "/policys/2/members").WithBody(`{"user_id":1,"index":100}`).Do()
	client.NewRequest("POST", "/policys/2/members").Do()
	client.NewRequest("PUT", "/policys/2/members/1").WithBody(`{"user_id":1,"index":100,"expiration":"2022-01-01T15:04:05Z"}`).Do()
	client.NewRequest("PUT", "/policys/2/members/1").Do()
	client.NewRequest("DELETE", "/policys/2/members/2").Do()
	client.NewRequest("GET", "/policys/reload/members").Do()
	client.NewRequest("GET", "/policys/members").Do()
	client.NewRequest("GET", "/policys/1/members").Do().Out()
	client.NewRequest("GET", "/policys/runtime").Do()

	defer func() {
		db.Close()
		client.NewRequest("GET", "/policys/reload/policys").Do()
		client.NewRequest("GET", "/policys").Do()
		client.NewRequest("GET", "/policys/1").Do().Out()
		client.NewRequest("GET", "/policys/reload/members").Do()
		client.NewRequest("GET", "/policys/members").Do()
		client.NewRequest("GET", "/policys/1/members").Do().Out()
	}()

	client.AddHeaderValue(eudore.HeaderAuthorization, policys.NewBearer(1, "", "", time.Now().Add(87600*time.Hour).Unix()))
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
