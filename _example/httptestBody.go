package main

import (
	"bytes"
	"io"
	"strings"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("%s", ctx.Body())
	})

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").WithBody(nil).Do()
	client.NewRequest("GET", "/").WithBody(123).Do()
	client.NewRequest("GET", "/").WithBody("string body").Do()
	client.NewRequest("GET", "/").WithBody([]byte("byte body")).Do()
	client.NewRequest("GET", "/").WithBody(strings.NewReader("reader body")).Do()
	client.NewRequest("GET", "/").WithBody(bytes.NewBufferString("reader body")).Do()
	client.NewRequest("GET", "/").WithBody(bytes.NewReader([]byte("reader body"))).Do()
	client.NewRequest("GET", "/").WithBody(htttestReader{}).Do()

	client.NewRequest("GET", "/").WithBodyString("string body").Do()
	client.NewRequest("GET", "/").WithBodyBytes([]byte("byte body")).Do()
	client.NewRequest("GET", "/").WithBodyJSON(httptestRequestJSON{"json name"}).WithBodyJSONValue("key1", "val1").Do()
	client.NewRequest("GET", "/").WithBodyJSONValue("key1", "val1").WithBodyJSONValue("key2", "val2").Do()

	client.NewRequest("GET", "/").WithBodyFormValues(map[string][]string{"name": {"eudore"}}).Do()
	client.NewRequest("GET", "/").WithBodyFormFile("file1", "name", 7666).Do()
	client.NewRequest("GET", "/").WithBodyFormFile("file1", "name", "file body").Do()
	client.NewRequest("GET", "/").WithBodyFormLocalFile("file1", "name", "app.go").Do()
	client.NewRequest("GET", "/").WithBodyFormLocalFile("file1", "name", "appNew.go").Do()

	app.CancelFunc()
	app.Run()
}

type httptestRequestJSON struct {
	Name string `json:"name"`
}

type htttestReader struct{}

func (htttestReader) Read([]byte) (int, error) {
	return 0, io.EOF
}
