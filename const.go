package eudore

// const定义全部全局变量和常量

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"time"
)

var (
	// ContextKeyApp 定义获取app的Key。
	ContextKeyApp = NewContextKey("app")
	// ContextKeyLogger 定义获取logger的Key。
	ContextKeyLogger = NewContextKey("logger")
	// ContextKeyConfig 定义获取config的Key。
	ContextKeyConfig = NewContextKey("config")
	// ContextKeyDatabase 定义获取database的Key。
	ContextKeyDatabase = NewContextKey("database")
	// ContextKeyClient 定义获取client的Key。
	ContextKeyClient = NewContextKey("client")
	// ContextKeyClientTrace 定义获取client-trace的Key。
	ContextKeyClientTrace = NewContextKey("client-trace")
	// ContextKeyServer 定义获取server的Key。
	ContextKeyServer = NewContextKey("server")
	// ContextKeyRouter 定义获取router的Key。
	ContextKeyRouter = NewContextKey("router")
	// ContextKeyContextPool 定义获取context-pool的Key。
	ContextKeyContextPool = NewContextKey("context-pool")
	// ContextKeyError 定义获取error的Key。
	ContextKeyError = NewContextKey("error")
	// ContextKeyBind 定义获取bind的Key。
	ContextKeyBind = NewContextKey("bind")
	// ContextKeyValidate 定义获取validate的Key。
	ContextKeyValidate = NewContextKey("validate")
	// ContextKeyFilte 定义获取filte的Key。
	ContextKeyFilte = NewContextKey("filte")
	// ContextKeyRender 定义获取render的Key。
	ContextKeyRender = NewContextKey("render")
	// ContextKeyFuncCreator 定义获取func-creator的Key。
	ContextKeyFuncCreator = NewContextKey("func-creator")
	// ContextKeyTemplate 定义获取templdate的Key。
	ContextKeyTemplate = NewContextKey("templdate")
	// ContextKeyTrace 定义获取trace的Key。
	ContextKeyTrace = NewContextKey("trace")

	// DefaultBindFormTags 定义bind form使用tags。
	DefaultBindFormTags = []string{"form", "alias"}
	// DefaultBindHeaderTags 定义bind header使用tags。
	DefaultBindHeaderTags = []string{"header", "alias"}
	// DefaultBindURLTags 定义bind url使用tags。
	DefaultBindURLTags = []string{"url", "alias"}
	// DefaultClientBodyContextType 定义NewClientBody默认使用的内容类型。
	DefaultClientBodyContextType = MimeApplicationJSONCharsetUtf8
	// DefaultClientHost 定义clientStd默认使用的Host。
	DefaultClientHost = "localhost:80"
	// DefaultClientParseErrStar 定义NewClientParseErr解析err的状态码范围
	DefaultClientParseErrStar = 500
	// DefaultClientParseErrEnd 定义NewClientParseErr解析err的状态码范围
	DefaultClientParseErrEnd = 500
	// DefaultClientInternalHost 定义clientStd使用内部连接的Host。
	DefaultClientInternalHost = "127.0.0.10:80"
	// DefaultConfigAllParseFunc 定义ConfigMap和ConfigEudore默认使用的解析函数。
	DefaultConfigAllParseFunc = []ConfigParseFunc{NewConfigParseJSON("config"), NewConfigParseArgs(nil),
		NewConfigParseEnvs("ENV_"), NewConfigParseWorkdir("workdir"), NewConfigParseHelp("help")}
	// DefaultConfigGetSetTags 定义ConfigStd默认使用GetSet的tag。
	DefaultConfigGetSetTags = []string{"alias"}
	// DefaultContextMaxHandler 定义请求上下文handler数量上限，需要小于该值。
	DefaultContextMaxHandler = 0xff
	// DefaultContextFormMaxMemory 默认解析From body使用内存。
	DefaultContextFormMaxMemory int64 = 32 << 20 // 32 MB
	// DefaultEmbedCacheControl 定义默认NewHandlerEmbedFunc使用的Cache-Control缓存策略
	DefaultEmbedCacheControl = "no-cache"
	// DefaultEmbedTime 设置http返回embed文件的最后修改时间，默认为服务启动时间。
	// 如果服务存在多副本部署，通过设置相同的值保持多副本间的时间版本一致。
	DefaultEmbedTime = time.Now()
	// DefaultFuncCreator 定义全局默认FuncCreator,RouetrCoreStd默认使用。
	DefaultFuncCreator = NewFuncCreator()
	// DefaultGodocServer 定义应用默认使用的godoc服务器域名。
	DefaultGodocServer = "https://golang.org"
	// DefaultHandlerExtend 为默认的函数扩展处理者，是RouterStd使用的最顶级的函数扩展处理者。
	DefaultHandlerExtend = NewHandlerExtendBase()
	// DefaultHandlerExtendAllowType 定义handlerExtendBase允许使用的参数类型。
	DefaultHandlerExtendAllowType = map[reflect.Kind]struct{}{reflect.Func: {}, reflect.Interface: {},
		reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {}, reflect.Struct: {}}
	// DefaultLoggerDepth 定义GetPanicStack函数默认显示栈最大层数。
	DefaultLoggerDepth = 64
	// DefaultLoggerLevelString 定义日志级别输出字符串。
	DefaultLoggerLevelString = [5]string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}
	// DefaultLoggerNull 定义空日志输出器。
	DefaultLoggerNull = NewLoggerNull()
	// DefaultLoggerSyncDuration 定义LoggerStd默认Sync写入日志间隔时间，在Mount时使用。
	DefaultLoggerSyncDuration = time.Millisecond * 80
	// DefaultLoggerTimeFormat 定义默认日志输出和contextBase.WriteError的时间格式
	DefaultLoggerTimeFormat = "2006-01-02 15:04:05"
	// DefaultRenderFunc 定义默认使用的Render函数。
	DefaultRenderFunc = RenderJSON
	// DefaultRenderHTMLTemplate 定义RenderHTML的默认通用模板。
	DefaultRenderHTMLTemplate *template.Template
	// DefaultRouterAllMethod 定义路由器允许注册的全部方法，注册其他方法别忽略,前六种方法始终存在。
	DefaultRouterAllMethod = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions, MethodConnect, MethodTrace}
	// DefaultRouterAnyMethod 定义Any方法的注册使用的方法。
	DefaultRouterAnyMethod = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch}
	// DefaultServerShutdownWait 定义Server优雅退出等待时间。
	DefaultServerShutdownWait = 30 * time.Second
	// DefaultTraceServer 定义应用默认使用的jaeger链路追踪服务器域名。
	DefaultTraceServer = ""
	// DefaultValidateTag 定义NewValidateField获取校验规则的结构体tag。
	DefaultValidateTag = "validate"
	// defaultRouterAnyMethod 定义routerStd默认存储的6种方法处理对象。
	defaultRouterAnyMethod = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch}

	// DefaultGetSetTags 定义Get/Set函数使用的默认tag。
	DefaultGetSetTags = []string{"alias"}
	// DefaultConvertTags 定义默认转换使用的结构体tags。
	DefaultConvertTags = []string{"alias"}

	// 定义HandlerData返回error使用的Status和Code

	// StatucBindFail 定义Bind返回错误的状态码。
	StatucBindFail = 0
	// StatucValidateFail 定义Validate返回错误的状态码。
	StatucValidateFail = 0
	// StatucFilteFail 定义Filte返回错误的状态码。
	StatucFilteFail = 0
	// StatucRenderFail 定义Render返回错误的状态码。
	StatucRenderFail = 0
	// CodeBindFail 定义Bind返回错误的Code。
	CodeBindFail = 0
	// CodeValidateFail 定义Validate返回错误的Code。
	CodeValidateFail = 0
	// CodeFilteFail 定义Filte返回错误的Code。
	CodeFilteFail = 0
	// CodeRenderFail 定义Render返回错误的Code。
	CodeRenderFail = 0
)

