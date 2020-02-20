package eudore

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

type (
	// Stream 定义请求流，抽象websocket处理。
	Stream interface {
		io.ReadWriteCloser
		StreamID() string
		GetType() int
		SetType(int)
		SendMsg(m interface{}) error
		RecvMsg(m interface{}) error
	}
	// Header 定义eudore.Header为http.Header。
	Header = http.Header
	// RequestReader 对象为请求信息的载体。
	RequestReader = http.Request
	// ResponseWriter 接口用于写入http请求响应体status、header、body。
	//
	// net/http.response实现了flusher、hijacker、pusher接口。
	ResponseWriter interface {
		// http.ResponseWriter
		Header() http.Header
		Write([]byte) (int, error)
		WriteHeader(int)
		// http.Flusher
		Flush()
		// http.Hijacker
		Hijack() (net.Conn, *bufio.ReadWriter, error)
		// http.Pusher
		Push(string, *http.PushOptions) error
		Size() int
		Status() int
	}
	// ResponseWriterHTTP 是对net/http.ResponseWriter接口封装
	ResponseWriterHTTP struct {
		http.ResponseWriter
		code int
		size int
	}

	// SetCookie 定义响应返回的set-cookie header的数据生成
	SetCookie = http.Cookie
	// Cookie 定义请求读取的cookie header的键值对数据存储
	Cookie struct {
		Name  string
		Value string
	}

	// Params 定义请求上下文中的参数接口。
	Params interface {
		Get(string) string
		Add(string, string)
		Set(string, string)
	}
	// ParamsArray 使用数组实现Params
	ParamsArray struct {
		Keys []string
		Vals []string
	}
)

var (
	sriRegexpScript, _                    = regexp.Compile(`\s*<script.*src=([\"\'])(\S*\.js)([\"\']).*></script>`)
	sriRegexpCSS, _                       = regexp.Compile(`\s*<link.*href=([\"\'])(\S*\.css)([\"\']).*>`)
	sriRegexpImg, _                       = regexp.Compile(`\s*<img.*src=([\"\'])(\S*)([\"\']).*>`)
	sriRegexpIntegrity, _                 = regexp.Compile(`.*\s+integrity=[\"\'](\S*)[\"\'].*`)
	cachePushFile                         = make(map[string][]string)
	_                      ResponseWriter = (*ResponseWriterHTTP)(nil)
	responseWriterHTTPPool                = sync.Pool{
		New: func() interface{} {
			return &ResponseWriterHTTP{}
		},
	}
)

// Clone 方法深复制一个ParamArray对象。
func (p *ParamsArray) Clone() *ParamsArray {
	params := &ParamsArray{
		Keys: make([]string, len(p.Keys)),
		Vals: make([]string, len(p.Vals)),
	}
	copy(params.Keys, p.Keys)
	copy(params.Vals, p.Vals)
	return params
}

func (p *ParamsArray) String() string {
	var b bytes.Buffer
	for i := range p.Keys {
		if p.Keys[i] != "" && p.Vals[i] != "" {
			if b.Len() != 0 {
				b.WriteString(" ")
			}
			fmt.Fprintf(&b, "%s=%s", p.Keys[i], p.Vals[i])
		}
	}
	return b.String()
}

// Get 方法返回一个参数的值。
func (p *ParamsArray) Get(key string) string {
	for i, str := range p.Keys {
		if str == key {
			return p.Vals[i]
		}
	}
	return ""
}

// Add 方法添加一个参数。
func (p *ParamsArray) Add(key string, val string) {
	if key != "" {
		p.Keys = append(p.Keys, key)
		p.Vals = append(p.Vals, val)
	}
}

// Set 方法设置一个参数的值。
func (p *ParamsArray) Set(key string, val string) {
	for i, str := range p.Keys {
		if str == key {
			p.Vals[i] = val
			return
		}
	}
	p.Add(key, val)
}

// Reset 方法重置ResponseWriterHTTP对象。
func (w *ResponseWriterHTTP) Reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.code = http.StatusOK
	w.size = 0
}

// Write 方法实现io.Writer接口。
func (w *ResponseWriterHTTP) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size = w.size + n
	return n, err
}

// WriteHeader 方法实现写入http请求状态码。
func (w *ResponseWriterHTTP) WriteHeader(codeCode int) {
	w.code = codeCode
	w.ResponseWriter.WriteHeader(w.code)
}

// Flush 方法实现刷新缓冲，将缓冲的请求发送给客户端。
func (w *ResponseWriterHTTP) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

// Hijack 方法实现劫持http连接。
func (w *ResponseWriterHTTP) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, ErrResponseWriterHTTPNotHijacker
}

// Push 方法实现http Psuh，如果ResponseWriterHTTP实现http.Push接口，则Push资源。
func (w *ResponseWriterHTTP) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, &http.PushOptions{})
	}
	return nil
}

// Size 方法获得写入的数据长度。
func (w *ResponseWriterHTTP) Size() int {
	return w.size
}

// Status 方法获得设置的http状态码。
func (w *ResponseWriterHTTP) Status() int {
	return w.code
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	i := int(r)
	return !(i < len(isTokenTable) && isTokenTable[i])
}

var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}

// HandlerPush 根据文件名称自动push其中的资源
func HandlerPush(ctx Context, path string) {
	if ctx.Request().Proto != "HTTP/2.0" {
		return
	}
	files, ok := cachePushFile[path]
	if !ok {
		files, _ = getStatic(path)
		cachePushFile[path] = files
	}
	// push file
	for _, file := range files {
		ctx.Push(file, nil)
	}
}

func getStatic(path string) ([]string, error) {
	// 检测文件大小
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() > 10<<20 {
		return nil, fmt.Errorf("%s file is to long, size: %d", path, fileInfo.Size())
	}
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var statics = []string{"/favicon.ico"}
	br := bufio.NewReader(file)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// match script
		params := sriRegexpScript.FindStringSubmatch(line)
		// match css
		if len(params) == 0 {
			params = sriRegexpCSS.FindStringSubmatch(line)
		}
		if len(params) == 0 {
			params = sriRegexpImg.FindStringSubmatch(line)
		}
		// 判断是否匹配数据
		if len(params) > 1 {
			statics = append(statics, params[2])
		}
	}
	return statics, nil
}
