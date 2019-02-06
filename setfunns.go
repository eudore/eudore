package eudore



var (
	handleFuncs				map[string]HandlerFunc
	loggerFormatFuncs		map[string]LoggerFormatFunc
)

func init() {
	loggerFormatFuncs = make(map[string]LoggerFormatFunc)
	loggerFormatFuncs["default"] = LoggerFormatDefault
	loggerFormatFuncs["json"] = LoggerFormatJson
	loggerFormatFuncs["jsonindent"] = LoggerFormatJsonIndent
	loggerFormatFuncs["xml"] = LoggerFormatXml
	handleFuncs = make(map[string]HandlerFunc)
}

/*func ConfigRegisterHandler() {

}*/

func ConfigSaveHandleFunc(name string, fn HandlerFunc) {
	handleFuncs[name] = fn
}

func ConfigLoadHandleFunc(name string) HandlerFunc {
	fn, ok := handleFuncs[name]
	if ok {
		return fn
	}
	return nil
}

func ConfigSaveLoggerFormatFunc(name string, fn LoggerFormatFunc) {
	loggerFormatFuncs[name] = fn
}

func ConfigLoadLoggerFormatFunc(name string) LoggerFormatFunc {
	fn, ok := loggerFormatFuncs[name]
	if ok {
		return fn
	}
	return nil
}