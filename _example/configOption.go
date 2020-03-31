package main

/*
 */

import (
	"errors"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.ParseOption(func([]eudore.ConfigParseFunc) []eudore.ConfigParseFunc {
		return []eudore.ConfigParseFunc{parseError}
	})
	app.Parse()
	app.Run()
}

func parseError(eudore.Config) error {
	return errors.New("throws a parse test error")
}
