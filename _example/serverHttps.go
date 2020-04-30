package main

/*
ListenTLS方法一般均默认开启了h2，如果需要仅开启https，需要手动listen监听然后使用app.Serve启动服务。
*/

import (
	"crypto/tls"
	"net/http"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug("istls:", ctx.Istls())
	})
	// 使用空证书会自动签发私人证书。
	app.ListenTLS(":8089", "", "")

	client := httptest.NewClient(app)
	client.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.NewRequest("GET", "https://localhost:8089/").Do().CheckStatus(200).Out()

	app.CancelFunc()
	app.Run()
}
