package main

/*
实现参考eudore.ConfigParseRead和eudore.ConfigParseConfig内容
*/

import (
	"github.com/eudore/eudore"
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
	app.Set("keys.config", filepath)
	app.Set("keys.help", true)
	app.Run()
}
