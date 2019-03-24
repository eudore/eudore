package eudore

/*
保存各种全局函数，用于根据名称获得对应的函数。
*/

import (
	"regexp"
	"strconv"
)

var (
	globalHandler				map[string]Handler
	globalHandleFuncs				map[string]HandlerFunc
	// globalMiddlewares				map[string]Middleware
	globalLoggerHandleFuncs		map[string]LoggerHandleFunc
	globalRouterCheckFunc		map[string]RouterCheckFunc
	globalRouterNewCheckFunc		map[string]RouterNewCheckFunc
	globalConfigReadFunc			map[string]ConfigReadFunc
	globalPoolGetFunc				map[string]PoolGetFunc
)

func init() {
	globalHandler = make(map[string]Handler)
	globalHandleFuncs = make(map[string]HandlerFunc)
	// globalMiddlewares = make(map[string]Middleware)
	globalLoggerHandleFuncs = make(map[string]LoggerHandleFunc)
	globalRouterCheckFunc = make(map[string]RouterCheckFunc)
	globalRouterNewCheckFunc = make(map[string]RouterNewCheckFunc)
	// LoggerHandle
	globalLoggerHandleFuncs["default"] = LoggerHandleDefault
	globalLoggerHandleFuncs["json"] = LoggerHandleJson
	globalLoggerHandleFuncs["jsonindent"] = LoggerHandleJsonIndent
	globalLoggerHandleFuncs["xml"] = LoggerHandleXml
	// RouterCheckFunc
	globalRouterCheckFunc["isnum"] = RouterCheckFuncIsnm
	// RouterNewCheckFunc
	globalRouterNewCheckFunc["min"] = RouterNewCheckFuncMin
	globalRouterNewCheckFunc["regexp"] = RouterNewCheckFuncRegexp
	globalConfigReadFunc = make(map[string]ConfigReadFunc)
	globalConfigReadFunc["default"] = ReadFile
	globalConfigReadFunc["file"] = ReadFile
	globalConfigReadFunc["https"] = ReadHttp
	globalConfigReadFunc["http"] = ReadHttp
	globalPoolGetFunc = make(map[string]PoolGetFunc)
}

// Handler
func ConfigSaveHandler(name string, fn Handler) {
	globalHandler[name] = fn
}

func ConfigLoadHandler(name string) Handler {
	return globalHandler[name]
}

// HandleFunc
func ConfigSaveHandleFunc(name string, fn HandlerFunc) {
	globalHandleFuncs[name] = fn
}

func ConfigLoadHandleFunc(name string) HandlerFunc {
	return globalHandleFuncs[name]
}

// Middleware
/*func ConfigSaveMiddleware(name string, fn Middleware) {
	globalMiddlewares[name] = fn
}

func ConfigLoadMiddleware(name string) Middleware {
	return globalMiddlewares[name]
}
*/

// LoggerHandleFunc
func ConfigSaveLoggerHandleFunc(name string, fn LoggerHandleFunc) {
    globalLoggerHandleFuncs[name] = fn
}

func ConfigLoadLoggerHandleFunc(name string) LoggerHandleFunc {
    return globalLoggerHandleFuncs[name]
}

// RouterCheckFunc
func ConfigSaveRouterCheckFunc(name string, fn RouterCheckFunc) {
	globalRouterCheckFunc[name] = fn
}

func ConfigLoadRouterCheckFunc(name string) RouterCheckFunc {
	return globalRouterCheckFunc[name]
}

func RouterCheckFuncIsnm(arg string) bool {
	_, err := strconv.Atoi(arg)
	return err == nil
}

// RouterNewCheckFunc
func ConfigSaveRouterNewCheckFunc(name string, fn RouterNewCheckFunc) {
	globalRouterNewCheckFunc[name] = fn
}

func ConfigLoadRouterNewCheckFunc(name string) RouterNewCheckFunc {
	return globalRouterNewCheckFunc[name]
}

func RouterNewCheckFuncMin(str string) RouterCheckFunc {
	n, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num >= n
	}
}

func RouterNewCheckFuncRegexp(str string) RouterCheckFunc {
	// 创建正则表达式
	re, err := regexp.Compile(str)
	if err != nil {
		return nil
	}
	// 返回正则匹配校验函数
	return func(arg string) bool {
		return re.MatchString(arg)
	}
}


// ConfigReadFunc
func ConfigSaveConfigReadFunc(name string, fn ConfigReadFunc) {
	globalConfigReadFunc[name] = fn
}

func ConfigLoadConfigReadFunc(name string) ConfigReadFunc {
	return globalConfigReadFunc[name]
}

// ConfigReadFunc
func ConfigSavePoolGetFunc(name string, fn PoolGetFunc) {
	globalPoolGetFunc[name] = fn
}

func ConfigLoadPoolGetFunc(name string) PoolGetFunc {
	return globalPoolGetFunc[name]
}