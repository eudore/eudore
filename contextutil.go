package eudore

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"
)

type contextBaseValue struct {
	sync.RWMutex
	context.Context
	Logger
	Error  error
	Values []any
}

var baseCtxKey int

func (ctx *contextBaseValue) Reset(c context.Context, conf *contextBaseConfig) {
	ctx.Context = c
	ctx.Logger = conf.Logger
	ctx.Error = nil
	ctx.Values = ctx.Values[0:0]
}

func (ctx *contextBaseValue) Clone() context.Context {
	ctx.RLock()
	defer ctx.RUnlock()
	base := &contextBaseValue{
		Context: ctx.Context,
		Logger:  ctx.Logger,
		Error:   ctx.Error,
		Values:  make([]any, len(ctx.Values)),
	}
	copy(base.Values, ctx.Values)
	return base
}

func (ctx *contextBaseValue) SetValue(key, val any) {
	ctx.Lock()
	defer ctx.Unlock()
	switch key {
	case ContextKeyLogger:
		ctx.Logger, _ = val.(Logger)
	case ContextKeyError:
		ctx.Error, _ = val.(error)
	default:
		for i := 0; i < len(ctx.Values); i += 2 {
			if ctx.Values[i] == key {
				ctx.Values[i+1] = val
				return
			}
		}
		ctx.Values = append(ctx.Values, key, val)
	}
}

func (ctx *contextBaseValue) Value(key any) any {
	ctx.RLock()
	defer ctx.RUnlock()
	switch key {
	case &baseCtxKey:
		return ctx
	case ContextKeyLogger:
		return ctx.Logger
	}
	for i := 0; i < len(ctx.Values); i += 2 {
		if ctx.Values[i] == key {
			return ctx.Values[i+1]
		}
	}
	return ctx.Context.Value(key)
}

func (ctx *contextBaseValue) Err() error {
	ctx.RLock()
	defer ctx.RUnlock()
	if ctx.Error != nil {
		return ctx.Error
	}
	return ctx.Context.Err()
}

func (ctx *contextBaseValue) String() string {
	ctx.RLock()
	defer ctx.RUnlock()
	var meta []string
	for i := 0; i < len(ctx.Values); i += 2 {
		meta = append(meta,
			fmt.Sprintf("%v=%v", ctx.Values[i], ctx.Values[i+1]),
		)
	}
	if ctx.Error != nil {
		meta = append(meta, "error="+ctx.Error.Error())
	}
	return fmt.Sprintf("%v.WithEudoreContext(%s)",
		ctx.Context, strings.Join(meta, ", "),
	)
}

// contextBaseEntry implements the Wrap [Logger] used by [contextBase].
type contextBaseEntry struct {
	Logger
	writeFatal func(error)
}

// The Fatal method writes the Error log and ends the request.
func (e *contextBaseEntry) Fatal(args ...any) {
	err := getMessagError(args)
	e.writeFatal(err)
	e.Error(err.Error())
}

// The Fatalf method writes the Error log and ends the request.
func (e *contextBaseEntry) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	e.writeFatal(errors.New(msg))
	e.Error(msg)
}

// The WithField method adds a logger field.
func (e *contextBaseEntry) WithField(key string, value any) Logger {
	e.Logger = e.Logger.WithField(key, value)
	return e
}

// The WithFields method adds multiple logger field.
func (e *contextBaseEntry) WithFields(keys []string, fields []any) Logger {
	e.Logger = e.Logger.WithFields(keys, fields)
	return e
}

// responseWriterHTTP is a wrapper for the [http.ResponseWriter] interface.
type responseWriterHTTP struct {
	http.ResponseWriter
	code int
	size int
}

// The Reset method resets the responseWriterHTTP object.
func (w *responseWriterHTTP) Reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.code = http.StatusOK
	w.size = 0
}

// The Unwrap method returns the original [http.ResponseWrite].
func (w *responseWriterHTTP) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// The Write method implements the [io.Writer] interface.
func (w *responseWriterHTTP) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

// The WriteString method implements the [io.StringWriter] interface.
func (w *responseWriterHTTP) WriteString(data string) (int, error) {
	n, err := io.WriteString(w.ResponseWriter, data)
	w.size += n
	return n, err
}

func (w *responseWriterHTTP) WriteStatus(code int) {
	if w.code > 0 {
		if code > 0 {
			w.code = code
		} else {
			w.ResponseWriter.WriteHeader(w.code)
			w.code = -w.code
		}
	}
}

// The WriteHeader method implements writing status code and [http.Header].
func (w *responseWriterHTTP) WriteHeader(code int) {
	if w.code > 0 && code > 0 {
		w.ResponseWriter.WriteHeader(code)
		w.code = -code
	}
}

