/*
Handler

Handler接口定义了Context对象请求处理函数。

文件：handler.go
*/

package eudore

import (
	"os"
	"io"
	"fmt"
	"mime"
	"sync"
	"bufio"
	"regexp"
	"strings"
	"strconv"
	"errors"
	"mime/multipart"
	"net/textproto"
	"path/filepath"
	"crypto/sha512"
	"net/http"
	"net/http/httptest"
	"github.com/eudore/eudore/protocol"
)

type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}
	HandlerFuncs	[]HandlerFunc
)
var (
	sriRegexpScript, _		= regexp.Compile(`\s*<script.*src=([\"\'])(\S*\.js)([\"\']).*></script>`)
	sriRegexpCss, _			= regexp.Compile(`\s*<link.*href=([\"\'])(\S*\.css)([\"\']).*>`)
	sriRegexpImg, _			= regexp.Compile(`\s*<img.*src=([\"\'])(\S*)([\"\']).*>`)
	sriRegexpIntegrity, _	= regexp.Compile(`.*\s+integrity=[\"\'](\S*)[\"\'].*`)
	sriHashPool				=	sync.Pool {
		New: func() interface{} {
			return sha512.New()
		},
	}
	cachePushFile			=	make(map[string][]string)
	cacheFileType			=	make(map[string]string)
	errNoOverlap = errors.New("invalid range: failed to overlap")
)


func CombineHandlers(hs1, hs2 HandlerFuncs) HandlerFuncs {
	// if nil
	if len(hs1) == 0 {
		return hs2
	}
	if len(hs2) == 0 {
		return hs1
	}
	// combine
	const abortIndex int8 = 63
	finalSize := len(hs1) + len(hs2)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	hs := make(HandlerFuncs, finalSize)
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

func TestHttpHandler(h http.Handler, method, path string) {
	r := httptest.NewRequest(method, path, nil)	
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
}



func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}

// 根据文件名称自动push其中的资源
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

func HandlerError(ctx Context, error string, code int) {
	ctx.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
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

func HandlerFile(ctx Context, path string) (error) {
	f, err := os.Open(path)
	if err != nil {
		ctx.Error(err)
		msg, code := handlerErrorStatus(err)
		HandlerError(ctx, msg, code)
		return err
	}
	defer f.Close()

	desc, err := f.Stat()
	if err != nil {
		ctx.Error(err)
		msg, code := handlerErrorStatus(err)
		HandlerError(ctx, msg, code)
		return err
	}

	// TODO: not test
	// index page
	if desc.IsDir() {
		ctx.Redirect(307, path + "index.html")
		return nil
	}

	return handlerContext(ctx, path, f)
}

func handlerContext(ctx Context, path string, content *os.File) error {
	desc, _ := content.Stat()
	// If Content-Type isn't set, use the file's extension to find it, but
	// if the Content-Type is unset explicitly, do not sniff the type.
	h := ctx.Response().Header()
	h.Set("Last-Modified", desc.ModTime().UTC().Format(TimeFormat))
/*	ctype := h.Get(HeaderContentType)
	if len(ctype) == 0 {
		ctype = getFileType(path)
		h.Set("Content-Type", ctype)
	}*/
	ctype := getFileType(path)
	h.Set("Content-Type", ctype)


	// handle Content-Range header.
	sendSize := desc.Size()
	var sendContent io.Reader = content
	if sendSize >= 0 {
		ranges, err := parseRange(ctx.GetHeader("Range"), sendSize)
		if err != nil {
			if err == errNoOverlap {
				ctx.SetHeader("Content-Range", fmt.Sprintf("bytes */%d", sendSize))
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
			ctx.SetHeader("Content-Range", ra.contentRange(sendSize))
			ctx.WriteHeader(StatusPartialContent)
			sendSize = ra.length
		default:
			ctx.WriteHeader(StatusPartialContent)
			pr, pw := io.Pipe()
			mw := multipart.NewWriter(pw)
			ctx.SetHeader("Content-Type", "multipart/byteranges; boundary="+mw.Boundary())
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
		"Content-Range": {r.contentRange(size)},
		"Content-Type":  {contentType},
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
		return nil, errors.New("invalid range")
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
			return nil, errors.New("invalid range")
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i < 0 {
				return nil, errors.New("invalid range")
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
					return nil, errors.New("invalid range")
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
		return nil, errNoOverlap
	}
	return ranges, nil
}

func getStatic(path string) ([]string, error) {
	// 检测文件大小
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() > 10 << 20 {
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
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
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
	source.Range(func(key, val string){{
		if isEndHeader(key) {
			target.Add(key, val)
		}
	}})
}

func HandlerProxy(addr string) HandlerFunc {
	return func(ctx Context) {
		req := NewClientHttp().NewRequest(ctx.Method(), addr + ctx.Request().RequestURI(), ctx)
		copyheader(ctx.Request().Header(), req.Header())

		req.Header().Set("X-Forwarded-For", ctx.RemoteAddr())
		if ctx.GetHeader("Te") == "trailers" {
			req.Header().Add("Te", "trailers")
		}
		// After stripping all the hop-by-hop connection headers above, add back any
		// necessary for protocol upgrades, such as for websockets.
		if upType := ctx.GetHeader("Upgrade"); len(upType) > 0 {
			req.Header().Add("Connection", "Upgrade")
			req.Header().Add("Upgrade", upType)
		}


		resp, err := req.Do()
		if err != nil {
			ctx.Error(err)
			ctx.WriteHeader(502)
			return
		}
		if resp.Statue() == StatusSwitchingProtocols  {
			// handle Upgrade Response
			return
		}

		// 
		ctx.WriteHeader(resp.Statue())
		copyheader(resp.Header(), ctx.Response().Header())
		io.Copy(ctx, resp)
	}
}


