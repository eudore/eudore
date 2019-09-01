package eudore

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/eudore/eudore/protocol"
)

type (
	// Params 定义请求上下文中的参数接口。
	Params interface {
		GetParam(string) string
		AddParam(string, string)
		SetParam(string, string)
	}
	// Querys 定义请求的uri参数。
	Querys interface {
		Get(string) string
		Set(string, string)
		Add(string, string)
		Del(string)
		Len() int
		Range(func(string, string))
	}
	// ParamsArray 使用数组实现Params
	ParamsArray struct {
		Keys []string
		Vals []string
	}
	// QueryUrl 实现uri参数解析。
	QueryUrl struct {
		Keys []string
		Vals []string
	}
	// SetCookie 定义响应返回的set-cookie header的数据生成
	SetCookie = http.Cookie
	// Cookie 定义请求读取的cookie header的键值对数据存储
	Cookie struct {
		Name  string
		Value string
	}
)

const (
	// sniffLen 是读取数据长度。
	sniffLen = 512
	// timeFormat 定义GMT时间格式化使用的格式。
	timeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
	// RFC6455: The value of this header field MUST be a nonce consisting of a
	// randomly selected 16-byte value that has been base64-encoded (see
	// Section 4 of [RFC4648]).  The nonce MUST be selected randomly for each
	// connection.
	nonceKeySize = 16
	nonceSize    = 24 // base64.StdEncoding.EncodedLen(nonceKeySize)

	// RFC6455: The value of this header field is constructed by concatenating
	// /key/, defined above in step 4 in Section 4.2.2, with the string
	// "258EAFA5- E914-47DA-95CA-C5AB0DC85B11", taking the SHA-1 hash of this
	// concatenated value to obtain a 20-byte value and base64- encoding (see
	// Section 4 of [RFC4648]) this 20-byte hash.
	acceptSize = 28 // base64.StdEncoding.EncodedLen(sha1.Size)
)

// 定义websocket握手时的错误。
var (
	ErrHandshakeBadProtocol     = NewErrorCode(StatusHTTPVersionNotSupported, "handshake error: bad HTTP protocol version")
	ErrHandshakeBadMethod       = NewErrorCode(StatusMethodNotAllowed, "handshake error: bad HTTP request method")
	ErrHandshakeBadHost         = NewErrorCode(StatusBadRequest, "handshake error: bad Host heade")
	ErrHandshakeBadUpgrade      = NewErrorCode(StatusBadRequest, "handshake error: bad Upgrade header")
	ErrHandshakeBadConnection   = NewErrorCode(StatusBadRequest, "handshake error: bad Connection header")
	ErrHandshakeBadSecAccept    = NewErrorCode(StatusBadRequest, "handshake error: bad Sec-Websocket-Accept header")
	ErrHandshakeBadSecKey       = NewErrorCode(StatusBadRequest, "handshake error: bad Sec-Websocket-Key header")
	ErrHandshakeBadSecVersion   = NewErrorCode(StatusBadRequest, "handshake error: bad Sec-Websocket-Version header")
	ErrHandshakeUpgradeRequired = NewErrorCode(StatusUpgradeRequired, "handshake error: bad Sec-Websocket-Version header")
	sriRegexpScript, _          = regexp.Compile(`\s*<script.*src=([\"\'])(\S*\.js)([\"\']).*></script>`)
	sriRegexpCss, _             = regexp.Compile(`\s*<link.*href=([\"\'])(\S*\.css)([\"\']).*>`)
	sriRegexpImg, _             = regexp.Compile(`\s*<img.*src=([\"\'])(\S*)([\"\']).*>`)
	sriRegexpIntegrity, _       = regexp.Compile(`.*\s+integrity=[\"\'](\S*)[\"\'].*`)
	sriHashPool                 = sync.Pool{
		New: func() interface{} {
			return sha512.New()
		},
	}
	cachePushFile = make(map[string][]string)
	cacheFileType = make(map[string]string)
)

// GetParam 方法返回一个参数的值。
func (p *ParamsArray) GetParam(key string) string {
	for i, str := range p.Keys {
		if str == key {
			return p.Vals[i]
		}
	}
	return ""
}

