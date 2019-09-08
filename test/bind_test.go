package test

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"testing"
)

type (
	Request1 struct {
		Field1 int `set:"field1"`
		Field2 string
		Field3 bool `set:"field3"`
	}
)

func TestCacheMap(t *testing.T) {
	app := eudore.NewCore()
	app.GetFunc("/get", BindUrl)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get?field1=2&field3=1").Do()
}

func BindUrl(ctx eudore.Context) {
	var req Request1
	ctx.Bind(&req)
	eudore.JSON(req)
}
