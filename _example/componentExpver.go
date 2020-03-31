package main

/*
访问路径 /eudore/debug/vars
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/expvar"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.GetFunc("/eudore/debug/vars", expvar.Expvar)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/vars").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().OutBody()
	for client.Next() {
		app.Error(client.Error())
	}
	app.Run()
}
