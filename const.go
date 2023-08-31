package eudore

// const定义全部全局变量和常量

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"
)

var (
	// 定义各种类型的反射类型。
	typeAny           = reflect.TypeOf((*any)(nil)).Elem()
	typeError         = reflect.TypeOf((*error)(nil)).Elem()
	typeContext       = reflect.TypeOf((*Context)(nil)).Elem()
	typeHandlerFunc   = reflect.TypeOf((*HandlerFunc)(nil)).Elem()
	typeTimeDuration  = reflect.TypeOf((*time.Duration)(nil)).Elem()
	typeTimeTime      = reflect.TypeOf((*time.Time)(nil)).Elem()
	typeFmtStringer   = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	typeJSONMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	typeTextMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	// 检测各类接口。
	_ Client          = (*clientStd)(nil)
	_ Config          = (*configStd)(nil)
	_ Context         = (*contextBase)(nil)
	_ Controller      = (*ControllerAutoRoute)(nil)
	_ Controller      = (*controllerError)(nil)
	_ FuncCreator     = (*funcCreatorBase)(nil)
	_ FuncCreator     = (*funcCreatorExpr)(nil)
	_ HandlerExtender = (*handlerExtenderBase)(nil)
	_ HandlerExtender = (*handlerExtenderTree)(nil)
	_ HandlerExtender = (*handlerExtenderWarp)(nil)
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
	_ Router          = (*RouterStd)(nil)
	_ RouterCore      = (*routerCoreHost)(nil)
	_ RouterCore      = (*routerCoreLock)(nil)
	_ RouterCore      = (*routerCoreStd)(nil)
	_ Server          = (*serverFcgi)(nil)
	_ Server          = (*serverStd)(nil)
)

