package eudore_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	. "github.com/eudore/eudore"
)

func TestRouterStdAny(t *testing.T) {
	// 扩展RouterStd允许的方法
	DefaultRouterAllMethod = append(DefaultRouterAllMethod, "LOCK", "UNLOCK")
	DefaultRouterAnyMethod = append(DefaultRouterAnyMethod, "LOCK", "UNLOCK")
	defer func() {
		DefaultRouterAllMethod = DefaultRouterAllMethod[:len(DefaultRouterAllMethod)-2]
		DefaultRouterAnyMethod = DefaultRouterAnyMethod[:len(DefaultRouterAnyMethod)-2]
	}()

	r, c := newCSR(nil)
	// Any方法覆盖
	r.GetFunc("/get/:val")
	r.GetFunc("/get/:val", func(ctx Context) {
		ctx.WriteString("method is get1")
	})
	r.AnyFunc("/get/:val", func(ctx Context) {
		ctx.WriteString("method is any")
	})
	r.GetFunc("/get/:val", func(ctx Context) {
		ctx.WriteString("method is get2")
	})
	r.PostFunc("/get/:val", func(ctx Context) {
		ctx.WriteString("method is post")
	})
	r.AddHandler("LOCK", "/get/:val", func(ctx Context) {
		ctx.WriteString("method is lock")
	})
	r.GetFunc("/index", HandlerEmpty)
	r.AddHandler("404,444", "", HandlerRouter404)
	r.AddHandler("405", "", HandlerRouter405)

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/get/1", "method is get"},
		{"POST", "/get/2", "method is post"},
		{"PUT", "/get/3", "method is any"},
		{"LOCK", "/get/4", "method is lock"},
		{"COPY", "/get/5", "405"},
		{"GET", "/get", "404"},
		{"POST", "/get", "404"},
		{"PUT", "/get", "404"},
		{"POST", "/index", "405"},
		{"GET", "/index", ""},
	}
	for _, route := range routes {
		err := c.NewRequest(route.method, route.path,
			NewClientCheckBody(route.body),
		)
		if err != nil {
			t.Error(route.method, route.path, err)
		}
	}
}

func TestRouterStdCheck(t *testing.T) {
	r, c := newCSR(NewRouterCoreMux())
	r.AnyFunc("/1/:num|num version=1", HandlerEmpty)
	r.AnyFunc("/1/222", HandlerEmpty)
	r.AnyFunc("/2/:num|num", HandlerEmpty)
	r.AnyFunc("/2/:num|", HandlerEmpty)
	r.AnyFunc("/2/:", HandlerEmpty)
	r.AnyFunc("/3/:num|num/22", HandlerEmpty)
	r.AnyFunc("/3/:num|num/*", HandlerEmpty)
	r.AnyFunc("/4/*num|num", HandlerEmpty)
	r.AnyFunc("/4/*num|num", HandlerEmpty)
	r.AnyFunc("/4/*", HandlerEmpty)
	r.AnyFunc("/5/*num|num", HandlerEmpty)
	r.AnyFunc("/api/v1/2", HandlerEmpty)
	r.AnyFunc("/api/v1/1", HandlerEmpty)
	r.AnyFunc("/*num|^\\d+$", HandlerEmpty)
	r.AnyFunc("/api/v1/*|{^0/api\\S+$}", HandlerEmpty)
	r.AnyFunc("/api/v1/*|{\\s+{}}", HandlerEmpty)
	r.AnyFunc("{/api/v1/*\\}}", HandlerEmpty)
	r.AnyFunc("/api/v1/{{*}}", HandlerEmpty)
	r.AnyFunc("/api/*", HandlerEmpty)
	r.AnyFunc("/api/*", HandlerEmpty)
	r.AddHandler(MethodOptions, "/", HandlerEmpty)
	r.AddHandler(MethodConnect, "/", HandlerEmpty)
	r.AddHandler(MethodTrace, "/", HandlerEmpty)

	// 请求测试
	c.NewRequest("GET", "/1/1")
	c.NewRequest("POST", "/1/222")
	c.NewRequest("PUT", "/2/3")
	c.NewRequest("PUT", "/3/11/3")
	c.NewRequest("PUT", "/3/11/22")
	c.NewRequest("PUT", "/4/22")
	c.NewRequest("PUT", "/5/22")
	c.NewRequest("PUT", "/:{num}")
}

