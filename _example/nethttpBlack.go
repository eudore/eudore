package main

import (
	"net/http"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	srv := &http.Server{
		Handler: middleware.NewNetHTTPBlackFunc(mux, map[string]bool{
			"127.0.0.1/8":    true,
			"192.168.0.0/16": true,
			"10.0.0.0/8":     false,
		}),
	}

	client := httptest.NewClient(srv.Handler)
	client.NewRequest("GET", "/eudore/debug/black/ui").Do()
	client.NewRequest("GET", "/eudore/debug/black/ui").Do()
	client.NewRequest("PUT", "/eudore/debug/black/black/10.127.87.0?mask=24").Do()
	client.NewRequest("PUT", "/eudore/debug/black/white/10.127.87.0?mask=24").Do()
	client.NewRequest("GET", "/eudore/debug/black/data").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24").Do()

	client.NewRequest("GET", "/eudore").Do()
	client.NewRequest("GET", "/eudore").WithHeaderValue(eudore.HeaderXForwardedFor, "192.168.1.4 192.168.1.1").Do()
	client.NewRequest("GET", "/eudore").WithRemoteAddr("127.0.0.1:29398").Do()
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.1:8298").Do()
	client.NewRequest("GET", "/eudore").WithRemoteAddr("10.1.1.1:2334").Do()
}
