package eudore_test

import (
	"context"
	"errors"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"net/http"
	"testing"
)

func TestContext2(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/ctx", func(ctx eudore.Context) {
		ctx.WithContext(context.WithValue(ctx.GetContext(), "num", 66666))
		ctx.Debug("context:", ctx.GetContext())
	})
	app.AnyFunc("/handler", func(ctx eudore.Context) {
		h, ok := ctx.(interface {
			GetHandler() (int, eudore.HandlerFuncs)
		})
		if ok {
			ctx.Debug(h.GetHandler())
		}
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/ctx").Do()
	client.NewRequest("GET", "/handler").Do()
	client.NewRequest("GET", "/err").Do()

	app.CancelFunc()
	app.Run()
}

type noReadRequest struct{}

func (noReadRequest) Read([]byte) (int, error) {
	return 0, errors.New("test disable read")
}
func (noReadRequest) Close() error {
	return nil
}

type noWriteResponse struct {
	eudore.ResponseWriter
}

func (noWriteResponse) Write([]byte) (int, error) {
	return 0, errors.New("test disable write")
}

func (noWriteResponse) Push(target string, opts *http.PushOptions) error {
	return errors.New("test error no push")
}

func TestReadWriteError2(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/r", func(ctx eudore.Context) {
		req := ctx.Request()
		req = req.WithContext(ctx.GetContext())
		req.Body = &noReadRequest{}
		ctx.SetRequest(req)

		ctx.Body()

		var data map[string]interface{}
		ctx.Bind(&data)
	})
	app.AnyFunc("/w", func(ctx eudore.Context) {
		ctx.Push("/index", nil)
		ctx.SetResponse(&noWriteResponse{ctx.Response()})

		ctx.Write([]byte("wirte byte"))
		ctx.WriteString("wirte string")
		ctx.Push("/index", nil)
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/w").Do().Out()
	client.NewRequest("PUT", "/r").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do()
	client.NewRequest("PUT", "/r").WithBodyFormValue("name", "eudore").Do()

	app.CancelFunc()
	app.Run()
}
