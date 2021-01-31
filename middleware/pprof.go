package middleware

import (
	"bytes"
	"expvar"
	"fmt"
	"net/http/pprof"
	"runtime"
	"strings"

	"github.com/eudore/eudore"
)

var goroot string = runtime.GOROOT()

type pprofController struct {
	eudore.ControllerAutoRoute
}

// NewPprofController 函数定义一个pprof控制器注入pprof处理函数。
func NewPprofController() eudore.Controller {
	return new(pprofController)
}

func (pprofController) Inject(_ eudore.Controller, r eudore.Router) error {
	r = r.Group("/pprof")
	server := r.Params().Get("godoc")
	if server == "" {
		server = "https://godoc.org"
	}
	r.AddMiddleware(pprofMiddleware(server))
	r.AnyFunc("/", pprof.Index)
	r.AnyFunc("/cmdline", pprof.Cmdline)
	r.AnyFunc("/profile", pprof.Profile)
	r.AnyFunc("/symbol", pprof.Symbol)
	r.AnyFunc("/trace", pprof.Trace)
	r.AnyFunc("/allocs", pprof.Handler("allocs"))
	r.AnyFunc("/block", pprof.Handler("block"))
	r.AnyFunc("/goroutine", pprof.Handler("goroutine"))
	r.AnyFunc("/heap", pprof.Handler("heap"))
	r.AnyFunc("/mutex", pprof.Handler("mutex"))
	r.AnyFunc("/threadcreate", pprof.Handler("threadcreate"))
	r.AnyFunc("/expvar", HandlerExpvar)
	return nil
}

type response struct {
	eudore.ResponseWriter
	*bytes.Buffer
}

func (w *response) Write(p []byte) (int, error) {
	return w.Buffer.Write(p)
}

func pprofMiddleware(server string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		w := &response{ctx.Response(), bytes.NewBuffer(nil)}
		ctx.SetResponse(w)
		// fixpprof 如果X-Content-Type-Options=nosniff可能直接返回html文本。
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		ctx.Next()

		if ctx.GetQuery("debug") == "" || ctx.GetQuery("m") == "txt" {
			w.ResponseWriter.Write(w.Buffer.Bytes())
			return
		}
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		w.ResponseWriter.Write([]byte("<pre>"))
		renderPprofHTML(w, server)
		w.ResponseWriter.Write([]byte("</pre>"))
	}
}

func renderPprofHTML(w *response, server string) {
	for {
		line, err := w.ReadString('\n')
		if err != nil {
			w.ResponseWriter.Write([]byte(line))
			break
		}
		first := strings.Index(line, "/src/")
		end := strings.LastIndex(line, ".go:")
		// 处理 src .go文件
		if first != -1 && end != -1 {
			first += 5
			for end += 4; '0' <= line[end] && line[end] <= '9' && end < len(line); end++ {
			}

			if server == "https://godoc.org" {
				if strings.Contains(line, goroot) {
					line = fmt.Sprintf("%s<a href='https://golang.org/src/%s'>%s</a>%s", line[:first], strings.Replace(line[first:end], ".go:", ".go#L", 1), line[first:end], line[end:])
				} else if strings.HasPrefix(line[first:end], "github.com/") {
					args := strings.Split(strings.Replace(line[first:end], ".go:", ".go#L", 1), "/")
					args[2] = args[2] + "/blob/master"
					line = fmt.Sprintf("%s<a href='https://%s'>%s</a>%s", line[:first], strings.Join(args, "/"), line[first:end], line[end:])
				}
			} else {
				line = fmt.Sprintf("%s<a href='%s/src/%s'>%s</a>%s", line[:first], server, strings.Replace(line[first:end], ".go:", ".go#L", 1), line[first:end], line[end:])
			}

			// 处理 debug=1 pkg文件
			if strings.HasPrefix(line, "#") {
				pos := strings.Index(line[11:], "\t") + 11
				line = line[:11] + createlink(line[11:pos], server) + line[pos:]
			}
		} else {
			// 处理 debug=2 pkg文件
			switch {
			case strings.HasPrefix(line, "created by "):
				line = "created by " + createlink(line[11:], server)
			case strings.HasSuffix(line, ")\n"):
				line = createlink(line, server)
			}
		}
		w.ResponseWriter.Write([]byte(line))
	}
}

func createlink(str, route string) string {
	pos := strings.LastIndexByte(str, '/')
	if pos == -1 {
		pos = 0
	}
	pos = strings.IndexByte(str[pos:], '.') + pos
	pkg := str[:pos]
	if pkg == "main" {
		return str
	}
	str = str[pos:]

	pos, hash := getLinkHash(str)
	return fmt.Sprintf("<a href='%s/pkg/%s%s'>%s</a>%s", route, pkg, hash, pkg+str[:pos], str[pos:])
}

// getLinkHash 函数获取对象的有效连接长度和锚点。
func getLinkHash(str string) (int, string) {
	strs := strings.Split(str, ".")
	name, length := getName(strs[1])
	if name == "" {
		return 0, ""
	}
	hash := "#" + name
	if len(strs) > 2 {
		name, length2 := getName(strs[2])
		if name != "" {
			hash = hash + "." + name
			return length + length2 + 2, hash
		}
	}
	return length + 1, hash
}

// getName 函数获取对象名称，处理指针、参数、偏移
func getName(str string) (string, int) {
	if str == "" {
		return "", 0
	}
	length := len(str)
	if strings.HasPrefix(str, "(") {
		str = str[1 : len(str)-1]
		if str[0] == '*' {
			str = str[1:]
		}
	}
	if strings.HasSuffix(str, ")\n") {
		pos := strings.LastIndexByte(str, '(')
		str = str[:pos]
		length = len(str)
	}
	pos := strings.LastIndex(str, "+0x")
	if pos != -1 {
		str = str[:pos]
		length = len(str)
	}
	if str[0] < 'A' || str[0] > 'Z' {
		return "", 0
	}
	return str, length
}

// HandlerExpvar 方法实现expvar处理。
func HandlerExpvar(ctx eudore.Context) {
	ctx.SetHeader("Content-Type", "application/json; charset=utf-8")
	ctx.SetHeader("X-Eudore-Admin", "expvar")
	ctx.WriteHeader(200)
	ctx.Write([]byte("{\n"))
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			ctx.Write([]byte(",\n"))
		}
		first = false
		fmt.Fprintf(ctx, "%q: %s", kv.Key, kv.Value)
	})
	ctx.Write([]byte("\n}\n"))
}
