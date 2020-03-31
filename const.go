package eudore

import (
	"errors"
	"net"
	"reflect"
	"time"
)

var (
	// AppContextKey 定义从context.Value中获取app实例对象的key，如果app支持的话。
	AppContextKey = struct{}{}
	// AppServerGrace 定义从context.Value中获取是否进行热重启的key。
	AppServerGrace = struct{}{}
	// DefaultBodyMaxMemory 默认Body解析占用内存。
	DefaultBodyMaxMemory int64 = 32 << 20 // 32 MB
	// DefaultConvertTags 定义默认转换使用的结构体tags。
	DefaultConvertTags = []string{"alias"}
	// DefaultConvertFormTags 定义bind form使用tags。
	DefaultConvertFormTags = []string{"form", "alias"}
	// DefaultConvertURLTags 定义bind url使用tags。
	DefaultConvertURLTags = []string{"url", "alias"}
	// LogLevelString 定义日志级别输出字符串。
	LogLevelString = [5]string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}
	// RouterAllMethod 定义全部的方法，影响Any方法的注册。
	RouterAllMethod = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
)

// 定义各种类型的反射类型。
var (
	typeBool      = reflect.TypeOf((*bool)(nil)).Elem()
	typeString    = reflect.TypeOf((*string)(nil)).Elem()
	typeError     = reflect.TypeOf((*error)(nil)).Elem()
	typeInterface = reflect.TypeOf((*interface{})(nil)).Elem()

	typeContext           = reflect.TypeOf((*Context)(nil)).Elem()
	typeController        = reflect.TypeOf((*Controller)(nil)).Elem()
	typeHandlerFunc       = reflect.TypeOf((*HandlerFunc)(nil)).Elem()
	typeValidateInterface = reflect.TypeOf((*validateInterface)(nil)).Elem()

	typeNetListener  = reflect.TypeOf((*net.Listener)(nil)).Elem()
	typeTimeDuration = reflect.TypeOf((*time.Duration)(nil)).Elem()
	typeTimeTime     = reflect.TypeOf((*time.Time)(nil)).Elem()
)

// 检测各类接口
var (
	_ Context    = (*ContextBase)(nil)
	_ Config     = (*ConfigMap)(nil)
	_ Config     = (*ConfigEudore)(nil)
	_ Logger     = (*LoggerInit)(nil)
	_ Logger     = (*LoggerStd)(nil)
	_ Server     = (*ServerStd)(nil)
	_ Server     = (*ServerFcgi)(nil)
	_ Server     = (*ServerGrace)(nil)
	_ Router     = (*RouterStd)(nil)
	_ RouterCore = (*RouterCoreRadix)(nil)
	_ RouterCore = (*RouterCoreFull)(nil)

	_ ResponseWriter  = (*ResponseWriterHTTP)(nil)
	_ Controller      = (*ControllerBase)(nil)
	_ Controller      = (*ControllerData)(nil)
	_ Controller      = (*ControllerSingleton)(nil)
	_ Controller      = (*ControllerView)(nil)
	_ ControllerPool  = (*controllerPoolSync)(nil)
	_ ControllerPool  = (*controllerPoolSingleton)(nil)
	_ HandlerExtender = (*handlerExtendBase)(nil)
	_ HandlerExtender = (*handlerExtendWarp)(nil)
	_ HandlerExtender = (*handlerExtendTree)(nil)
	_ Validater       = (*validaterBase)(nil)
	_ Params          = (*ParamsArray)(nil)
)

