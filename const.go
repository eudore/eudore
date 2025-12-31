package eudore

// const defines all global variables and constants

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

var (
	// defines reflect type.
	typeAny           = reflect.TypeOf((*any)(nil)).Elem()
	typeError         = reflect.TypeOf((*error)(nil)).Elem()
	typeContext       = reflect.TypeOf((*Context)(nil)).Elem()
	typeHandlerFunc   = reflect.TypeOf((*HandlerFunc)(nil)).Elem()
	typeTimeDuration  = reflect.TypeOf((*time.Duration)(nil)).Elem()
	typeTimeTime      = reflect.TypeOf((*time.Time)(nil)).Elem()
	typeFmtStringer   = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	typeJSONMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	typeTextMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	// check interface.
	_ Client          = (*clientStd)(nil)
	_ ClientHook      = (*clientHookCookie)(nil)
	_ ClientHook      = (*clientHookTimeout)(nil)
	_ ClientHook      = (*clientHookRedirect)(nil)
	_ ClientHook      = (*clientHookRetry)(nil)
	_ ClientHook      = (*clientHookLogger)(nil)
	_ ClientHook      = (*clientHookDigest)(nil)
	_ ClientBody      = (*bodyDecoder)(nil)
	_ ClientBody      = (*bodyFile)(nil)
	_ ClientBody      = (*bodyForm)(nil)
	_ Config          = (*configStd)(nil)
	_ Context         = (*contextBase)(nil)
	_ Controller      = (*ControllerAutoRoute)(nil)
	_ Controller      = (*ControllerAutoType[any])(nil)
	_ Controller      = (*controllerError)(nil)
	_ FuncCreator     = (*funcCreatorBase)(nil)
	_ FuncCreator     = (*funcCreatorExpr)(nil)
	_ HandlerExtender = (*handlerExtenderBase)(nil)
	_ HandlerExtender = (*handlerExtenderTree)(nil)
	_ HandlerExtender = (*handlerExtenderWrap)(nil)
	_ Logger          = (*loggerStd)(nil)
	_ LoggerHandler   = (*loggerFormatterJSON)(nil)
	_ LoggerHandler   = (*loggerFormatterText)(nil)
	_ LoggerHandler   = (*loggerHandlerInit)(nil)
	_ LoggerHandler   = (*loggerHookFilter)(nil)
	_ LoggerHandler   = (*loggerHookMeta)(nil)
	_ LoggerHandler   = (*loggerWriterFile)(nil)
	_ LoggerHandler   = (*loggerWriterRotate)(nil)
	_ LoggerHandler   = (*loggerWriterStdoutColor)(nil)
	_ LoggerHandler   = (*loggerWriterStdout)(nil)
	_ ResponseWriter  = (*responseWriterHTTP)(nil)
	_ Router          = (*routerStd)(nil)
	_ RouterCore      = (*routerCoreMux)(nil)
	_ RouterCore      = (*routerCoreHost)(nil)
	_ Server          = (*serverStd)(nil)
	_ Server          = (*serverFcgi)(nil)
)

