//go:build go1.16
// +build go1.16

package eudore_test

import (
	"embed"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

//go:embed *.go
var root embed.FS

func TestUtilPatch16(t *testing.T) {
	eudore.NewHandlerEmbedFunc(root, ".")
	eudore.DefaultEmbedTime = time.Now()

	app := eudore.NewApp()
	app.GetFunc("/static/*", root)

	app.NewRequest(nil, "GET", "/static/app_test.go")
	app.NewRequest(nil, "GET", "/static/none_test.go")

	app.CancelFunc()
	app.Run()
}
