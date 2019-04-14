package eudore

import (
	"fmt"
	"path"
	"strings"
	"strconv"
	"net/url"
	"net/http"
	"unicode/utf8"
)

/*
当前文件定义各种ctx处理函数
*/

// Redirect a Context.
//
// 重定向一个Context。
func HandlerRedirect(ctx Context, redirectUrl string, code int) {
	u, err := url.Parse(redirectUrl); 
	if err != nil {
		ctx.WithField("error", "redirect").Error(err)
		return
	}


	oldpath := ctx.Path()
	if oldpath == "" { // should not happen, but avoid a crash if it does
		oldpath = "/"
	}

	// no leading http://server
	if redirectUrl == "" || redirectUrl[0] != '/' {
		// make relative path absolute
		olddir, _ := path.Split(oldpath)
		redirectUrl = olddir + redirectUrl
	}

	var query string
	if i := strings.Index(redirectUrl, "?"); i != -1 {
		redirectUrl, query = redirectUrl[:i], redirectUrl[i:]
	}

	// clean up but preserve trailing slash
	trailing := strings.HasSuffix(redirectUrl, "/")
	redirectUrl = path.Clean(redirectUrl)
	if trailing && !strings.HasSuffix(redirectUrl, "/") {
		redirectUrl += "/"
	}
	redirectUrl += query

	// 判断是否内部重定向
	if u.Scheme == "" && u.Host == "" {
		HandlerRedirectInternal(ctx, redirectUrl, code)
	}
	HandlerRedirectInternal(ctx, redirectUrl, code)
}

// Request internal redirects.
//
// 请求内部重定向。
func HandlerRedirectInternal(ctx Context, redirectUrl string, code int) {
	var method string
	switch code {
	case 301:
		fallthrough
	case 302:
		method = MethodGet
	case 307:
		fallthrough
	case 308:
		method = ctx.Request().Method()
	default:
		return
	}
	ctx.Debug(method, redirectUrl, "未完成")
	// ctx.App().EudoreHTTP(ctx.Response(), NewRedirectRequest(ctx.Request(), method, redirectUrl))
	ctx.End()
}

// Request external redirects.
//
// 请求外部重定向。
func HandlerRedirectExternal(ctx Context, redirectUrl string, code int) {
	method := ctx.Request().Method()
	h := ctx.Response().Header()

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	// Do it only if the request didn't already have a Content-Type header.
	hadCT := len(h.Get(HeaderContentType)) > 0

	h.Set(HeaderLocation, hexEscapeNonASCII(redirectUrl))
	if !hadCT && (method == MethodGet || method == MethodHead) {
		h.Set(HeaderContentType, "text/html; charset=utf-8")
	}
	ctx.Response().WriteHeader(code)

	// Shouldn't send the body for POST or HEAD; that leaves GET.
	if !hadCT && method == MethodGet {
		body := "<a href=\"" + htmlEscape(redirectUrl) + "\">" + http.StatusText(code) + "</a>.\n"
		fmt.Fprintln(ctx.Response(), body)
	}
}



func hexEscapeNonASCII(s string) string {
	newLen := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			newLen += 3
		} else {
			newLen++
		}
	}
	if newLen == len(s) {
		return s
	}
	b := make([]byte, 0, newLen)
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			b = append(b, '%')
			b = strconv.AppendInt(b, int64(s[i]), 16)
		} else {
			b = append(b, s[i])
		}
	}
	return string(b)
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

func htmlEscape(s string) string {
	return htmlReplacer.Replace(s)
}