package eudore_test

import (
	"context"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func TestHttptestClientParam(t *testing.T) {
	client := httptest.NewClient(nil)
	client.AddQuerys(url.Values{
		"cmd": []string{"httptest"},
	})
	client.AddHeaders(http.Header{
		"Host": []string{"localhost"},
		"Form": []string{"httptest"},
	})
	client.AddCookie("/", "name", "eudore")
	client.GetCookie("/", "name")
	client.AddCookie("http://192.168.0.%31/", "name", "eudore")
	client.GetCookie("http://192.168.0.%31/", "name")
	client.NewRequest("GET", "/").WithHeaders(http.Header{
		"Name": []string{"eudore"},
	})
}

func TestHttptestClientDial(t *testing.T) {
	app := eudore.NewApp()
	app.Listen(":8088")
	app.ListenTLS(":8089", "", "")
	client := httptest.NewClient(nil)
	client.NewRequest("GET", "http://127.0.0.1:/").WithWebsocket(closeTestConn).Do()
	client.Client.Transport = &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			conn, _ := net.Pipe()
			conn.Close()
			return conn, nil
		}}
	client.NewRequest("GET", "http://127.0.0.1:8088/").WithWebsocket(closeTestConn).Do()
	client.Client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	client.NewRequest("GET", "http://127.0.0.1:8088/").WithWebsocket(closeTestConn).Do()
	client.Client.Transport = &http.Transport{
		DialTLS:         net.Dial,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.NewRequest("GET", "https://127.0.0.1:8089/").WithWebsocket(closeTestConn).Do()
	app.CancelFunc()
	app.Run()
}

func closeTestConn(conn net.Conn) {
	conn.Close()
}

func TestHttptestClientResponse(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/wserr", func(ctx eudore.Context) {
		_, rw, err := ctx.Response().Hijack()
		if err != nil {
			ctx.Fatal(err)
			return
		}
		rw.Write([]byte("HTTP/1.1-101\r\n\r\n"))
		rw.Flush()
		// conn.Close()
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/wserr").WithWebsocket(closeTestConn).Do().Out()
	client.NewRequest("GET", "/").Do().HandleRespone(&http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(readErr{}),
	})

	app.CancelFunc()
	app.Run()
}

type readErr struct{}

func (readErr) Read([]byte) (int, error) {
	return 0, errors.New("test Response read error")
}

func TestHttptestClientCheck(t *testing.T) {
	jsondata := map[string]interface{}{
		"name":   "eudore",
		"action": "httptest check",
	}
	app := eudore.NewApp()
	app.AnyFunc("/json", func(ctx eudore.Context) interface{} {
		return jsondata
	})
	app.AnyFunc("/gzip", middleware.NewGzipFunc(5), func(ctx eudore.Context) {
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/gziperr", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentEncoding, "gzip")
		ctx.WriteString("gzip body")
	})

	client := httptest.NewClient(app)
	client.AddHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON)
	client.NewRequest("GET", "/json").Do().CheckBodyJSON(jsondata)
	client.NewRequest("GET", "/json").Do().CheckBodyJSON("ss")
	client.NewRequest("GET", "/json").Do().CheckBodyJSON(struct{ Fn func() }{})
	client.NewRequest("GET", "/json").Do().OutStatus().OutHeader().CheckHeader(eudore.HeaderContentType, "json", eudore.HeaderContentType, "text")
	client.NewRequest("GET", "/gzip").Do().CheckBodyContainString("gzip")
	client.NewRequest("GET", "/gziperr").Do().CheckBodyContainString("gzip")

	app.CancelFunc()
	app.Run()
}
