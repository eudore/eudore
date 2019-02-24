package config


import (
	"os"
	"testing"
	"eudore/config"
)

type (
	Config struct {
		Path	string	`description:"path"`
		Log 	*LogT	`description:"log"`
		Num		[]*NumT
		Mapi	map[int]NumT
		Maps	map[string]NumT
	}
	LogT struct {
		Level		int	`description:"log.level"`
		Depth		int	`description:"log.depth"`
		Format		string	`description:"log.format"`
	}
	NumT struct {
		Name	string
	}
)

func TestGetData(t *testing.T) {
	data := &Config{
		Path:	"/data/web",
		Log:	&LogT{
			Level:	3,
			Depth:	2,
		},
		Num:	[]*NumT{
			&NumT{
				Name: "111",
			},
			&NumT{
				Name: "222",
			},
			&NumT{
				Name: "333",
			},
		},
		Mapi:	map[int]NumT{
			1:	NumT{
				Name: "m111",
			},
		},
		Maps:	map[string]NumT{
			"m1":	NumT{
				Name: "m111",
			},
		},
	}
	t.Log(config.GetData(data, "num.1.name"))
	t.Log(config.GetData(data, "num.-2.name"))
	t.Log(config.GetData(data, "num.#len-2*1.name"))
	t.Log(config.GetData(data, "num.11.name"))
	t.Log(config.GetData(data, "maps.m1.name"))
	t.Log(config.GetData(data, "mapi.0.name"))
	t.Log(config.GetData(data, "mapi.1.name"))
	i, err := config.GetData(data, "log.level")
	if err != nil {
		t.Log(err)	
	}
	config.Json(i)
}


func TestHelp(t *testing.T) {
	var i interface{} = &Config{}
	err := config.Help(i, "  --", os.Stdout)
	if err != nil {
		t.Log(err)	
	}
}