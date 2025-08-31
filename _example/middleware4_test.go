package eudore_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
)

func TestMiddlewareCORS(t *testing.T) {
	corsPattern := []string{"localhost"}
	corsHeaders := map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
		"access-control-Max-Age":           "1000",
	}
	NewCORSFunc(nil, corsHeaders)

	app := NewApp()
	app.AddMiddleware("global", NewCORSFunc(corsPattern, corsHeaders))
	app.AnyFunc("/*", HandlerEmpty)

	reqs := []struct {
		method string
		origin string
		status int
	}{
		{MethodOptions, DefaultClientInternalHost, 405},
		{MethodGet, DefaultClientInternalHost, 200},
		{MethodOptions, DefaultClientHost, 204},
		{MethodGet, DefaultClientHost, 200},
		{MethodGet, "127.0.0.1", 403},
	}
	for i, r := range reqs {
		err := app.NewRequest(r.method, "/"+strconv.Itoa(i),
			NewClientHeader(HeaderOrigin, "http://"+r.origin),
			NewClientCheckStatus(r.status),
		)
		if err != nil {
			t.Error(err)
		}
	}
	app.GetRequest("/", NewClientHeader(HeaderOrigin, "a"))
	app.GetRequest("/", NewClientHeader(HeaderOrigin, "https://a"))
	app.CancelFunc()
	app.Run()
}

