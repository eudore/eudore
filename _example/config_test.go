package eudore_test

import (
	"errors"
	"github.com/eudore/eudore"
	"testing"
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
	conf.Set("print", t.Log)
	t.Log(conf.Set("hh", "aa"))
	t.Logf("%#v", conf.Get(""))

	conf.Set("", struct{ Name, Message string }{"eudore", "msg"})
	t.Logf("%#v", conf.Get(""))

	conf.ParseOption(func([]eudore.ConfigParseFunc) []eudore.ConfigParseFunc {
		return []eudore.ConfigParseFunc{func(eudore.Config) error {
			return errors.New("throws a parse test error")
		}}
	})
	conf.Parse()
}