// AddParam 方法添加一个参数。
func (p *ParamsArray) AddParam(key string, val string) {
	p.Keys = append(p.Keys, key)
	p.Vals = append(p.Vals, val)
}

// SetParam 方法设置一个参数的值。
func (p *ParamsArray) SetParam(key string, val string) {
	for i, str := range p.Keys {
		if str == key {
			p.Vals[i] = val
			return
		}
	}
	p.AddParam(key, val)
}

// Get 方法获得一个uri参数的值。
func (q *QueryUrl) Get(key string) string {
	for i, k := range q.Keys {
		if k == key {
			return q.Vals[i]
		}
	}
	return ""
}

// Set 方法设置一个uri参数的值。
func (q *QueryUrl) Set(key string, val string) {
	for i, k := range q.Keys {
		if k == key {
			q.Vals[i] = val
			return
		}
	}
	q.Add(key, val)
}

// Add 方法新增一个uri参数的值。
func (q *QueryUrl) Add(key string, val string) {
	q.Keys = append(q.Keys, key)
	q.Vals = append(q.Vals, val)
}

// Del 方法删除一个uri参数的值。
func (q *QueryUrl) Del(key string) {
	for i, k := range q.Keys {
		if k == key {
			q.Keys[i] = ""
			return
		}
	}
}

// Len 方法返回uri参数的数量。
func (q *QueryUrl) Len() (n int) {
	for _, k := range q.Keys {
		if k != "" {
			n++
		}
	}
	return
}

// Range 方法实现遍历uri参数。
func (q *QueryUrl) Range(fn func(string, string)) {
	for i, k := range q.Keys {
		if k != "" {
			fn(k, q.Vals[i])
		}
	}
}

func (q *QueryUrl) readQuery(query string) (err error) {
	q.Keys = q.Keys[0:0]
	q.Vals = q.Vals[0:0]
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
		q.Keys = append(q.Keys, key)
		q.Vals = append(q.Vals, value)
	}
	return err
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

// HandlerUpgradeHttp 函数实现http upgrade成websocket实现，最后返回net.Conn对象。
//
// source: github.com/gobwas/ws
func HandlerUpgradeHttp(ctx Context) (net.Conn, error) {
	conn, err := ctx.Response().Hijack()
	if err != nil {
		return nil, err
	}

	rw := bufio.NewWriter(conn)
	var nonce string
	if ctx.Method() != MethodGet {
		err = ErrHandshakeBadMethod
	} else if ctx.Request().Proto() != "HTTP/1.1" {
		err = ErrHandshakeBadProtocol
	} else if ctx.Host() == "" {
		err = ErrHandshakeBadHost
	} else if ctx.GetHeader("Upgrade") != "websocket" {
		err = ErrHandshakeBadUpgrade
	} else if ctx.GetHeader("Connection") != "Upgrade" {
		err = ErrHandshakeBadConnection
	} else if v := ctx.GetHeader("Sec-Websocket-Version"); v != "13" {
		if v != "" {
			err = ErrHandshakeUpgradeRequired
		} else {
			err = ErrHandshakeBadSecVersion
		}
	} else if nonce = ctx.GetHeader("Sec-Websocket-Key"); nonce == "" {
		err = ErrHandshakeBadSecKey
	}

	if err == nil {
		httpWriteResponseUpgrade(rw, []byte(nonce))
		err = rw.Flush()
	} else {
		var code int = 500
		if err2, ok := err.(*ErrorCode); ok {
			code = err2.Code()
		}
		ctx.WriteHeader(code)
		ctx.WriteString(err.Error())
		err = rw.Flush()
	}
	return conn, err
}

func httpWriteResponseUpgrade(bw *bufio.Writer, nonce []byte) {
	const textHeadUpgrade = "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
	bw.WriteString(textHeadUpgrade)
	bw.WriteString("Sec-Websocket-Accept: ")
	writeAccept(bw, nonce)
	bw.WriteString("\r\n\r\n")
}

func writeAccept(w io.Writer, nonce []byte) (int, error) {
	var b [acceptSize]byte
	bp := uintptr(unsafe.Pointer(&b))
	bts := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: bp,
		Len:  acceptSize,
		Cap:  acceptSize,
	}))

	initAcceptFromNonce(bts, nonce)

	return w.Write(bts)
}