func TestMiddlewareReferer(t *testing.T) {
	app := NewApp()
	app.AddMiddleware(NewRefererCheckFunc(map[string]bool{
		"*":                      false,
		"":                       true,
		"origin":                 true,
		"eudore.cn/*":            true,
		"www.eudore.cn/*":        true,
		"*.eudore.cn/*":          true,
		"*.eudore.cn/*/public/*": true,
	}))
	app.AnyFunc("/*", HandlerEmpty)

	reqs := []struct {
		referer string
		status  int
	}{
		{"http://" + DefaultClientInternalHost + "/", 200},
		{"http://eudore.cn/", 200},
		{"http://www.eudore.cn/", 200},
		{"http://godoc.eudore.cn/", 200},
		{"http://godoc.eudore.cn/pkg/public/app.js", 200},
		{"http://www.example.com/", 403},
		{"", 200},
		{"origin", 403},
	}
	for i, r := range reqs {
		err := app.GetRequest("/"+strconv.Itoa(i),
			NewClientHeader(HeaderReferer, r.referer),
			NewClientCheckStatus(r.status),
		)
		if err != nil {
			t.Error(err)
		}
	}

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRewrite(t *testing.T) {
	rewritedata := map[string]string{
		"/public/*":    "/static/$0",
		"/api/v1/*":    "/api/v3/$0",
		"/api/v2/*":    "/api/v3/$0",
		"/user/*":      "/users/$0",
		"/user/*/icon": "/users/icon/$0",
		"/split1/*":    "$0",
		"/split2/*":    "/$0-$1-$",
		"/u1":          "",
		"/u2":          "",
		"/u3":          "",
		"/u4":          "",
	}

	app := NewApp()
	app.AddMiddleware("global", NewRewriteFunc(rewritedata))
	app.AnyFunc("*", func(ctx Context) {
		ctx.WriteString(ctx.Path())
	})

	reqs := []struct {
		path   string
		target string
	}{
		{"/api/v1/name", "/api/v3/name"},
		{"/api/v2/name", "/api/v3/name"},
		{"/user/eudore", "/users/eudore"},
		{"/user/eudore/icon", "/users/icon/eudore"},
		{"/public/app.js", "/static/app.js"},
		{"/metrics", "/metrics"},
		{"/split1/name", "name"},
		{"/split2/name", "/name-$1-$"},
	}
	for _, r := range reqs {
		err := app.GetRequest(r.path,
			NewClientCheckStatus(200),
			NewClientCheckBody(r.target),
		)
		if err != nil {
			t.Error(err)
		}
	}

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRouter(*testing.T) {
	routerdata := map[string]interface{}{
		"/api/:v/*": func(ctx Context) {
			ctx.Request().URL.Path = "/api/v3/" + ctx.GetParam("*")
		},
		"GET /api/:v/*": func(ctx Context) {
			ctx.WriteHeader(403)
			ctx.End()
		},
	}

	app := NewApp()
	app.AddMiddleware(NewRoutesFunc(routerdata))
	app.AnyFunc("/*", HandlerEmpty)

	app.GetRequest("/api/v1/user")
	app.PutRequest("/api/v1/user")
	app.PutRequest("/api/v2/user")

	app.CancelFunc()
	app.Run()
}

type conditionError1 struct {
	F func()
}

func (cond conditionError1) Match(ctx Context) bool { return true }

func TestMiddlewarePolicyConditions(t *testing.T) {
	data := []string{
		"sourceip", `["127.0.0.1","172.17.0.0/16"]`, ``,
		"sourceip", `[127]`, `policy conditions unmarshal json sourceip error: json: cannot unmarshal number into Go value of type string`,
		"sourceip", `["127.0.0.1/64"]`, `policy conditions sourceip parse cidr error: invalid CIDR address: 127.0.0.1/64`,
		"date", `{"after":"2025-08-31","before":"2025-08-31"}`, "",
		"date", `{"after":"2025-08-31 04:00:00","before":"2025-08-31 04:00:00"}`, "",
		"date", `{"after":true}`, `policy conditions unmarshal json date error: json: cannot unmarshal bool into Go struct field .after of type string`,
		"date", `{"after":"2025-08-31T"}`, `policy conditions date parse after error: parsing time "2025-08-31T" as "2006-01-02 15:04:05": cannot parse "T" as " "`,
		"date", `{"before":"2025-08-31T"}`, `policy conditions date parse before error: parsing time "2025-08-31T" as "2006-01-02 15:04:05": cannot parse "T" as " "`,
		"time", `{"after":"12:00:00"}`, "",
		"time", `{"before":"12:00:00"}`, "",
		"time", `{"after":"08:00:00","before":"12:00:00"}`, "",
		"time", `{"after":"12:00:00","before":"08:00:00"}`, "",
		"time", `{"after":true}`, `policy conditions unmarshal json time error: json: cannot unmarshal bool into Go struct field .after of type string`,
		"time", `{"after":"12:00:00T"}`, `policy conditions time parse after error: parsing time "12:00:00T": extra text: "T"`,
		"time", `{"before":"12:00:00T"}`, `policy conditions time parse before error: parsing time "12:00:00T": extra text: "T"`,
		"rate", `{"speed":1,"max":3}`, ``,
		"rate", `{"speed":{}}`, `json: cannot unmarshal object into Go struct field .speed of type int64`,
		"version", `{"name":"brower","version":[{"name":"Chrome","min":120}]}`, `json: cannot unmarshal number into Go struct field conditionVersionValueString.version.min of type string`,
		"version", `{"name":"brower","version":[{"name":"Chrome","min":"120.0.0.0"}]}`, ``,
		"and", `x`, `invalid character 'x' looking for beginning of value`,
		"and", `{"find":x}`, `invalid character 'x' looking for beginning of value`,
		"and", `{"find":true}`, `invalid character 'true' looking for beginning of value`,
		"and", `{"find":""}`, `policy conditions unmarshal json and error: undefined condition find`,
		"and", `{"method":""}`, `policy conditions parse method error: json: cannot unmarshal string into Go value of type middleware.conditionMethod`,
		"and", `{"method":["GET"],"path":[""]}`, ``,
		"or", `{"method":""}`, `policy conditions parse method error: json: cannot unmarshal string into Go value of type middleware.conditionMethod`,
		"or", `{"method":["GET"],"path":[""]}`, ``,
	}
	for i := 0; i < len(data); i += 3 {
		val := DefaultPolicyConditions[data[i]]()
		err := json.Unmarshal([]byte(data[i+1]), val)
		if err == nil {
			body, _ := json.Marshal(val)
			t.Log(i/3, string(body))
		} else if data[i+2] != err.Error() {
			t.Log(i/3, err)
		}
	}
}

func TestMiddlewarePolicyStatements(t *testing.T) {
	policys := []string{
		`{"user": "test", "data":["conditionMethod", "GetMenu"], "policy": [
			"conditionMethod",
			"conditionPath",
			"conditionParams",
			"conditionDate",
			"conditionTime1",
			"conditionTime2",
			"conditionSourceIP1",
			"conditionSourceIP2",
			"conditionRate",
			"conditionOr1",
			"conditionOr2",
			"conditionVersion1",
			"conditionVersion2",
			"conditionVersion3",
			"conditionTestUser"
		]}`,
		`{"user":"<Guest User>"}`,
		`{"name":"conditionMethod","statement":[{"effect":true,"conditions":{"method":["PUT"]}}]}`,
		`{"name":"conditionPath","statement":[{"effect":true,"conditions":{"path":["/api/v1/users/path"]}}]}`,
		`{"name":"conditionParams","statement":[{"effect":true,"conditions":{"params":{"username":["eudore"]}}}]}`,
		`{"name":"conditionDate","statement":[{"effect":true,"resource":["/users/date"],"conditions":{"date":{"after":"2020-01-01 12:00:00"}}}]}`,
		`{"name":"conditionTime1","statement":[{"effect":true,"resource":["/users/time"],"conditions":{"time":{"after":"12:00:00"}}}]}`,
		`{"name":"conditionTime2","statement":[{"effect":true,"resource":["/users/time"],"conditions":{"time":{"before":"12:00:00"}}}]}`,
		`{"name":"conditionSourceIP1","statement":[{"effect":true,"resource":["/users/ip1"],"conditions":{"sourceip":["127.0.0.0/24"]}}]}`,
		`{"name":"conditionSourceIP2","statement":[{"effect":true,"resource":["/users/ip2"],"conditions":{"sourceip":["10.0.0.0/24"]}}]}`,
		`{"name":"conditionRate","statement":[{"effect":true,"resource":["/users/rate"],"conditions":{"rate":{"max":3}}}]}`,
		`{"name":"conditionOr1","statement":[{"effect":true,"resource":["/users/or1"],"conditions":{"or":{"or":{}}}}]}`,
		`{"name":"conditionOr2","statement":[{"effect":true,"resource":["/users/or2"],"conditions":{"or":{"method":["GET"]}}}]}`,
		`{"name":"conditionVersion1","statement":[{"effect":true,"resource":["/users/version"],"conditions":{"version":{"name":"brower","version":[{"name":"Chrome","max":"120.0.0.0"}]}}}]}`,
		`{"name":"conditionVersion2","statement":[{"effect":true,"resource":["/users/version"],"conditions":{"version":{"name":"brower","version":[{"name":"Chrome","min":"140.0.0.0"}]}}}]}`,
		`{"name":"conditionVersion3","statement":[{"effect":true,"resource":["/users/version"],"conditions":{"version":{"name":"brower","version":[{"name":"Chrome","min":"120.0.0.0"}]}}}]}`,
		`{"name":"conditionTestUser","statement":[{"effect":false,"action":["test:Users:GetByUsername"],"conditions":{"params":{"Userid":["test"]}}}]}`,
		`{"name":"GetMenu","statement":[{"data":{"menu":["Get*"]},"conditions":{"method":["GET"]}}]}`,
	}

	ch := make(chan string)
	app := NewApp()
	app.AddMiddleware(
		func(ctx Context) {
			ctx.SetParam(ParamUserid, ctx.GetQuery("user"))
			ctx.SetParam("brower", ctx.GetQuery("brower"))
		},
		NewResourceFunc("/api/v1"),
		NewResourceFunc("/api/v1", "/api/v1"),
		NewSecurityPolicysFunc(policys, NewOptionSecurityPolicysChan(ch)),
	)
	ch <- ""
	app.AnyFunc("/*", HandlerEmpty)
	app.AnyFunc("/api/v1/users action=test:Users:Get", HandlerEmpty)
	app.AnyFunc("/api/v1/users/:username action=test:Users:GetByUsername", HandlerEmpty)

	app.PutRequest("/api/v1/users/method?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/path?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/eudore?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/date?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/time?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/ip1?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/ip2?user=test", NewClientCheckStatus(403))
	app.GetRequest("/api/v1/users/rate?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/or1?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/or2?user=test", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/version?user=test&brower=Chrome/138.0.0.0", NewClientCheckStatus(200))
	app.GetRequest("/api/v1/users/version?user=test&brower=Safari/18.3", NewClientCheckStatus(403))
	app.GetRequest("/api/v1/users/version?user=test", NewClientCheckStatus(403))
	app.GetRequest("/api/v1/users/undefined?user=test", NewClientCheckStatus(403))
	app.GetRequest("/api/v1/users/undefined?user=guest", NewClientCheckStatus(403))
	app.GetRequest("/api/v1/users/undefined?", NewClientCheckStatus(401))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewarePolicyAPI(t *testing.T) {
	app := NewApp()
	app.AddMiddleware(NewLoggerLevelFunc(func(Context) int { return 4 }))
	app.AddMiddleware(NewSecurityPolicysFunc(nil, NewOptionRouter(app.Group(" loggerkind=~handler"))))

	h := http.Header{HeaderContentType: []string{MimeApplicationJSON}}
	app.Client = app.Client.NewClient(NewClientHeader(HeaderAccept, MimeApplicationJSON))
	app.PutRequest("/policys/new", h, strings.NewReader(`{"statement":[{"effect":true,"action":["GetIndex"]}]}`))
	app.PutRequest("/policys/new", h, strings.NewReader(`{"statement":[{"effect":true,"action":["GetIndex"]}]}`))
	app.PutRequest("/policys/new", h, strings.NewReader(`{"statement":[{"conditions":"x"}]}`), NewClientCheckStatus(500))
	app.GetRequest("/policys/new", NewClientCheckBody(`{"name":"new","statement":[{"effect":true,"action":["GetIndex"]}]}`))
	app.GetRequest("/policys/admin", NewClientCheckStatus(500))
	app.GetRequest("/policys", NewClientCheckBody(`[{"name":"new","statement":[{"effect":true,"action":["GetIndex"]}]}]`))
	app.DeleteRequest("/policys/new")
	app.DeleteRequest("/policys/new")

	app.PostRequest("/members/append", h, strings.NewReader(`x`), NewClientCheckStatus(500))
	app.PostRequest("/members/append", h, strings.NewReader(`{"policy":["new1", "new2", "new3"]}`), NewClientCheckStatus(200))
	app.PostRequest("/members/append", h, strings.NewReader(`{"policy":["new3", "new4"],"data":["data1"]}`), NewClientCheckStatus(200))
	app.PutRequest("/members/new", h, strings.NewReader(`{"policy":["new"]}`), NewClientCheckStatus(200))
	app.PutRequest("/members/new", h, strings.NewReader(`{"policy":["new"]}`), NewClientCheckStatus(200))
	app.PutRequest("/members/new", h, strings.NewReader(`x`), NewClientCheckStatus(500))
	app.GetRequest("/members/new", NewClientCheckBody(`{"user":"new","policy":["new"]}`))
	app.GetRequest("/members/admin", NewClientCheckStatus(500))
	app.GetRequest("/members", NewClientCheckBody(`{"user":"new","policy":["new"]}`))
	app.DeleteRequest("/members/new")
	app.DeleteRequest("/members/new")
	app.PutRequest("/members/"+DefaultPolicyGuestUser, h, strings.NewReader(`{"policy":["new"]}`), NewClientCheckStatus(200))

	defer func() {
		recover()
	}()
	NewSecurityPolicysFunc([]string{"policy"})

	app.CancelFunc()
	app.Run()
}
