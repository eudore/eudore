package main

import (
	"fmt"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/set", func(ctx eudore.Context) {
		ctx.SetCookie(&eudore.SetCookie{
			Name:     "set1",
			Value:    "val1",
			Path:     "/",
			HttpOnly: true,
		})
		ctx.SetCookieValue("name", "eudore", 600)
	})
	app.AnyFunc("/get", func(ctx eudore.Context) {
		ctx.Infof("cookie name value is: %s", ctx.GetCookie("name"))
		for _, i := range ctx.Cookies() {
			fmt.Fprintf(ctx, "%s: %s\n", i.Name, i.Value)
		}
	})
	app.Listen(":8088")
	app.Run()
}