// initAcceptFromNonce fills given slice with accept bytes generated from given
// nonce bytes. Given buffer should be exactly acceptSize bytes.
func initAcceptFromNonce(dst, nonce []byte) {
	if len(dst) != acceptSize {
		panic("accept buffer is invalid")
	}
	if len(nonce) != nonceSize {
		panic("nonce is invalid")
	}

	sha := acquireSha1()
	defer releaseSha1(sha)

	sha.Write(nonce)
	sha.Write(webSocketMagic)

	var sb [sha1.Size]byte
	sh := uintptr(unsafe.Pointer(&sb))
	sum := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: sh,
		Len:  0,
		Cap:  sha1.Size,
	}))
	sum = sha.Sum(sum)

	base64.StdEncoding.Encode(dst, sum)
}

var webSocketMagic = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

var sha1Pool sync.Pool

// nonce helps to put nonce bytes on the stack and then retrieve stack-backed
// slice with unsafe.
type nonce [nonceSize]byte

func (n *nonce) bytes() []byte {
	h := uintptr(unsafe.Pointer(n))
	b := &reflect.SliceHeader{Data: h, Len: nonceSize, Cap: nonceSize}
	return *(*[]byte)(unsafe.Pointer(b))
}
func acquireSha1() hash.Hash {
	if h := sha1Pool.Get(); h != nil {
		return h.(hash.Hash)
	}
	return sha1.New()
}

func releaseSha1(h hash.Hash) {
	h.Reset()
	sha1Pool.Put(h)
}

