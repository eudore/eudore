package config_test

import (
	"eudore/config"
)

type(
	App struct {
		Config		*config.Config
	}
	Config struct {
		Name	string	`description:"App name"`
		Config 	string	`description:"Config info"`

	}
)


func ExampleNew() {
	g := &Config{}
	app := &App{
		Config:	config.NewConfig(g),
	}
	app.Config.Help()
	config.Json(app)
}