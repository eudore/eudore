package test

import (
	"context"
	"fmt"
	"github.com/eudore/eudore"
	"testing"
	"time"
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
	app.PostFunc("/get", eudore.HandlerRpc(hanele1))
	app.PostFunc("/2", eudore.HandlerRpc(hanele2))

	req, _ := eudore.NewRequestReaderTest("POST", "/get", `{"Name": "han1"}`)
	req.Header().Add(eudore.HeaderContentType, eudore.MimeApplicationJson)
	resp := eudore.NewResponseWriterTest()
	app.EudoreHTTP(context.Background(), resp, req)

	req, _ = eudore.NewRequestReaderTest("POST", "/2", `{"Name": "han2"}`)
	req.Header().Add(eudore.HeaderContentType, eudore.MimeApplicationJson)
	resp = eudore.NewResponseWriterTest()
	app.EudoreHTTP(context.Background(), resp, req)

	time.Sleep(time.Second)

}

func hanele1(ctx eudore.Context, req map[string]string) (resp map[string]string, err error) {
	eudore.Json(req)
	fmt.Println("hanele1", ctx.Path())
	resp = map[string]string{
		"name": "hanele1",
	}
	// resp["name"] = "hanele1"
	return
}

func hanele2(ctx eudore.Context, req *Request) (resp Response, err error) {
	resp.Name = req.Name
	return
}