// HandlerPush 根据文件名称自动push其中的资源
func HandlerPush(ctx Context, path string) {
	if ctx.Request().Proto() != "HTTP/2.0" {
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

// HandlerError 函数处理请求上下文错误码。
func HandlerError(ctx Context, error string, code int) {
	ctx.Response().Header().Set(HeaderContentType, "text/plain; charset=utf-8")
	ctx.Response().Header().Set("X-Content-Type-Options", "nosniff")
	ctx.WriteHeader(code)
	ctx.WriteString(error)
}

func handlerErrorStatus(err error) (string, int) {
	if os.IsNotExist(err) {
		return "404 page not found", StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", StatusInternalServerError
}

// HandlerFile 函数实现返回一个文件内容，支持304和Range。
func HandlerFile(ctx Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		msg, code := handlerErrorStatus(err)
		HandlerError(ctx, msg, code)
		return err
	}
	defer f.Close()

	desc, err := f.Stat()
	if err != nil {
		msg, code := handlerErrorStatus(err)
		HandlerError(ctx, msg, code)
		return err
	}

	// index page
	if desc.IsDir() {
		ctx.Redirect(307, path+"index.html")
		return nil
	}

	return HandlerContent(ctx, f, desc.ModTime(), desc.Size(), getFileType(path))
}

// HandlerContent 函数返回指定的文件内容。
func HandlerContent(ctx Context, content io.ReadSeeker, modtime time.Time, sendSize int64, ctype string) error {
	if checkPreconditions(ctx, modtime) {
		return nil
	}
	// If Content-Type isn't set, use the file's extension to find it, but
	// if the Content-Type is unset explicitly, do not sniff the type.
	h := ctx.Response().Header()
	h.Set("Last-Modified", modtime.UTC().Format(timeFormat))
	h.Set(HeaderContentType, ctype)

	// handle Content-Range header.
	var sendContent io.Reader = content
	if sendSize >= 0 {
		ranges, err := parseRange(ctx.GetHeader("Range"), sendSize)
		if err != nil {
			if err == ErrHandlerInvalidRange {
				ctx.SetHeader(HeaderContentRange, fmt.Sprintf("bytes */%d", sendSize))
			}
			HandlerError(ctx, err.Error(), StatusRequestedRangeNotSatisfiable)
			return err
		}
		if sumRangesSize(ranges) > sendSize {
			// The total number of bytes in all the ranges
			// is larger than the size of the file by
			// itself, so this is probably an attack, or a
			// dumb client. Ignore the range request.
			ranges = nil
		}
		switch len(ranges) {
		case 0:
		case 1:
			ra := ranges[0]
			if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
				HandlerError(ctx, err.Error(), StatusRequestedRangeNotSatisfiable)
				return err
			}
			ctx.SetHeader(HeaderContentRange, ra.contentRange(sendSize))
			ctx.WriteHeader(StatusPartialContent)
			sendSize = ra.length
		default:
			ctx.WriteHeader(StatusPartialContent)
			pr, pw := io.Pipe()
			mw := multipart.NewWriter(pw)
			ctx.SetHeader(HeaderContentType, "multipart/byteranges; boundary="+mw.Boundary())
			sendContent = pr
			defer pr.Close()
			go func() {
				for _, ra := range ranges {
					part, err := mw.CreatePart(ra.mimeHeader(ctype, sendSize))
					if err != nil {
						pw.CloseWithError(err)
						return
					}
					if _, err := content.Seek(ra.start, io.SeekStart); err != nil {
						pw.CloseWithError(err)
						return
					}
					if _, err := io.CopyN(part, content, ra.length); err != nil {
						pw.CloseWithError(err)
						return
					}
				}
				mw.Close()
				pw.Close()
			}()
		}
	}

	h.Set("Accept-Ranges", "bytes")
	// 内容编码为空时写入长度
	// 否则其他编码下长度不对。
	if h.Get("Content-Encoding") == "" {
		h.Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}
	_, err := io.CopyN(ctx, sendContent, sendSize)
	return err
}

func checkPreconditions(ctx Context, modtime time.Time) bool {
	ch := checkIfMatch(ctx)
	if ch == condNone {
		ch = checkIfUnmodifiedSince(ctx, modtime)
	}
	if ch == condFalse {
		ctx.WriteHeader(StatusPreconditionFailed)
		return true
	}
	switch checkIfNoneMatch(ctx) {
	case condFalse:
		if ctx.Method() == "GET" || ctx.Method() == "HEAD" {
			writeNotModified(ctx)
			return true
		}
		ctx.WriteHeader(StatusPreconditionFailed)
		return true
	case condNone:
		if checkIfModifiedSince(ctx, modtime) == condFalse {
			writeNotModified(ctx)
			return true
		}
	}
	return false
}

type condResult int

const (
	condNone condResult = iota
	condTrue
	condFalse
)

func checkIfMatch(ctx Context) condResult {
	im := ctx.GetHeader("If-Match")
	if im == "" {
		return condNone
	}
	for {
		im = textproto.TrimString(im)
		if len(im) == 0 {
			break
		}
		if im[0] == '.' {
			im = im[1:]
			continue
		}
		if im[0] == '*' {
			return condTrue
		}
		etag, remian := scanETag(im)
		if etag == "" {
			break
		}
		if etagStrongMatch(etag, ctx.Response().Header().Get("Etag")) {
			return condTrue
		}
		im = remian
	}
	return condFalse
}
func checkIfUnmodifiedSince(ctx Context, modtime time.Time) condResult {
	ius := ctx.GetHeader("If-Unmodified-Since")
	if ius == "" || isZeroTime(modtime) {
		return condNone
	}
	if t, err := http.ParseTime(ius); err == nil {
		if modtime.Before(t.Add(1 * time.Second)) {
			return condTrue
		}
		return condFalse
	}
	return condNone
}

func checkIfNoneMatch(ctx Context) condResult {
	inm := ctx.GetHeader("If-None-Match")
	if inm == "" {
		return condNone
	}
	buf := inm
	for {
		buf = textproto.TrimString(buf)
		if len(buf) == 0 {
			break
		}
		if buf[0] == ',' {
			buf = buf[1:]
		}
		if buf[0] == '*' {
			return condFalse
		}
		etag, remain := scanETag(buf)
		if etag == "" {
			break
		}
		if etagWeakMatch(etag, ctx.Response().Header().Get("Etag")) {
			return condFalse
		}
		buf = remain
	}
	return condTrue
}

func checkIfModifiedSince(ctx Context, modtime time.Time) condResult {
	if ctx.Method() != "GET" && ctx.Method() != "HEAD" {
		return condNone
	}
	ims := ctx.GetHeader("If-Modified-Since")
	if ims == "" || isZeroTime(modtime) {
		return condNone
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return condNone
	}
	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if modtime.Before(t.Add(1 * time.Second)) {
		return condFalse
	}
	return condTrue
}

func checkIfRange(ctx Context, modtime time.Time) condResult {
	if ctx.Method() != "GET" && ctx.Method() != "HEAD" {
		return condNone
	}
	ir := ctx.GetHeader("If-Range")
	if ir == "" {
		return condNone
	}
	etag, _ := scanETag(ir)
	if etag != "" {
		if etagStrongMatch(etag, ctx.Response().Header().Get("Etag")) {
			return condTrue
		}
		return condFalse
	}
	// The If-Range value is typically the ETag value, but it may also be
	// the modtime date. See golang.org/issue/8367.
	if modtime.IsZero() {
		return condFalse
	}
	t, err := http.ParseTime(ir)
	if err != nil {
		return condFalse
	}
	if t.Unix() == modtime.Unix() {
		return condTrue
	}
	return condFalse
}

func writeNotModified(ctx Context) {
	h := ctx.Response().Header()
	h.Del("Content-Type")
	h.Del("Content-Length")
	if h.Get("Etag") != "" {
		h.Del("Last-Modified")
	}
	ctx.WriteHeader(StatusNotModified)
}

// etagStrongMatch reports whether a and b match using strong ETag comparison.
// Assumes a and b are valid ETags.
func etagStrongMatch(a, b string) bool {
	return a == b && a != "" && a[0] == '"'
}

// etagWeakMatch reports whether a and b match using weak ETag comparison.
// Assumes a and b are valid ETags.
func etagWeakMatch(a, b string) bool {
	return strings.TrimPrefix(a, "W/") == strings.TrimPrefix(b, "W/")
}

// scanETag determines if a syntactically valid ETag is present at s. If so,
// the ETag and remaining text after consuming ETag is returned. Otherwise,
// it returns "", "".
func scanETag(s string) (etag string, remain string) {
	s = textproto.TrimString(s)
	start := 0
	if strings.HasPrefix(s, "W/") {
		start = 2
	}
	if len(s[start:]) < 2 || s[start] != '"' {
		return "", ""
	}
	// ETag is either W/"text" or "text".
	// See RFC 7232 2.3.
	for i := start + 1; i < len(s); i++ {
		c := s[i]
		switch {
		// Character values allowed in ETags.
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
		case c == '"':
			return s[:i+1], s[i+1:]
		default:
			return "", ""
		}
	}
	return "", ""
}

var unixEpochTime = time.Unix(0, 0)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

func sumRangesSize(ranges []httpRange) (size int64) {
	for _, ra := range ranges {
		size += ra.length
	}
	return
}

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

func (r httpRange) contentRange(size int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.start, r.start+r.length-1, size)
}

