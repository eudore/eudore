package main

import (
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	type Data struct {
		Name string
	}
	app := eudore.NewApp()
	// 修改默认客户端全部Hook
	app.SetValue(eudore.ContextKeyClient, eudore.NewClientCustom(
		eudore.NewClientHookCookie(nil),
		eudore.NewClientHookTimeout(time.Second),
		eudore.NewClientHookRedirect(nil),
		eudore.NewClientHookRetry(3, nil, nil),
		eudore.NewClientHookLogger(eudore.LoggerInfo, time.Millisecond*20),
	))
	app.SetValue(eudore.ContextKeyClient, app.NewClient(
		eudore.NewClientOptionHost("eudore.cn"),
	))
	app.AddMiddleware(
		"global",
		middleware.NewLoggerFunc(app),
		middleware.NewRequestIDFunc(nil),
	)
	app.GetFunc("/data", func() any {
		return &Data{"eudore"}
	})

	app.GetRequest("user",
		eudore.NewClientOptionURL("/api"),
		eudore.NewClientCheckStatus(404),
		eudore.NewClientCheckBody("404"),
	)

	var data Data
	app.GetRequest("/data",
		eudore.NewClientParseErr(),
		eudore.NewClientParseIf(200, &data),
	)
	app.Debug(data)

	app.Listen(":8088")
	app.Run()
}

/*
func NewClientOption(options []any) *ClientOption
func NewClientQuery(key, val string) url.Values
func NewClientHeader(key, val string) http.Header
func NewClientOptionHost(host string) *ClientOption
func NewClientOptionURL(host string) *ClientOption
func NewClientOptionUserAgent(ua string) http.Header
func NewClientOptionBasicauth(username, password string) http.Header
func NewClientOptionBearer(bearer string) http.Header
func NewClientOptionEventID(id int) http.Header
func NewClientCheckStatus(status ...int) func(*http.Response) error
func NewClientCheckBody(str string) func(*http.Response) error
func NewClientParse(data any) func(*http.Response) error
func NewClientParseErr() func(*http.Response) error
func NewClientParseIf(status int, data any) func(*http.Response) error
func NewClientParseIn(star, end int, data any) func(*http.Response) error
func NewClientEventHandler[T any](fn func(e *Event[T]) error) *ClientOption
func NewClientEventChan[T any](events chan *Event[T]) *ClientOption
func NewClientEventCancel(cancel func()) *ClientOption
*/
