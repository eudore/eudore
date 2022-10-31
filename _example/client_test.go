package eudore_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func TestClientRequest(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("server body:", string(ctx.Body()))
		ctx.Write(ctx.Body())
	})
	app.AnyFunc("/ctx", func(ctx eudore.Context) {
		client := ctx.Value(eudore.ContextKeyClient).(eudore.Client).WithClient(
			eudore.NewClientHeader(eudore.HeaderAuthorization, ctx.GetHeader(eudore.HeaderAuthorization)),
		)
		ctx.SetValue(eudore.ContextKeyClient, client)
		ctx.NewRequest("GET", "/")
	})

	client := app.GetClient()
	tp, ok := client.Transport.(*http.Transport)
	if ok {
		tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	app.Debug(app.NewRequest(nil, "GET", "/",
		"eudore body",
		url.Values{"name": []string{"eudore"}},
		http.Header{"Client": []string{"eudore-client"}},
		&http.Cookie{Name: "name", Value: "eudore"},
		func(*http.Request) {},
		func(*http.Response) error { return nil },
		eudore.NewClientQuery("name", "eudore"),
		eudore.NewClientQuerys(url.Values{"state": []string{"active"}}),
		eudore.NewClientHeader("Client", "eudore"),
		eudore.NewClientHeaders(http.Header{"Accept": []string{"application/json"}}),
		eudore.NewClientCookie("id", "c6e2ada8-8715-465b-af25-f992723b5b0a"),
		eudore.NewClientBasicAuth("eudore", "pass"),
		eudore.NewClientTrace(),
		eudore.NewClientDumpBody(),
	))
	app.Debug(app.NewRequest(nil, "GET", "/", []byte("eudore bytes")))
	app.Debug(app.NewRequest(nil, "GET", "/", bytes.NewBufferString("eudore buffer")))
	app.Debug(app.NewRequest(nil, "GET", "/", ioutil.NopCloser(bytes.NewBufferString("eudore buffer"))))
	app.Debug(app.NewRequest(nil, "GET", "\u007f"))
	app.Debug(app.NewRequest(nil, "GET", ""))
	app.Debug(app.NewRequest(nil, "GET", "/",
		func(*http.Response) error { return fmt.Errorf("eudore client test error") },
	))
	app.Debug(app.NewRequest(nil, "GET", "/ctx"))

	app.CancelFunc()
	app.Run()
}

func TestClientRequestBody(t *testing.T) {
	type Body struct {
		Name string
	}
	app := eudore.NewApp()
	app.AnyFunc("/redirect", func(ctx eudore.Context) {
		ctx.Body()
		ctx.Redirect(308, "/")
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("server body:", string(ctx.Body()))
		ctx.Write(ctx.Body())
	})

	app.Debug(app.NewRequest(nil, "GET", "/body/string", eudore.NewClientBodyString("eudore body string")))
	app.Debug(app.NewRequest(nil, "GET", "/body/json", Body{"eudor"}))
	app.Debug(app.NewRequest(nil, "GET", "/body/jsonstruct", eudore.NewClientBodyJSON(struct{ Name string }{"eudore"})))
	app.Debug(app.NewRequest(nil, "GET", "/body/jsonmap", eudore.NewClientBodyJSON(map[string]interface{}{"name": "eudore"})))
	app.Debug(app.NewRequest(nil, "GET", "/body/jsonvalue", eudore.NewClientBodyJSONValue("name", "eudore")))
	app.Debug(app.NewRequest(nil, "GET", "/bdoy/formvalue",
		eudore.NewClientBodyFormValue("name", "eudore"),
		eudore.NewClientBodyFormValues(map[string]string{"server": "eudore"}),
	))
	app.Debug(app.NewRequest(nil, "GET", "/body/formfile",
		eudore.NewClientBodyFormFile("file", "string.txt", "file string"),
		eudore.NewClientBodyFormFile("file", "bytes.txt", []byte("file bytes")),
		eudore.NewClientBodyFormFile("file", "buffer.txt", bytes.NewBufferString("file buffer")),
		eudore.NewClientBodyFormFile("file", "rc.txt", ioutil.NopCloser(bytes.NewBufferString("file rc"))),
		eudore.NewClientBodyFormFile("file", "none.txt", nil),
		eudore.NewClientBodyFormLocalFile("file", "", "appNew.go"),
	))
	app.Debug(app.NewRequest(nil, "PUT", "/body/json", eudore.NewClientHeader(eudore.HeaderContentType, eudore.MimeApplicationJSON), eudore.NewClientBody(Body{"eudore"})))
	app.Debug(app.NewRequest(nil, "PUT", "/body/json", eudore.NewClientHeader(eudore.HeaderContentType, eudore.MimeApplicationJSON), eudore.NewClientBody([]Body{{"eudore"}})))
	app.Debug(app.NewRequest(nil, "PUT", "/body/xml", eudore.NewClientHeader(eudore.HeaderContentType, eudore.MimeApplicationXML), eudore.NewClientBody(Body{"eudore"})))
	app.Debug(app.NewRequest(nil, "PUT", "/body/pb", eudore.NewClientHeader(eudore.HeaderContentType, eudore.MimeApplicationProtobuf), eudore.NewClientBody(&Body{"eudore"})))
	app.Debug(app.NewRequest(nil, "PUT", "/body/pb", eudore.NewClientHeader(eudore.HeaderContentType, "pb"), eudore.NewClientBody(Body{"eudore"})))

	app.Debug(app.NewRequest(nil, "PUT", "/redirect", ioutil.NopCloser(bytes.NewBufferString("buffer rc"))))
	app.Debug(app.NewRequest(nil, "PUT", "/redirect", eudore.NewClientBody(Body{"eudore"})))
	app.Debug(app.NewRequest(nil, "PUT", "/redirect", eudore.NewClientBodyString("eudore body string")))
	app.Debug(app.NewRequest(nil, "PUT", "/redirect", eudore.NewClientBodyJSONValue("name", "eudore")))
	app.Debug(app.NewRequest(nil, "PUT", "/redirect", eudore.NewClientBodyFormValue("name", "eudore")))

	app.CancelFunc()
	app.Run()
}