func (r httpRange) mimeHeader(contentType string, size int64) textproto.MIMEHeader {
	return textproto.MIMEHeader{
		HeaderContentRange: {r.contentRange(size)},
		HeaderContentType:  {contentType},
	}
}

// parseRange parses a Range header string as per RFC 7233.
// errNoOverlap is returned if none of the ranges overlap.
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, ErrHandlerInvalidRange
	}
	var ranges []httpRange
	noOverlap := false
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, ErrHandlerInvalidRange
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, ErrHandlerInvalidRange
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, ErrHandlerInvalidRange
			}
			if i >= size {
				// If the range begins after the size of the content,
				// then it does not overlap.
				noOverlap = true
				continue
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, ErrHandlerInvalidRange
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	if noOverlap && len(ranges) == 0 {
		// The specified ranges did not overlap with the content.
		return nil, ErrHandlerInvalidRange
	}
	return ranges, nil
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
			params = sriRegexpCss.FindStringSubmatch(line)
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

func getFileType(path string) string {
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		//
		ctype = cacheFileType[path]
		if len(ctype) > 0 {
			return ctype
		}
		//
		f, err := os.Open(path)
		if err != nil {
			return ""
		}
		defer f.Close()
		// read a chunk to decide between utf-8 text and binary
		var buf [sniffLen]byte
		n, _ := io.ReadFull(f, buf[:])
		ctype = http.DetectContentType(buf[:n])
		if err != nil {
			return ""
		}
		//
		cacheFileType[path] = ctype
	}
	return ctype
}

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	HeaderConnection,
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	HeaderUpgrade,
}