// Define global constants.
const (
	// EnvEudoreDaemonListeners defines the address of listen fd.
	EnvEudoreDaemonListeners = "EUDORE_DAEMON_LISTENERS"
	// EnvEudoreDaemonParentPID defines the pid of the parent process that the
	// child process kills when restarting.
	EnvEudoreDaemonParentPID = "EUDORE_DAEMON_RESTART_ID"
	// EnvEudoreDaemonEnable defines whether to start daemon,
	// which will disable [Logger] stdout output.
	EnvEudoreDaemonEnable = "EUDORE_DAEMON_ENABLE"
	// EnvEudoreDaemonTimeout defines the timeout in seconds that the
	// daemon waits for the restart and stop commands to complete.
	EnvEudoreDaemonTimeout = "EUDORE_DAEMON_TIMEOUT"

	// default http method by rfc2616.

	MethodAny        = "ANY"
	MethodTest       = "TEST"
	MethodNotFound   = "NOTFOUND"
	MethodNotAllowed = "NOTALLOWED"
	MethodGet        = "GET"
	MethodPost       = "POST"
	MethodPut        = "PUT"
	MethodDelete     = "DELETE"
	MethodHead       = "HEAD"
	MethodPatch      = "PATCH"
	MethodOptions    = "OPTIONS"
	MethodConnect    = "CONNECT"
	MethodTrace      = "TRACE"

	// Status.

	StatusContinue                      = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols            = 101 // RFC 7231, 6.2.2
	StatusProcessing                    = 102 // RFC 2518, 10.1
	StatusEarlyHints                    = 103 // RFC 8297
	StatusOK                            = 200 // RFC 7231, 6.3.1
	StatusCreated                       = 201 // RFC 7231, 6.3.2
	StatusAccepted                      = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo          = 203 // RFC 7231, 6.3.4
	StatusNoContent                     = 204 // RFC 7231, 6.3.5
	StatusResetContent                  = 205 // RFC 7231, 6.3.6
	StatusPartialContent                = 206 // RFC 7233, 4.1
	StatusMultiStatus                   = 207 // RFC 4918, 11.1
	StatusAlreadyReported               = 208 // RFC 5842, 7.1
	StatusIMUsed                        = 226 // RFC 3229, 10.4.1
	StatusMultipleChoices               = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently              = 301 // RFC 7231, 6.4.2
	StatusFound                         = 302 // RFC 7231, 6.4.3
	StatusSeeOther                      = 303 // RFC 7231, 6.4.4
	StatusNotModified                   = 304 // RFC 7232, 4.1
	StatusUseProxy                      = 305 // RFC 7231, 6.4.5
	StatusTemporaryRedirect             = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect             = 308 // RFC 7538, 3
	StatusBadRequest                    = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                  = 401 // RFC 7235, 3.1
	StatusPaymentRequired               = 402 // RFC 7231, 6.5.2
	StatusForbidden                     = 403 // RFC 7231, 6.5.3
	StatusNotFound                      = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed              = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                 = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired             = 407 // RFC 7235, 3.2
	StatusRequestTimeout                = 408 // RFC 7231, 6.5.7
	StatusConflict                      = 409 // RFC 7231, 6.5.8
	StatusGone                          = 410 // RFC 7231, 6.5.9
	StatusLengthRequired                = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed            = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge         = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong             = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType          = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable  = 416 // RFC 7233, 4.4
	StatusExpectationFailed             = 417 // RFC 7231, 6.5.14
	StatusTeapot                        = 418 // RFC 7168, 2.3.3
	StatusMisdirectedRequest            = 421 // RFC 7540, 9.1.2
	StatusUnprocessableEntity           = 422 // RFC 4918, 11.2
	StatusLocked                        = 423 // RFC 4918, 11.3
	StatusFailedDependency              = 424 // RFC 4918, 11.4
	StatusTooEarly                      = 425 // RFC 8470, 5.2.
	StatusUpgradeRequired               = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired          = 428 // RFC 6585, 3
	StatusTooManyRequests               = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge   = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons    = 451 // RFC 7725, 3
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
	StatusClientClosedRequest           = 499 // nginx

	// Header.

	HeaderAccept                          = "Accept"
	HeaderAcceptCharset                   = "Accept-Charset"
	HeaderAcceptEncoding                  = "Accept-Encoding"
	HeaderAcceptLanguage                  = "Accept-Language"
	HeaderAcceptRanges                    = "Accept-Ranges"
	HeaderAcceptPost                      = "Accept-Post"
	HeaderAcceptPatch                     = "Accept-Patch"
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
	HeaderContentDisposition              = "Content-Disposition"
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
	HeaderETag                            = "Etag"
	HeaderEarlyData                       = "Early-Data"
	HeaderExpect                          = "Expect"
	HeaderExpectCT                        = "Expect-Ct"
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
	HeaderLastEventID                     = "Last-Event-Id"
	HeaderLastRetry                       = "Last-Retry"
	HeaderLastModified                    = "Last-Modified"
	HeaderLocation                        = "Location"
	HeaderOrigin                          = "Origin"
	HeaderPragma                          = "Pragma"
	HeaderProxyAuthenticate               = "Proxy-Authenticate"
	HeaderProxyAuthorization              = "Proxy-Authorization"
	HeaderProxyConnection                 = "Proxy-Connection"
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
	HeaderTE                              = "Te"
	HeaderTimingAllowOrigin               = "Timing-Allow-Origin"
	HeaderTk                              = "Tk"
	HeaderTrailer                         = "Trailer"
	HeaderTransferEncoding                = "Transfer-Encoding"
	HeaderUpgrade                         = "Upgrade"
	HeaderUpgradeInsecureRequests         = "Upgrade-Insecure-Requests"
	HeaderUserAgent                       = "User-Agent"
	HeaderVary                            = "Vary"
	HeaderVia                             = "Via"
	HeaderWWWAuthenticate                 = "Www-Authenticate"
	HeaderWarning                         = "Warning"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXCSRFToken                      = "X-Csrf-Token" // #nosec G101
	HeaderXDNSPrefetchControl             = "X-Dns-Prefetch-Control"
	HeaderXForwardedFor                   = "X-Forwarded-For"
	HeaderXForwardedHost                  = "X-Forwarded-Host"
	HeaderXForwardedProto                 = "X-Forwarded-Proto"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderXXSSProtection                  = "X-Xss-Protection"
	HeaderXRateLimit                      = "X-Rate-Limit"
	HeaderXRateReset                      = "X-Rate-Reset"
	HeaderXRateRemaining                  = "X-Rate-Remaining"
	HeaderXRealIP                         = "X-Real-Ip"
	HeaderXRequestID                      = "X-Request-Id"
	HeaderXTraceID                        = "X-Trace-Id"
	HeaderXEudoreRoute                    = "X-Eudore-Route"

	HeaderValueChunked   = "chunked"
	HeaderValueClose     = "close"
	HeaderValueKeepAlive = "keep-alive"
	HeaderValueNoCache   = "no-cache"
	HeaderValueUpgrade   = "Upgrade"

	// Mime.

	MimeAll                        = "*/*"
	MimeText                       = "text/*"
	MimeTextPlain                  = "text/plain"
	MimeTextMarkdown               = "text/markdown"
	MimeTextJavascript             = "text/javascript"
	MimeTextHTML                   = "text/html"
	MimeTextCSS                    = "text/css"
	MimeTextXML                    = "text/xml"
	MimeTextEventStream            = "text/event-stream"
	MimeApplicationYAML            = "application/yaml"
	MimeApplicationXML             = "application/xml"
	MimeApplicationProtobuf        = "application/protobuf"
	MimeApplicationJSON            = "application/json"
	MimeApplicationForm            = "application/x-www-form-urlencoded"
	MimeApplicationOctetStream     = "application/octet-stream"
	MimeMultipartForm              = "multipart/form-data"
	MimeMultipartMixed             = "multipart/mixed"
	MimeCharsetUtf8                = "charset=utf-8"
	MimeTextPlainCharsetUtf8       = MimeTextPlain + "; " + MimeCharsetUtf8
	MimeTextMarkdownCharsetUtf8    = MimeTextMarkdown + "; " + MimeCharsetUtf8
	MimeTextJavascriptCharsetUtf8  = MimeTextJavascript + "; " + MimeCharsetUtf8
	MimeTextHTMLCharsetUtf8        = MimeTextHTML + "; " + MimeCharsetUtf8
	MimeTextCSSCharsetUtf8         = MimeTextCSS + "; " + MimeCharsetUtf8
	MimeTextXMLCharsetUtf8         = MimeTextXML + "; " + MimeCharsetUtf8
	MimeApplicationYAMLCharsetUtf8 = MimeApplicationYAML + "; " + MimeCharsetUtf8
	MimeApplicationXMLCharsetUtf8  = MimeApplicationXML + "; " + MimeCharsetUtf8
	MimeApplicationJSONCharsetUtf8 = MimeApplicationJSON + "; " + MimeCharsetUtf8
	MimeApplicationFormCharsetUtf8 = MimeApplicationForm + "; " + MimeCharsetUtf8

	// Router Param.

	ParamAction          = "Action"
	ParamPolicy          = "Policy"
	ParamResource        = "Resource"
	ParamUserid          = "Userid"
	ParamUsername        = "Username"
	ParamAllow           = "allow"           // HandlerRouter405
	ParamAutoIndex       = "autoindex"       // NewHandlerFileSystem
	ParamBrowser         = "browser"         // middlewae.NewUaserAgentFunc
	ParamControllerGroup = "controllergroup" // ControllerInjectAutoRoute
	ParamLoggerKind      = "loggerkind"      // Router.Group
	ParamRouteHost       = "route-host"      // NewRouterCoreHost
	ParamRoute           = "route"           // NewRouter
	ParamTemplate        = "template"        // NewHandlerDataRenderTemplates

	// Logger Field.

	FieldCaller     = "caller"
	FieldDepth      = "depth"
	FieldError      = "error"
	FieldFile       = "file"
	FieldFunc       = "func"
	FieldLogger     = "logger"
	FieldStack      = "stack"
	FieldTime       = "time"
	FieldXRequestID = "x_request_id"
	FieldXTraceID   = "x_trace_id"
)

