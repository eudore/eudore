package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

// HandlerAdmin 函数返回Admin UI界面。
func HandlerAdmin(ctx eudore.Context) {
	ctx.SetHeader("X-Eudore-Admin", "ui")
	ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
	http.ServeContent(ctx.Response(), ctx.Request(), "admin.html", now, strings.NewReader(AdminStatic))
}

// NewBasicAuthFunc 创建一个Basic auth认证中间件。
//
// names为保存用户密码的map。
//
// 注意: BasicAuth需要放置在CORS之后。
func NewBasicAuthFunc(names map[string]string) eudore.HandlerFunc {
	checks := make(map[string]string, len(names))
	for name, pass := range names {
		checks[base64.StdEncoding.EncodeToString([]byte(name+":"+pass))] = name
	}
	return func(ctx eudore.Context) {
		auth := ctx.GetHeader(eudore.HeaderAuthorization)
		if len(auth) > 5 && auth[:6] == "Basic " {
			name, ok := checks[auth[6:]]
			if ok {
				ctx.SetParam("basicauth", name)
				return
			}
		}
		ctx.SetHeader(eudore.HeaderWWWAuthenticate, "Basic")
		ctx.WriteHeader(401)
		ctx.End()
	}
}

// NewBodyLimitFunc 函数创建显示请求body长度的处理中间件。
func NewBodyLimitFunc(size int64) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		req := ctx.Request()
		if req.ContentLength > size {
			ctx.WriteHeader(http.StatusRequestEntityTooLarge)
			ctx.Render(struct {
				Status  int    `json:"status" xml:"status"`
				Message string `json:"message" xml:"message"`
				Size    int64  `json:"size" xml:"size"`
			}{Status: 413, Message: "Request Entity Too Large", Size: req.ContentLength})
			ctx.End()
			return
		}

		req.Body = &limitedReader{req.Body, size}
	}
}

type limitedReader struct {
	io.ReadCloser       // underlying reader
	N             int64 // max bytes remaining
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.ReadCloser.Read(p)
	l.N -= int64(n)
	return
}

// NewHeaderFunc 函数创建响应header写入中间件。
func NewHeaderFunc(h http.Header) eudore.HandlerFunc {
	if len(h) == 0 {
		return nil
	}
	return func(ctx eudore.Context) {
		header := ctx.Response().Header()
		for k, v := range h {
			header[k] = append(header[k], v...)
		}
	}
}

// NewHeaderWithSecureFunc 函数创建响应header写入中间件，并额外附加基本安全header。
func NewHeaderWithSecureFunc(h http.Header) eudore.HandlerFunc {
	header := http.Header{
		"X-XSS-Protection":       []string{"1; mode=block"},
		"X-Frame-Options":        []string{"SAMEORIGIN"},
		"X-Content-Type-Options": []string{"nosniff"},
	}
	for k, v := range h {
		header[k] = append(header[k], v...)
	}
	return NewHeaderFunc(header)
}

// NewLoggerFunc 函数创建一个请求日志记录中间件。
//
// app参数传入*eudore.App需要使用其Logger输出日志，paramsh获取Context.Params如果不为空则添加到输出日志条目中
//
// 状态码如果为40x、50x输出日志级别为Error。
func NewLoggerFunc(app *eudore.App, params ...string) eudore.HandlerFunc {
	keys := []string{"method", "path", "realip", "proto", "host", "status", "request-time", "size"}
	headerkeys := [...]string{eudore.HeaderXRequestID, eudore.HeaderXTraceID, eudore.HeaderLocation}
	headernames := [...]string{"x-request-id", "x-trace-id", "location"}
	return func(ctx eudore.Context) {
		now := time.Now()
		ctx.Next()
		status := ctx.Response().Status()
		// 连续WithField保证field顺序
		out := app.WithFields(keys, []interface{}{
			ctx.Method(), ctx.Path(), ctx.RealIP(), ctx.Request().Proto,
			ctx.Host(), status, time.Now().Sub(now).String(), ctx.Response().Size(),
		})

		for _, param := range params {
			val := ctx.GetParam(param)
			if val != "" {
				out = out.WithField(strings.ToLower(param), val)
			}
		}

		if xforward := ctx.GetHeader(eudore.HeaderXForwardedFor); len(xforward) > 0 {
			out = out.WithField("x-forward-for", xforward)
		}
		headers := ctx.Response().Header()
		for i, key := range headerkeys {
			val := headers.Get(key)
			if val != "" {
				out = out.WithField(headernames[i], val)
			}
		}

		if status < 500 {
			out.Info()
		} else {
			if err := ctx.Err(); err != nil {
				out = out.WithField("error", err.Error())
			}
			out.Error()
		}
	}
}

// NewLoggerLevelFunc 函数创建一个设置一次请求日志级别的中间件。
//
// 通过一个函数处理请求，返回一个0-4,代表日志级别Debug-Fatal,默认处理函数使用debug参数转换成日志级别数字。
func NewLoggerLevelFunc(fn func(ctx eudore.Context) int) eudore.HandlerFunc {
	if fn == nil {
		fn = func(ctx eudore.Context) int {
			level := ctx.GetQuery("eudore_debug")
			if level != "" {
				return eudore.GetStringInt(level)
			}
			return -1
		}
	}
	return func(ctx eudore.Context) {
		l := fn(ctx)
		if -1 < l && l < 5 {
			log := ctx.Logger().WithFields(nil, nil)
			log.SetLevel(eudore.LoggerLevel(l))
			ctx.SetLogger(log)
		}
	}
}

// NewRecoverFunc 函数创建一个错误捕捉中间件，并返回500。
func NewRecoverFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := eudore.GetPanicStack(5)
			ctx.WithField("error", "recover error").WithField("stack", stack).Error(err)

			if ctx.Response().Size() == 0 {
				ctx.WriteHeader(500)
			}
			ctx.Render(map[string]interface{}{
				"error":        err.Error(),
				"stack":        stack,
				"status":       ctx.Response().Status(),
				"x-request-id": ctx.RequestID(),
			})
		}()
		ctx.Next()
	}
}

// NewRequestIDFunc 函数创建一个请求ID注入处理函数，不给定请求ID创建函数，默认使用时间戳和随机数。
func NewRequestIDFunc(fn func(eudore.Context) string) eudore.HandlerFunc {
	if fn == nil {
		fn = func(eudore.Context) string {
			randkey := make([]byte, 3)
			io.ReadFull(rand.Reader, randkey)
			return fmt.Sprintf("%d-%x", time.Now().UnixNano(), randkey)

		}
	}
	return func(ctx eudore.Context) {
		requestID := ctx.GetHeader(eudore.HeaderXRequestID)
		if requestID == "" {
			requestID = fn(ctx)
			ctx.Request().Header.Add(eudore.HeaderXRequestID, requestID)
		}
		ctx.SetHeader(eudore.HeaderXRequestID, requestID)
		ctx.SetLogger(ctx.Logger().WithField("x-request-id", requestID).WithFields(nil, nil))
	}
}
