package pprof

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/eudore/eudore"
)

// GodocServer 定义默认外置godoc server地址，如果为空会启动一个内置godoc(go1.11)或者使用golang.org为默认地址。
var GodocServer = os.Getenv("GODOC")

type response struct {
	eudore.ResponseWriter
	*bytes.Buffer
	status int
}

func (w *response) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *response) Write(p []byte) (int, error) {
	return w.Buffer.Write(p)
}

// Init 函数实现注入pprof路由。
func Init(r eudore.Router) {
	r = r.Group("/pprof")
	server := GodocServer
	route := r.GetParam("route")
	if server == "" {
		// go.11创建内置godoc并启动
		godoc := NewGodoc(route + "/godoc")
		if godoc != nil {
			server = route + "/godoc"
			r.AnyFunc("/godoc/*", fixgodoc(server), godoc)
		}
	}
	if server == "" {
		// 1.9 1.10没配置godoc就跳转到官网去，如果非标准库就404算了。
		server = "https://golang.org"
	}

	r.AddMiddleware(pprofMiddleware(server))
	r.AnyFunc("/", pprof.Index)
	r.AnyFunc("/*", fixpath, pprof.Index)
	r.AnyFunc("/cmdline", pprof.Cmdline)
	r.AnyFunc("/profile", pprof.Profile)
	r.AnyFunc("/symbol", pprof.Symbol)
	r.AnyFunc("/trace", pprof.Trace)
	r.AnyFunc("/expvar", Expvar)
	r.AnyFunc("/look godoc="+server, Look)
	r.AnyFunc("/look/* godoc="+server, Look)
}

func fixgodoc(route string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctx.Request().URL.Path = "/" + ctx.GetParam("*")
		w := &response{ctx.Response(), bytes.NewBuffer(nil), 200}
		ctx.SetResponse(w)
		ctx.Next()
		loca := ctx.Response().Header().Get(eudore.HeaderLocation)
		if w.status == 301 && !strings.HasPrefix(loca, route) {
			http.Redirect(w.ResponseWriter, ctx.Request(), route+loca, 301)
		} else {
			w.ResponseWriter.WriteHeader(w.status)
			w.ResponseWriter.Write(w.Buffer.Bytes())
		}
	}
}

// 修复pprof前缀要求
func fixpath(ctx eudore.Context) {
	ctx.Request().URL.Path = "/debug/pprof/" + ctx.GetParam("*")
}

func pprofMiddleware(server string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		w := &response{ctx.Response(), bytes.NewBuffer(nil), 200}
		ctx.SetResponse(w)
		// fixpprof 如果X-Content-Type-Options=nosniff可能直接返回html文本。
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		ctx.Next()

		if ctx.GetQuery("debug") != "" && ctx.GetQuery("m") != "txt" {
			ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
			w.ResponseWriter.WriteHeader(w.status)
			w.ResponseWriter.Write([]byte("<pre>"))
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
					ends := getPos(line[end+4:])
					end2 := end + 4 + len(ends)
					line = fmt.Sprintf("%s<a href='%s%s.go#L%s'>%s</a>%s", line[:first+5], server, line[first:end], ends, line[first+5:end2], line[end2:])
					// 处理 debug=1 pkg文件
					if strings.HasPrefix(line, "#") {
						p1, p2 := getPkgPos(line)
						if p1 != -1 {
							line = line[:p1+1] + createlink(line[p1+1:p2], server) + line[p2:]
						}
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
			w.ResponseWriter.Write([]byte("</pre>"))
		} else {
			w.ResponseWriter.WriteHeader(w.status)
			w.ResponseWriter.Write(w.Buffer.Bytes())
		}
	}
}

func getPos(str string) string {
	for i := range str {
		if str[i] < '0' || str[i] > '9' {
			return str[:i]
		}
	}
	return str
}

func getPkgPos(str string) (int, int) {
	var n, star int
	for i, s := range str {
		if s == 9 {
			n++
			if n == 2 {
				star = i
			}
			if n == 3 {
				return star, i
			}
		}
	}
	return -1, -1
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
