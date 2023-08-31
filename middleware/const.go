package middleware

import (
	"errors"
	"net/http"
	"net/http/pprof"
	"time"
)

const (
	CompressBufferLength = 1024
	CompressStateUnknown = iota
	CompressStateBuffer
	CompressStateEnable
	CompressStateDisable
	CompressNameGzip     = "gzip"
	CompressNameBrotli   = "br"
	CompressNameDeflate  = "deflate"
	CompressNameIdentity = "identity"
	MimeValueJSON        = "value/json"
	MimeValueHTML        = "value/html"
	MimeValueText        = "value/text"
	QueryFormatJSON      = "json"
	QueryFormatHTML      = "html"
	QueryFormatText      = "text"
)

var (
	// DefaultBlackInvalidAddress 定义解析无效地址时使用的默认地址，127.0.0.2。
	DefaultBlackInvalidAddress uint64 = 2130706434
	DefaultCacheSaveTime              = time.Second * 10
	// DefaultComoressBrotliFunc 指定brotli压缩构造函数。
	//
	// import "github.com/andybalholm/brotli"
	//
	// middleware.DefaultComoressBrotliFunc = func() any {return brotli.NewWriter(io.Discard)} .
	DefaultComoressBrotliFunc  func() any
	DefaultComoressDisableMime = map[string]bool{
		"application/gzip":             true, // gz
		"application/zip":              true, // zip
		"application/x-compressed-tar": true, // tar.gz
		"application/x-7z-compressed":  true, // 7z
		"application/x-rar-compressed": true, // rar
		"image/gif":                    true, // gif
		"image/jpeg":                   true, // jpeg
		"image/png":                    true, // png
		"image/svg+xml":                true, // svg
		"image/webp":                   true, // webp
		"font/woff2":                   true, // woff2
	}

	DefaultPprofHandlers = map[string]http.Handler{
		"cmdline":      http.HandlerFunc(pprof.Cmdline),
		"profile":      http.HandlerFunc(pprof.Profile),
		"symbol":       http.HandlerFunc(pprof.Symbol),
		"trace":        http.HandlerFunc(pprof.Trace),
		"allocs":       pprof.Handler("allocs"),
		"block":        pprof.Handler("block"),
		"heap":         pprof.Handler("heap"),
		"mutex":        pprof.Handler("mutex"),
		"threadcreate": pprof.Handler("threadcreate"),
	}

	ErrRateReadWaitLong  = errors.New("if the github.com/eudore/eudore/middleware speed limit waiting time is too long, it will time out")
	ErrRateWriteWaitLong = errors.New("if the github.com/eudore/eudore/middleware speed limit waits for write time is too long, it will wait for timeout")
)
