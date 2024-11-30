package middleware

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	"github.com/eudore/eudore"
)

const (
	CompressionBufferLength  = 1024
	CompressionNameGzip      = "gzip"
	CompressionNameBrotli    = "br"
	CompressionNameZstandard = "zstd"
	CompressionNameDeflate   = "deflate"
	CompressionNameIdentity  = "identity"
	QueryFormatJSON          = "json"
	QueryFormatHTML          = "html"
	QueryFormatText          = "text"

	headerContentType = eudore.HeaderContentType
)

var (
	DefaultCacheAllowAccept = map[string]struct{}{
		eudore.MimeAll:                 {},
		eudore.MimeApplicationJSON:     {},
		eudore.MimeApplicationProtobuf: {},
		eudore.MimeApplicationXML:      {},
	}
	// DefaultCompressionDisableMime global defines the Mime type that disables
	// compression. The data is in compressed format.
	DefaultCompressionDisableMime = map[string]struct{}{
		"application/gzip":             {}, // gz
		"application/zip":              {}, // zip
		"application/x-compressed-tar": {}, // tar.gz
		"application/x-7z-compressed":  {}, // 7z
		"application/x-rar-compressed": {}, // rar
		"image/gif":                    {}, // gif
		"image/jpeg":                   {}, // jpeg
		"image/png":                    {}, // png
		"image/svg+xml":                {}, // svg
		"image/webp":                   {}, // webp
		"font/woff2":                   {}, // woff2
	}
	// DefaultCompressionEncoder defines the default supported hybrid
	// compression methods.
	DefaultCompressionEncoder = map[string]func() any{
		CompressionNameGzip:    compressionWriterGzip,
		CompressionNameDeflate: compressionWriterFlate,
	}
	// DefaultCompressionOrder global  defines the priority order of available
	// compression,
	// compression methods outside the list have the lowest priority.
	DefaultCompressionOrder = []string{
		CompressionNameZstandard,
		CompressionNameBrotli,
		CompressionNameGzip,
		CompressionNameDeflate,
	}
	DefaultLoggerFixedFields = [...]string{
		"host", "method", "path", "proto", "realip", "route",
		"status", "bytes-out", "duration",
	}
	DefaultLoggerOptionalFields = [...]string{
		"remote-addr", "scheme", "querys", "byte-in",
	}
	DefaultPageAdmin          = adminStatic
	DefaultPageBasicAuth      = "401 Unauthorized"
	DefaultPageBodyLimit      = "413 Request Entity Too Large: body limit {{value}} bytes."
	DefaultPageBlack          = "403 Forbidden: your IP is blacklisted {{value}}."
	DefaultPageCircuitBreaker = "503 Service Unavailable: breaker triggered {{value}}."
	DefaultPageCORS           = ""
	DefaultPageCSRF           = "403 Forbidden: invalid CSRF token {{value}}."
	DefaultPageHealth         = "unhealthy: {{value}}"
	DefaultPageRate           = "429 Too Many Requests: rate limit exceeded {{value}}."
	DefaultPageReferer        = "403 Forbidden: invalid Referer header {{value}}."
	DefaultPageTimeout        = "503 Service Unavailable"
	// DefaultPProfHandlers global defines pprof route.
	DefaultPProfHandlers = map[string]http.Handler{
		"cmdline":      http.HandlerFunc(pprof.Cmdline),
		"profile":      http.HandlerFunc(pprof.Profile),
		"symbol":       http.HandlerFunc(pprof.Symbol),
		"trace":        http.HandlerFunc(pprof.Trace),
		"allocs":       pprof.Handler("allocs"),
		"block":        pprof.Handler("block"),
		"heap":         pprof.Handler("heap"),
		"mutex":        pprof.Handler("mutex"),
		"threadcreate": pprof.Handler("threadcreate"),
		"expvar":       expvar.Handler(),
	}
	DefaultRateRetryMin = 3
)
