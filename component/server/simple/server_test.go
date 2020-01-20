package simple

import (
	"github.com/eudore/eudore/component/server/simple"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	ln, err := net.Listen("tcp", ":8085")
	if err != nil {
		t.Log(err)
		return
	}
	server := &simple.Server{
		Handler: func(w *simple.Response, r *simple.Request) {
			w.Header().Add("Server", "simple server")
			w.Write([]byte("hello http server. your remote addr is " + r.RemoteAddr()))
		},
	}
	t.Log(server.Serve(ln))
}
