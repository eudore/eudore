package middleware

import (
	"net/http"
	"strings"
	"time"
	"unsafe"

	"github.com/eudore/eudore"
)

// The NewLoggerFunc function creates middleware to implement
// output access logs.
//
// Output these fields [DefaultLoggerFixedFields] by default,
//
// You can set params to customize additional fields:
// param:<name> request:<header-name> response:<header-name> cookie:<name>
//
// Default params: response:X-Request-Id response:X-Trace-Id
//
// If the response status is 50x, the output log level is [eudore.LoggerError].
//
// Note that if diff params output duplicate log fields,
// parsing the JSON log will fail.
//
// This middleware needs to be placed before [NewRecoveryFunc],
// and does not handle panic situations.
func NewLoggerFunc(log eudore.Logger, params ...string) Middleware {
	call := loggerInit(log, params)
	return func(ctx eudore.Context) {
		now := time.Now()
		ctx.Next()
		call(ctx, now)
	}
}

// The NewLoggerWithEventFunc function creates middleware to implement handle
// Sever-send-event output access logs.
//
// If it is an SSE request, output the log at the first
// [eudore.ResponseWriter].Flush.
func NewLoggerWithEventFunc(log eudore.Logger, params ...string) Middleware {
	call := loggerInit(log, params)
	return func(ctx eudore.Context) {
		now := time.Now()
		if ctx.GetHeader(eudore.HeaderAccept) != eudore.MimeTextEventStream {
			ctx.Next()
			call(ctx, now)
			return
		}

		w := &responseWriteFlush{
			ResponseWriter: ctx.Response(),
			ctx:            ctx,
			now:            now,
			call:           call,
		}
		ctx.SetResponse(w)
		ctx.Next()
		w.flush()
	}
}

type responseWriteFlush struct {
	eudore.ResponseWriter
	ctx  eudore.Context
	now  time.Time
	call func(eudore.Context, time.Time)
}

func (w *responseWriteFlush) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriteFlush) Flush() {
	w.flush()
	w.ResponseWriter.Flush()
}

func (w *responseWriteFlush) flush() {
	if w.call != nil {
		w.call(w.ctx, w.now)
		w.call = nil
	}
}

func loggerInit(log eudore.Logger, params []string) func(eudore.Context, time.Time) {
	log = log.WithField(eudore.FieldDepth, eudore.DefaultLoggerDepthKindDisable).
		WithField(eudore.FieldLogger, true)
	if params == nil {
		params = []string{"response:X-Request-Id", "response:X-Trace-Id"}
	}
	return func(ctx eudore.Context, now time.Time) {
		r, w := ctx.Request(), ctx.Response()
		// const fields
		out := log.WithField(eudore.FieldTime, now).
			WithFields(DefaultLoggerFixedFields[:], []any{
				r.Host, r.Method, r.URL.Path, r.Proto, ctx.RealIP(),
				ctx.GetParam(eudore.ParamRoute), w.Status(), w.Size(),
				eudore.GetStringDuration(time.Since(now) / 1000),
			})

		rh, wh := r.Header, w.Header()
		for _, key := range params {
			pos := strings.IndexByte(key, ':')
			switch pos {
			case 5: // param:
				key = key[6:]
				out = loggerValue(out, loggerName(key), ctx.GetParam(key))
			case 7: // request:
				key = key[8:]
				out = loggerValue(out, loggerName(key), rh.Get(key))
			case 8: // response:
				key = key[9:]
				out = loggerValue(out, loggerName(key), wh.Get(key))
			case 6: // cookie:
				key = key[7:]
				out = loggerValue(out, loggerName(key), ctx.GetCookie(key))
			default:
				fields := DefaultLoggerOptionalFields
				switch key {
				case fields[0]:
					out = out.WithField(fields[0], loggerRemote(r.RemoteAddr))
				case fields[1]:
					out = out.WithField(fields[1], loggerScheme(r.TLS != nil))
				case fields[2]:
					out = out.WithField(fields[2], r.URL.RawQuery)
				case fields[3]:
					out = out.WithField(fields[3], r.ContentLength)
				}
			}
		}

		if w.Status() < 500 {
			out.Info()
		} else {
			err := ctx.Err()
			if err != nil {
				out = out.WithField(eudore.FieldError, err.Error())
			}
			out.Error()
		}
	}
}

func loggerName(name string) string {
	buf := make([]byte, 0, len(name))
	for i := range name {
		c := name[i]
		switch {
		case 0x40 < c && c < 0x5B:
			buf = append(buf, c+0x20)
		case c == '-':
			buf = append(buf, '_')
		default:
			buf = append(buf, c)
		}
	}
	return unsafe.String(unsafe.SliceData(buf), len(buf))
}

// can inline with cost 70.
func loggerValue(log eudore.Logger, key, val string) eudore.Logger {
	if val == "" {
		return log
	}
	return log.WithField(key, val)
}

func loggerScheme(b bool) string {
	if b {
		return "https"
	}
	return "http"
}

// can inline with cost 64.
func loggerRemote(addr string) string {
	pos := strings.LastIndexByte(addr, ':')
	if pos != -1 {
		addr = addr[:pos]
		// ipv6
		if len(addr) > 1 && addr[0] == '[' {
			addr = addr[1 : len(addr)-1]
		}
	}
	return addr
}

// The NewLoggerLevelFunc function creates middleware to implement
// set the request [eudore.LoggerLevel].
//
// The conversion function returns a 0-4 show the log level
// [eudore.LoggerDebug] - [eudore.LoggerFatal].
//
// The default processing function uses the uri parameter 'eudore_debug'
// to convert it into a log level.
//
//go:noinline
func NewLoggerLevelFunc(fn func(ctx eudore.Context) int) Middleware {
	if fn == nil {
		name := DefaultLoggerLevelQueryName
		fn = func(ctx eudore.Context) int {
			level := ctx.GetQuery(name)
			if level != "" {
				return eudore.GetAnyByString[int](level)
			}
			return -1
		}
	}
	return func(ctx eudore.Context) {
		l := fn(ctx)
		if -1 < l && l < 6 {
			log := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger)
			level := eudore.LoggerLevel(l)
			if log.GetLevel() != level {
				log = log.WithField(eudore.FieldLogger, true)
				log.SetLevel(level)
				ctx.SetValue(eudore.ContextKeyLogger, log)
			}
		}
	}
}
