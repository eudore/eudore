package eudore_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/policy"
)

func TestPolicyUnmarshalJSON(t *testing.T) {
	policybody := []string{
		`{"policy_id":0,"statement":[{"effect":true, "conditions":[]}]}`,
		`{"policy_id":1,"statement":[{"effect":true, "conditions":{"ok":[]}}]}`,
		`{"policy_id":2,"statement":[{"effect":true, "conditions":{"and":[]}}]}`,
		`{"policy_id":3,"statement":[{"effect":true, "conditions":{"method":["GET"]}}]}`,
		`{"policy_id":4,"statement":[{"effect":true,"data":{"ok":["Home","Index"]}}]}`,
		`{"policy_id":5,"statement":[{"effect":true,"data":{"menu":[12,13]}}]}`,
		`{"policy_id":6,"statement":[{"effect":true,"data":{"menu":["Home","Index"]}}]}`,
		`{"policy_id":7,"statement":[{"effect":true,"conditions":{"and":{"method":["GET"]}}}]}`,
		`{"policy_id":8,"statement":[{"effect":true,"conditions":{"or":{"method":["GET"]}}}]}`,
		`{"policy_id":9,"statement":[{"effect":true,"conditions":{"or":""}}]}`,
		`{"policy_id":10,"statement":[{"effect":true,"conditions":{"sourceip":[127]}}]}`,
		`{"policy_id":11,"statement":[{"effect":true,"conditions":{"sourceip":["127.0.0.1"]}}]}`,
		`{"policy_id":12,"statement":[{"effect":true,"conditions":{"sourceip":["127.0.0.1/33"]}}]}`,
		`{"policy_id":13,"statement":[{"effect":true,"conditions":{"date":{"before":123}}}]}`,
		`{"policy_id":14,"statement":[{"effect":true,"conditions":{"date":{"before":"2022"}}}]}`,
		`{"policy_id":15,"statement":[{"effect":true,"conditions":{"date":{"after":"2022"}}}]}`,
		`{"policy_id":16,"statement":[{"effect":true,"conditions":{"date":{"after":"2022-12-31"}}}]}`,
		`{"policy_id":17,"statement":[{"effect":true,"conditions":{"time":{"before":123}}}]}`,
		`{"policy_id":18,"statement":[{"effect":true,"conditions":{"time":{"before":"2022"}}}]}`,
		`{"policy_id":19,"statement":[{"effect":true,"conditions":{"time":{"after":"2022"}}}]}`,
		`{"policy_id":20,"statement":[{"effect":true,"conditions":{"time":{"after":"16:00:00"}}}]}`,
		`{"policy_id":21,"statement":[{"effect":true,"conditions":{"method":["GET"]}}]}`,
		`{"policy_id":22,"statement":[{"effect":true,"conditions":{"params":{"userid":["1","2"]}}}]}`,
	}
	for i := range policybody {
		_, err := policy.NewPolicy(policybody[i])
		if err != nil {
			t.Log(i, err)
		}
	}
}

func TestPolicyPbacParse(t *testing.T) {
	pbac := policy.NewPolicys()
	app := eudore.NewApp()
	app.SetValue("policy", pbac)
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "resource", "Userid"))
	app.AddMiddleware(pbac)
	app.AnyFunc("/static/*", eudore.HandlerEmpty)
	app.AnyFunc("/ action=Home", eudore.HandlerEmpty)

	now := time.Now().Add(time.Hour).Unix()
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, "000"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, "Bearer 000"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOjEwLCJwb2xpY3kiOiJiYXNlNjQiLCJleHBpcmF0aW9uIjoxNjQ5MTQwMzkwfQ.2mqeTZZizrP"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.{"userid":10,"expiration":1649140575}.ffikvNJyZVA8u01PtZ_3fUwQJQ5aGjw_0uCKhoKDr9w`))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiIxMCIsImV4cGlyYXRpb24iOjE2NDkxNDA1NzV9.LgfnJJ-UknB1hOJIA1FrYbpeCNJ2cRuSj_r_bJo8vA8`))

	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, "", time.Now().Add(time.Hour*-1).Unix())))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, "", now)))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, "Bearer "+pbac.Signaturer.Signed(&policy.SignatureUser{UserID: 10, Policy: "base64", Expiration: now})))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, "base64", now)))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, `[{"effect":true,"action":["Home"]}]`, now)))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, `[{"effect":false,"action":["Home"]}]`, now)))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, `[{"effect":true,"action":["Index"]}]`, now)))

	option := func(req *http.Request) {
		req.Header.Set(eudore.HeaderAuthorization, pbac.NewBearer(10, "", now))
	}
	app.NewRequest(nil, "GET", "/", option)
	app.NewRequest(nil, "GET", "/static/1.js", option)

	app.CancelFunc()
	app.Run()
}

