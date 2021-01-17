package main

import (
	"net/http"

	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	srv := &http.Server{
		Handler: middleware.NewNetHTTPRateRequestFunc(mux, 1, 3, func(req *http.Request) string {
			// 自定义限流key
			return req.UserAgent()
		}),
	}

	client := httptest.NewClient(srv.Handler)
	client.NewRequest("GET", "/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").Do().CheckStatus(200)
	client.NewRequest("GET", "/").Do().CheckStatus(200)
}