// 定义各种类型的反射类型。
var (
	typeBool      = reflect.TypeOf((*bool)(nil)).Elem()
	typeBytes     = reflect.TypeOf((*[]byte)(nil)).Elem()
	typeError     = reflect.TypeOf((*error)(nil)).Elem()
	typeInterface = reflect.TypeOf((*interface{})(nil)).Elem()
	typeString    = reflect.TypeOf((*string)(nil)).Elem()

	typeContext        = reflect.TypeOf((*Context)(nil)).Elem()
	typeController     = reflect.TypeOf((*Controller)(nil)).Elem()
	typeControllerName = reflect.TypeOf((*controllerName)(nil)).Elem()
	typeHandlerFunc    = reflect.TypeOf((*HandlerFunc)(nil)).Elem()
	typeStringer       = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	typeTimeTime       = reflect.TypeOf((*time.Time)(nil)).Elem()
)

// 检测各类接口
var (
	_ Logger          = (*LoggerStd)(nil)
	_ LoggerStdData   = (*loggerStdDataJSON)(nil)
	_ LoggerStdData   = (*loggerStdDataInit)(nil)
	_ Config          = (*configStd)(nil)
	_ Client          = (*clientStd)(nil)
	_ Server          = (*serverStd)(nil)
	_ Server          = (*serverFcgi)(nil)
	_ Router          = (*RouterStd)(nil)
	_ RouterCore      = (*routerCoreStd)(nil)
	_ RouterCore      = (*routerCoreDebug)(nil)
	_ RouterCore      = (*routerCoreHost)(nil)
	_ RouterCore      = (*routerCoreLock)(nil)
	_ Context         = (*contextBase)(nil)
	_ ResponseWriter  = (*responseWriterHTTP)(nil)
	_ Controller      = (*ControllerAutoRoute)(nil)
	_ Controller      = (*controllerError)(nil)
	_ HandlerExtender = (*handlerExtendBase)(nil)
	_ HandlerExtender = (*handlerExtendWarp)(nil)
	_ HandlerExtender = (*handlerExtendTree)(nil)
	_ FuncCreator     = (*funcCreator)(nil)
)

