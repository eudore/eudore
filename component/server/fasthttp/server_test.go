package fasthttp

import (
	"testing"

	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/server/fasthttp"
	"github.com/eudore/eudore/protocol"
)

type Handler struct {
	int
}

func (h *Handler) EudoreHTTP(_ context.Context, w protocol.ResponseWriter, _ protocol.RequestReader) {
	w.Write([]byte("start fasthttp server, this default page."))
}

func TestStart(*testing.T) {
	srv := fasthttp.NewServer(nil)
	ln, err := eudore.ListenWithFD(":8084")
	if err != nil {
		panic(err)
	}
	srv.AddListener(ln)
	srv.AddHandler(&Handler{})
	srv.Start()
}

func TestEudore(*testing.T) {
	app := eudore.NewCore()
	app.Server = fasthttp.NewServer(nil)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.")
	})
	app.Listen(":8084")
	app.Run()
}
