package fasthttp

import (
	"testing"

	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/server/fasthttp"
	"github.com/eudore/eudore/protocol"
)

func TestStart(t *testing.T) {
	srv, _ := fasthttp.NewServer(nil)
	eudore.Set(srv, "config.http.+.addr", ":8084")
	srv.Set("config.handler", protocol.HandlerFunc(func(_ context.Context, w protocol.ResponseWriter, _ protocol.RequestReader) {
		w.Write([]byte("start fasthttp server, this default page."))
	}))
	t.Log(srv.Start())
}

func TestEudore(t *testing.T) {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.")
	})
	srv, _ := fasthttp.NewServer(nil)
	eudore.Set(srv, "config.http.+.addr", ":8084")
	// eudore.Set(srv, "config.http.+.addr", ":8085")
	srv.Set("config.handler", app)
	t.Log(srv.Start())
}
