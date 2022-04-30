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
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.GetFunc("/static/*", root)

	client.NewRequest("GET", "/static/app_test.go").Do()
	client.NewRequest("GET", "/static/none_test.go").Do()

	app.CancelFunc()
	app.Run()
}
