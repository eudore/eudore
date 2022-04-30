package eudore_test

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func init() {
	eudore.DefaulerServerShutdownWait = time.Microsecond * 100
}

func TestClientRequest(t *testing.T) {
	type Data struct {
		Name string `json:"name"`
	}
	app := eudore.NewApp()
	client := eudore.NewClientWarp(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}, func(tp http.RoundTripper) http.RoundTripper {
		return tp
	})
	app.SetValue(eudore.ContextKeyClient, client)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info(ctx.Method(), ctx.Path(), string(ctx.Body()))
	})
	app.AnyFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore" + ctx.Host())
	})
	app.AnyFunc("/redirect/*", func(ctx eudore.Context) {
		ctx.Redirect(308, "/"+ctx.GetParam("*"))
	})
	app.Listen(":8088")
	time.Sleep(5 * time.Millisecond)

	client.AddBasicAuth("eudore", "pass")
	client.AddHeader("Client", "eudore")
	client.AddHeaders(http.Header{"Debug": []string{"true"}})
	client.AddQuery("client", "eudore")
	client.AddQuerys(url.Values{"debug": []string{"true"}})
	client.AddCookie("", "none", "93df237641ca921e1bacda0eca191030")
	client.AddCookie("%2", "none", "93df237641ca921e1bacda0eca191030")

	app.Info("Client GetCookie / :", client.GetCookie("", "none"))
	app.Info("Client GetCookie qq.com/ :", client.GetCookie("qq.com/", "none"))
	app.Info("Client GetCookie %2 :", client.GetCookie("%2", "none"))

	client.NewRequest("GET", "http://127.0.0.1:8088/hello").Do()
	client.NewRequest("GET", "/hello").Do()

	client.NewRequest("GET", "/hello").AddQuery("name", "eudore").Do()
	client.NewRequest("GET", "/hello").AddHeader("name", "eudore").Do()
	client.NewRequest("GET", "/hello").AddHeader("Host", "eudore").Do()
	client.NewRequest("GET", "/hello").AddHeaders(http.Header{"name": []string{"eudore"}}).Do()

	client.NewRequest("PUT", "/redirect/bodystring").Body("trace").Do()
	client.NewRequest("PUT", "/redirect/bodybytes").Body([]byte("trace")).Do()
	client.NewRequest("PUT", "/redirect/bodyjson").Body(Data{"eudore"}).Do()
	client.NewRequest("PUT", "/redirect/bodyreadr").Body(strings.NewReader("file body")).Do()
	client.NewRequest("PUT", "/redirect/bodyreadcloser").Body(ioutil.NopCloser(strings.NewReader("file body"))).Do()

	client.NewRequest("PUT", "/redirect/body/string").BodyString("trace").Do()
	client.NewRequest("PUT", "/redirect/body/bytes").BodyBytes([]byte("trace")).Do()
	client.NewRequest("PUT", "/redirect/body/json/struct").BodyJSON(Data{"eudore"}).Do()
	client.NewRequest("PUT", "/redirect/body/json/map").BodyJSON(map[string]interface{}{"nanme": "eudore"}).Do()
	client.NewRequest("PUT", "/redirect/body/json/value").BodyJSONValue("k", "v").Do()
	client.NewRequest("PUT", "/redirect/body/form/value").BodyFormValue("name", "eudore").Do()
	client.NewRequest("PUT", "/redirect/body/form/values").BodyFormValues(map[string][]string{"name": {"eudore"}}).Do()
	client.NewRequest("PUT", "/redirect/body/file/string").BodyFormFile("file", "file", "file body").Do()
	client.NewRequest("PUT", "/redirect/body/file/bytes").BodyFormFile("file", "file", []byte("file body")).Do()
	client.NewRequest("PUT", "/redirect/body/file/reader").BodyFormFile("file", "file", strings.NewReader("file body")).Do()
	client.NewRequest("PUT", "/redirect/body/file/readcloser").BodyFormFile("file", "file", ioutil.NopCloser(strings.NewReader("file body"))).Do()
	client.NewRequest("PUT", "/redirect/body/file/nil").BodyFormFile("file", "file", 99).Do()
	client.NewRequest("PUT", "/redirect/body/file/local").BodyFormLocalFile("file", "", "./appNew.go").Do()

	time.Sleep(200 * time.Millisecond)
	app.CancelFunc()
	app.Run()
}

func TestClientCheck(t *testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.AnyFunc("/200", func(ctx eudore.Context) {
		ctx.WriteHeader(200)
		ctx.WriteString("hello")
	})
	app.AnyFunc("/400", func(ctx eudore.Context) {
		ctx.WriteHeader(400)
	})
	app.AnyFunc("/500", func(ctx eudore.Context) {
		ctx.WriteHeader(500)
	})
	app.AnyFunc("/cookie", func(ctx eudore.Context) {
		ctx.SetCookieValue("name", "eudore", -1)
		ctx.WriteJSON(map[string]string{"name": "eudore"})
	})

	client.NewRequest("GET", "/200").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("GET", "/400").Do().Callback(eudore.NewResponseReaderCheckStatus(400))
	client.NewRequest("GET", "/500").Do().Callback(eudore.NewResponseReaderCheckStatus(500))
	client.NewRequest("GET", "/500").Do().Callback(eudore.NewResponseReaderCheckStatus(501))
	client.NewRequest("GET", "/200").Do().Callback(eudore.NewResponseReaderCheckBody("hello"))
	client.NewRequest("GET", "/200").Do().Callback(eudore.NewResponseReaderCheckBody("hello eudore"))
	client.NewRequest("GET", "/200").Do().Callback(eudore.NewResponseReaderOutHead())
	client.NewRequest("GET", "/200").Do().Callback(eudore.NewResponseReaderOutBody())
	client.NewRequest("GET", "/cookie").Do().Callback(func(resp eudore.ResponseReader, req *http.Request, log eudore.Logger) error {
		var data map[string]interface{}
		json.NewDecoder(resp).Decode(&data)
		log.Info("cookies:", resp.Cookies())
		log.Info("data:", data)
		return nil
	})

	eudore.NewClientWarp().NewRequest("GET", "/500").Do().Callback(eudore.NewResponseReaderCheckStatus(501))

	app.CancelFunc()
	app.Run()
}