func TestClientResponse(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("server body:", string(ctx.Body()))
		ctx.Write(ctx.Body())
	})
	app.AnyFunc("/trace", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderXTraceID, "558ac45caefc87c517a7c1cf49918f1aeudore")
	})
	app.GetFunc("/body", func(eudore.Context) interface{} {
		return eudore.LoggerStdConfig{
			Std:   true,
			Path:  "/tmp/client.log",
			Level: eudore.LoggerInfo,
		}
	})
	app.GetFunc("/err", func(eudore.Context) error {
		return fmt.Errorf("test err")
	})

	app.Debug(app.NewRequest(nil, "GET", "https://goproxy.cn",
		eudore.NewClientTimeout(time.Second),
		eudore.NewClientTrace(),
		eudore.NewClientDumpHead(),
	))
	app.Debug(app.NewRequest(nil, "GET", "https://golang.org",
		eudore.NewClientTimeout(time.Second),
		eudore.NewClientTrace(),
		eudore.NewClientDumpHead(),
	))
	app.Debug(app.NewRequest(context.Background(), "GET", "/",
		eudore.NewClientTimeout(time.Second),
		eudore.NewClientDumpHead(),
	))

	app.NewRequest(nil, "GET", "/check/status", eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/check/status", eudore.NewClientCheckStatus(201))

	app.NewRequest(nil, "GET", "/trace", eudore.NewClientCheckBody("201"))
	app.NewRequest(nil, "GET", "/trace", NewClientBodyError(), eudore.NewClientCheckBody("201"))

	app.NewRequest(nil, "GET", "/trace",
		NewClientBodyError(),
		eudore.NewClientDumpBody(),
	)

	var conf eudore.LoggerStdConfig
	err := app.NewRequest(nil, "GET", "/body",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON),
		eudore.NewClientParse(&conf),
	)
	app.Debugf("%v %v", conf, err)
	err = app.NewRequest(nil, "GET", "/body",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationXML),
		eudore.NewClientTrace(),
		eudore.NewClientDumpBody(),
		eudore.NewClientParse(&conf),
	)
	app.Debugf("%v %v", conf, err)
	app.NewRequest(nil, "GET", "/body",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML),
		eudore.NewClientParse(&conf),
	)
	app.NewRequest(nil, "GET", "/body",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationProtobuf),
		eudore.NewClientParse(&conf),
		eudore.NewClientParseErr(),
	)
	app.NewRequest(nil, "GET", "/err",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON),
		eudore.NewClientParseIf(200, &conf),
		eudore.NewClientParseIn(200, 200, &conf),
		eudore.NewClientParseErr(),
	)
	app.NewRequest(nil, "GET", "/err",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML),
		eudore.NewClientParseErr(),
	)

	app.CancelFunc()
	app.Run()
}

func NewClientBodyError() eudore.ClientResponseOption {
	return func(w *http.Response) error {
		w.Body = &responseBody{}
		return nil
	}
}

type responseBody struct{}

func (r *responseBody) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("test error")
}
func (r *responseBody) Close() error {
	return nil
}