// 定义默认错误
var (
	// ErrApplicationStop 在app正常退出时返回。
	ErrApplicationStop = errors.New("stop application success")
	// ErrConverterInputDataNil 在Converter方法时，输出参数是空。
	ErrConverterInputDataNil = errors.New("Converter input value is nil")
	// ErrConverterInputDataNotPtr 在Converter方法时，输出参数是空。
	ErrConverterInputDataNotPtr = errors.New("Converter input value not is ptr")
	// ErrConverterTargetDataNil 在Converter方法时，目标参数是空。
	ErrConverterTargetDataNil = errors.New("Converter target data is nil")
	// ErrEudoreInitIgnore 在Eudore的Init函数返回该错误，忽略执行后续init函数。
	ErrEudoreInitIgnore = errors.New("eudore ignore the remaining init function")
	// ErrHandlerInvalidRange 在http使用range分开请求文件出现错误时返回。
	ErrHandlerInvalidRange = errors.New("invalid range: failed to overlap")
	// ErrHandlerProxyBackNotWriter HandlerProxy函数处理101时，后端连接不支持io.Writer。
	ErrHandlerProxyBackNotWriter = errors.New("HandlerProxy error: Back conn ResponseReader not suppert io.Writer, Need to implement the io.ReadWriteCloser interface")
	// ErrListenerNotImplementsServerListenFiler 在获取net.Listener的fd文件时，没有实现serverListenFiler接口，无法获得fd文件。
	ErrListenerNotImplementsServerListenFiler = errors.New("Listener is not implements ServerListenFiler")
	// ErrLoggerLevelUnmarshalText 日志级别解码错误，请检查输出的[]byte是否有效。
	ErrLoggerLevelUnmarshalText = errors.New("logger level UnmarshalText error")
	// ErrNewHandlerFuncParamNotFunc 调用NewHandlerFunc函数时参数错误，参数必须是一个函数对象。
	ErrNewHandlerFuncParamNotFunc = errors.New("NewHandlerFuncs input parameter must be a function")
	// ErrNotToServerListener newServerListens函数参数异常，无法解析并转换成serverListener类型。
	ErrNotToServerListener = errors.New("Parameters cannot be converted to serverListener type")
	// ErrRegisterControllerHandlerFuncParamNotFunc RegisterControllerHandlerFunc第一个参数必须是一个函数。
	ErrRegisterControllerHandlerFuncParamNotFunc = errors.New("The parameter type of RegisterControllerHandlerFunc must be a function")
	// ErrRegisterNewHandlerParamNotFunc 调用RegisterHandlerExtend函数时，参数必须是一个函数。
	ErrRegisterNewHandlerParamNotFunc = errors.New("The parameter type of RegisterNewHandler must be a function")
	// ErrResponseWriterHTTPNotHijacker ResponseWriterHTTP对象没有实现http.Hijacker接口。
	ErrResponseWriterHTTPNotHijacker = errors.New("http.Hijacker interface is not supported")
	// ErrResponseWriterTestNotSupportHijack ResponseWriterTest对象的Hijack不支持。
	ErrResponseWriterTestNotSupportHijack = errors.New("ResponseWriterTest no support hijack")
	// ErrServerStdStateException Server启动状态检查异常，需要启动时状态为ServerStateInit，该Server可能已经启动导致状态异常。
	ErrServerStdStateException = errors.New("ServerStd state not is ServerStateInit")
	// ErrServerNotAddListener Server没有添加net.Listner监听对象。
	ErrServerNotAddListener = errors.New("Server not add net.Listener")
	// ErrSeterNotSupportField Seter对象不支持设置当前属性。
	ErrSeterNotSupportField = errors.New("Converter seter not support set field")

	// ErrFormatBindDefaultNotSupportContentType BindDefault函数不支持当前的Content-Type Header。
	ErrFormatBindDefaultNotSupportContentType = "BindDefault not support content type header: %s"
	// ErrFormatControllerBaseParseFuncNotSupport ControllerBaseParseFunc函数解析的控制器不支持ControllerRoute接口，无法解析。
	ErrFormatControllerBaseParseFuncNotSupport = "%s not support ControllerBaseParseFunc, ControllerRoute interface is not implemented"
	// ErrFormatConverterCheckZeroUnknownType checkValueIsZero方法处理未定义的类型。
	ErrFormatConverterCheckZeroUnknownType = "reflect: call of reflect.Value.IsZero on %s Value"
	// ErrFormatConverterGetWithTags 在Get方法时，无法或到值，返回错误描述。
	ErrFormatConverterGetWithTags = "Get or GetWithTags func cannot get the value of the attribute '%s', error description: %v"
	// ErrFormatConverterNotGetValue 在Get方法时，getValue无法继续查找新的属性值。
	ErrFormatConverterNotGetValue = "The getValue method cannot continue to obtain a value, the current type is %s, and the remaining path is: %v"
	// ErrFormatConverterNotCanset 在Set方法时，结构体不支持该项属性。
	ErrFormatConverterNotCanset = "The attribute '%s' of structure %s is not set, please use public field"
	// ErrFormatConverterSetArrayIndexInvalid 在Set方法时，设置数组的索引的无效
	ErrFormatConverterSetArrayIndexInvalid = "the Set function obtained array index '%s' is invalid, array len is %d"
	// ErrFormatConverterSetStringUnknownType setWithString函数遇到未定义的反射类型
	ErrFormatConverterSetStringUnknownType = "setWithString unknown type %s"
	// ErrFormatConverterSetStructNotField 在Set时，结构体没有当前属性。
	ErrFormatConverterSetStructNotField = "Setting the structure has no attribute '%s', or this attribute is not exportable"
	// ErrFormatConverterSetTypeError 在Set时，类型异常，无法继续设置值。
	ErrFormatConverterSetTypeError = "The type of the set value is %s, which is not configurable, key: %v, val: %s"
	// ErrFormatConverterSetWithValue setWithValue函数中类型无法赋值。
	ErrFormatConverterSetWithValue = "The setWithValue method type %s cannot be assigned to type %s"
	// ErrFormatHandlerProxyConnHijack HandlerProxy函数处理101时，请求连接不支持hijack，返回当前错误。
	ErrFormatHandlerProxyConnHijack = "HandlerProxy error: Conn hijack error: %v"
	// ErrFormatNewContrllerExecFuncTypeNotFunc NewContrllerExecFunc函数的参数类型函数为注册，需要先使用RegisterControllerExecFunc注册函数类型。
	ErrFormatNewContrllerExecFuncTypeNotFunc = "The NewContrllerExecFunc parameter type is %s, which is an unregistered handler type"
	// ErrFormatRegisterHandlerExtendInputParamError RegisterHandlerExtend函数注册的函数参数错误。
	ErrFormatRegisterHandlerExtendInputParamError = "The '%s' input parameter is illegal and should be one"
	// ErrFormatRegisterHandlerExtendOutputParamError RegisterHandlerExtend函数注册的函数返回值错误。
	ErrFormatRegisterHandlerExtendOutputParamError = "The '%s' output parameter is illegal and should be a HandlerFunc object"
	// ErrFormatRouterStdAddController RouterStd控制器路由注入错误
	ErrFormatRouterStdAddController = "The RouterStd.AddController Inject error: %v"
	// ErrFormatRouterStdAddHandlerExtend RouterStd添加扩展错误
	ErrFormatRouterStdAddHandlerExtend = "The RouterStd.AddHandlerExtend path is '%s' RegisterHandlerExtend error: %v"
	// ErrFormatRouterStdRegisterHandlersMethodInvalid RouterStd.registerHandlers 的添加的是无效的，全部有效方法为RouterAllMethod。
	ErrFormatRouterStdRegisterHandlersMethodInvalid = "The RouterStd.registerHandlers arg method '%s' is invalid, complete method: '%s', add fullpath: '%s'"
	// ErrFormatRouterStdRegisterHandlersRecover RouterStd出现panic。
	ErrFormatRouterStdRegisterHandlersRecover = "The RouterStd.registerHandlers arg method is '%s' and path is '%s', recover error: %v"
	// ErrFormatRouterStdNewHandlerFuncsUnregisterType RouterStd添加处理对象或中间件的第n个参数类型未注册，需要先使用RegisterHandlerExtend或AddHandlerExtend注册该函数类型。
	ErrFormatRouterStdNewHandlerFuncsUnregisterType = "The RouterStd.newHandlerFuncs path is '%s', %dth handler parameter type is '%s', this is the unregistered handler type"
	// ErrFormatStartNewProcessError 在StartNewProcess函数fork启动新进程错误。
	ErrFormatStartNewProcessError = "StartNewProcess failed to forkexec error: %v"
	// ErrFormatUnknownTypeBody 在transbody函数解析参数成io.Reader时，参数类型是非法的。
	ErrFormatUnknownTypeBody = "unknown type used for body: %+v"
)

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
	// EnvEudoreGracefulAddrs 按顺序记录fork多端口fd对应的地址。
	EnvEudoreGracefulAddrs = "EnvEudoreGracefulAddrs"

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
