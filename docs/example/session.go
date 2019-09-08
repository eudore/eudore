package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/session"
	// _ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
)

func main() {
	// 创建session，并注册转换函数。
	provider := session.NewSessionMap()
	// provider := session.NewSessionSql("mysql","root:@/jass?parseTime=true")
	// provider := session.NewSessionSql("postgres", "host=localhost port=5432 user=jass password=abcd dbname=jass sslmode=disable")
	eudore.RegisterHandlerFunc(func(fn func(session.ContextSession)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(session.ContextSession{
				Context: ctx,
				Session: provider,
			})
		}
	})

	app := eudore.NewCore()
	app.GetFunc("/set", func(ctx session.ContextSession) {
		// 读取会话
		data := ctx.GetSession()
		data["key1"] = 22
		ctx.Debugf("session set key1 value: %v", data["key1"])

		// 保存会话数据
		ctx.SetSession(data)
	})
	app.GetFunc("/get", func(ctx session.ContextSession) {
		data := ctx.GetSession()
		ctx.Debugf("session get key1 value: %v", data["key1"])
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do()
	client.NewRequest("GET", "/set").Do()
	// 如果第二次get还是nil，是httptest库还未正确处理cookie，请使用阅览器测试。
	client.NewRequest("GET", "/get").Do()

	app.Listen(":8085")
	app.Run()
}
