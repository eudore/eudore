package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	// 修改第一个解析函数为readHttp解析http请求
	app.ParseOption(append([]eudore.ConfigParseFunc{readHttp}, eudore.ConfigAllParseFunc...))
	app.Set("config-remote", []string{"http://127.0.0.1:8089/xxx", "http://127.0.0.1:8088/xxx"})
	app.Set("help", true)

	go func(app2 *eudore.App) {
		app := eudore.NewApp()
		app.AnyFunc("/*", func(ctx eudore.Context) {
			ctx.WriteJSON(map[string]interface{}{
				"route": "/*",
				"name":  "eudore",
			})
		})
		app.Listen(":8088")

		app2.Options(app2.Parse())
		app2.CancelFunc()
		app.CancelFunc()
		app.Run()
	}(app)

	app.Run()
}

// 自定义一个解析http请求的配置解析函数
func readHttp(c eudore.Config) error {
	for _, path := range eudore.GetStrings(c.Get("config-remote")) {
		if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
			continue
		}
		resp, err := http.Get(path)
		if err == nil {
			err = json.NewDecoder(resp.Body).Decode(c)
			resp.Body.Close()
		}
		if err == nil {
			c.Set("print", "read http succes json config by "+path)
			return nil
		}
		c.Set("print", "read http fatal "+path+" error: "+err.Error())
	}
	return nil
}
