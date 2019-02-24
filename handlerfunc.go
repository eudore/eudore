package eudore

import (
	"os"
	"io"
	"fmt"
	"path"
	"mime"
	"strings"
	"strconv"
	"net/url"
	"net/http"
	"unicode/utf8"
	"path/filepath"
)

/*
当前文件定义各种ctx处理函数
*/

// Redirect a Context.
//
// 重定向一个Context。
func Redirect(ctx Context, redirectUrl string, code int) {
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
		RedirectInternal(ctx, redirectUrl, code)
	}
	RedirectInternal(ctx, redirectUrl, code)
}

// Request internal redirects.
//
// 请求内部重定向。
func RedirectInternal(ctx Context, redirectUrl string, code int) {
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
func RedirectExternal(ctx Context, redirectUrl string, code int) {
	method := ctx.Request().Method()
	h := ctx.Response().Header()

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	// Do it only if the request didn't already have a Content-Type header.
	_, hadCT := h[HeaderContentType]

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


func ServeFile(ctx Context, path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return 0, err
	}
	h := ctx.Response().Header()

	ctypes, haveType := h["Content-Type"]
	var ctype string
	if !haveType {
		ctype = mime.TypeByExtension(filepath.Ext(path))
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [sniffLen]byte
			n, _ := io.ReadFull(f, buf[:])
			ctype = http.DetectContentType(buf[:n])
			_, err := f.Seek(0, io.SeekStart) // rewind to output whole file
			if err != nil {
				ctx.Error("seeker can't seek", http.StatusInternalServerError)
				return 0, nil
			}
		}
		h.Set("Content-Type", ctype)
	} else if len(ctypes) > 0 {
		ctype = ctypes[0]
	}


	// h.Set("Content-Type", "multipart/byteranges; boundary=")
	h.Set("Last-Modified", d.ModTime().UTC().Format(TimeFormat))
	h.Set("Accept-Ranges", "bytes")
	if h.Get("Content-Encoding") == "" {
		// h.Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}
	// h(HeaderContentType, htmlContentType)
	// http.ServeFile(c.Response(), c.Request(), path)
	n, err := io.Copy(ctx, f)
	return int(n), err
}




func readQuery(query string, p Params) (err error) {
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		p.AddParam(key, value)
	}
	return err
}