func TestPolicyPbacHandler(t *testing.T) {
	pbac := policy.NewPolicys()
	app := eudore.NewApp()
	app.SetValue("policy", pbac)
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "resource", "Userid"))
	app.AddMiddleware(pbac)

	for i := 1; i < 11; i++ {
		app.AnyFunc(fmt.Sprintf("/%d action=Num:%d", i, i), eudore.HandlerEmpty)
		pbac.AddMember(&policy.Member{UserID: 10, PolicyID: i})
	}
	app.AnyFunc("/menu action=Menu", func(ctx eudore.Context) {
		menu := ctx.Value(eudore.NewContextKey("policy-menu"))
		if menu == nil {
			menu = []string{"*"}
		}
		ctx.Render(menu)
	})
	app.AnyFunc("/runtime", pbac.HandleRuntime)
	app.AnyFunc("/has action=Menu", func(ctx eudore.Context) (interface{}, error) {
		type Data struct {
			UserID   int    `json:"user_id"`
			Action   string `json:"action"`
			Resource string `json:"resource"`
		}
		type Statement struct {
			Effect         bool     `json:"effect"`
			Action         []string `json:"action"`
			Resource       []string `json:"resource"`
			MatchAction    bool
			MatchResource  bool
			MatchCondition bool
			MatchData      map[string][]interface{}
		}
		type Response struct {
			PolicyID    int    `json:"policy_id"`
			PolicyName  string `json:"policy_name"`
			MemberIndex int    `json:"member_index"`
			Description string
			Statement   []Statement
		}
		var data Data
		data.UserID = eudore.GetStringInt(ctx.GetParam(eudore.ParamUserid))
		err := ctx.Bind(&data)
		if err != nil {
			return nil, err
		}

		var resps []*Response
		for _, m := range pbac.GetMember(data.UserID) {
			resp := &Response{PolicyID: m.Policy.PolicyID, PolicyName: m.Policy.PolicyName,
				MemberIndex: m.Index, Description: m.Policy.Description}
			for _, stmt := range m.Policy.Statement {
				resp.Statement = append(resp.Statement, Statement{Effect: stmt.Effect, Action: stmt.Action, Resource: stmt.Resource,
					MatchAction: stmt.MatchAction(data.Action), MatchResource: stmt.MatchResource(data.Action),
					MatchCondition: stmt.MatchCondition(ctx), MatchData: stmt.MatchData()})
			}
			resps = append(resps, resp)
		}
		return resps, nil
	})
	pbac.AddMember(&policy.Member{UserID: 10, PolicyID: 11, Expiration: time.Now().Add(time.Hour)})
	pbac.AddMember(&policy.Member{UserID: 10, PolicyID: 11, Expiration: time.Now().Add(time.Hour * -1)})

	pbac.AddPolicyString(`{"policy_id":1,"statement":[{"effect":true,"action":["Num:1"],"conditions":{"and":{"method":["GET"],"sourceip":["127.0.0.1"]}}}]}`)
	pbac.AddPolicyString(`{"policy_id":2,"statement":[{"effect":true,"action":["Num:2"],"conditions":{"or":{"method":["GET"],"sourceip":["127.0.0.1"]}}}]}`)
	pbac.AddPolicyString(`{"policy_id":3,"statement":[{"effect":true,"action":["Num:3"],"conditions":{"sourceip":["127.0.0.1"]}}]}`)
	pbac.AddPolicyString(`{"policy_id":4,"statement":[{"effect":true,"action":["Num:4"],"conditions":{"date":{"before":"2030-12-31"}}}]}`)
	pbac.AddPolicyString(`{"policy_id":5,"statement":[{"effect":true,"action":["Num:5"],"conditions":{"time":{"before":"23:59:59"}}}]}`)
	pbac.AddPolicyString(`{"policy_id":6,"statement":[{"effect":true,"action":["Num:6"],"conditions":{"method":["GET"]}}]}`)
	pbac.AddPolicyString(`{"policy_id":7,"statement":[{"effect":true,"action":["Num:7"],"conditions":{"params":{"action":["Num:7"]}}}]}`)
	pbac.AddPolicyString(`{"policy_id":8,"statement":[{"effect":false,"action":["Num:8"]}]}`)
	pbac.AddPolicyString(`{"policy_id":9,"statement":[{"effect":true,"action":["Menu"],"data":{"menu":["Home"]}}]}`)
	pbac.AddPolicyString(`{"policy_id":10,"statement":[{"effect":true,"action":["Menu"],"data":{"menu":["Index"]}}]}`)
	pbac.AddPolicyString(`{"policy_id":12}`)
	pbac.AddPolicyString(`{"policy_id":13,}`)

	app.Client = app.WithClient(eudore.NewClientHeader(eudore.HeaderAuthorization, pbac.NewBearer(10, "", time.Now().Add(time.Hour).Unix())))
	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "PUT", "/1")
	app.NewRequest(nil, "GET", "/2")
	app.NewRequest(nil, "PUT", "/2")
	app.NewRequest(nil, "GET", "/3", eudore.NewClientHeader(eudore.HeaderXRealIP, "127.0.0.1"))
	app.NewRequest(nil, "GET", "/3", eudore.NewClientHeader(eudore.HeaderXRealIP, "172.17.1.3"))
	app.NewRequest(nil, "GET", "/4")
	app.NewRequest(nil, "GET", "/5")
	app.NewRequest(nil, "PUT", "/6")
	app.NewRequest(nil, "GET", "/6")
	app.NewRequest(nil, "GET", "/7")
	app.NewRequest(nil, "GET", "/8")
	app.NewRequest(nil, "GET", "/menu")
	app.NewRequest(nil, "PUT", "/has", eudore.NewClientBodyJSON(map[string]interface{}{"action": "Menu", "resource": "/has"}))
	app.NewRequest(nil, "GET", "/runtime")

	app.CancelFunc()
	app.Run()
}

