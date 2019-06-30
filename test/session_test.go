package test

import (
	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/logger"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	app := eudore.NewCore()
	c, err := app.RegisterComponent("session-cache", &eudore.SessionCacheConfig{
		Cache: app.Cache,
	})
	s, ok := c.(eudore.Session)
	app.Debug(s, ok)
	app.Debug(c, err)
	app.Debug(app.Session.Version())

	app.AddMiddleware(logger.NewLogger(eudore.GetRandomString).Handle)
	app.Router.Get("/se", eudore.HandlerDataFunc(func(ctx eudore.ContextData) {}))
	app.GetFunc("/set", func(ctx eudore.Context) {
		t.Log("set")
		sess := ctx.GetSession()
		sess.Set("key1", 1)
		ctx.SetSession(sess)
	})
	app.GetFunc("/get", func(ctx eudore.Context) {
		t.Log("get")
		sess := ctx.GetSession()
		t.Log(sess.Get("key1"))
	})

	// test
	req, _ := eudore.NewRequestReaderTest("GET", "/set", nil)
	resp := eudore.NewResponseWriterTest()
	app.EudoreHTTP(context.Background(), resp, req)
	req, _ = eudore.NewRequestReaderTest("GET", "/get", nil)
	resp = eudore.NewResponseWriterTest()
	app.EudoreHTTP(context.Background(), resp, req)

	// wait logger flush
	time.Sleep(time.Second)
}