var (
	// ContextKeyApp 定义获取app的Key。
	ContextKeyApp = NewContextKey("app")
	// ContextKeyAppKeys 定义获取app全部可获取数据keys的Key。
	ContextKeyAppKeys = NewContextKey("app-keys")
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
	// ContextKeyValidater 定义获取validate的Key。
	ContextKeyValidater = NewContextKey("validater")
	// ContextKeyFilter 定义获取filte的Key。
	ContextKeyFilter = NewContextKey("filter")
	// ContextKeyFilterRules 定义获取filte-data的Key。
	ContextKeyFilterRules = NewContextKey("filter-rules")
	// ContextKeyRender 定义获取render的Key。
	ContextKeyRender = NewContextKey("render")
	// ContextKeyHandlerExtender 定义获取handler-extender的Key。
	ContextKeyHandlerExtender = NewContextKey("handler-extender")
	// ContextKeyFuncCreator 定义获取func-creator的Key。
	ContextKeyFuncCreator = NewContextKey("func-creator")
	// ContextKeyTemplate 定义获取templdate的Key。
	ContextKeyTemplate = NewContextKey("templdate")
	// ContextKeyTrace 定义获取trace的Key。
	ContextKeyTrace = NewContextKey("trace")
	// ContextKeyDaemonCommand 定义获取daemon-command的Key。
	ContextKeyDaemonCommand = NewContextKey("daemon-command")
	// ContextKeyDaemonSignal 定义获取daemon-signal的Key。
	ContextKeyDaemonSignal = NewContextKey("daemon-signal")
	// ContextKeyDatabaseRuntime 定义获取database-runtime的Key。
	ContextKeyDatabaseRuntime = NewContextKey("database-runtime")
	// DefaultClientDialKeepAlive 定义默认DialContext超时时间。
	DefaultClientDialKeepAlive = 30 * time.Second
	// DefaultClientDialTimeout 定义默认DialContext超时时间。
	DefaultClientDialTimeout = 30 * time.Second
	// DefaultClientHost 定义clientStd默认使用的Host。
	DefaultClientHost = "localhost:80"
	// DefaultClientInternalHost 定义clientStd使用内部连接的Host。
	DefaultClientInternalHost = "127.0.0.10:80"
	// DefaultClientParseErrStar 定义NewClientParseErr解析err的状态码范围。
	DefaultClientParseErrStar = 500
	// DefaultClientParseErrEnd 定义NewClientParseErr解析err的状态码范围。
	DefaultClientParseErrEnd = 500
	// DefaultClientTimeout 定义客户端默认超时时间。
	DefaultClientTimeout = 30 * time.Second
	// DefaultClinetHopHeaders 定义Hop to Hop Header。
	DefaultClinetHopHeaders = [...]string{
		HeaderConnection,
		HeaderUpgrade,
		HeaderKeepAlive,
		HeaderProxyConnection,
		HeaderProxyAuthenticate,
		HeaderProxyAuthorization,
		HeaderTE,
		HeaderTrailer,
		HeaderTransferEncoding,
	}
	// DefaultClinetLoggerLevel 定义Client默认最小输出日志级别。
	DefaultClinetLoggerLevel = LoggerError
	// DefaultClinetRetryStatus 定义NewClientRetryNetwork重试状态码。
	DefaultClinetRetryStatus = map[int]bool{
		StatusTooManyRequests:     true,
		StauusClientClosedRequest: true,
		StatusBadGateway:          true,
		StatusServiceUnavailable:  true,
		StatusGatewayTimeout:      true,
	}
	// DefaultConfigAllParseFunc 定义Config默认使用的解析函数。
	DefaultConfigAllParseFunc = []ConfigParseFunc{
		NewConfigParseEnvFile(),
		NewConfigParseDefault(),
		NewConfigParseJSON("config"),
		NewConfigParseArgs(nil),
		NewConfigParseEnvs("ENV_"),
		NewConfigParseWorkdir("workdir"),
		NewConfigParseHelp("help"),
	}
	// DefaultConfigEnvFiles 定义NewConfigParseEnvFile函数默认读取ENV文件。
	DefaultConfigEnvFiles = ".env"
	// DefaultContextMaxHandler 定义请求上下文handler数量上限，需要小于该值。
	DefaultContextMaxHandler = 0xff
	// DefaultContextMaxApplicationFormSize 默认解析ApplicationFrom时body限制长度；
	// 如果Body实现Limit() int64方法忽略该值。
	DefaultContextMaxApplicationFormSize int64 = 10 << 20 // 10M
	// DefaultContextMaxMultipartFormMemory 默认解析MultipartFrom时body使用内存大小。
	DefaultContextMaxMultipartFormMemory int64 = 32 << 20 // 32 MB
	// DefaultContextPushNotSupportedError 定义Context.Push时是否输出http.ErrNotSupported错误。
	DefaultContextPushNotSupportedError = true
	// DefaultFuncCreator 定义全局默认FuncCreator, RouetrCoreStd默认使用。
	DefaultFuncCreator = NewFuncCreator()
	// DefaultHandlerBindFormTags 定义bind form使用tags。
	DefaultHandlerBindFormTags = []string{"form", "alias"}
	// DefaultHandlerBindHeaderTags 定义bind header使用tags。
	DefaultHandlerBindHeaderTags = []string{"header", "alias"}
	// DefaultHandlerBindURLTags 定义bind url使用tags。
	DefaultHandlerBindURLTags = []string{"url", "alias"}
	// DefaultHandlerDataCode 定义Bind/Validate/Filter/Render返回错误时使用的自定义Code。
	DefaultHandlerDataCode = [4]int{}
	// DefaultHandlerDataStatus 定义Bind/Validate/Filter/Render返回错误时使用的自定义Status。
	DefaultHandlerDataStatus = [4]int{}
	// DefaultHandlerRenderFunc 定义默认使用的Render函数。
	DefaultHandlerRenderFunc = RenderJSON
	// DefaultHandlerValidateTag 定义NewValidateField获取校验规则的结构体tag。
	DefaultHandlerValidateTag = "validate"
	// DefaultHandlerEmbedCacheControl 定义默认NewHandlerEmbedFunc使用的Cache-Control缓存策略。
	DefaultHandlerEmbedCacheControl = "no-cache"
	// DefaultHandlerEmbedTime 设置http返回embed文件的最后修改时间，默认为服务启动时间。
	// 如果服务存在多副本部署，通过设置相同的值使多副本间的时间版本一致，保证启用304缓存。
	DefaultHandlerEmbedTime = time.Now()
	// DefaultHandlerExtender 为默认的函数扩展处理者。
	DefaultHandlerExtender = NewHandlerExtender()
	// DefaultHandlerExtenderAllowType 定义handlerExtenderBase允许使用的参数类型。
	DefaultHandlerExtenderAllowType = map[reflect.Kind]struct{}{
		reflect.Func: {}, reflect.Interface: {},
		reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {}, reflect.Struct: {},
	}
	// DefaultHandlerExtenderFuncs 定义NewHandlerExtender默认注册的扩展函数。
	DefaultHandlerExtenderFuncs = []any{
		NewHandlerEmbed,
		NewHandlerFunc,
		NewHandlerFuncContextError,
		NewHandlerFuncContextAnyError,
		NewHandlerFuncContextRender,
		NewHandlerFuncContextRenderError,
		NewHandlerFuncError,
		NewHandlerFuncRPC,
		NewHandlerFuncRPCMap,
		NewHandlerFuncRender,
		NewHandlerFuncRenderError,
		NewHandlerFuncString,
		NewHandlerHTTP,
		NewHandlerHTTPFileSystem,
		NewHandlerHTTPFunc1,
		NewHandlerHTTPFunc2,
		NewHandlerHTTPHandler,
		NewHandlerStringer,
	}
	// DefaultLoggerDepthMaxStack 定义GetCallerStacks函数默认显示栈最大层数。
	DefaultLoggerDepthMaxStack = 0x4f
	// DefaultLoggerNull 定义空日志输出器。
	DefaultLoggerNull            = NewLoggerNull()
	DefaultLoggerEnableHookFatal = false
	DefaultLoggerEnableHookMeta  = false
	DefaultLoggerEnableStdColor  = true
	// DefaultLoggerEntryBufferLength 定义默认LoggerEntry缓冲长度。
	DefaultLoggerEntryBufferLength = 2048
	// DefaultLoggerEntryFieldsLength 定义默认LoggerEntry Field数量。
	DefaultLoggerEntryFieldsLength = 4
	// DefaultLoggerFormatter 定义Logger默认日志格式化格式。
	DefaultLoggerFormatter = "json"
	// DefaultLoggerFormatterFormatTime 定义默认日志输出和contextBase.WriteError的时间格式。
	DefaultLoggerFormatterFormatTime = "2006-01-02 15:04:05.000"
	// DefaultLoggerFormatterKeyLevel 定义默认Level字段输出名称。
	DefaultLoggerFormatterKeyLevel = "level"
	// DefaultLoggerFormatterKeyMessage 定义默认Message字段输出名称。
	DefaultLoggerFormatterKeyMessage = "message"
	// DefaultLoggerFormatterKeyTime 定义默认Time字段输出名称。
	DefaultLoggerFormatterKeyTime = "time"
	// DefaultLoggerLevelStrings 定义日志级别输出字符串。
	DefaultLoggerLevelStrings = [...]string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL", "DISCARD"}
	// DefaultLoggerWriterRotateDataKeys 定义日期滚动时/天/月/年的关键字，顺序不可变化。
	DefaultLoggerWriterRotateDataKeys = [...]string{"hh", "dd", "mm", "yyyy"}
	// DefaultLoggerWriterStdoutWindowsColor 定义GOOS=windows时是否使用彩色level字段。
	DefaultLoggerWriterStdoutWindowsColor = false
	// DefaultRouterAllMethod 定义路由器允许注册的全部方法，前六种方法在RouterCore始终存在。
	DefaultRouterAllMethod = []string{
		MethodGet, MethodPost, MethodPut,
		MethodDelete, MethodHead, MethodPatch,
		MethodOptions, MethodConnect, MethodTrace,
	}
	// DefaultRouterAnyMethod 定义Any方法的注册使用的方法。
	DefaultRouterAnyMethod = append([]string{}, DefaultRouterAllMethod[0:6]...)
	// DefaultRouterCoreMethod 定义routerCoreStd实现中默认存储的6种方法处理对象。
	DefaultRouterCoreMethod = append([]string{}, DefaultRouterAllMethod[0:6]...)
	// DefaultRouterLoggerKind 定义默认RouterStd输出那些类型日志。
	DefaultRouterLoggerKind = "all"
	// DefaultServerListen 定义ServerListenConfig使用Listen函数，用于hook listen。
	DefaultServerListen            = net.Listen
	DefaultServerReadTimeout       = 60 * time.Second
	DefaultServerReadHeaderTimeout = 60 * time.Second
	DefaultServerWriteTimeout      = 60 * time.Second
	DefaultServerIdleTimeout       = 60 * time.Second
	// DefaultServerShutdownWait 定义Server优雅退出等待时间。
	DefaultServerShutdownWait = 30 * time.Second
	// DefaultTemplateNameStaticIndex 定义默认渲染静态目录模板名称。
	DefaultTemplateNameStaticIndex = "eudore-embed-index"
	// DefaultTemplateNameRenderData 定义默认RenderHTML模板名称。
	DefaultTemplateNameRenderData = "eudore-render-data"
	// DefaultTemplateContentStaticIndex 定义默认渲染静态目录模板内容。
	DefaultTemplateContentStaticIndex = templateEmbedIndex
	// DefaultTemplateContentRenderData 定义默认RenderHTML模板内容。
	DefaultTemplateContentRenderData = tempdateRenderData
	// DefaultTemplateInit 定义App默认加载模板内容。
	DefaultTemplateInit = fmt.Sprintf(`{{- define "%s" -}}%s{{- end -}}{{- define "%s" -}}%s{{- end -}}`,
		DefaultTemplateNameStaticIndex, DefaultTemplateContentStaticIndex,
		DefaultTemplateNameRenderData, DefaultTemplateContentRenderData,
	)
	// DefaultValueGetSetTags 定义Get/SetAny默认的tag。
	DefaultValueGetSetTags = []string{"alias"}
	// DefaultValueParseTimeFormats 定义尝试解析的时间格式。
	DefaultValueParseTimeFormats = []string{
		"2006-01-02",
		"20060102",
		"15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999999999Z07:00",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	}
	// DefaultValueParseTimeFixed 定义预定义时间格式长度是否固定，避免解析长度不相同的时间格式。
	DefaultValueParseTimeFixed = []bool{
		true, true, true, true, true, true,
		false, false, true, true, true, true, true, true, true, true,
	}
	// DefaultDaemonPidfile 定义daemon默认使用的pid文件。
	DefaultDaemonPidfile = "/var/run/eudore.pid"
	// DefaultGodocServer 定义应用默认使用的godoc服务器域名。
	DefaultGodocServer = "https://golang.org"
	// DefaultTraceServer 定义应用默认使用的jaeger链路追踪服务器域名。
	DefaultTraceServer = ""

	// ErrClientBodyFormNotGetBody 定义ClientBodyForm无法获取复制对象错误。
	ErrClientBodyFormNotGetBody = errors.New("client bodyForm contains files that cannot be copied, cannot copy body")
	// ErrFuncCreatorNotFunc 定义FuncCreator无法获取或创建函数。
	ErrFuncCreatorNotFunc = errors.New("not found or create func")
	// ErrHandlerExtenderParamNotFunc 定义调用RegisterHandlerExtender函数时，参数必须是一个函数。
	ErrHandlerExtenderParamNotFunc = errors.New("the parameter type of RegisterNewHandler must be a function")
	// ErrLoggerLevelUnmarshalText 日志级别解码错误，请检查输出的[]byte是否有效。
	ErrLoggerLevelUnmarshalText = errors.New("logger level UnmarshalText error")
	// ErrRenderHandlerSkip 定义Renders执行Render时无法渲染，跳过当前Render。
	ErrRenderHandlerSkip = errors.New("render hander skip")
	// ErrResponseWriterNotHijacker ResponseWriterHTTP对象没有实现http.Hijacker接口。
	ErrResponseWriterNotHijacker = errors.New("http.Hijacker interface is not supported")
	// ErrValueInputDataNil 在Converter方法时，输出参数是空。
	ErrValueInputDataNil = errors.New("converter input value is nil")
	// ErrValueInputDataNotPtr 在Converter方法时，输出参数是空。
	ErrValueInputDataNotPtr = errors.New("converter input value not is ptr")

	// ErrFormatBindDefaultNotSupportContentType BindDefault函数不支持当前的Content-Type Header。
	ErrFormatBindDefaultNotSupportContentType = "BindDefault: not support content type header: %s"
	// ErrFormatClintCheckStatusError 定义Client检查status不匹配错误。
	ErrFormatClintCheckStatusError = "clint check status is %d not in %v"
	// ErrFormatClintParseBodyError 定义Client解析Body时无法解析Content-Type错误。
	ErrFormatClintParseBodyError = "eudore client parse not suppert Content-Type: %s"
	// ErrFormatContextParseFormNotSupportContentType Context解析Form时时，不支持Content-Type。
	ErrFormatContextParseFormNotSupportContentType = "eudore.Context: parse form not supported Content-Type: %s"
	// ErrFormatContextRedirectInvalid Context.Redirect方法使用了无效的状态码。
	ErrFormatContextRedirectInvalid = "eudore.Context: invalid redirect status code %d"
	// ErrFormatContextPushFailed Context.Push方法推送资源错误。
	ErrFormatContextPushFailed = "eudore.Context: push resource %s failed: %w"
	// ErrFormatFuncCreatorRegisterInvalidType fc注册函数类似是无效的。
	ErrFormatFuncCreatorRegisterInvalidType = "Register func '%s' type is %T, must 'func(T) bool' or 'func(string) (func(T) bool, error)'"
	// ErrFormatHandlerExtenderInputParamError RegisterHandlerExtender函数注册的函数参数错误。
	ErrFormatHandlerExtenderInputParamError = "the '%s' input parameter is illegal and should be one func/interface/ptr/struct"
	// ErrFormatHandlerExtenderOutputParamError RegisterHandlerExtender函数注册的函数返回值错误。
	ErrFormatHandlerExtenderOutputParamError = "the '%s' output parameter is illegal and should be a HandlerFunc object"
	// ErrFormatRouterStdAddController RouterStd控制器路由注入错误。
	ErrFormatRouterStdAddController = "the RouterStd.AddController Inject %s error: %w"
	// ErrFormatRouterStdAddHandlerExtender RouterStd添加扩展处理函数错误。
	ErrFormatRouterStdAddHandlerExtender = "the RouterStd.AddHandlerExtender path is '%s' RegisterHandlerExtender error: %w"
	// ErrFormatRouterStdaddHandlerMethodInvalid RouterStd.addHandler 的添加的是无效的，全部有效方法为RouterAnyMethod。
	ErrFormatRouterStdAddHandlerMethodInvalid = "the RouterStd.addHandler arg method '%s' is invalid, add fullpath: '%s'"
	// ErrFormatRouterStdaddHandlerRecover RouterStd注册路由时恢复panic。
	ErrFormatRouterStdAddHandlerRecover = "the RouterStd.addHandler arg method is '%s' and path is '%s', recover error: %w"
	// ErrFormarRouterStdLoadInvalidFunc RouterStd无法加载路径对应的校验函数。
	ErrFormarRouterStdLoadInvalidFunc = "loadCheckFunc path is invalid, load path '%s' error: %v "
	// ErrFormatRouterStdNewHandlerFuncsUnregisterType RouterStd添加处理对象或中间件的第n个参数类型未注册，需要先使用RegisterHandlerExtender或AddHandlerExtender注册该函数类型。
	ErrFormatRouterStdNewHandlerFuncsUnregisterType = "the RouterStd.newHandlerFuncs path is '%s', %dth handler parameter type is '%s', this is the unregistered handler type"
	// ErrFormatProtobufDecodeNilInteface 定义protobuf解码到空接口。
	ErrFormatProtobufDecodeNilInteface    = "protobuf decode %s interface %s is nil"
	ErrFormatProtobufDecodeInvalidFlag    = "protobuf decode %s invalid flag %d"
	ErrFormatProtobufDecodeInvalidKind    = "protobuf decode %s invalid kind %s"
	ErrFormatProtobufDecodeReadError      = "protobuf decode %s read %s error: %w"
	ErrFormatProtobufDecodeReadInvalid    = "protobuf decode %s read length %d invalid has data %d"
	ErrFormatProtobufDecodeMessageNotRead = "protobuf decode message has %d not read"
	ErrFormatProtobufTypeMustSturct       = "protobuf encdoe/decode kind must struct, current type %s"
	// ErrFormatParseValidateFieldError 定义Validate校验失败时输出Error格式。
	ErrFormatValidateErrorFormat = "validate %s.%s field %s check %s rule fatal, value: %%#v"
	// ErrFormatValidateParseFieldError Validate解析结构体规则错误。
	ErrFormatValidateParseFieldError = "validateField %s.%s parse field %s create rule %s error: %s"
	// ErrFormatValueError 定义Value操作错误。
	ErrFormatValueError = "value %s path '%s' error: %w"
	// ErrFormatValueTypeNil 定义Value对象为空。
	ErrFormatValueTypeNil           = "is nil"
	ErrFormatValueAnonymousField    = " is anonymous field"
	ErrFormatValueNotField          = "not found field '%s'"
	ErrFormatValueArrayIndexInvalid = "parse index '%s' is invalid, length is %d"
	ErrFormatValueMapIndexInvalid   = "parse index '%s' is invalid"
	ErrFormatValueMapValueInvalid   = "obtained index '%s' value is invalid"
	ErrFormatValueStructUnexported  = "field '%s' is unexported"
	ErrFormatValueStructNotCanset   = "field '%s' is not canset "
	// ErrFormatConverterSetStringUnknownType setWithString函数遇到未定义的反射类型。
	ErrFormatValueSetStringUnknownType = "setWithString unknown type %s"
	// ErrFormatConverterSetWithValue setWithValue函数中类型无法赋值。
	ErrFormatValueSetWithValue = "the setWithValue method type %s cannot be assigned to type %s"
)

