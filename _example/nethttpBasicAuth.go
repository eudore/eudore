package main

import (
	"net/http"

	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	data := map[string]string{"user": "pw"}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	middleware.NewNetHTTPBasicAuthFunc(mux, "", data)
	srv := &http.Server{
		Handler: middleware.NewNetHTTPBasicAuthFunc(mux, "Eudore", data),
	}

	client := httptest.NewClient(srv.Handler)
	client.NewRequest("GET", "/1").Do()
	client.NewRequest("GET", "/2").WithHeaderValue("Authorization", "Basic dXNlcjpwdw==").Do()
}
