package eudore_test

import (
	"errors"
	"os"
	"testing"

	"github.com/eudore/eudore"
)

func TestConfigMapPrint2(t *testing.T) {
	conf := eudore.NewConfigMap(map[string]interface{}{
		"int": "int",
	})
	conf.Set("print", conf.Get("int"))

	conf.Set("print", t.Log)
	conf.Set("print", conf.Get("int"))
}

func TestConfigEudore2(t *testing.T) {
	// 默认 map[string]interface{}
	conf := eudore.NewConfigEudore(nil)
	conf.Set("print", nil)
	conf.Set("print", t.Log)
	conf.Set("print", "nil2")
	t.Log(conf.Set("hh", "aa"))
	t.Logf("%#v", conf.Get(""))

	conf.Set("", struct{ Name, Message string }{"eudore", "msg"})
	t.Logf("%#v", conf.Get(""))

	conf.ParseOption([]eudore.ConfigParseFunc{func(eudore.Config) error {
		return errors.New("throws a parse test error")
	}})
	conf.Parse()
}

func TestConfigNoread2(t *testing.T) {
	conf := eudore.NewConfigEudore(nil)
	conf.Set("print", t.Log)
	conf.Set("config", "notfound-file")
	conf.Parse()
}

func TestConfigReadError2(t *testing.T) {
	filename := "testconfig.json"
	os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	defer os.Remove(filename)

	app := eudore.NewApp()
	app.Set("config", filename)
	app.Set("help", true)
	app.Options(app.Parse())
	app.CancelFunc()
	app.Run()
}