// 定义eudore定义各种常量。
const (
	// EnvEudoreListeners 定义启动fd的地址。
	EnvEudoreDaemonListeners = "EUDORE_DAEMON_LISTENERS"
	// EnvEudoreDaemonRestartID 定义重启时父进程的pid，由子进程kill。
	EnvEudoreDaemonRestartID = "EUDORE_DAEMON_RESTART_ID"
	// EnvEudoreDaemonEnable 用于表示是否fork后台启动，会禁用Logger stdout输出。
	EnvEudoreDaemonEnable = "EUDORE_DAEMON_ENABLE"
	// EnvEudoreDaemonTimeout 定义daemon等待restart和stop命令完成的超时秒数。
	EnvEudoreDaemonTimeout = "EUDORE_DAEMON_TIMEOUT"

	// Status.

	StatusContinue                      = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols            = 101 // RFC 7231, 6.2.2
	StatusProcessing                    = 102 // RFC 2518, 10.1
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
	StauusClientClosedRequest           = 499 // nginx

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
	HeaderXEudoreCache                    = "X-Eudore-Cache"
	HeaderXEudoreRoute                    = "X-Eudore-Route"

	// default http method by rfc2616.

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

	// Mime.

	MimeText                       = "text/*"
	MimeTextPlain                  = "text/plain"
	MimeTextMarkdown               = "text/markdown"
	MimeTextJavascript             = "text/javascript"
	MimeTextHTML                   = "text/html"
	MimeTextCSS                    = "text/css"
	MimeTextXML                    = "text/xml"
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
	// Param.

	ParamAction          = "action"
	ParamAllow           = "allow"
	ParamAutoIndex       = "autoindex"
	ParamBasicAuth       = "basicauth"
	ParamCaller          = "caller"
	ParamControllerGroup = "controllergroup"
	ParamDepth           = "depth"
	ParamLoggerKind      = "loggerkind"
	ParamPrefix          = "prefix"
	ParamRegister        = "register"
	ParamTemplate        = "template"
	ParamRoute           = "route"
	ParamUserid          = "Userid"
	ParamUsername        = "Username"
	ParamPolicy          = "Policy"
	ParamResource        = "Resource"
)

