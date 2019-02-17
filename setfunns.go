package eudore



var (
	globalhandler				map[string]Handler
	globalhandleFuncs				map[string]HandlerFunc
	globalmiddlewares				map[string]Middleware
	globalloggerFormatFuncs		map[string]LoggerFormatFunc
)

func init() {
	globalhandler = make(map[string]Handler)
	globalhandleFuncs = make(map[string]HandlerFunc)
	globalmiddlewares = make(map[string]Middleware)
	globalloggerFormatFuncs = make(map[string]LoggerFormatFunc)
	globalloggerFormatFuncs["default"] = LoggerFormatDefault
	globalloggerFormatFuncs["json"] = LoggerFormatJson
	globalloggerFormatFuncs["jsonindent"] = LoggerFormatJsonIndent
	globalloggerFormatFuncs["xml"] = LoggerFormatXml
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

// LoggerFormatFunc
func ConfigSaveLoggerFormatFunc(name string, fn LoggerFormatFunc) {
	globalloggerFormatFuncs[name] = fn
}

func ConfigLoadLoggerFormatFunc(name string) LoggerFormatFunc {
	return globalloggerFormatFuncs[name]
}