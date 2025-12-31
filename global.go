package eudore

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net"
	"os"
	"reflect"
	"time"
)

var (
	ContextKeyApp             = NewContextKey("app")
	ContextKeyAppCancel       = NewContextKey("app-Cancel")
	ContextKeyAppValues       = NewContextKey("app-values")
	ContextKeyLogger          = NewContextKey("logger")
	ContextKeyConfig          = NewContextKey("config")
	ContextKeyClient          = NewContextKey("client")
	ContextKeyClientTrace     = NewContextKey("client-trace")
	ContextKeyServer          = NewContextKey("server")
	ContextKeyRouter          = NewContextKey("router")
	ContextKeyError           = NewContextKey("error")
	ContextKeyContextPool     = NewContextKey("context-pool")
	ContextKeyContextUser     = NewContextKey("context-user")
	ContextKeyHandlerExtender = NewContextKey("handler-extender")
	ContextKeyBind            = NewContextKey("handler-bind")
	ContextKeyRender          = NewContextKey("handler-render")
	ContextKeyHTTPHandler     = NewContextKey("http-handler")
	ContextKeyFuncCreator     = NewContextKey("func-creator")
	ContextKeyFilterRules     = NewContextKey("filter-rules")
	ContextKeyDaemonCommand   = NewContextKey("daemon-command")
	ContextKeyDaemonSignal    = NewContextKey("daemon-signal")
	ContextKeyEventHub        = NewContextKey("event-hub")
	ContextKeyTrace           = NewContextKey("trace")
	// DefaultClientCheckBodyLength global defines the max length of the
	// [NewClientCheckBody] output string.
	DefaultClientCheckBodyLength = 128
	// DefaultClientDialKeepAlive defines the client [net.Dialer.KeepAlive].
	DefaultClientDialKeepAlive = 30 * time.Second
	// DefaultClientDialTimeout defines the client [net.Dialer.Timeout].
	DefaultClientDialTimeout = 30 * time.Second
	// DefaultClientHost global defines default Host used by [Client].
	DefaultClientHost = "localhost"
	// DefaultClientInternalHost global defines default Host used by [Client]
	// internal connections.
	DefaultClientInternalHost = "internalhost"
	// DefaultClientOptionLoggerError global defines whether
	// [ClientOption.ResponseHooks] outputs [LoggerError] logs.
	DefaultClientOptionLoggerError = true
	// DefaultClientParseErrRange global defines the status code range of
	// [NewClientParseErr] parsing error.
	DefaultClientParseErrRange = [2]int{
		StatusBadRequest,
		StatusNetworkAuthenticationRequired,
	}
	// DefaultClientTimeout defines the default client timeout.
	DefaultClientTimeout = 30 * time.Second
	// DefaultClinetHopHeaders defines Hop to Hop Header, not used.
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
	// DefaultClinetRetryStatus defines the status code when [NewClientRetryNetwork] is retrying.
	DefaultClinetRetryStatus = map[int]struct{}{
		StatusTooManyRequests:     {},
		StatusClientClosedRequest: {},
		StatusBadGateway:          {},
		StatusServiceUnavailable:  {},
		StatusGatewayTimeout:      {},
	}
	// DefaultClinetRetryInterval defines the retry interval and random of
	// [NewClientHookRetry].
	// If the number is exceeded, the last interval is used.
	DefaultClinetRetryInterval = []time.Duration{
		100 * time.Millisecond, 25 * time.Millisecond,
		200 * time.Millisecond, 50 * time.Millisecond,
		400 * time.Millisecond, 100 * time.Millisecond,
		800 * time.Millisecond, 200 * time.Millisecond,
		1600 * time.Millisecond, 400 * time.Millisecond,
		3200 * time.Millisecond, 800 * time.Millisecond,
		6400 * time.Millisecond, 1600 * time.Millisecond,
	}
	// DefaultConfigAllParseFunc defines the all parse used by [NewConfig].
	DefaultConfigAllParseFunc = []ConfigParseFunc{
		NewConfigParseEnvFile(),
		NewConfigParseDefault(),
		NewConfigParseJSON("config"),
		NewConfigParseEnvs(""),
		NewConfigParseArgs(),
		NewConfigParseWorkdir("workdir"),
	}
	// DefaultConfigEnvFiles defines the [NewConfigParseEnvFile] to
	// read the ENV file.
	DefaultConfigEnvFiles = ".env"
	// DefaultConfigEnvPrefix defines the default parameters for
	// [NewConfigParseEnvs], and [NewConfigParseDefault] [NewConfigParseDecoder]
	// will also use it to find a value to look up env.
	DefaultConfigEnvPrefix = "ENV_"
	// DefaultConfigParseTimeout global defines the [Config.Parse] method
	// execution timeout.
	DefaultConfigParseTimeout = time.Second * 60
	// DefaultContextMaxHandler global defines the upper limit of the number
	// of [Context] handlers.
	DefaultContextMaxHandler = 0xff
	// DefaultContextMaxApplicationFormSize defaults to the length limit
	// of the body when parsing [MimeApplicationForm];
	// If Body implements Limit() int64 method, this value is ignored.
	DefaultContextMaxApplicationFormSize int64 = 10 << 20 // 10M
	// DefaultContextMaxMultipartFormMemory The memory size used by the body
	// when parsing [MimeMultipartForm].
	DefaultContextMaxMultipartFormMemory int64 = 32 << 20 // 32 MB
	// DefaultContextFormatTime defines the contextMessage Time format.
	// Modification affects the API response.
	DefaultContextFormatTime = "2006-01-02 15:04:05.000"
	// DefaultControllerParam global defines the controller injection [Params]
	// format and is replaced using [strings.ReplaceAll].
	DefaultControllerParam = "controller={{Package}}.{{Name}} controllermethod={{Method}}"
	// DefaultFuncCreator defines the global default [FuncCreator]
	// used by [NewRouterCoreMux].
	DefaultFuncCreator = NewFuncCreator()
	// DefaultHandlerDataBindFormTags global defines the form tags
	// for [HandlerDataBindForm].
	DefaultHandlerDataBindFormTags = []string{"form", "alias"}
	// DefaultHandlerDataBindURLTags global defines the url tags
	// for [HandlerDataBindURL].
	DefaultHandlerDataBindURLTags = []string{"url", "alias"}
	// DefaultHandlerDataBinds defines all [HandlerDataFuncs] processed
	// by [NewHandlerDataBinds].
	DefaultHandlerDataBinds = map[string]HandlerDataFunc{
		"":                         HandlerDataBindURL,
		MimeApplicationOctetStream: HandlerDataBindURL,
		MimeApplicationJSON:        HandlerDataBindJSON,
		MimeApplicationForm:        HandlerDataBindForm,
		MimeMultipartForm:          HandlerDataBindForm,
		MimeApplicationXML:         HandlerDataBindXML,
	}
	// DefaultHandlerDataRenders defines all [HandlerDataFuncs] processed
	// by [NewHandlerDataRenders].
	DefaultHandlerDataRenders = map[string]HandlerDataFunc{
		MimeAll:             HandlerDataRenderJSON,
		MimeText:            HandlerDataRenderText,
		MimeTextPlain:       HandlerDataRenderText,
		MimeTextHTML:        NewHandlerDataRenderTemplates(nil, nil),
		MimeApplicationJSON: HandlerDataRenderJSON,
	}
	// DefaultHandlerDataRenderTemplateAppend defines the non-existent template
	// to be append when Render the template.
	DefaultHandlerDataRenderTemplateAppend, _ = template.New("").Parse(
		fmt.Sprintf(`{{- define "%s" -}}%s{{- end -}}`,
			DefaultHandlerEmbedTemplateName,
			templateEmbedIndex,
		),
	)
	DefaultHandlerDataRenderTemplateHeaders = map[string]string{
		HeaderXContentTypeOptions: "nosniff",
		HeaderXFrameOptions:       "SAMEORIGIN",
		HeaderXXSSProtection:      "1; mode=block",
	}
	// DefaultHandlerDataTemplateReload defines
	// [NewHandlerDataRenderTemplates] enables template Reload.
	DefaultHandlerDataTemplateReload = true
	// DefaultHandlerValidateTag global defines the struct tag of
	// [NewHandlerDataValidateStruct] to get the validation rules.
	DefaultHandlerValidateTag = "valid"
	// DefaultHandlerEmbedCacheControl defines the [HeaderCacheControl]
	// cache strategy used by [NewHandlerHTTPFileSystem].
	DefaultHandlerEmbedCacheControl = "no-cache"
	// DefaultHandlerEmbedTemplateName global defines the template name used by
	// [NewHandlerFileSystem].
	DefaultHandlerEmbedTemplateName = "eudore-embed-index"
	// DefaultHandlerEmbedTime sets the [HeaderLastModified] Header of the
	// embed file returned by [NewHandlerHTTPFileSystem].
	//
	// The default is the service startup time.
	// If the service is deployed in multiple copies,
	// set the same value makes the [HeaderLastModified] consistent
	// and ensures that [StatusNotModified] caching is enabled.
	DefaultHandlerEmbedTime = time.Now()
	// DefaultHandlerExtender defines global default function [HandlerExtender].
	DefaultHandlerExtender = NewHandlerExtender()
	// DefaultHandlerExtenderAllowKind defines the parameter types allowed by
	// [NewHandlerExtenderBase].
	DefaultHandlerExtenderAllowKind = map[reflect.Kind]struct{}{
		reflect.Func: {}, reflect.Interface: {},
		reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {}, reflect.Struct: {},
	}
	// DefaultHandlerExtenderShowName global defines whether [HandlerFunc]
	// displays the extended function name.
	DefaultHandlerExtenderShowName = true
	// DefaultHandlerExtenderFuncs defines the extension functions registered by
	// [NewHandlerExtender].
	DefaultHandlerExtenderFuncs = []any{
		NewHandlerFunc,
		NewHandlerFuncAny,
		NewHandlerFuncError,
		NewHandlerFuncAnyError,
		NewHandlerFuncContextAny,
		NewHandlerFuncContextError,
		NewHandlerFuncContextAnyError,
		NewHandlerFuncContextMapAnyError,
		NewHandlerHTTPFunc1,
		NewHandlerHTTPFunc2,
		NewHandlerHTTPHandler,
		NewHandlerFileEmbed,
		NewHandlerFileIOFS,
		NewHandlerFileSystem,
		NewHandlerAnyContextTypeAnyError,
	}
	DefaultLoggerDepthKindEnable  = "enable"
	DefaultLoggerDepthKindDisable = "disable"
	DefaultLoggerDepthKindStack   = "stack" // non-fixed
	// DefaultLoggerDepthMaxStack defines the max number of stack layers
	// displayed by the [GetCallerStacks] function.
	DefaultLoggerDepthMaxStack = 0x4f
	// DefaultLoggerNull defines a null log outputter.
	DefaultLoggerNull = NewLoggerNull()
	// DefaultLoggerEntryBufferLength defines the [LoggerEntry] buffer length.
	DefaultLoggerEntryBufferLength = 2048
	// DefaultLoggerEntryFieldsLength defines the number of
	// [LoggerEntry] Fields.
	DefaultLoggerEntryFieldsLength = 4
	// DefaultLoggerFormatter defines the log format for Logger.
	DefaultLoggerFormatter = "json"
	// DefaultLoggerFormatterFormatTime defines the time format for log output.
	DefaultLoggerFormatterFormatTime = "2006-01-02 15:04:05.000"
	// DefaultLoggerFormatterKeyLevel defines the level field output name.
	DefaultLoggerFormatterKeyLevel = "level"
	// DefaultLoggerFormatterKeyMessage defines the message field output name.
	DefaultLoggerFormatterKeyMessage = "message"
	// DefaultLoggerFormatterKeyTime defines the Time field output name.
	DefaultLoggerFormatterKeyTime = "time"
	// DefaultLoggerHookFatal defines whether HookFatal is enabled by default.
	DefaultLoggerHookFatal = false
	// DefaultLoggerLevelStrings global defines the log level output strings.
	//
	// Logger Formatter does not use this variable.
	DefaultLoggerLevelStrings = [...]string{
		"DEBUG", "INFO", "WARNING", "ERROR", "FATAL", "DISCARD",
	}
	// DefaultLoggerWriterRotateDataKeys global defines the keywords for
	// date rolling time/day/month/year, the order cannot be changed.
	DefaultLoggerWriterRotateDataKeys = [...]string{"hh", "dd", "mm", "yyyy"}
	// DefaultLoggerWriterStdout defines whether to output to [os.Stdout].
	DefaultLoggerWriterStdout = os.Getenv(EnvEudoreDaemonEnable) == ""
	// DefaultLoggerWriterStdoutColor defines whether the color level
	// field is used when the OS supports it.
	//
	// sh bash git-bash goland vsc uses the environment TERM.
	DefaultLoggerWriterStdoutColor = os.Getenv("TERM") != ""
	DefaultLoggerPriorityInit      = 100
	// DefaultLoggerPriorityFormatter defines the log formatter priority.
	// Text and JSON share this value.
	DefaultLoggerPriorityFormatter    = 30
	DefaultLoggerPriorityHookCaller   = 20
	DefaultLoggerPriorityHookFatal    = 101
	DefaultLoggerPriorityHookFilter   = 10
	DefaultLoggerPriorityHookMeta     = 60
	DefaultLoggerPriorityWriterAsync  = 80
	DefaultLoggerPriorityWriterStdout = 90
	DefaultLoggerPriorityWriterFile   = 100
	// DefaultRouterAllMethod defines all methods that the router is allowed.
	//
	// Used global in [ControllerInjectAutoRoute].
	DefaultRouterAllMethod = []string{
		MethodGet, MethodPost, MethodPut,
		MethodDelete, MethodHead, MethodPatch,
		MethodOptions, MethodConnect, MethodTrace,
	}
	// DefaultRouterAnyMethod defines the http method used for Any method.
	DefaultRouterAnyMethod = append([]string{}, DefaultRouterAllMethod[0:6]...)
	// DefaultRouterLoggerKind defines the types of logs that the Router
	// outputs.
	DefaultRouterLoggerKind = "all"
	// DefaultServerListen defines [ServerListenConfig] to use the
	// [net.Listen] function for hooking listen.
	DefaultServerListen            = net.Listen
	DefaultServerReadTimeout       = 60 * time.Second
	DefaultServerReadHeaderTimeout = 60 * time.Second
	DefaultServerWriteTimeout      = 60 * time.Second
	// DefaultServerIdleTimeout defines the connection reuse waiting time,
	// which is equal to ReadTimeout when it is 0.
	DefaultServerIdleTimeout = time.Duration(0)
	// DefaultServerShutdownWait global defines the waiting time for the
	// Server to exit gracefully.
	DefaultServerShutdownWait = 30 * time.Second // non-fixed
	// DefaultServerTLSConfig defines the default [tls.Config] used by [ServerListenConfig].
	DefaultServerTLSConfig = &tls.Config{
		NextProtos: []string{"http/1.1"},
		MinVersion: tls.VersionTLS12,
	}
	// DefaultValueGetSetTags global defines the tags for
	// [GetAnyByPath]/[SetAnyByPath].
	DefaultValueGetSetTags = []string{"alias"} // non-fixed
	// DefaultValueParseTimeFormats global defines the time formats to attempt
	// parse in [GetAnyByString].
	DefaultValueParseTimeFormats = []string{ // non-fixed
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"20060102",
		"15:04:05",
	}
	// DefaultValueParseTimeFixed global defines [GetAnyByString] whether the
	// length of the predefined time format is fixed
	// and ignores time formats with different lengths.
	DefaultValueParseTimeFixed = []bool{ // non-fixed
		true, false, true, true, true, true,
	}
	// DefaultValueTimeLocation global defines the timezone used for parsing times.
	DefaultValueTimeLocation = time.Local //nolint:gosmopolitan
	// DefaultDaemonPidfile defines the default pid file
	// used by [daemon.Command].
	DefaultDaemonPidfile = "/var/run/eudore.pid"
	// DefaultGodocServer defines the godoc server domain name used by the app.
	DefaultGodocServer = "https://golang.org"

	ErrLoggerLevelUnmarshalText = "LoggerLevel: UnmarshalText invalid data: %s"
	ErrLoggerInitUnmounted      = errors.New("Logger: loggerInit has been Unmounted, please check the logger initialization order")

	ErrConfigParseDecoder = "Config: decoder %s parse file '%s' error: %w"
	ErrConfigParseError   = "Config: parse func %v error: %v"

	ErrRouterAddController              = "Router: AddController inject %s error: %w"
	ErrRouterAddHandlerExtender         = "Router: AddHandlerExtender path is '%s' RegisterHandlerExtender error: %w"
	ErrRouterAddHandlerMethodInvalid    = "Router: addHandler method '%s' is invalid, add fullpath: '%s'"
	ErrRouterAddHandlerRecover          = "Router: addHandler method is '%s' and path is '%s', recover error: %v"
	ErrRouterHandlerFuncsUnregisterType = "Router: newHandlerFuncs path is '%s', %dth handler parameter type is '%s', this is the unregistered handler type"
	ErrRouterMuxLoadInvalidFunc         = "routerCoreMux: load path '%s' is invalid, error: %w"

	ErrClientBodyNotGetBody    = errors.New("ClientBody: cannot copy body")
	ErrClientOptionInvalidType = "ClientOption: invalid option type %T"
	ErrClientCheckStatusError  = "Client: check %s %s status is %d not in %v"
	ErrClientParseBodyError    = "Client: parse not suppert Content-Type: %s"
	ErrClientParseEventInvalid = "Client: parse event invalid data: %s"

	ErrContextParseFormNotSupportContentType = "Context: parse form not support Content-Type: %s"
	ErrContextRedirectInvalid                = "Context: invalid redirect status code %d"
	ErrContextNotHijacker                    = errors.New("ResponseWriter: http.Hijacker interface is not supported")

	ErrHandlerDataBindNotSupportContentType = "HandlerData: bind not support Content-Type: %s"
	ErrHandlerDataBindMustSturct            = "HandlerData: bind value type %s must be a struct"
	ErrHandlerDataRenderTemplateNotFound    = "HandlerData: render not found template %s"
	ErrHandlerDataRenderTemplateNotLoad     = "Unable to load template at %s: patterns: %v"
	ErrHandlerDataRenderTemplateNeedName    = errors.New("HandlerData: render template need eudore.Context param 'template'")
	ErrHandlerDataValidateCheckFormat       = "Validate: %s.%s field %s check rule %s fatal, value: %%#v"
	ErrHandlerDataValidateCreateRule        = "Validate: %s.%s field %s create rule %s error: %w"

	ErrHandlerExtenderParamNotFunc = errors.New("HandlerExtender: registration function must be a function type")
	ErrHandlerExtenderInputParam   = "HandlerExtender: parameter kind of the registered function %s must be one of func/interface/ptr/struct "
	ErrHandlerExtenderOutputParam  = "HandlerExtender: return type of the registered function %s must be of HandlerFunc type"
	ErrHandlerFuncsCombineTooMany  = "NewHandlerFuncsCombine: too many handlers %d"

	ErrValueNil                  = errors.New("value is nil")
	ErrValueNotSet               = errors.New("value not can set")
	ErrValueNotFound             = errors.New("value not found")
	ErrValueStructNotField       = NewErrorWithWrapped(ErrValueNotFound, "struct field not found")
	ErrValueStructUnexported     = NewErrorWithWrapped(ErrValueNotFound, "struct field unexported")
	ErrValueMapIndexInvalid      = NewErrorWithWrapped(ErrValueNotFound, "map index invalid")
	ErrValueSliceIndexOutOfRange = NewErrorWithWrapped(ErrValueNotFound, "slice index out of range")

	ErrValueLookNil         = "look type %s value %s: %w"
	ErrValueLookType        = "look value type %s path '%s': %w"
	ErrValueLookStruct      = "look struct type %s field '%s': %w"
	ErrValueLookMap         = "look map type %s key '%s': %w"
	ErrValueLookSlice       = "look slice type %s index '%s', len is %d, error: %w"
	ErrValueParseMapKey     = "parse map key '%s' error: %w"
	ErrValueParseSliceIndex = "parse slice index '%s', len is %d error: %w"

	ErrValueSetStringUnknownType = "the SetValueString unknown type %s"
	ErrValueSetValuePtr          = "the SetValuePtr method type %s cannot be assigned to type %s"

	ErrFuncCreatorNotFunc                   = errors.New("FuncCreator: not found or create func")
	ErrFormatFuncCreatorRegisterInvalidType = "Register func '%s' type is %T, must 'func(T) bool' or 'func(string) (func(T) bool, error)'"
)
