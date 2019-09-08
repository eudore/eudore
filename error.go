package eudore

import (
	"errors"
	"fmt"
)

// 定义默认错误
var (
	// ErrApplicationStop 在app正常退出时返回。
	ErrApplicationStop = errors.New("stop application success")
	// ErrConverterInputDataNil 在Converter方法时，输出参数是空。
	ErrConverterInputDataNil = errors.New("Converter input value is nil")
	// ErrConverterTargetDataNil 在Converter方法时，目标参数是空。
	ErrConverterTargetDataNil = errors.New("Converter target data is nil")
	// ErrEudoreIgnoreInit 在Eudore的Init函数返回该错误，忽略执行后续错误。
	ErrEudoreIgnoreInit = errors.New("eudore ignore the remaining init function")
	// ErrHandlerInvalidRange 在http使用range分开请求文件出现错误时返回。
	ErrHandlerInvalidRange = errors.New("invalid range: failed to overlap")
	// ErrHandlerProxyBackNotWriter HandlerProxy函数处理101时，后端连接不支持io.Writer。
	ErrHandlerProxyBackNotWriter = errors.New("HandlerProxy error: Back conn ResponseReader not suppert io.Writer, Need to implement the io.ReadWriteCloser interface")
	// ErrListenerNotImplementsServerListenFiler 在获取net.Listener的fd文件时，没有实现serverListenFiler接口，无法获得fd文件。
	ErrListenerNotImplementsServerListenFiler = errors.New("Listener is not implements ServerListenFiler")
	// ErrLoggerLevelUnmarshalText 日志级别解码错误，请检查输出的[]byte是否有效。
	ErrLoggerLevelUnmarshalText = errors.New("logger level UnmarshalText error")
	// ErrNewHandlerFuncParamNotFunc 调用NewHandlerFunc函数时参数错误，参数必须是一个函数对象。
	ErrNewHandlerFuncParamNotFunc = errors.New("NewHandlerFunc input parameter must be a function")
	// ErrNotToServerListener newServerListens函数参数异常，无法解析并转换成serverListener类型。
	ErrNotToServerListener = errors.New("Parameters cannot be converted to serverListener type")
	// ErrRegisterControllerExecFuncParamNotFunc RegisterControllerHandlerFunc第一个参数必须是一个函数。
	ErrRegisterControllerHandlerFuncParamNotFunc = errors.New("The parameter type of RegisterControllerHandlerFunc must be a function")
	// ErrRegisterNewHandlerParamNotFunc 调用RegisterHandlerFunc函数时，参数必须是一个函数。
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
	// ErrFormatConverterNotCanset 在Set时，结构体不支持该项属性。
	ErrFormatConverterNotCanset = "The attribute %s of structure %s is not set, please use pointer"
	// ErrFormatConverterSetStringUnknownType setWithString函数遇到未定义的反射类型
	ErrFormatConverterSetStringUnknownType = "setWithString unknown type %s"
	// ErrFormatConverterSetStructNotField 在Set时，结构体没有当前属性。
	ErrFormatConverterSetStructNotField = "Setting the structure has no attribute %s"
	// ErrFormatConverterSetTypeError 在Set时，类型异常，无法继续设置值。
	ErrFormatConverterSetTypeError = "The type of the set value is %s, which is not configurable, key: %v,val: %s"
	// ErrFormatHandlerProxyConnHijack HandlerProxy函数处理101时，请求连接不支持hijack，返回当前错误。
	ErrFormatHandlerProxyConnHijack = "HandlerProxy error: Conn hijack error: %v"
	// ErrFormatNewContrllerExecFuncTypeNotFunc NewContrllerExecFunc函数的参数类型函数为注册，需要先使用RegisterControllerExecFunc注册函数类型。
	ErrFormatNewContrllerExecFuncTypeNotFunc = "The NewContrllerExecFunc parameter type is %s, which is an unregistered handler type"
	// ErrFormatNewHandlerFuncsTypeNotFunc NewHandlerFuncs函数参数错误，参数必须是一个切片或者函数类型。
	ErrFormatNewHandlerFuncsTypeNotFunc = "The NewHandlerFuncs parameter is of the wrong type and must be an slice or a function. The current type is %s"
	// ErrFormatNewHandlerFuncsUnregisterType NewHandlerFuncs函数的参数类型函数未注册，需要先使用RegisterHandlerFunc注册该函数类型。
	ErrFormatNewHandlerFuncsUnregisterType = "The NewHandlerFunc parameter type is %s, which is an unregistered handler type"
	// ErrFormatRegisterHandlerFuncInputParamError RegisterHandlerFunc函数注册的函数参数错误。
	ErrFormatRegisterHandlerFuncInputParamError = "The '%s' input parameter is illegal and should be one"
	// ErrFormatRegisterHandlerFuncOutputParamError RegisterHandlerFunc函数注册的函数返回值错误。
	ErrFormatRegisterHandlerFuncOutputParamError = "The '%s' output parameter is illegal and should be a HandlerFunc object"
	// ErrFormatStartNewProcessError 在StartNewProcess函数fork启动新进程错误。
	ErrFormatStartNewProcessError = "StartNewProcess failed to forkexec error: %v"
	// ErrFormatUnknownTypeBody 在transbody函数解析参数成io.Reader时，参数类型是非法的。
	ErrFormatUnknownTypeBody = "unknown type used for body: %+v"
)

type (
	// Errors 实现多个error组合。
	Errors struct {
		errs []error
	}
	// ErrorCode 实现具有错误信息和错误码的error。
	ErrorCode struct {
		code    int
		message string
	}
)

// NewErrors 创建Errors对象。
func NewErrors() *Errors {
	return &Errors{}
}

// HandleError 实现处理多个错误，如果非空则保存错误。
func (e *Errors) HandleError(errs ...error) {
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
}

// Error 方法实现error接口，返回错误描述。
func (e *Errors) Error() string {
	switch len(e.errs) {
	case 0:
		return ""
	case 1:
		return e.errs[0].Error()
	default:
		return fmt.Sprint(e.errs)
	}
}

// GetError 方法返回错误，如果没有保存的错误则返回空。
func (e *Errors) GetError() error {
	if len(e.errs) == 0 {
		return nil
	}
	return e
}

// NewErrorCode 创建一个ErrorCode对象。
func NewErrorCode(c int, str string) *ErrorCode {
	return &ErrorCode{code: c, message: str}
}

// Error 方法实现error接口，返回错误信息。
func (e *ErrorCode) Error() string {
	return e.message
}

// Code 方法返回错误状态码。
func (e *ErrorCode) Code() int {
	return e.code
}
