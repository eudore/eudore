package eudore_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/policy"
)

func TestPolicyRegister2(*testing.T) {
	policy.RegisterExpression("time", nil)
	policy.RegisterValue("header", nil)
}

func TestPolicyNoAction2(*testing.T) {
	app := eudore.NewApp()
	policys := policy.NewPolicys()
	app.AddMiddleware(policys.HandleHTTP)
	app.GetFunc("/policys", eudore.HandlerEmpty)

	client := httptest.NewClient(app).AddHeaderValue("Accept", eudore.MimeApplicationJSONUtf8)
	client.NewRequest("GET", "/policys").Do().Out()

	app.CancelFunc()
	app.Run()
}

func TestPolicyNoAuth2(*testing.T) {
	app := eudore.NewApp()
	policys := policy.NewPolicys()
	app.AddMiddleware(policys.HandleHTTP)
	app.GetFunc("/policys action=GetPolicys", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.AddHeaderValue("Accept", eudore.MimeApplicationJSONUtf8)
	client.NewRequest("GET", "/policys").Do().Out()

	app.CancelFunc()
	app.Run()
}

func TestPolicyBearer2(*testing.T) {
	app := eudore.NewApp()
	policys := policy.NewPolicys()
	app.AddMiddleware(policys.HandleHTTP)
	app.GetFunc("/policys action=GetPolicys", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.AddHeaderValue("Accept", eudore.MimeApplicationJSONUtf8)
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, "Basic yYHGHS==").Do().Out()
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, "Bearer xxxx.xxxxx").Do().Out()
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, "Bearer xxxx.xxxxx.xxxxxxx").Do().Out()
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.{"userid": 1}.IcX-L8Fbict3ITS6SW4EqtEm0wuBeKEesTrAbCunc6g`).Do().Out()
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiAxeHh9.l6n5pZ1_WXYdlSDORJJBRWxBk8nj0BsH2EPqAX7IGZ0").Do().Out()

	sign1 := policys.NewBearer(10, "", time.Now().Add(-100*time.Hour).Unix())
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, sign1).Do().Out()

	sign2 := policys.NewBearer(10, "x", time.Now().Add(100*time.Hour).Unix())
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, sign2).Do().Out()

	sign3 := policys.NewBearer(10, "[]", time.Now().Add(100*time.Hour).Unix())
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, sign3).Do().Out()

	sign4 := policys.NewBearer(10, `[{"effect":true,"action":["*"], "resource":["*"]}]`, time.Now().Add(100*time.Hour).Unix())
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, sign4).Do().Out()

	sign5 := policys.NewBearer(10, `[{"effect":true,"action":["GET*"], "resource":["*"]},{"effect":false,"action":["*"], "resource":["*"]}]`, time.Now().Add(100*time.Hour).Unix())
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, sign5).Do().Out()

	sign6 := "Bearer " + policys.Signaturer.Signed(map[string]interface{}{
		"userid":     10,
		"username":   "eudore",
		"policy":     "[]",
		"expiration": time.Now().Add(100 * time.Hour).Unix(),
	})
	client.NewRequest("GET", "/policys").WithHeaderValue(eudore.HeaderAuthorization, sign6).Do().Out()

	app.CancelFunc()
	app.Run()
}

func TestPolicyAddMember2(*testing.T) {
	app := eudore.NewApp()
	policys := policy.NewPolicys()
	policys.AddMember(&policy.Member{UserID: 10, PolicyID: 1})
	policys.AddPolicy(&policy.Policy{PolicyID: 1, Statement: []byte(`[{"effect":true,"action":["1"],"resource":["*"]}]`)})
	policys.AddPolicy(&policy.Policy{PolicyID: 1, Statement: []byte(`xxx`)})
	policys.AddPolicy(&policy.Policy{PolicyID: 2})
	policys.AddMember(&policy.Member{UserID: 10, PolicyID: 2})
	policys.AddMember(&policy.Member{UserID: 10, PolicyID: 2, Index: 1})

	app.CancelFunc()
	app.Run()
}

func TestPolicyAddRole2(*testing.T) {
	app := eudore.NewApp()
	policys := policy.NewPolicys()
	policys.AddPermission(100, "GetPolicy")
	policys.AddPermission(100, "GetPolicys")
	policys.AddPermission(100, "PuttPolicy")
	policys.AddRole(10, 100)
	policys.DeletePermission(100, "PuttPolicy")
	policys.DeletePermission(101, "PuttPolicy")
	policys.DeleteRole(10, 100)

	app.CancelFunc()
	app.Run()
}

func TestPolicyPrint2(t *testing.T) {
	p := &policy.Policy{
		PolicyID: 1,
		Statement: []byte(`[
			{"conditions":{"or":{"time":{"after":"2019-9-27"}},"hh":[]}},
			{"effect":true,"conditions":{"or":{"time":{"after":"2019-9-27"},"method":["GET"],"sourceip":["127.0.0.1"]}}},
			{"conditions":{"and":{"params":{"route":"/*"},"method":["GET"],"or":[]}}}
		]`),
	}
	t.Log(p.StatementUnmarshal())
	body, err := json.Marshal(p)
	t.Log(string(body), err)
	t.Log(p.StatementMarshal())
	t.Log(string(p.Statement))
}

func TestPolicyUnmarshal2(t *testing.T) {
	p := &policy.Policy{
		PolicyID:  1,
		Statement: []byte(`[{"resource":"*"}]`),
	}
	t.Log(p.StatementUnmarshal())

	var raw *policy.RawMessage
	t.Log(raw.UnmarshalJSON(nil))
}

func TestPolicyTree2(t *testing.T) {
	p := &policy.Policy{
		PolicyID:   1,
		PolicyName: "PolicyTree",
		Statement: []byte(`[
			{"effect":true, "action": ["test:Trace:Get*","test:Trace:GetSpan","test:*:Get","match1:Get*Name:2","eudore:*:Get*","eudore:**Group:Get"], "resource": ["*"]},
			{"action": ["data:*", "data:*Group:List"], "data": []}
		]`),
	}
	t.Log(p.StatementUnmarshal())

	ctx := eudore.NewContextMock(nil, "GET", "/")
	t.Log(p.Match(ctx, "test:Trace:Get", "/"))
	t.Log(p.Match(ctx, "test:Trace:GetSpan", "/"))
	t.Log(p.Match(ctx, "test:Trace:GetSpanName", "/"))
	t.Log(p.Match(ctx, "test:Span:Get", "/"))
	t.Log(p.Match(ctx, "test:SpanGet", "/"))
	t.Log(p.Match(ctx, "match1:GetName:2", "/"))
	t.Log(p.Match(ctx, "eudore:Policy:Get", "/"))
	t.Log(p.Match(ctx, "eudore:Policy:GetMemebr", "/"))
	t.Log(p.Match(ctx, "test:Trace:PutSpan", "/"))
	t.Log(p.Match(ctx, "test:Span:GetName", "/"))
	t.Log(p.Match(ctx, "1", "/"))

	t.Log(p.Match(ctx, "data:UserGet", "/"))
	t.Log(p.Match(ctx, "", "/"))
}

func TestPolicyData2(t *testing.T) {
	p := &policy.Policy{
		PolicyID:  1,
		Statement: []byte(`[{"effect":true,"data":[{"kind":"value","name":"id","value":["value:param:Userid"]},{"kind":"value","name":"id","value":[3,4]}]}]`),
	}
	t.Log(p.StatementUnmarshal())

	ctx := eudore.NewContextMock(nil, "GET", "/")
	t.Log(p.Match(ctx, "test:Trace:Get", "/"))
}

func TestPolicyDataParse2(t *testing.T) {
	// new
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":{}}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"name": 112}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"name": "id", "kind": "new-kind"}]}]`),
	}).StatementUnmarshal())

	// value
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"value","name":"user_id"}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"value","name":"user_id","value":1}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"value","not":true,"name":"user_id","value":[1,2,3]}]}]`),
	}).StatementUnmarshal())

	// range
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"range","name":"user_id","not":1}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"range","name":"user_id"}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"range","name":"user_id","min":1}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"range","name":"user_id","max":1}]}]`),
	}).StatementUnmarshal())
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"range","name":"user_id","not":true,"max":1}]}]`),
	}).StatementUnmarshal())

	// sql
	t.Log((&policy.Policy{
		Statement: []byte(`[{"effect":true,"data":[{"kind":"sql","sql":["user_id"]}]}]`),
	}).StatementUnmarshal())
}
