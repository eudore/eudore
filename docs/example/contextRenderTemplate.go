package main

import (
	"github.com/eudore/eudore"
	"html/template"
)

func main() {
	app := eudore.NewCore()
	app.Renderer = eudore.NewHTMLWithRender(app.Renderer, nil)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.SetParam("template", "view/index.html")
		ctx.Render(map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})
	app.AnyFunc("/2/*path", func(ctx eudore.Context) interface{} {
		ctx.SetParam("template", "view/index.html")
		return map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		}
	})
	app.AnyFunc("/template/*", func(ctx eudore.Context) error {
		t, err := template.ParseFiles("view/index.html")
		if err != nil {
			return err
		}
		return t.Execute(ctx, map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})
	app.Listen(":8088")
	app.Run()
}
