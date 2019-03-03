package eudore

import (
	"io"
	"os"
	"fmt"
	"time"
	"bufio"
	"strings"
	"strconv"
	"runtime"
	"encoding/json"
	"encoding/xml"
	"text/template"
)

const (
	LogDebug	LoggerLevel = iota
	LogInfo
	LogWarning
	LogError
	LogFatal
	numSeverity = 5
)

var LogLevelString = [5]string{"DEBUG", "INFO" ,"WARNING" ,"ERROR" ,"FATAL"}

type (
	// 日志级别
	LoggerTime time.Time
	LoggerLevel int
	Fields map[string]interface{}
	LoggerHandleFunc		func(io.Writer, Entry) 
	// 日志输出接口
	LogOut interface {
		Debug(...interface{})
		Info(...interface{})
		Warning(...interface{})
		Error(...interface{})
		Fatal(...interface{})
		WithField(key string, value interface{}) LogOut
		WithFields(fields Fields) LogOut
	}

	Entry interface {
		LogOut
		GetLevel() LoggerLevel
	}

	Logger interface {
		Component
		LogOut
		// SetLevel(Level)
		// SetFromat(LoggerFormatFunc)
		Handle(Entry)
	}
	// The initial log processor necessary interface to process all logs of the current record using the new log processor.
	//
	// 初始日志处理器必要接口，使用新日志处理器处理当前记录的全部日志。
	LoggerInitHandler interface {
		NextHandler(Logger)
	}
	// The initial log processor only records the log. After setting the log processor,
	// it will forward the log of the current record to the new log processor for processing the log generated before the program is initialized.
	//
	// 初始日志处理器仅记录日志，再设置日志处理器后，
	// 会将当前记录的日志交给新日志处理器处理，用于处理程序初始化之前产生的日志。
	LoggerInit struct {
		data		[]Entry
	}
	LoggerStdConfig struct {
		Std			bool
		Path		string
		Level		LoggerLevel `json:"level"`
		Format		string
	}
	LoggerStd struct {
		*LoggerStdConfig
		LoggerHandleFunc
		out			*bufio.Writer
		// out				io.Writer
	}

	LoggerMultiConfig struct {
		Configs		[]interface{}
	}
	LoggerMulti struct {
		*LoggerMultiConfig
		Loggers		[]Logger
	}

	// 标准日志条目
	EntryStd struct {
		Logger		Logger		`json:"-" xml:"-" yaml:"-"`
		Level		LoggerLevel		`json:"level"`
		Fields		Fields		`json:"fields,omitempty"`
		Timestamp	time.Time	`json:"timestamp"`
		Message		string		`json:"message,omitempty"`
	}
	// Context使用日志条目，重写Fatal方法。
	EntryContext struct {
		ctx		Context			`json:"-"`
		Entry					`json:"-"`
		Fields		Fields		`json:"fields,omitempty"`
	}
)