var templateEmbedIndex = `<!DOCTYPE html><html>
<head>
	<meta charset="utf-8">
	<meta name="referrer" content="always">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="color-scheme" content="light dark">
	<meta name="google" value="notranslate">
	<title id="title">Index of {{.Path}}</title>
</head>
<body bgcolor="white" from="chrome file">
<h1>Index of {{.Path}}</h1>
{{- if ne .Path "/"}}<div id="dir-link" style="display: block;"><a class="icon up" href="../">[parent directory]</a></div>{{end}}
<table>
	<thead><tr><th>Name</th><th>Size</th><th>Date Modified</th></tr></thead>
	<tbody>
		{{- range $index, $file := .Files}}{{if $file.IsDir}}
		<tr><td data-value="{{$file.Name}}"><a class="icon dir" href="{{$file.Name}}/">{{$file.Name}}/</a></td><td data-value="0" class="column"></td><td data-value="{{$file.UnixTime}}" class="column">{{$file.ModTime}}</td></tr>
		{{- else }}
		<tr><td data-value="{{$file.Name}}"><a class="icon file" draggable="true" href="{{$file.Name}}">{{$file.Name}}</a></td><td data-value="{{$file.Size}}" class="column">{{$file.SizeFormat}}</td><td data-value="{{$file.UnixTime}}" class="column">{{$file.ModTime}}</td></tr>
		{{- end }}{{end}}
	</tbody>
</table><style>
h1 {margin-bottom: 10px; padding-bottom: 10px;color: #000; border-bottom: 1px solid #c0c0c0; white-space: nowrap;}
table {border-collapse: collapse;color: #000;}
th {cursor: pointer;}
td.column {padding-inline-start: 2em; text-align: end; white-space: nowrap;}
a.icon {padding-inline-start: 1.5em; text-decoration: none; user-select: auto; color: #00e;}
a.icon:hover {text-decoration: underline; }
a.file {background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAIAAACQkWg2AAAABnRSTlMAAAAAAABupgeRAAABEElEQVR42nRRx3HDMBC846AHZ7sP54BmWAyrsP588qnwlhqw/k4v5ZwWxM1hzmGRgV1cYqrRarXoH2w2m6qqiqKIR6cPtzc3xMSML2Te7XZZlnW7Pe/91/dX47WRBHuA9oyGmRknzGDjab1ePzw8bLfb6WRalmW4ip9FDVpYSWZgOp12Oh3nXJ7nxoJSGEciteP9y+fH52q1euv38WosqA6T2gGOT44vry7BEQtJkMAMMpa6JagAMcUfWYa4hkkzAc7fFlSjwqCoOUYAF5RjHZPVCFBOtSBGfgUDji3c3jpibeEMQhIMh8NwshqyRsBJgvF4jMs/YlVR5KhgNpuBLzk0OcUiR3CMhcPaOzsZiAAA/AjmaB3WZIkAAAAASUVORK5CYII=") left top no-repeat; }
a.dir {background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAABt0lEQVR42oxStZoWQRCs2cXdHTLcHZ6EjAwnQWIkJyQlRt4Cd3d3d1n5d7q7ju1zv/q+mh6taQsk8fn29kPDRo87SDMQcNAUJgIQkBjdAoRKdXjm2mOH0AqS+PlkP8sfp0h93iu/PDji9s2FzSSJVg5ykZqWgfGRr9rAAAQiDFoB1OfyESZEB7iAI0lHwLREQBcQQKqo8p+gNUCguwCNAAUQAcFOb0NNGjT+BbUC2YsHZpWLhC6/m0chqIoM1LKbQIIBwlTQE1xAo9QDGDPYf6rkTpPc92gCUYVJAZjhyZltJ95f3zuvLYRGWWCUNkDL2333McBh4kaLlxg+aTmyL7c2xTjkN4Bt7oE3DBP/3SRz65R/bkmBRPGzcRNHYuzMjaj+fdnaFoJUEdTSXfaHbe7XNnMPyqryPcmfY+zURaAB7SHk9cXSH4fQ5rojgCAVIuqCNWgRhLYLhJB4k3iZfIPtnQiCpjAzeBIRXMA6emAqoEbQSoDdGxFUrxS1AYcpaNbBgyQBGJEOnYOeENKR/iAd1npusI4C75/c3539+nbUjOgZV5CkAU27df40lH+agUdIuA/EAgDmZnwZlhDc0wAAAABJRU5ErkJggg==") left top no-repeat; }
a.up {background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAACM0lEQVR42myTA+w1RxRHz+zftmrbdlTbtq04qRGrCmvbDWp9tq3a7tPcub8mj9XZ3eHOGQdJAHw77/LbZuvnWy+c/CIAd+91CMf3bo+bgcBiBAGIZKXb19/zodsAkFT+3px+ssYfyHTQW5tr05dCOf3xN49KaVX9+2zy1dX4XMk+5JflN5MBPL30oVsvnvEyp+18Nt3ZAErQMSFOfelCFvw0HcUloDayljZkX+MmamTAMTe+d+ltZ+1wEaRAX/MAnkJdcujzZyErIiVSzCEvIiq4O83AG7LAkwsfIgAnbncag82jfPPdd9RQyhPkpNJvKJWQBKlYFmQA315n4YPNjwMAZYy0TgAweedLmLzTJSTLIxkWDaVCVfAbbiKjytgmm+EGpMBYW0WwwbZ7lL8anox/UxekaOW544HO0ANAshxuORT/RG5YSrjlwZ3lM955tlQqbtVMlWIhjwzkAVFB8Q9EAAA3AFJ+DR3DO/Pnd3NPi7H117rAzWjpEs8vfIqsGZpaweOfEAAFJKuM0v6kf2iC5pZ9+fmLSZfWBVaKfLLNOXj6lYY0V2lfyVCIsVzmcRV9Y0fx02eTaEwhl2PDrXcjFdYRAohQmS8QEFLCLKGYA0AeEakhCCFDXqxsE0AQACgAQp5w96o0lAXuNASeDKWIvADiHwigfBINpWKtAXJvCEKWgSJNbRvxf4SmrnKDpvZavePu1K/zu/due1X/6Nj90MBd/J2Cic7WjBp/jUdIuA8AUtd65M+PzXIAAAAASUVORK5CYII=") left top no-repeat; }
#dir-link {margin-bottom: 10px; padding-bottom: 10px; }
#upload {display: none}
</style><script>"use strict";
function sortTable(column) {
	let thead = document.querySelector("thead>tr"); let tbody = document.querySelector("tbody");
	let oldOrder = parseInt(thead.cells[column].dataset.order || '1', 10); let newOrder = - oldOrder;
	let rows = tbody.rows; let list = []; thead.cells[column].dataset.order = newOrder;
	for (let i = 0; i < rows.length; i++) {list.push(rows[i]); }
	list.sort(function(row1, row2) {
		let a = row1.cells[column].dataset.value; let b = row2.cells[column].dataset.value;
		if (column>0) {a = parseInt(a, 10); b = parseInt(b, 10); return a > b ? newOrder : a < b ? oldOrder : 0; }
		if (a > b) return newOrder; if (a < b) return oldOrder; return 0;
	});
	for (let i = 0; i < list.length; i++) {tbody.appendChild(list[i]); }
}
document.querySelectorAll("thead th").forEach((e,i)=>{e.onclick=()=>sortTable(i)});
</script></body></html>`
