package main

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

type Body struct {
	Name string
}

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, app.NewClient(
		eudore.NewClientHookLogger(eudore.LoggerInfo, time.Millisecond*20),
	))
	app.AddMiddleware(
		middleware.NewLoggerFunc(app),
		middleware.NewRequestIDFunc(nil),
	)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		body, _ := ctx.Body()
		ctx.Write(body)
		ctx.Info(ctx.GetHeader(eudore.HeaderContentType), string(body))
	})

	// io.Reader
	app.PutRequest("/body/reader", strings.NewReader("eudore body string"))

	// file
	file, err := os.Open("README.md")
	if err != nil {
		app.Error(err)
	} else {
		app.PutRequest("/body/file", eudore.NewClientBodyFile("", file))
	}

	// json
	app.GetRequest("/body/json", eudore.NewClientBodyJSON(&Body{"eudore"}))
	app.GetRequest("/body/json", eudore.NewClientBodyJSON(map[string]interface{}{"name": "eudore"}))
	bodyJSON := eudore.NewClientBodyJSON(nil)
	bodyJSON.AddValue("name", "eudore")
	app.GetRequest("/body/json", bodyJSON)

	// form
	app.GetRequest("/body/form", eudore.NewClientBodyForm(url.Values{
		"name": []string{"eudore"},
	}))
	bodyForm := eudore.NewClientBodyForm(nil)
	bodyForm.AddValue("name", "eudore")
	bodyForm.AddFile("file", "README.md", "README.md") // open README.md
	bodyForm.AddFile("file", "README.md", []byte("file context"))
	app.GetRequest("/body/form", bodyForm)

	// other
	app.GetRequest("/body/pb", eudore.NewClientBodyProtobuf(&Body{"eudore"}))
	app.GetRequest("/body/xml", eudore.NewClientBodyXML(&Body{"eudore"}))

	app.Error(app.GetRequest("/body/json",
		eudore.NewClientBodyJSON(&Body{"eudore"}),
		eudore.NewClientCheckBody("example"),
	))

	app.Listen(":8088")
	app.Run()
}
