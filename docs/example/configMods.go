package main

import (
	"os"
	"time"
	"github.com/eudore/eudore"


/*	"fmt"
	"github.com/kr/pretty"*/
)

func main() {
	modstd()
	modeudore()
}


func modstd() {
	content := []byte(`{
	"keys.default": true,
	"keys.help": true,
	"mods.debug": {
		"keys.debug": true
	}
}
`)
	tmpfile, _ := os.Create("example.json")
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)


	app := eudore.NewCore()
	app.Config.Set("keys.config", "example.json")
	app.Config.Set("enable", []string{"debug"})

	app.Info(app.Parse())
	// fmt.Printf("struct: %# v\n", pretty.Formatter(app.Config))
	time.Sleep(100 * time.Millisecond)
}


type conf struct {
	Keys	map[string]interface{} `set:"keys"`
	Enable	[]string	`set:"enable"`
	Mods	map[string]*conf	`set:"mods"`
}

func modeudore() {
	content := []byte(`{
	"keys": {
		"default": true,
		"help": true
	},
	"mods": {
		"debug": {
			"keys": {
				"debug": true
			}
		}
	}
}
`)
	tmpfile, _ := os.Create("example.json")
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)


	app := eudore.NewCore()
	app.RegisterComponent("config-eudore", new(conf))
	app.Config.Set("keys.config", "example.json")
	app.Config.Set("enable", []string{"debug"})

	app.Info(app.Parse())
	// fmt.Printf("struct: %# v\n", pretty.Formatter(app.Config))
	time.Sleep(100 * time.Millisecond)
}