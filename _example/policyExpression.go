package main

import (
	"fmt"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/policy"
)

func main() {
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	policys := policy.NewPolicys()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "Policy", "Resource", "Userid"))
	app.AddMiddleware(policys.HandleHTTP)
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), handlerExprUser)
	}
	app.AnyFunc("/* action=hello", handlerExprUser)

	policys.AddPolicy(&policy.Policy{
		PolicyID:    1,
		Description: "and",
		Statement: []byte(`[
			{"effect":true,"action":["1"],"data":[
				{"kind":"and","data":[
					{"kind":"value","name":"group_id","value":[1]},
					{"kind":"value","name":"user_id","value":[1,"value:param:id"]}
				]}
			]}
		]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    2,
		Description: "or",
		Statement: []byte(`[
			{"effect": true, "action": ["2"], "data": [
				{"kind": "or", "data": [
					{"kind": "and", "data": [
						{"kind": "value", "name": "user_id", "value": [1]},
						{"kind": "value", "name": "user_id", "value": ["value:param:Userid"]}
					]},
					{"kind": "value", "name": "group_id", "value": [1]}
				]}
			]}
		]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    3,
		Description: "value",
		Statement: []byte(`[{"effect":true,"action":["1"],"data":[
			{"kind":"value","name":"user_id","value":[1]},
			{"kind":"range","table":"no_table","name":"group_id","min":1,"max":4},
			{"kind":"range","name":"group_id","min":1,"max":4}
		]}]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    4,
		Description: "value",
		Statement: []byte(`[
			{"effect":true,"action":["3"],"data":[
				{"kind":"value","name":"user_id","value":["value:param:Userid"]},
				{"kind":"value","table": "expr_user2","name":"user_id","value":["value:param:Userid"]},
				{"kind":"value","name":"group_name","value":["value:param:groupname"]},
				{"kind":"value","table": "expr_user2","name":"group_name","value":["value:param:groupname"]}
			]}
		]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    5,
		Description: "sql",
		Statement: []byte(`[
			{"effect":true,"action":["5"],"data":[
				{"kind":"sql","sql":"user_id=?","value":["value:param:Userid"]},
				{"kind":"sql","name":"user_id","sql":"user_id=?","value":["value:param:Userid"]}
			]}
		]`),
	})
	policys.AddPolicy(&policy.Policy{
		PolicyID:    6,
		Description: "no data",
		Statement: []byte(`[
			{"effect":true,"action":["4"]}
		]`),
	})
	policys.AddMember(&policy.Member{UserID: 1, PolicyID: 1})
	policys.AddMember(&policy.Member{UserID: 1, PolicyID: 2})
	policys.AddMember(&policy.Member{UserID: 1, PolicyID: 3})
	policys.AddMember(&policy.Member{UserID: 1, PolicyID: 4})
	policys.AddMember(&policy.Member{UserID: 1, PolicyID: 5})
	policys.AddMember(&policy.Member{UserID: 1, PolicyID: 6})

	client := httptest.NewClient(app)
	client.AddHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationJSONUtf8)
	client.AddHeaderValue(eudore.HeaderAuthorization, policys.NewBearer(1, "", "", time.Now().Add(87600*time.Hour).Unix()))
	client.NewRequest("GET", "/policys").Do()
	client.NewRequest("GET", "/1").Do()
	client.NewRequest("GET", "/2").Do()
	client.NewRequest("GET", "/3").Do()
	client.NewRequest("GET", "/4").Do()
	client.NewRequest("GET", "/5").Do()
	client.NewRequest("GET", "/6").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type exprUser struct {
	UserID   int
	UserName string
	GroupID  int
}

func handlerExprUser(ctx eudore.Context) {
	ctx.Info(policy.CreateExpressions(ctx, "expr_user", []string{"user_id", "user_name", "group_id"}, 0))
	ctx.Info(policy.CreateExpressions(ctx, "expr_user", []string{"user_id", "user_name", "group_id"}, 1))
}
