package main

/*
实现参考eudore.ConfigParseRead和eudore.ConfigParseConfig内容
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"os"
	"time"
)

var filepath = "example.json"

func main() {
	content := []byte(`{
	"keys.default": true,
	"keys.help": true,
	"mods.debug": {
		"keys.debug": true
	}
}
`)
	tmpfile, _ := os.Create(filepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)

	time.Sleep(100 * time.Millisecond)
	app := eudore.NewCore()
	err := app.Parse()
	if err != nil {
		panic(err)
	}
	httptest.NewClient(app).Stop(0)
	app.Set("keys.config", filepath)
	app.Set("keys.help", true)
	app.Listen(":8088")
	app.Run()
}
