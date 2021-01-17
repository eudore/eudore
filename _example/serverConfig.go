package main

/*
 */

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
