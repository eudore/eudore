package middleware

import (
	"encoding/base64"
	"errors"
	"expvar"
	"net/http"
	"net/http/pprof"

	"github.com/eudore/eudore"
)

var (
	_ eudore.ResponseWriter = (*responseWriterCache)(nil)
	_ eudore.ResponseWriter = (*responseWriterCompress)(nil)
	_ eudore.ResponseWriter = (*responseWriterDump)(nil)
	_ eudore.ResponseWriter = (*responseWriteFlush)(nil)
	_ eudore.ResponseWriter = (*responseWriterRate)(nil)
	_ eudore.ResponseWriter = (*responseWriterTimeout)(nil)
	_ eudore.ResponseWriter = (*responseWriterTiming)(nil)
)

var (
	base64Encoding    = base64.RawURLEncoding.Strict()
	headerContentType = eudore.HeaderContentType
	valueBasicAuth    = "Basic "
	valueBearerAuth   = "Bearer "
	valueStar         = "*"
	valueBoolTrue     = true
	valueBoolFalse    = false
	valueStruct       = struct{}{}
	valueSchemeHTTP   = "http://"
	valueSchemeHTTPS  = "https://"
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
)

var (
	// DefaultCacheAllowAccept global defines the [HeaderAccept] allowed
	// by Cache to distinguish response formats.
	DefaultCacheAllowAccept = map[string]struct{}{
		eudore.MimeAll:                    {},
		eudore.MimeText:                   {},
		eudore.MimeTextPlain:              {},
		eudore.MimeTextHTML:               {},
		eudore.MimeApplicationJSON:        {},
		eudore.MimeApplicationProtobuf:    {},
		eudore.MimeApplicationXML:         {},
		eudore.MimeApplicationOctetStream: {},
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
	// DefaultCompressionOrder global defines the priority order of available
	// compression,
	// compression methods outside the list have the lowest priority.
	DefaultCompressionOrder = []string{
		CompressionNameZstandard,
		CompressionNameBrotli,
		CompressionNameGzip,
		CompressionNameDeflate,
	}
	// DefaultLoggerFixedFields global defines fixed access log fields.
	DefaultLoggerFixedFields = [...]string{
		"host", "method", "path", "proto", "realip", "route",
		"status", "bytes_out", "duration",
	}
	// DefaultLoggerOptionalFields global defines optional access log fields.
	DefaultLoggerOptionalFields = [...]string{
		"remote_addr", "scheme", "querys", "byte_in",
	}
	// DefaultLoggerLevelQueryName defines the query name for change the default log level.
	DefaultLoggerLevelQueryName = "eudore_debug"
	// All page content.

	DefaultPageAdmin          = adminStatic
	DefaultPageBasicAuth      = "401 Unauthorized"
	DefaultPageBearerAuth     = "401 Unauthorized: bearer error {{value}}"
	DefaultPageBodyLimit      = "413 Request Entity Too Large: body limit {{value}} bytes."
	DefaultPageBlack          = "403 Forbidden: your IP is blacklisted {{value}}."
	DefaultPageCircuitBreaker = "503 Service Unavailable: breaker triggered {{value}}."
	DefaultPageCORS           = ""
	DefaultPageCSRF           = "403 Forbidden: invalid CSRF token {{value}}."
	DefaultPageDigestAuth     = "401 Unauthorized: {{value}}"
	DefaultPageHealth         = "unhealthy: {{value}}"
	DefaultPageRate           = "429 Too Many Requests: rate limit exceeded {{value}}."
	DefaultPageReferer        = "403 Forbidden: invalid Referer header {{value}}."
	DefaultPageTimeout        = "503 Service Unavailable"
	// DefaultPolicyConditions global defines the conditions that policy parsing allows.
	//
	// Returns any object that implements the 'Match(ctx eudore.Context) bool' method.
	DefaultPolicyConditions = map[string]func() any{
		"and":      func() any { return &conditionAnd{} },
		"or":       func() any { return &conditionOr{} },
		"sourceip": func() any { return &conditionSourceIP{} },
		"date":     func() any { return &conditionDate{} },
		"time":     func() any { return &conditionTime{} },
		"method":   func() any { return &conditionMethod{} },
		"path":     func() any { return &conditionPath{} },
		"params":   func() any { return &conditionParams{} },
		"rate":     func() any { return &conditionRate{} },
		"version":  func() any { return &conditionVersion{} },
	}
	// DefaultPolicyGuestUser global defines the Guest user for Policy access.
	DefaultPolicyGuestUser = "<Guest User>"
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
	// DefaultRecoveryErrorFormat global defines the format of the recover data.
	DefaultRecoveryErrorFormat = "%v"
	DefaultRateRetryMin        = 3
	// DefaultUserAgentMapping defines the mapping to replace device codes
	// with normalized names during User-Agent analysis.
	DefaultUserAgentMapping = userAgentMapping
	// DefaultUserAgentRules defines the parsing and analysis rules for
	// [eudore.HeaderUserAgent] strings.
	//
	// The rules are stored as a string slice in the format [pattern, path, ...]:
	// pattern (User-Agent Name): Used to name the matched User-Agent type and its version.
	//   - $N: Replaced by the N-th matched value of '$' in the path (counting starts from 1).
	//   - $$: Replaced by the result matched by the next node.
	//
	// path (Matching Path): Used to match a segment of the User-Agent string.
	//   - ${char...}: Matches until any character within the braces is encountered, then returns the current matched value.
	//   - $char: Is the shorthand for '${char}char'.
	//   - $ at the end of the path: Is the shorthand for '${ }', indicating a match until a space or the end of the string.
	//   - Next Node Logic: If the character set matched by the last '$' in the path has a corresponding child node,
	//     that child node is used as the next node for subsequent matching.
	//
	// Note: If your User-Agent information cannot be parsed, please add the
	// necessary rules data yourself. Do not create an issue or pull request.
	DefaultUserAgentRules = userAgentRules

	ErrBearerTokenInvalid             = errors.New("bearer token is invalid")
	ErrBearerTokenNotValid            = "bearer token not valid before %v"
	ErrBearerTokenExpired             = "bearer token expired at %v" // #nosec G101
	ErrBearerSignatureInvalid         = errors.New("bearer signature is invalid")
	ErrCompressMissingEncoder         = "compress missing encoder function for compression '%s'"
	ErrCompressInvalidEncoder         = "compress invalid encoder function for compression '%s'"
	ErrPolicyConditionsUnmarshalError = "policy conditions unmarshal json %s error: %v"
	ErrPolicyConditionsParseError     = "policy conditions parse %s error: %v"
	ErrPolicyConditionParseError      = "policy conditions %s parse %s error: %v"
)