// The Flush method implements the [http.Flusher] interface.
func (w *responseWriterHTTP) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// The Hijack method implements the [http.Hijacker] interface.
func (w *responseWriterHTTP) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		w.code = -StatusSwitchingProtocols
		return hijacker.Hijack()
	}
	return nil, nil, ErrContextNotHijacker
}

// The Push method implements the [http.Pusher] interface.
func (w *responseWriterHTTP) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return nil
}

// The Size method returns the length of the data written.
func (w *responseWriterHTTP) Size() int {
	return w.size
}

// The Status method returns the set http status code.
func (w *responseWriterHTTP) Status() int {
	// abs
	m := w.code >> 31
	return (w.code + m) ^ m
}

type contextMessage struct {
	Time       string `json:"time" protobuf:"1,name=time" yaml:"time"`
	Host       string `json:"host" protobuf:"2,name=host" yaml:"host"`
	Method     string `json:"method" protobuf:"3,name=method" yaml:"method"`
	Path       string `json:"path" protobuf:"4,name=path" yaml:"path"`
	Route      string `json:"route" protobuf:"5,name=route" yaml:"route"`
	Status     int    `json:"status" protobuf:"6,name=status" yaml:"status"`
	Code       int    `json:"code,omitempty" protobuf:"7,name=code" yaml:"code,omitempty"`
	XRequestID string `json:"x-request-id,omitempty" protobuf:"8,name=x-request-id" yaml:"xRequestId,omitempty"`
	XTraceID   string `json:"x-trace-id,omitempty" protobuf:"9,name=x-trace-id" yaml:"xTraceId,omitempty"`
	Error      string `json:"error,omitempty" protobuf:"10,name=error" yaml:"error,omitempty"`
	Message    any    `json:"message,omitempty" protobuf:"11,name=message" yaml:"message,omitempty"`
}

// The NewContextMessgae method creates a message of an error or object
// and records information related to the [Context].
//
// If the error type is [http.MaxBytesError], status set to
// [StatusRequestEntityTooLarge].
func NewContextMessgae(ctx Context, err error, message any) any {
	h := ctx.Response().Header()
	msg := contextMessage{
		Time:       time.Now().Format(DefaultContextFormatTime),
		Host:       ctx.Host(),
		Method:     ctx.Method(),
		Path:       ctx.Path(),
		Route:      ctx.GetParam(ParamRoute),
		XRequestID: h.Get(HeaderXRequestID),
		XTraceID:   h.Get(HeaderXTraceID),
		Status:     ctx.Response().Status(),
		Message:    message,
	}
	if err != nil {
		msg.Code = getErrorCode(err)
		msg.Error = err.Error()
	}
	return msg
}

func getErrorStatus(err error) int {
	var statusErr interface{ Status() int }
	if errors.As(err, &statusErr) {
		return statusErr.Status()
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return StatusRequestEntityTooLarge
	}

	return StatusInternalServerError
}

func getErrorCode(err error) int {
	var codeErr interface{ Code() int }
	if errors.As(err, &codeErr) {
		return codeErr.Code()
	}
	return 0
}

func (ctx *contextBase) wrapLogger() Logger {
	return ctx.logger().WithField(ParamDepth, 1)
}

func (ctx *contextBase) Debug(args ...any) {
	ctx.wrapLogger().Debug(args...)
}

func (ctx *contextBase) Info(args ...any) {
	ctx.wrapLogger().Info(args...)
}

func (ctx *contextBase) Warning(args ...any) {
	ctx.wrapLogger().Warning(args...)
}

func (ctx *contextBase) Error(args ...any) {
	if hasMessagError(args) {
		ctx.wrapLogger().Error(args...)
	}
}

func (ctx *contextBase) Fatal(args ...any) {
	err := getMessagError(args)
	ctx.writeFatal(err)
	ctx.wrapLogger().Error(err.Error())
}

func hasMessagError(args []any) bool {
	if len(args) == 1 {
		err, ok := args[0].(error)
		if ok {
			return err != nil
		}
		return args[0] != nil
	}
	return true
}

func getMessagError(args []any) error {
	if len(args) == 1 {
		err, ok := args[0].(error)
		if ok {
			return err
		}
	}
	msg := fmt.Sprintln(args...)
	msg = msg[:len(msg)-1]
	return errors.New(msg)
}

func (ctx *contextBase) Debugf(format string, args ...any) {
	ctx.wrapLogger().Debug(fmt.Sprintf(format, args...))
}

func (ctx *contextBase) Infof(format string, args ...any) {
	ctx.wrapLogger().Info(fmt.Sprintf(format, args...))
}

func (ctx *contextBase) Warningf(format string, args ...any) {
	ctx.wrapLogger().Warning(fmt.Sprintf(format, args...))
}

func (ctx *contextBase) Errorf(format string, args ...any) {
	ctx.wrapLogger().Error(fmt.Sprintf(format, args...))
}

