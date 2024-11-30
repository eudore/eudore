package middleware

import (
	"net/http"
	"strings"
	"time"

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
	log = log.WithField(
		eudore.ParamDepth,
		eudore.DefaultLoggerDepthKindDisable,
	).WithField("logger", true)
	if params == nil {
		params = []string{"response:X-Request-Id", "response:X-Trace-Id"}
	}
	return func(ctx eudore.Context, now time.Time) {
		r, w := ctx.Request(), ctx.Response()
		status := w.Status()
		// const fields
		out := log.WithField("time", now).
			WithFields(DefaultLoggerFixedFields[:], []any{
				r.Host, r.Method, r.URL.Path, r.Proto,
				ctx.RealIP(), ctx.GetParam(eudore.ParamRoute),
				status, w.Size(),
				eudore.GetStringDuration(time.Since(now) / 1000),
			})

		rh, wh := r.Header, w.Header()
		for _, key := range params {
			pos := strings.IndexByte(key, ':')
			switch pos {
			case 5: // param:
				out = loggerValue(out, key[6:], ctx.GetParam(key[6:]))
			case 7: // request:
				out = loggerValue(out, strings.ToLower(key[8:]), rh.Get(key[8:]))
			case 8: // response:
				out = loggerValue(out, strings.ToLower(key[9:]), wh.Get(key[9:]))
			case 6: // cookie:
				out = loggerValue(out, key[7:], ctx.GetCookie(key[7:]))
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
// [eudore.LoggerDebug]-[eudore.LoggerFatal].
//
// The default processing function uses the uri parameter 'eudore_debug'
// to convert it into a log level.
//
//go:noinline
func NewLoggerLevelFunc(fn func(ctx eudore.Context) int) Middleware {
	if fn == nil {
		fn = func(ctx eudore.Context) int {
			level := ctx.GetQuery("eudore_debug")
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
				log = log.WithField("logger", true)
				log.SetLevel(level)
				ctx.SetValue(eudore.ContextKeyLogger, log)
			}
		}
	}
}
