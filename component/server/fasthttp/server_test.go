package fasthttp

import (
	"testing"
	
	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/component/server/fasthttp"
)

func TestStart(t *testing.T) {
	srv, _ := fasthttp.NewServer()
	eudore.Set(srv, "config.http.+.addr", ":8084")
	srv.Set("config.handler", protocol.HandlerFunc(func(ctx context.Context, w protocol.ResponseWriter, r protocol.RequestReader) {
		w.Write([]byte("start eudore server, this default page."))
	}))
	t.Log(srv.Start())
}