func (ctx *contextBase) Fatalf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ctx.writeFatal(errors.New(msg))
	ctx.wrapLogger().Errorf(msg)
}

func (ctx *contextBase) WithField(key string, value any) Logger {
	return &contextBaseEntry{
		Logger:     ctx.logger().WithField(key, value),
		writeFatal: ctx.writeFatal,
	}
}

func (ctx *contextBase) WithFields(keys []string, fields []any) Logger {
	return &contextBaseEntry{
		Logger:     ctx.logger().WithFields(keys, fields),
		writeFatal: ctx.writeFatal,
	}
}

// The parseForm function parses the form data and does not copy the
// PostForm data to the Form.
//
// If Body is ContentLength=0, PostForm = Form.
func (ctx *contextBase) parseForm(r *http.Request) error {
	if r.ContentLength == 0 {
		if r.Form == nil {
			var err error
			r.Form, err = url.ParseQuery(r.URL.RawQuery)
			if err != nil {
				return err
			}
		}
		r.PostForm = r.Form
		return nil
	}

	t, params, err := mime.ParseMediaType(r.Header.Get(HeaderContentType))
	if err != nil {
		return err
	}
	switch t {
	case MimeApplicationForm:
		var reader io.Reader = r.Body
		if reflect.TypeOf(reader).String() != "*http.maxBytesReader" {
			reader = io.LimitReader(r.Body, ctx.config.MaxApplicationFormSize)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		val, err := url.ParseQuery(string(body))
		if err != nil {
			return err
		}
		r.PostForm = val
	case MimeMultipartForm, MimeMultipartMixed:
		boundary, ok := params["boundary"]
		if !ok {
			return http.ErrMissingBoundary
		}

		reader := multipart.NewReader(r.Body, boundary)
		form, err := reader.ReadForm(ctx.config.MaxMultipartFormMemory)
		if err != nil {
			return err
		}
		r.PostForm = form.Value
		r.MultipartForm = form
	default:
		return fmt.Errorf(ErrContextParseFormNotSupportContentType, t)
	}
	return nil
}

func (ctx *contextBase) parseCookies() {
	if len(ctx.cookies) > 0 {
		return
	}
	for _, line := range ctx.RequestReader.Header[HeaderCookie] {
		line = textproto.TrimString(line)
		var part string
		for len(line) > 0 { // continue since we have rest
			part, line, _ = strings.Cut(line, ";")
			part = textproto.TrimString(part)
			if part == "" {
				continue
			}
			name, val, _ := strings.Cut(part, "=")
			if !isCookieNameValid(name) {
				continue
			}
			val, ok := parseCookieValue(val)
			if !ok {
				continue
			}
			ctx.cookies = append(ctx.cookies, Cookie{Name: name, Value: val})
		}
	}
}

var cookieNameSanitizer = strings.NewReplacer("\n", "-", "\r", "-")

// The String method returns the Cookie formatted string.
func (c Cookie) String() string {
	v := sanitizeCookieValue(c.Value)
	if strings.ContainsAny(v, " ,") {
		return `"` + v + `"`
	}
	return cookieNameSanitizer.Replace(c.Name) + "=" + v
}

func sanitizeCookieValue(v string) string {
	for i := 0; i < len(v); i++ {
		if validCookieValueByte(v[i]) {
			continue
		}

		buf := make([]byte, 0, len(v))
		buf = append(buf, v[:i]...)
		for ; i < len(v); i++ {
			if b := v[i]; validCookieValueByte(b) {
				buf = append(buf, b)
			}
		}
		return string(buf)
	}
	return v
}

func parseCookieValue(raw string) (string, bool) {
	if len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
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
	return !(i < len(tableCookie) && tableCookie[i])
}

var tableCookie = [127]bool{
	'!': true, '#': true, '$': true, '%': true, '&': true, '\'': true,
	'*': true, '+': true, '-': true, '.': true, '0': true, '1': true,
	'2': true, '3': true, '4': true, '5': true, '6': true, '7': true,
	'8': true, '9': true, 'A': true, 'B': true, 'C': true, 'D': true,
	'E': true, 'F': true, 'G': true, 'H': true, 'I': true, 'J': true,
	'K': true, 'L': true, 'M': true, 'N': true, 'O': true, 'P': true,
	'Q': true, 'R': true, 'S': true, 'T': true, 'U': true, 'W': true,
	'V': true, 'X': true, 'Y': true, 'Z': true, '^': true, '_': true,
	'`': true, 'a': true, 'b': true, 'c': true, 'd': true, 'e': true,
	'f': true, 'g': true, 'h': true, 'i': true, 'j': true, 'k': true,
	'l': true, 'm': true, 'n': true, 'o': true, 'p': true, 'q': true,
	'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true,
	'x': true, 'y': true, 'z': true, '|': true, '~': true,
}