var (
	templateEmbedIndex = `<!DOCTYPE html><html>
<head>
	<meta charset="utf-8">
	<meta name="color-scheme" content="light dark">
	<meta name="google" value="notranslate">
	<title id="title">Index of {{.Path}}</title>
</head>
<body bgcolor="white" from="chrome file">
<h1>Index of {{.Path}} {{if .Upload}}<label for="upload">Upload</label><input id="upload" type="file" name="files" multiple="multiple">{{end}}</h1>
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
</table><style>h1 {border-bottom: 1px solid #c0c0c0; margin-bottom: 10px; padding-bottom: 10px; white-space: nowrap; }
table {border-collapse: collapse; }
th {cursor: pointer; }
td.column {padding-inline-start: 2em; text-align: end; white-space: nowrap; }
a.icon {padding-inline-start: 1.5em; text-decoration: none; user-select: auto; }
a.icon:hover {text-decoration: underline; }
a.file {background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAIAAACQkWg2AAAABnRSTlMAAAAAAABupgeRAAABEElEQVR42nRRx3HDMBC846AHZ7sP54BmWAyrsP588qnwlhqw/k4v5ZwWxM1hzmGRgV1cYqrRarXoH2w2m6qqiqKIR6cPtzc3xMSML2Te7XZZlnW7Pe/91/dX47WRBHuA9oyGmRknzGDjab1ePzw8bLfb6WRalmW4ip9FDVpYSWZgOp12Oh3nXJ7nxoJSGEciteP9y+fH52q1euv38WosqA6T2gGOT44vry7BEQtJkMAMMpa6JagAMcUfWYa4hkkzAc7fFlSjwqCoOUYAF5RjHZPVCFBOtSBGfgUDji3c3jpibeEMQhIMh8NwshqyRsBJgvF4jMs/YlVR5KhgNpuBLzk0OcUiR3CMhcPaOzsZiAAA/AjmaB3WZIkAAAAASUVORK5CYII=") left top no-repeat; }
a.dir {background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAABt0lEQVR42oxStZoWQRCs2cXdHTLcHZ6EjAwnQWIkJyQlRt4Cd3d3d1n5d7q7ju1zv/q+mh6taQsk8fn29kPDRo87SDMQcNAUJgIQkBjdAoRKdXjm2mOH0AqS+PlkP8sfp0h93iu/PDji9s2FzSSJVg5ykZqWgfGRr9rAAAQiDFoB1OfyESZEB7iAI0lHwLREQBcQQKqo8p+gNUCguwCNAAUQAcFOb0NNGjT+BbUC2YsHZpWLhC6/m0chqIoM1LKbQIIBwlTQE1xAo9QDGDPYf6rkTpPc92gCUYVJAZjhyZltJ95f3zuvLYRGWWCUNkDL2333McBh4kaLlxg+aTmyL7c2xTjkN4Bt7oE3DBP/3SRz65R/bkmBRPGzcRNHYuzMjaj+fdnaFoJUEdTSXfaHbe7XNnMPyqryPcmfY+zURaAB7SHk9cXSH4fQ5rojgCAVIuqCNWgRhLYLhJB4k3iZfIPtnQiCpjAzeBIRXMA6emAqoEbQSoDdGxFUrxS1AYcpaNbBgyQBGJEOnYOeENKR/iAd1npusI4C75/c3539+nbUjOgZV5CkAU27df40lH+agUdIuA/EAgDmZnwZlhDc0wAAAABJRU5ErkJggg==") left top no-repeat; }
a.up {background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAACM0lEQVR42myTA+w1RxRHz+zftmrbdlTbtq04qRGrCmvbDWp9tq3a7tPcub8mj9XZ3eHOGQdJAHw77/LbZuvnWy+c/CIAd+91CMf3bo+bgcBiBAGIZKXb19/zodsAkFT+3px+ssYfyHTQW5tr05dCOf3xN49KaVX9+2zy1dX4XMk+5JflN5MBPL30oVsvnvEyp+18Nt3ZAErQMSFOfelCFvw0HcUloDayljZkX+MmamTAMTe+d+ltZ+1wEaRAX/MAnkJdcujzZyErIiVSzCEvIiq4O83AG7LAkwsfIgAnbncag82jfPPdd9RQyhPkpNJvKJWQBKlYFmQA315n4YPNjwMAZYy0TgAweedLmLzTJSTLIxkWDaVCVfAbbiKjytgmm+EGpMBYW0WwwbZ7lL8anox/UxekaOW544HO0ANAshxuORT/RG5YSrjlwZ3lM955tlQqbtVMlWIhjwzkAVFB8Q9EAAA3AFJ+DR3DO/Pnd3NPi7H117rAzWjpEs8vfIqsGZpaweOfEAAFJKuM0v6kf2iC5pZ9+fmLSZfWBVaKfLLNOXj6lYY0V2lfyVCIsVzmcRV9Y0fx02eTaEwhl2PDrXcjFdYRAohQmS8QEFLCLKGYA0AeEakhCCFDXqxsE0AQACgAQp5w96o0lAXuNASeDKWIvADiHwigfBINpWKtAXJvCEKWgSJNbRvxf4SmrnKDpvZavePu1K/zu/due1X/6Nj90MBd/J2Cic7WjBp/jUdIuA8AUtd65M+PzXIAAAAASUVORK5CYII=") left top no-repeat; }
#dir-link {margin-bottom: 10px; padding-bottom: 10px; }
#upload {display: none}
</style><script>"use strict";
let input = document.querySelector('#upload')
function uploadFiles(){
	let data = new FormData()
	for(let f of input.files) {data.append("files", f)}
	fetch('.', {method: 'POST', body: data}).then(()=>location.reload())
}
if(input!=null){input.addEventListener('change', uploadFiles, false);}
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
	tempdateRenderData = `<!DOCTYPE html><html>
<head>
	<meta charset="utf-8">
	<title>Eudore Render</title>
	<meta name="author" content="eudore">
	<meta name="referrer" content="always">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="Eudore render text/html">
	<style>
		*{margin:0;padding:0}
		main {display: flex;flex-direction: column;padding: 10px;}
		fieldset {border: 0;}
		.title {font-weight: 900;}
		.name {color: #5f6368;font-weight: 900;margin-left: 20px;}
	</style>
</head>
<body>
	<main>
		<fieldset>
			<legend class="title">General</legend>
			<div><span class="name">Request URL: </span><span>{{.Host}}{{.Path}}</span></div>
			<div><span class="name">Request Method: </span><span>{{.Method}}</span></div>
			<div><span class="name">Status Code: </span><span>{{.Status}}</span></div>
			<div><span class="name">Remote Address: </span><span>{{.RemoteAddr}}</span></div>
			<div><span class="name">Local Address: </span><span>{{.LocalAddr}}</span></div>
		</fieldset>
		{{- if ne (len .Query) 0 }}
		<fieldset>
			<legend class="title">Requesst Querys</legend>
			{{- range $key, $vals := .Query -}}
			{{- range $i, $val := $vals }}
			<div><span class="name">{{$key}}: </span><span>{{$val}}</span></div>
			{{- end }}
			{{- end }}
		</fieldset>
		{{- end }}
		<fieldset>
			<legend class="title">Requesst Params</legend>
			{{- $iskey := true }}
			{{- range $i,$val := .Params}}
			{{- if $iskey}}
			<div><span class="name">{{$val}}: </span>{{- else}}<span>{{$val}}</span></div>{{end}}
			{{- $iskey = not $iskey}}
			{{- end}}
		</fieldset>
		<fieldset>
			<legend class="title">Request Headers</legend>
			{{- range $key, $vals := .RequestHeader -}}
			{{- range $i, $val := $vals }}
			<div><span class="name">{{$key}}: </span><span>{{$val}}</span></div>
			{{- end }}
			{{- end }}
		</fieldset>
		<fieldset>
			{{- $trace := .TraceServer }}
			<legend class="title">Response Headers</legend>
			{{- range $key, $vals := .ResponseHeader -}}
			{{- range $i, $val := $vals }}
			{{- if and (eq $key "X-Trace-Id") (ne $trace "")}}
			<div><span class="name">{{$key}}: </span><span><a href="{{$trace}}/trace/{{$val}}">{{$val}}</a></span></div>
			{{- else }}
			<div><span class="name">{{$key}}: </span><span>{{$val}}</span></div>
			{{- end }}
			{{- end }}
			{{- end }}
		</fieldset>
		<fieldset>
			<legend class="title">Response Data</legend>
			<pre class="name">{{.Data}}</pre>
		</fieldset>
	</main>
</body></html>
`
)