type User023Controller struct {
	eudore.ControllerAutoRoute
	policy.ControllerAction
}

func (*User023Controller) Get(eudore.Context)     {}
func (*User023Controller) GetIcon(eudore.Context) {}

func TestPolicyutil(t *testing.T) {
	pbac := policy.NewPolicys()
	pbac.ActionFunc = func(ctx eudore.Context) string { return ctx.GetQuery("action") }
	pbac.AddPolicyString(`{
		"policy_id": 0,
		"statement": [
		  {
			"effect": false,
			"action": [
			  "eudore:user:Get**",
			  "eudore:user:Get2",
			  "eudore:user:Get1",
			  "eudore:user:*",
			  "eudore:group:*",
			  "eudore:group:",
			  "eudore:*:Get",
			  "*:*:Get"
			]
		  }
		]
	  }`)
	pbac.AddMember(&policy.Member{UserID: 0, PolicyID: 0})

	app := eudore.NewApp()
	app.SetValue("policy", pbac)

	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "resource", "Userid"))
	app.AddMiddleware(pbac)
	app.AnyFunc("/*", eudore.HandlerEmpty)
	app.AddController(&User023Controller{})
	app.Info((User023Controller{}).ControllerParam("github.com/eudore/eudore", "User023Controller", "Get"))

	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:user:Get1"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:user:Get2"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:user:Get3"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:group:22"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:group:"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:ns:Get"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientQuery("action", "eudore:ns:Get2"))

	app.CancelFunc()
	app.Run()
}