func isEndHeader(key string) bool {
	for _, i := range hopHeaders {
		if i == key {
			return false
		}
	}
	return true
}

func copyheader(source protocol.Header, target protocol.Header) {
	source.Range(func(key, val string) {
		{
			if isEndHeader(key) {
				target.Add(key, val)
			}
		}
	})
}

// HandlerProxy 创建一个反向代理处理函数，支持101处理。
func HandlerProxy(addr string) HandlerFunc {
	return func(ctx Context) {
		req, err := http.NewRequest(ctx.Method(), addr+ctx.Request().RequestURI(), ctx)
		if err != nil {
			ctx.Error(err)
			ctx.WriteHeader(502)
			return
		}
		copyheader(ctx.Request().Header(), HeaderMap(req.Header))

		if clientIP, _, err := net.SplitHostPort(ctx.Request().RemoteAddr()); err == nil {
			if prior := ctx.GetHeader(HeaderXForwardedFor); prior != "" {
				clientIP = prior + ", " + clientIP
			}
			req.Header.Set(HeaderXForwardedFor, clientIP)
		}
		if ctx.GetHeader("Te") == "trailers" {
			req.Header.Add("Te", "trailers")
		}
		// After stripping all the hop-by-hop connection headers above, add back any
		// necessary for protocol upgrades, such as for websockets.
		if upType := ctx.GetHeader(HeaderUpgrade); len(upType) > 0 {
			req.Header.Add(HeaderConnection, HeaderUpgrade)
			req.Header.Add(HeaderUpgrade, upType)
		}

		// send proxy
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Error(err)
			ctx.WriteHeader(502)
			return
		}
		ctx.WriteHeader(resp.StatusCode)
		copyheader(HeaderMap(resp.Header), ctx.Response().Header())

		if resp.StatusCode == StatusSwitchingProtocols {
			// handle Upgrade Response
			err = handleUpgradeResponse(ctx, resp)
			if err != nil {
				ctx.Fatal(err)
			}
			return
		}

		//  handle http body
		io.Copy(ctx, resp.Body)
		resp.Body.Close()
	}
}

func handleUpgradeResponse(ctx Context, resp *http.Response) error {
	backConn, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		return ErrHandlerProxyBackNotWriter
	}
	defer backConn.Close()

	conn, err := ctx.Response().Hijack()
	if err != nil {
		return fmt.Errorf(ErrFormatHandlerProxyConnHijack, err)
	}
	defer conn.Close()

	// return ws response
	h := ctx.Response().Header()
	h.Add("Connection", "Upgrade")
	h.Add("Upgrade", "websocket")
	ctx.Response().Flush()

	// start ws io cp
	errc := make(chan error, 1)
	spc := switchProtocolCopier{user: conn, backend: backConn}
	go spc.copyToBackend(errc)
	go spc.copyFromBackend(errc)
	return <-errc
}

// switchProtocolCopier exists so goroutines proxying data back and
// forth have nice names in stacks.
type switchProtocolCopier struct {
	user, backend io.ReadWriter
}

func (c switchProtocolCopier) copyFromBackend(errc chan<- error) {
	_, err := io.Copy(c.user, c.backend)
	errc <- err
}

func (c switchProtocolCopier) copyToBackend(errc chan<- error) {
	_, err := io.Copy(c.backend, c.user)
	errc <- err
}

// HandlerRedirect Redirect a Context.
//
// HandlerRedirect 重定向一个Context。
func HandlerRedirect(ctx Context, redirectUrl string, code int) {
	// parseURL is just url.Parse (url is shadowed for godoc).
	if u, err := url.Parse(redirectUrl); err == nil {
		// If url was relative, make its path absolute by
		// combining with request path.
		// The client would probably do this for us,
		// but doing it ourselves is more reliable.
		// See RFC 7231, section 7.1.2
		if u.Scheme == "" && u.Host == "" {
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
		}
	}

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