// 定义默认错误
var (
	// ErrConverterInputDataNil 在Converter方法时，输出参数是空。
	ErrConverterInputDataNil = errors.New("Converter input value is nil")
	// ErrConverterInputDataNotPtr 在Converter方法时，输出参数是空。
	ErrConverterInputDataNotPtr = errors.New("Converter input value not is ptr")
	// ErrConverterTargetDataNil 在Converter方法时，目标参数是空。
	ErrConverterTargetDataNil = errors.New("Converter target data is nil")
	// ErrLoggerLevelUnmarshalText 日志级别解码错误，请检查输出的[]byte是否有效。
	ErrLoggerLevelUnmarshalText = errors.New("logger level UnmarshalText error")
	ErrRenderHandlerSkip        = errors.New("render hander skip")
	// ErrRegisterNewHandlerParamNotFunc 调用RegisterHandlerExtend函数时，参数必须是一个函数。
	ErrRegisterNewHandlerParamNotFunc = errors.New("The parameter type of RegisterNewHandler must be a function")
	// ErrResponseWriterHTTPNotHijacker ResponseWriterHTTP对象没有实现http.Hijacker接口。
	ErrResponseWriterHTTPNotHijacker = errors.New("http.Hijacker interface is not supported")
	// ErrSeterNotSupportField Seter对象不支持设置当前属性。
	ErrSeterNotSupportField = errors.New("Converter seter not support set field")
	// ErrMiddlewareRequestEntityTooLarge middleware/BodyLimit 分段请求body读取时限制长队返回错误。
	ErrMiddlewareRequestEntityTooLarge = errors.New("Request Entity Too Large")

	// ErrFormatBindDefaultNotSupportContentType BindDefault函数不支持当前的Content-Type Header。
	ErrFormatBindDefaultNotSupportContentType = "BindDefault not support content type header: %s"
	// ErrFormatConverterGet 在Get方法路径查找返回错误。
	ErrFormatConverterGet = "Get path '%s' error: %s"
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
	// ErrFormatRegisterHandlerExtendInputParamError RegisterHandlerExtend函数注册的函数参数错误。
	ErrFormatRegisterHandlerExtendInputParamError = "The '%s' input parameter is illegal and should be one func/interface/ptr/struct"
	// ErrFormatRegisterHandlerExtendOutputParamError RegisterHandlerExtend函数注册的函数返回值错误。
	ErrFormatRegisterHandlerExtendOutputParamError = "The '%s' output parameter is illegal and should be a HandlerFunc object"
	// ErrFormatRouterStdAddController RouterStd控制器路由注入错误
	ErrFormatRouterStdAddController = "The RouterStd.AddController Inject %s error: %v"
	// ErrFormatRouterStdAddHandlerExtend RouterStd添加扩展处理函数错误
	ErrFormatRouterStdAddHandlerExtend = "The RouterStd.AddHandlerExtend path is '%s' RegisterHandlerExtend error: %v"
	// ErrFormatRouterStdRegisterHandlersMethodInvalid RouterStd.registerHandlers 的添加的是无效的，全部有效方法为RouterAnyMethod。
	ErrFormatRouterStdRegisterHandlersMethodInvalid = "The RouterStd.registerHandlers arg method '%s' is invalid, complete method: '%s', add fullpath: '%s'"
	// ErrFormatRouterStdRegisterHandlersRecover RouterStd注册路由时恢复panic。
	ErrFormatRouterStdRegisterHandlersRecover = "The RouterStd.registerHandlers arg method is '%s' and path is '%s', recover error: %v"
	// ErrFormatRouterStdNewHandlerFuncsUnregisterType RouterStd添加处理对象或中间件的第n个参数类型未注册，需要先使用RegisterHandlerExtend或AddHandlerExtend注册该函数类型。
	ErrFormatRouterStdNewHandlerFuncsUnregisterType = "The RouterStd.newHandlerFuncs path is '%s', %dth handler parameter type is '%s', this is the unregistered handler type"
	// ErrFormatProtobufDecodeNilInteface 定义protobuf解码到空接口
	ErrFormatProtobufDecodeNilInteface    = "protobuf decode %s interface %s is nil"
	ErrFormatProtobufDecodeInvalidFlag    = "protobuf decode %s invalid flag %d"
	ErrFormatProtobufDecodeInvalidKind    = "protobuf decode %s invalid kind %s"
	ErrFormatProtobufDecodeReadError      = "protobuf decode %s read %s error: %w"
	ErrFormatProtobufDecodeReadInvalid    = "protobuf decode %s read length %d invalid has data %d"
	ErrFormatProtobufDecodeMessageNotRead = "protobuf decode message has %d not read"
	ErrFormatProtobufTypeMustSturct       = "protobuf encdoe/decode kind must struct, current type %s"
	// ErrFormatMiddlewareRequestEntityTooLargeSzie BodyLimit请求长度超过限制。
	ErrFormatMiddlewareRequestEntityTooLargeSzie = "Request Entity Too Large, limit body size %d"
	// ErrFormarRouterStdLoadInvalidFunc RouterStd无法加载路径对应的校验函数。
	ErrFormarRouterStdLoadInvalidFunc = "loadCheckFunc path is invalid, load path '%s' error: %v "
	// ErrFormatParseValidateFieldError Validate解析结构体规则错误。
	ErrFormatParseValidateFieldError = "validateField %s.%s parse field %s create rule %s error: %s"
	// ErrFormatFuncCreatorRegisterInvalidType fc注册函数类似是无效的。
	ErrFormatFuncCreatorRegisterInvalidType = "Register func %s type is %T, must 'func(T) bool' or 'func(string) (func(T) bool, error)'"
	// ErrFormatFuncCreatorNotFunc 无法创建对应的校验函数。
	ErrFormatFuncCreatorNotFunc = "not found or create func %s"
)

