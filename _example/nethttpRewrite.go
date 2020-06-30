package main

import (
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"net/http"
)

func main() {
	rewritedata := map[string]string{
		"/js/*":                    "/public/js/$0",
		"/api/v1/users/*/orders/*": "/api/v3/user/$0/order/$1",
		"/d/*":           "/d/$0-$0",
		"/api/v1/*":      "/api/v3/$0",
		"/api/v2/*":      "/api/v3/$0",
		"/help/history*": "/api/v3/history",
		"/help/history":  "/api/v3/history",
		"/help/*":        "$0",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	srv := &http.Server{
		Handler: middleware.NewNetHTTPRewriteFunc(mux, rewritedata),
	}

	client := httptest.NewClient(srv.Handler)
	client.NewRequest("GET", "/").Do()
	client.NewRequest("GET", "/js/").Do()
	client.NewRequest("GET", "/js/index.js").Do()
	client.NewRequest("GET", "/api/v1/user").Do()
	client.NewRequest("GET", "/api/v1/user/new").Do()
	client.NewRequest("GET", "/api/v1/users/v3/orders/8920").Do()
	client.NewRequest("GET", "/api/v1/users/orders").Do()
	client.NewRequest("GET", "/api/v2").Do()
	client.NewRequest("GET", "/api/v2/user").Do()
	client.NewRequest("GET", "/d/3").Do()
	client.NewRequest("GET", "/help/history").Do()
	client.NewRequest("GET", "/help/historyv2").Do()
}
