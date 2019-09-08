package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	Request struct {
		Name string `json:"name"`
	}
	Response struct {
		Name string
	}
)

func TestRpc(t *testing.T) {
	app := eudore.NewCore()
	app.PostFunc("/get", hanele1)
	app.PostFunc("/2", eudore.NewRPCHandlerFunc(hanele2))

	client := httptest.NewClient(app).WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationJSON)
	client.NewRequest("POST", "/get").WithBodyString(`{"Name": "han1"}`).Do()
	client.NewRequest("POST", "/2").WithBodyString(`{"Name": "han2"}`).Do()

	time.Sleep(time.Second)
}

func hanele1(ctx eudore.Context, req map[string]interface{}) (resp map[string]interface{}, err error) {
	eudore.JSON(req)
	fmt.Println("hanele1", ctx.Path())
	resp = map[string]interface{}{
		"name": "hanele1",
	}
	// resp["name"] = "hanele1"
	return
}

func hanele2(ctx eudore.Context, req *Request) (resp Response, err error) {
	resp.Name = req.Name
	return
}
