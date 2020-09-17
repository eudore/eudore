package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
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

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/black/ui").Do()
	client.NewRequest("GET", "/eudore/debug/black/ui").Do()
	client.NewRequest("PUT", "/eudore/debug/black/black/10.127.87.0?mask=24").Do()
	client.NewRequest("PUT", "/eudore/debug/black/white/10.127.87.0?mask=24").Do()
	client.NewRequest("GET", "/eudore/debug/black/data").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24").Do()

	client.NewRequest("GET", "/eudore").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("127.0.0.1:29398").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.1:8298").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.3/28").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.0").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.1").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.77").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.148").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.100.222").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.4").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.5").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.75.6").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.99").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.100").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.101").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.102").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.103").Do().CheckStatus(200)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.104").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").WithRemoteAddr("192.168.1.105").Do().CheckStatus(403)
	client.NewRequest("GET", "/eudore").Do().CheckStatus(403)

	client.NewRequest("DELETE", "/eudore/debug/black/white/0.0.0.0?mask=0").Do()
	client.NewRequest("PUT", "/eudore/debug/black/white/192.168.75.4?mask=30").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/192.168.75.1").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/192.168.75.5").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/192.168.75.7").Do()
	client.NewRequest("PUT", "/eudore/debug/black/white/10.16.0.0?mask=16").Do()
	client.NewRequest("DELETE", "/eudore/debug/black/white/192.168.75.4?mask=30").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