func NewLogger(name string, arg interface{}) (Logger, error) {
	name = AddComponetPre(name, "logger")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	l, ok := c.(Logger)
	if ok {
		return l, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Logger type", name)
}


// Level type
func (l LoggerLevel) String() string {
	return LogLevelString[l]
}

func (l LoggerLevel) MarshalText() (text []byte, err error) {
	text = []byte(l.String())
	return
}

func (l LoggerLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	for i, s := range LogLevelString {
		if s == str {
			l = LoggerLevel(i)
			return nil
		}
	}
	n, err := strconv.Atoi(str)
	if err == nil && n < 5 && n > -1 {
		l = LoggerLevel(n)
		return nil
	}
	return fmt.Errorf("level UnmarshalText error")
}



func NewLoggerInit(interface{}) (Logger, error) {
	return &LoggerInit{}, nil
}

func (l *LoggerInit) NewEntryStd() Entry {
	return &EntryStd{
		Logger:		l,
		Timestamp:	time.Now(),
	}
}

func (l *LoggerInit) Handle(e Entry) {
	l.data = append(l.data, e)
}

func (l *LoggerInit) NextHandler(logger Logger) {
	for _, e := range l.data {
		logger.Handle(e)
	}
	l.data = l.data[0:0]
}

func (*LoggerInit) SetLevel( LoggerLevel) {
	// Do nothing because Initialization logger does not process entries
}

func (*LoggerInit) SetFromat( LoggerHandleFunc) {
	// Do nothing because Initialization logger does not process entries
}

func (l *LoggerInit) WithField(key string, value interface{}) LogOut {
	return l.NewEntryStd().WithField(key, value)
}

func (l *LoggerInit) WithFields(fields Fields) LogOut {
	return l.NewEntryStd().WithFields(fields)
}

func (l *LoggerInit) Debug(args ...interface{}) {
	l.NewEntryStd().Debug(args...)
}

func (l *LoggerInit) Info(args ...interface{}) {
	l.NewEntryStd().Info(args...)
}

func (l *LoggerInit) Warning(args ...interface{}) {
	l.NewEntryStd().Warning(args...)
}

func (l *LoggerInit) Error(args ...interface{}) {
	l.NewEntryStd().Error(args...)
}

func (l *LoggerInit) Fatal(args ...interface{}) {
	l.NewEntryStd().Fatal(args...)
}

func (l *LoggerInit) GetName() string {
	return ComponentLoggerInitName
}

func (l *LoggerInit) Version() string {
	return ComponentLoggerInitVersion
}



func NewLoggerStd(arg interface{}) (Logger, error) {
	// New
	l := &LoggerStd {
		LoggerHandleFunc: LoggerHandleDefault,
		LoggerStdConfig:		&LoggerStdConfig{
			Std:	true,
			Level:	LogDebug,
			Format:	"default",
		},
	}
	// Set config
	if arg != nil {
		cp, ok := arg.(*LoggerStdConfig)
		if ok {
			l.LoggerStdConfig = cp
		}else{
			err := MapToStruct(arg, &l.LoggerStdConfig)
			if err != nil {
				return nil, err
			}
		}
		// set logger format func
		fn := ConfigLoadLoggerHandleFunc(l.LoggerStdConfig.Format)
		if fn != nil {
			l.LoggerHandleFunc = fn
		}
		
	}
	// Init
	if len(l.Path) == 0 {
		l.out = bufio.NewWriter(os.Stdout) 
	}else {
		file, err := os.OpenFile(l.Path, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		if l.Std {
			l.out = bufio.NewWriter(io.MultiWriter(os.Stdout, file)) 
		}else{
			l.out = bufio.NewWriter(file) 
		}
	}
	go func(){
		for range time.NewTicker(time.Millisecond * 50).C {
			l.out.Flush()
		}
	}()
//	l.SetFromat(NewLoggerHandleTemplate(`[{{.Timestamp.Format "Jan 02, 2006 15:04:05 UTC"}}] {{.Level}}: {{.Message}}`))
	return l, nil
}

func (l *LoggerStd) Handle(e Entry) {
	if e.GetLevel() >= l.Level {
//		fmt.Fprintln(l.out, string(l.LoggerFormatFunc(e)))
		// e.Fields = nil
		l.LoggerHandleFunc(l.out, e)
	}
}

func (l *LoggerStd) Set(key string, val interface{}) error {
	switch i := val.(type) {
	case LoggerHandleFunc:
		l.LoggerHandleFunc = i
	case LoggerLevel:
		l.Level = i
	}
	return nil
}

func (l *LoggerStd) SetLevel(level LoggerLevel) {
	l.Level = level
}

func (l *LoggerStd) SetFromat(fn LoggerHandleFunc) {
	l.LoggerHandleFunc = fn
}


func (l *LoggerStd) WithField(key string, value interface{}) LogOut {
	return NewEntryStd(l).WithField(key, value)
}

func (l *LoggerStd) WithFields(fields Fields) LogOut {
	return NewEntryStd(l).WithFields(fields)
}

func (l *LoggerStd) Debug(args ...interface{}) {
	NewEntryStd(l).Debug(args...)
}

func (l *LoggerStd) Info(args ...interface{}) {
	NewEntryStd(l).Info(args...)
}

func (l *LoggerStd) Warning(args ...interface{}) {
	NewEntryStd(l).Warning(args...)
}

func (l *LoggerStd) Error(args ...interface{}) {
	NewEntryStd(l).Error(args...)
}

func (l *LoggerStd) Fatal(args ...interface{}) {
	NewEntryStd(l).Fatal(args...)
}

func (l *LoggerStdConfig) GetName() string {
	return ComponentLoggerStdName
}

func (l *LoggerStdConfig) Version() string {
	return ComponentLoggerStdVersion
}


func NewLoggerMulti(i interface{}) (Logger, error) {
	lc, ok := i.(*LoggerMultiConfig)
	if !ok {
		return nil, fmt.Errorf("The LoggerMulti configuration parameter type is not a LoggerMultiConfig pointer.")
	}
	l := &LoggerMulti{LoggerMultiConfig: lc,}
	l.Loggers = make([]Logger, len(lc.Configs))
	var err error
	for i, c := range lc.Configs {
		name := GetComponetName(c)
		if len(name) == 0 {
			return nil, fmt.Errorf("LoggerMulti %dth creation parameter could not get the corresponding component name", i)
		}
		l.Loggers[i], err = NewLogger(name, c)
		if err != nil {
			return nil, fmt.Errorf("LoggerMulti %dth creation Error: %v", i, err)
		}
	}
	return l, nil
}

func (l *LoggerMulti) Handle(e Entry) {
	for _, log := range l.Loggers {
		log.Handle(e)
	}
}
func (l *LoggerMulti) WithField(key string, value interface{}) LogOut {
	return NewEntryStd(l).WithField(key, value)
}

func (l *LoggerMulti) WithFields(fields Fields) LogOut {
	return NewEntryStd(l).WithFields(fields)
}

func (l *LoggerMulti) Debug(args ...interface{}) {
	NewEntryStd(l).Debug(args...)
}

func (l *LoggerMulti) Info(args ...interface{}) {
	NewEntryStd(l).Info(args...)
}

func (l *LoggerMulti) Warning(args ...interface{}) {
	NewEntryStd(l).Warning(args...)
}

func (l *LoggerMulti) Error(args ...interface{}) {
	NewEntryStd(l).Error(args...)
}

func (l *LoggerMulti) Fatal(args ...interface{}) {
	NewEntryStd(l).Fatal(args...)
}

func (*LoggerMultiConfig) GetName() string {
	return ComponentLoggerMultiName
}

func (*LoggerMultiConfig) Version() string {
	return ComponentLoggerMultiVersion
}





func NewEntryStd(log Logger) Entry {
	return &EntryStd{
		Logger:		log,
		Timestamp:	time.Now(),
	}
}

// func (e *EntryStd) MarshalJSON() ([]byte, error) {	
// 	if e.Fields == nil {
// 		e.Fields = make(Fields)
// 	}
// 	e.Fields["level"] = e.Level
// 	e.Fields["timestamp"] = e.Timestamp
// 	e.Fields["message"] = e.Message
// 	return json.Marshal(e.Fields)
// }

func (e *EntryStd) GetLevel() LoggerLevel {
	return e.Level
}

func (e *EntryStd) Debug(args ...interface{}) {
	e.Level = 0
	e.Message = fmt.Sprint(args...)
	e.Logger.Handle(e)
}

func (e *EntryStd) Info(args ...interface{}) {
	// fmt.Println(string(LogFormatStacks(true)))
	e.Level = 1
	e.Message = fmt.Sprint(args...)
	e.Logger.Handle(e)
}

func (e *EntryStd) Warning(args ...interface{}) {
	e.Level = 2
	e.Message = fmt.Sprint(args...)
	e.Logger.Handle(e)
}

func (e *EntryStd) Error(args ...interface{}) {
	e.Level = 3
	e.Message = fmt.Sprint(args...)
	e.Logger.Handle(e)
}

func (e *EntryStd) Fatal(args ...interface{}) {
	e.Level = 4
	e.Message = fmt.Sprint(args...)
	e.Logger.Handle(e)
}

func (e *EntryStd) WithField(key string, value interface{}) LogOut {
	if e.Fields == nil {
		e.Fields = make(Fields)
	}
	e.Fields[key] = value
	return e
}

func (e *EntryStd) WithFields(fields Fields) LogOut {
	e.Fields = fields
	return e
} 




func NewEntryContext(ctx Context, log Logger) Entry {
	file, line := LogFormatFileLine(0)
	f := Fields{
		HeaderXRequestID:	ctx.GetHeader(HeaderXRequestID),
		"file":				file,
		"line":				line,
	}
	return &EntryContext{
		ctx:	ctx,
		Entry:		&EntryStd{
			Logger:		log,
			Timestamp:	time.Now(),
		},
		Fields:		f,
	}
}

func (e *EntryContext) Debug(args ...interface{}) {
	e.Entry.WithFields(e.Fields).Debug(args...)
}

func (e *EntryContext) Info(args ...interface{}) {
	e.Entry.WithFields(e.Fields).Info(args...)
}

func (e *EntryContext) Warning(args ...interface{}) {
	e.Entry.WithFields(e.Fields).Warning(args...)
}

func (e *EntryContext) Error(args ...interface{}) {
	e.Entry.WithFields(e.Fields).Error(args...)
}

func (e *EntryContext) Fatal(args ...interface{}) {
	e.Entry.WithFields(e.Fields).Fatal(args...)
	// 结束Context
	e.ctx.WriteHeader(500)
	e.ctx.WriteRender(map[string]string{
		"status":	"500",
		"x-request-id":	e.ctx.RequestID(),
	})
	e.ctx.End()
}

func (e *EntryContext) WithField(key string, value interface{}) LogOut {
	if e.Fields == nil {
		e.Fields = make(Fields)
	}
	e.Fields[key] = value
	return e
}

func (e *EntryContext) WithFields(fields Fields) LogOut {
	e.Fields = fields
	return e
} 




func NewLoggerHandleTemplate(str string) LoggerHandleFunc {
	tmpl, err := template.New("test").Parse(str)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return func(w io.Writer, e Entry) {
		err := tmpl.Execute(w, e)
		fmt.Println(err)
	}
}

func LoggerHandleJson(w io.Writer, e Entry) {
	json.NewEncoder(w).Encode(e)
}

func LoggerHandleJsonIndent(w io.Writer, e Entry) {
	en := json.NewEncoder(w)
	en.SetIndent("", "\t")
	en.Encode(e)
}

func LoggerHandleXml(w io.Writer, e Entry) {
	xml.NewEncoder(w).Encode(e)
}

func LoggerHandleDefault(w io.Writer, e Entry) {
	fmt.Fprintln(w, e)
}


func LogFormatFileLine(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return file, line
}

func LogFormatFileLineArray(depth int) []string {
	f, l := LogFormatFileLine(depth + 1)
	return []string{
		fmt.Sprintf("file=%s", f),
		fmt.Sprintf("line=%d", l),
	}
}

func LogFormatStacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}
