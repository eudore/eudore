package main

/*
实现参考eudore.ConfigParseRead和eudore.ConfigParseConfig内容
*/

import (
	"github.com/eudore/eudore"
	"os"
)

var filepath = "example.json"

func main() {
	content := []byte(`{
	"default": true,
	"help": true,
	"mods.debug": {
		"debug": true
	}
}
`)
	tmpfile, _ := os.Create(filepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)

	app := eudore.NewApp()
	app.Set("config", filepath)
	app.Set("help", true)
	app.Options(app.Parse())
	// app.CancelFunc()
	app.Run()
}
