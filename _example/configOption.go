package main

/*

type ConfigParseFunc func(Config) error
type ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc

type Config interface {
	....
	ParseOption(ConfigParseOption)
	Parse() error
}

Config对象通过ParseOption来追加或设置ConfigParseFunc。
*/

import (
	"errors"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	// 设置配置解析函数为一个自定义函数返回错误。
	app.ParseOption(func([]eudore.ConfigParseFunc) []eudore.ConfigParseFunc {
		return []eudore.ConfigParseFunc{parseError}
	})
	app.Options(app.Parse())
	app.Run()
}

func parseError(eudore.Config) error {
	return errors.New("throws a parse test error")
}