// 定义eudore定义各种常量。
const (
	// Eudore environ

	// EnvEudoreIsDaemon 用于表示是否fork后台启动。
	EnvEudoreIsDaemon = "EUDORE_IS_DEAMON"
	// EnvEudoreIsNotify 表示使用使用了Notify组件。
	EnvEudoreIsNotify = "EUDORE_IS_NOTIFY"
	// EnvEudoreDisablePidfile 用于Command组件不写入pidfile，Notify组件启动的子程序不写入pidfile。
	EnvEudoreDisablePidfile = "EUDORE_DISABLE_PIDFILE"

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
	HeaderXCSRFToken                      = "X-Csrf-Token"
	HeaderXDNSPrefetchControl             = "X-Dns-Prefetch-Control"
	HeaderXForwardedFor                   = "X-Forwarded-For"
	HeaderXForwardedHost                  = "X-Forwarded-Host"
	HeaderXForwardedProto                 = "X-Forwarded-Proto"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderXXSSProtection                  = "X-Xss-Protection"
	HeaderXRealIP                         = "X-Real-Ip"
	HeaderXRequestID                      = "X-Request-Id"
	HeaderXTraceID                        = "X-Trace-Id"
	HeaderXEudoreAdmin                    = "X-Eudore-Admin"
	HeaderXEudoreRoute                    = "X-Eudore-Route"

	// default http method by rfc2616

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
	MimeMultipartForm              = "multipart/form-data"
	MimeText                       = "text/*"
	MimeTextPlain                  = "text/plain"
	MimeTextPlainCharsetUtf8       = MimeTextPlain + "; " + MimeCharsetUtf8
	MimeTextMarkdown               = "text/markdown"
	MimeTextMarkdownCharsetUtf8    = MimeTextMarkdown + "; " + MimeCharsetUtf8
	MimeTextJavascript             = "text/javascript"
	MimeTextJavascriptCharsetUtf8  = MimeTextJavascript + "; " + MimeCharsetUtf8
	MimeTextHTML                   = "text/html"
	MimeTextHTMLCharsetUtf8        = MimeTextHTML + "; " + MimeCharsetUtf8
	MimeTextCSS                    = "text/css"
	MimeTextCSSCharsetUtf8         = MimeTextCSS + "; " + MimeCharsetUtf8
	MimeTextXML                    = "text/xml"
	MimeTextXMLCharsetUtf8         = MimeTextXML + "; " + MimeCharsetUtf8
	MimeApplicationYAMLCharsetUtf8 = MimeApplicationYAML + "; " + MimeCharsetUtf8
	MimeApplicationYAML            = "application/yaml"
	MimeApplicationXMLCharsetUtf8  = MimeApplicationXML + "; " + MimeCharsetUtf8
	MimeApplicationXML             = "application/xml"
	MimeApplicationProtobuf        = "application/protobuf"
	MimeApplicationJSONCharsetUtf8 = MimeApplicationJSON + "; " + MimeCharsetUtf8
	MimeApplicationJSON            = "application/json"
	MimeApplicationFormCharsetUtf8 = MimeApplicationForm + "; " + MimeCharsetUtf8
	MimeApplicationForm            = "application/x-www-form-urlencoded"

	// Param

	ParamAction          = "action"
	ParamAllow           = "allow"
	ParamBasicAuth       = "basicauth"
	ParamCaller          = "caller"
	ParamControllerGroup = "controllergroup"
	ParamRegister        = "register"
	ParamTemplate        = "template"
	ParamRoute           = "route"
	ParamUserid          = "Userid"
	ParamPolicy          = "Policy"
	ParamResource        = "Resource"
)
