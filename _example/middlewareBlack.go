package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	middleware.NewBlackFunc(map[string]bool{
		"192.168.0.0/16": true,
		"0.0.0.0/0":      false,
	}, nil)

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewBlackFunc(map[string]bool{
		"192.168.100.0/24": true,
		"192.168.75.0/30":  true,
		"192.168.1.100/30": true,
		"127.0.0.1/32":     true,
		"10.168.0.0/16":    true,
		"0.0.0.0/0":        false,
	}, app.Group("/eudore/debug")))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "GET", "/eudore/debug/black/data")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")

	client.NewRequest("GET", "/eudore").CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("127.0.0.1:29398").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.1:8298").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.3/28").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.0").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.1").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.77").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.148").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.222").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.4").CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.5").CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.6").CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.99").CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.100").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.101").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.102").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.103").CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.104").CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.105").CheckStatus(403)
	client.NewRequest("GET", "/eudore").CheckStatus(403)

	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/0.0.0.0?mask=0")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/192.168.75.4?mask=30")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.1")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.5")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.7")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/10.16.0.0?mask=16")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.4?mask=30")

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
