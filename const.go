package eudore

// 定义eudore定义各种常量。
const (
	// Eudore environ

	// EnvEudoreIsNotify 表示使用使用了Notify组件。
	EnvEudoreIsNotify = "EUDORE_IS_NOTIFY"
	// EnvEudoreIsDaemon 用于表示是否fork来后台启动。
	EnvEudoreIsDaemon = "EUDORE_IS_DEAMON"
	// EnvEudoreDisablePidfile 用于Command组件不写入pidfile，Notify组件启动的子程序不写入pidfile。
	EnvEudoreDisablePidfile = "EUDORE_DISABLE_PIDFILE"
	// EnvEudoreDisableSignal 用于InitSignal不启用，Notify组件启动的子程序不注册信号。
	EnvEudoreDisableSignal = "EUDORE_DISABLE_SIGNAL"

	// Response statue

	StatusContinue           = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols = 101 // RFC 7231, 6.2.2
	StatusProcessing         = 102 // RFC 2518, 10.1

	StatusOK                   = 200 // RFC 7231, 6.3.1
	StatusCreated              = 201 // RFC 7231, 6.3.2
	StatusAccepted             = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo = 203 // RFC 7231, 6.3.4
	StatusNoContent            = 204 // RFC 7231, 6.3.5
	StatusResetContent         = 205 // RFC 7231, 6.3.6
	StatusPartialContent       = 206 // RFC 7233, 4.1
	StatusMultiStatus          = 207 // RFC 4918, 11.1
	StatusAlreadyReported      = 208 // RFC 5842, 7.1
	StatusIMUsed               = 226 // RFC 3229, 10.4.1

	StatusMultipleChoices  = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently = 301 // RFC 7231, 6.4.2
	StatusFound            = 302 // RFC 7231, 6.4.3
	StatusSeeOther         = 303 // RFC 7231, 6.4.4
	StatusNotModified      = 304 // RFC 7232, 4.1
	StatusUseProxy         = 305 // RFC 7231, 6.4.5

	StatusTemporaryRedirect = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect = 308 // RFC 7538, 3

	StatusBadRequest                   = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 = 401 // RFC 7235, 3.1
	StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
	StatusForbidden                    = 403 // RFC 7231, 6.5.3
	StatusNotFound                     = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
	StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
	StatusConflict                     = 409 // RFC 7231, 6.5.8
	StatusGone                         = 410 // RFC 7231, 6.5.9
	StatusLengthRequired               = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed           = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
	StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
	StatusTeapot                       = 418 // RFC 7168, 2.3.3
	StatusMisdirectedRequest           = 421 // RFC 7540, 9.1.2
	StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
	StatusLocked                       = 423 // RFC 4918, 11.3
	StatusFailedDependency             = 424 // RFC 4918, 11.4
	StatusTooEarly                     = 425 // RFC 8470, 5.2.
	StatusUpgradeRequired              = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         = 428 // RFC 6585, 3
	StatusTooManyRequests              = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // RFC 7725, 3

	StatusInternalServerError           = 500 // RFC 7231, 6.6.1
	StatusNotImplemented                = 501 // RFC 7231, 6.6.2
	StatusBadGateway                    = 502 // RFC 7231, 6.6.3
	StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
	StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
	StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
	StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           = 507 // RFC 4918, 11.5
	StatusLoopDetected                  = 508 // RFC 5842, 7.2
	StatusNotExtended                   = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6

	// Header

	HeaderAccept                          = "Accept"
	HeaderAcceptCharset                   = "Accept-Charset"
	HeaderAcceptEncoding                  = "Accept-Encoding"
	HeaderAcceptLanguage                  = "Accept-Language"
	HeaderAcceptRanges                    = "Accept-Ranges"
	HeaderAccessControlAllowCredentials   = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders       = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods       = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowOrigin        = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders      = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge             = "Access-Control-Max-Age"
	HeaderAccessControlRequestHeaders     = "Access-Control-Request-Headers"
	HeaderAccessControlRequestMethod      = "Access-Control-Request-Method"
	HeaderAge                             = "Age"
	HeaderAllow                           = "Allow"
	HeaderAltSvc                          = "Alt-Svc"
	HeaderAuthorization                   = "Authorization"
	HeaderCacheControl                    = "Cache-Control"
	HeaderClearSiteData                   = "Clear-Site-Data"
	HeaderConnection                      = "Connection"
	HeaderContentEncoding                 = "Content-Encoding"
	HeaderContentLanguage                 = "Content-Language"
	HeaderContentLength                   = "Content-Length"
	HeaderContentLocation                 = "Content-Location"
	HeaderContentRange                    = "Content-Range"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderContentType                     = "Content-Type"
	HeaderCookie                          = "Cookie"
	HeaderDate                            = "Date"
	HeaderETag                            = "ETag"
	HeaderEarlyData                       = "Early-Data"
	HeaderExpect                          = "Expect"
	HeaderExpectCT                        = "Expect-CT"
	HeaderExpires                         = "Expires"
	HeaderFeaturePolicy                   = "Feature-Policy"
	HeaderForwarded                       = "Forwarded"
	HeaderFrom                            = "From"
	HeaderHost                            = "Host"
	HeaderIfMatch                         = "If-Match"
	HeaderIfModifiedSince                 = "If-Modified-Since"
	HeaderIfNoneMatch                     = "If-None-Match"
	HeaderIfRange                         = "If-Range"
	HeaderIfUnmodifiedSince               = "If-Unmodified-Since"
	HeaderIndex                           = "Index"
	HeaderKeepAlive                       = "Keep-Alive"
	HeaderLastModified                    = "Last-Modified"
	HeaderLocation                        = "Location"
	HeaderOrigin                          = "Origin"
	HeaderPragma                          = "Pragma"
	HeaderProxyAuthenticate               = "Proxy-Authenticate"
	HeaderProxyAuthorization              = "Proxy-Authorization"
	HeaderPublicKeyPins                   = "Public-Key-Pins"
	HeaderPublicKeyPinsReportOnly         = "Public-Key-Pins-Report-Only"
	HeaderRange                           = "Range"
	HeaderReferer                         = "Referer"
	HeaderReferrerPolicy                  = "Referrer-Policy"
	HeaderRetryAfter                      = "Retry-After"
	HeaderSecWebSocketAccept              = "Sec-WebSocket-Accept"
	HeaderServer                          = "Server"
	HeaderServerTiming                    = "Server-Timing"
	HeaderSetCookie                       = "Set-Cookie"
	HeaderSourceMap                       = "SourceMap"
	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderTE                              = "TE"
	HeaderTimingAllowOrigin               = "Timing-Allow-Origin"
	HeaderTk                              = "Tk"
	HeaderTrailer                         = "Trailer"
	HeaderTransferEncoding                = "Transfer-Encoding"
	HeaderUpgradeInsecureRequests         = "Upgrade-Insecure-Requests"
	HeaderUserAgent                       = "User-Agent"
	HeaderVary                            = "Vary"
	HeaderVia                             = "Via"
	HeaderWWWAuthenticate                 = "WWW-Authenticate"
	HeaderWarning                         = "Warning"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXDNSPrefetchControl             = "X-DNS-Prefetch-Control"
	HeaderXForwardedFor                   = "X-Forwarded-For"
	HeaderXForwardedHost                  = "X-Forwarded-Host"
	HeaderXForwardedProto                 = "X-Forwarded-Proto"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXParentID                       = "X-Parent-ID"
	HeaderXRequestID                      = "X-Request-ID"

	// 默认http请求方法

	MethodAny     = "ANY"
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodHead    = "HEAD"
	MethodPatch   = "PATCH"
	MethodOptions = "OPTIONS"
	MethodConnect = "CONNECT"
	MethodTrace   = "TRACE"

	// Mime

	MimeCharsetUtf8                = "charset=utf-8"
	MimeText                       = "text/*"
	MimeTextPlain                  = "text/plain"
	MimeTextPlainCharsetUtf8       = MimeTextPlain + "; " + MimeCharsetUtf8
	MimeTextHTML                   = "text/html"
	MimeTextHTMLCharsetUtf8        = MimeTextHTML + "; " + MimeCharsetUtf8
	MimeTextCSS                    = "text/css"
	MimeTextCSSUtf8                = MimeTextCSS + "; " + MimeCharsetUtf8
	MimeTextJavascript             = "text/javascript"
	MimeTextJavascriptUtf8         = MimeTextJavascript + "; " + MimeCharsetUtf8
	MimeTextMarkdown               = "text/markdown"
	MimeTextMarkdownUtf8           = MimeTextMarkdown + "; " + MimeCharsetUtf8
	MimeTextXML                    = "text/xml"
	MimeTextXMLCharsetUtf8         = MimeTextXML + "; " + MimeCharsetUtf8
	MimeApplicationJSON            = "application/json"
	MimeApplicationJSONUtf8        = MimeApplicationJSON + "; " + MimeCharsetUtf8
	MimeApplicationXML             = "application/xml"
	MimeApplicationxmlCharsetUtf8  = MimeApplicationXML + "; " + MimeCharsetUtf8
	MimeApplicationForm            = "application/x-www-form-urlencoded"
	MimeApplicationFormCharsetUtf8 = MimeApplicationForm + "; " + MimeCharsetUtf8
	MimeMultipartForm              = "multipart/form-data"

	// Param

	ParamAction   = "action"
	ParamAllKeys  = "allkeys"
	ParamAllVals  = "allvals"
	ParamCaller   = "caller"
	ParamRAM      = "ram"
	ParamTemplate = "template"
	ParamRoute    = "route"
	ParamDeny     = "deny"
	ParamUID      = "UID"
	ParamUNAME    = "UNAME"
)