func newCSR(core RouterCore) (Router, Client) {
	s := NewServer(nil)
	c := NewClient()
	r := NewRouter(core)
	get := NewContextBaseFunc(context.Background())
	s.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := get()
		ctx.Reset(w, req)
		ctx.SetHandlers(-1, r.Match(ctx.Method(), ctx.Path(), ctx.Params()))
		ctx.Next()
	}))

	ctx := context.WithValue(context.Background(),
		ContextKeyServer, s,
	)
	ctx = context.WithValue(ctx,
		ContextKeyFuncCreator, DefaultFuncCreator,
	)

	r.(interface{ Mount(context.Context) }).Mount(ctx)
	c.(interface{ Mount(context.Context) }).Mount(ctx)
	return r, c
}

func TestRouterStdSplit(t *testing.T) {
	paths := []struct {
		path string
		sub  string
	}{
		{"/ a=1", "/"},
		{"/{} a=1", "/{}"},
		{"/{{} a=1", "/{{}"},
		{"/{}} a=1", "/{}}"},
		{"/{ } a=1", "/{ }"},
		{"/{ } \\d", "/{ }"},
	}
	for i := range paths {
		if getRoutePathT(paths[i].path) != paths[i].sub {
			t.Error(i, data[i], getRoutePathT(paths[i].path))
		}
	}

	datas := []struct {
		path string
		strs []string
	}{
		{"/", []string{"/"}},
		{"/api/note/", []string{"/api/note/"}},
		{"//api/*", []string{"//api/", "*"}},
		{"//api/*name", []string{"//api/", "*name"}},
		{"/api/get/", []string{"/api/get/"}},
		{"/api/get", []string{"/api/get"}},
		{"/api/:get", []string{"/api/", ":get"}},
		{"/api/:get/*", []string{"/api/", ":get", "/", "*"}},
		{"{/api/**{}:get/*}", []string{"/api/**{", ":get", "/", "*}"}},
		{"/api/:name/info/*", []string{"/api/", ":name", "/info/", "*"}},
		{"/api/:name|^\\d+$/info", []string{"/api/", ":name|^\\d+$", "/info"}},
		{"/api/*|{^0/api\\S+$}", []string{"/api/", "*|^0/api\\S+$"}},
		{"/api/*|^\\$\\d+$", []string{"/api/", "*|^\\$\\d+$"}},
	}
	for i := range datas {
		if fmt.Sprint(getSplitPathT(datas[i].path)) != fmt.Sprint(datas[i].strs) {
			t.Error(i, data[i])
		}
	}
}

func getRoutePathT(path string) string {
	var isblock bool
	var last rune
	for i, b := range path {
		if isblock {
			if b == '}' && last != '\\' {
				isblock = false
			}
			last = b
			continue
		}

		switch b {
		case '{':
			isblock = true
		case ' ':
			return path[:i]
		}
	}
	return path
}

func getSplitPathT(path string) []string {
	var strs []string
	bytes := make([]byte, 0, 64)
	var isblock, isconst bool
	for _, b := range path {
		// block pattern
		if isblock {
			if b == '}' {
				if len(bytes) != 0 && bytes[len(bytes)-1] != '\\' {
					isblock = false
					continue
				}
				// escaping }
				bytes = bytes[:len(bytes)-1]
			}
			bytes = append(bytes, string(b)...)
			continue
		}
		switch b {
		case '/':
			// constant mode, creates a new string in non-constant mode
			if !isconst {
				isconst = true
				strs = append(strs, string(bytes))
				bytes = bytes[:0]
			}
		case ':', '*':
			// variable pattern or wildcard pattern
			isconst = false
			strs = append(strs, string(bytes))
			bytes = bytes[:0]
		case '{':
			isblock = true
			continue
		}
		bytes = append(bytes, string(b)...)
	}
	strs = append(strs, string(bytes))
	if strs[0] == "" {
		strs = strs[1:]
	}
	return strs
}
