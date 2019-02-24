package eudore



var (
	globalhandler				map[string]Handler
	globalhandleFuncs				map[string]HandlerFunc
	globalmiddlewares				map[string]Middleware
	globalloggerHandleFuncs		map[string]LoggerHandleFunc
)

func init() {
	globalhandler = make(map[string]Handler)
	globalhandleFuncs = make(map[string]HandlerFunc)
	globalmiddlewares = make(map[string]Middleware)
	//
	globalloggerHandleFuncs = make(map[string]LoggerHandleFunc)
	globalloggerHandleFuncs["default"] = LoggerHandleDefault
	globalloggerHandleFuncs["json"] = LoggerHandleJson
	globalloggerHandleFuncs["jsonindent"] = LoggerHandleJsonIndent
	globalloggerHandleFuncs["xml"] = LoggerHandleXml
}

// Handler
func ConfigSaveHandler(name string, fn Handler) {
	globalhandler[name] = fn
}

func ConfigLoadHandler(name string) Handler {
	return globalhandler[name]
}

// HandleFunc
func ConfigSaveHandleFunc(name string, fn HandlerFunc) {
	globalhandleFuncs[name] = fn
}

func ConfigLoadHandleFunc(name string) HandlerFunc {
	return globalhandleFuncs[name]
}

// Middleware
func ConfigSaveMiddleware(name string, fn Middleware) {
	globalmiddlewares[name] = fn
}

func ConfigLoadMiddleware(name string) Middleware {
	return globalmiddlewares[name]
}


// LoggerHandleFunc
func ConfigSaveLoggerHandleFunc(name string, fn LoggerHandleFunc) {
    globalloggerHandleFuncs[name] = fn
}

func ConfigLoadLoggerHandleFunc(name string) LoggerHandleFunc {
    return globalloggerHandleFuncs[name]
}

