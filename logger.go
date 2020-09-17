package eudore

/*
Logger

Logger定义通用日志处理接口

文件: logger.go loggerstd.go
*/

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LoggerLevel 定义日志级别
type LoggerLevel int32

// Fields 定义多个日志属性
type Fields map[string]interface{}

// Logout 日志输出接口
type Logout interface {
	Debug(...interface{})
	Info(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	WithField(key string, value interface{}) Logout
	WithFields(fields Fields) Logout
}

// Logger 定义日志处理器定义
type Logger interface {
	Logout
	Sync() error
	SetLevel(LoggerLevel)
}

// loggerInitHandler 定义初始日志处理器必要接口，使用新日志处理器处理当前记录的全部日志。
type loggerInitHandler interface {
	NextHandler(Logger)
}

// LoggerInit the initial log processor only records the log. After setting the log processor,
// it will forward the log of the current record to the new log processor for processing the log generated before the program is initialized.
//
// LoggerInit 初始日志处理器仅记录日志，再设置日志处理器后，
// 会将当前记录的日志交给新日志处理器处理，用于处理程序初始化之前产生的日志。
type LoggerInit struct {
	data  []*entryInit
	Mutex sync.Mutex
	*entryInit
}
type entryInit struct {
	logger  *LoggerInit
	level   LoggerLevel
	time    time.Time
	message string
	fields  Fields
	logout  bool
}

// 定义日志级别
const (
	LogDebug LoggerLevel = iota
	LogInfo
	LogWarning
	LogError
	LogFatal
	logSetLevel
)

// NewLoggerInit 函数创建一个初始化日志处理器。
func NewLoggerInit() Logger {
	log := &LoggerInit{}
	log.entryInit = &entryInit{
		logger: log,
		logout: true,
	}
	return log
}

// NextHandler 方法实现loggerInitHandler接口。
func (log *LoggerInit) NextHandler(logger Logger) {
	logout := logger.WithField("depth", "disable")
	for _, entry := range log.data {
		switch entry.level {
		case LogDebug:
			logout.WithFields(entry.fields).WithField("time", entry.time).Debug(entry.message)
		case LogInfo:
			logout.WithFields(entry.fields).WithField("time", entry.time).Info(entry.message)
		case LogWarning:
			logout.WithFields(entry.fields).WithField("time", entry.time).Warning(entry.message)
		case LogError:
			logout.WithFields(entry.fields).WithField("time", entry.time).Error(entry.message)
		case LogFatal:
			logout.WithFields(entry.fields).WithField("time", entry.time).Fatal(entry.message)
		case logSetLevel:
			logger.SetLevel(entry.fields["level"].(LoggerLevel))
		}
	}
	logger.Sync()
	log.data = log.data[0:0]
}

// SetLevel 方法设置日志处理级别。
func (log *LoggerInit) SetLevel(level LoggerLevel) {
	entry := log.newEntry()
	entry.level = logSetLevel
	entry.WithField("level", level)
	entry.putEntry()
}

// Sync 方法
func (log *LoggerInit) Sync() error {
	return nil
}

func (entry *entryInit) newEntry() *entryInit {
	newentry := &entryInit{
		logger: entry.logger,
		time:   time.Now(),
	}
	if entry.fields != nil {
		newentry.fields = make(Fields, len(entry.fields))
	}
	for k, v := range entry.fields {
		newentry.fields[k] = v
	}
	return newentry
}

func (entry *entryInit) putEntry() {
	entry.logger.Mutex.Lock()
	entry.logger.data = append(entry.logger.data, entry)
	entry.logger.Mutex.Unlock()
}

// Debug 方法输出Debug级别日志。
func (entry *entryInit) Debug(args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 0
	entry.message = fmt.Sprintln(args...)
	entry.message = entry.message[:len(entry.message)-1]
	entry.putEntry()
}

// Info 方法输出Info级别日志。
func (entry *entryInit) Info(args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 1
	entry.message = fmt.Sprintln(args...)
	entry.message = entry.message[:len(entry.message)-1]
	entry.putEntry()
}

// Warning 方法输出Warning级别日志。
func (entry *entryInit) Warning(args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 2
	entry.message = fmt.Sprintln(args...)
	entry.message = entry.message[:len(entry.message)-1]
	entry.putEntry()
}

// Error 方法输出Error级别日志。
func (entry *entryInit) Error(args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 3
	entry.message = fmt.Sprintln(args...)
	entry.message = entry.message[:len(entry.message)-1]
	entry.putEntry()
}

// Fatal 方法输出Fatal级别日志。
func (entry *entryInit) Fatal(args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 4
	entry.message = fmt.Sprintln(args...)
	entry.message = entry.message[:len(entry.message)-1]
	entry.putEntry()
}

// Debugf 方法格式化输出Debug级别日志。
func (entry *entryInit) Debugf(format string, args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 0
	entry.message = fmt.Sprintf(format, args...)
	entry.putEntry()
}

// Infof 方法格式化输出Info级别日志。
func (entry *entryInit) Infof(format string, args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 1
	entry.message = fmt.Sprintf(format, args...)
	entry.putEntry()
}

// Warningf 方法格式化输出Warning级别日志。
func (entry *entryInit) Warningf(format string, args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 2
	entry.message = fmt.Sprintf(format, args...)
	entry.putEntry()
}

// Errorf 方法格式化输出Error级别日志。
func (entry *entryInit) Errorf(format string, args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 3
	entry.message = fmt.Sprintf(format, args...)
	entry.putEntry()
}

// Fatalf 方法格式化输出Fatal级别日志。
func (entry *entryInit) Fatalf(format string, args ...interface{}) {
	if entry.logout {
		entry = entry.newEntry()
	}
	entry.level = 4
	entry.message = fmt.Sprintf(format, args...)
	entry.putEntry()
}

// WithField 方法给日志新增一个属性。
func (entry *entryInit) WithField(key string, value interface{}) Logout {
	if entry.logout {
		entry = entry.newEntry()
	}
	if key == "depth" {
		return entry
	}
	if entry.fields == nil {
		entry.fields = make(Fields)
	}
	entry.fields[key] = value
	return entry
}

// WithFields 方法给日志新增多个属性。
func (entry *entryInit) WithFields(fields Fields) Logout {
	if fields == nil {
		entry = entry.newEntry()
		entry.logout = true
		return entry
	}
	if entry.logout {
		entry = entry.newEntry()
	}
	if entry.fields == nil {
		entry.fields = make(Fields)
	}
	for k, v := range fields {
		entry.fields[k] = v
	}
	return entry
}

// String 方法实现ftm.Stringer接口，格式化输出日志级别。
func (l LoggerLevel) String() string {
	return LogLevelString[l]
}

// MarshalText 方法实现encoding.TextMarshaler接口，用于编码日志级别。
func (l LoggerLevel) MarshalText() (text []byte, err error) {
	text = []byte(l.String())
	return
}

// UnmarshalText 方法实现encoding.TextUnmarshaler接口，用于解码日志级别。
func (l *LoggerLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	for i, s := range LogLevelString {
		if s == str {
			*l = LoggerLevel(i)
			return nil
		}
	}
	n, err := strconv.Atoi(str)
	fmt.Println(n, err)
	if err == nil && n < 5 && n > -1 {
		*l = LoggerLevel(n)
		return nil
	}
	return ErrLoggerLevelUnmarshalText
}

// NewPrintFunc 函数使用Logout创建一个输出函数。
//
// 如果第一个参数Fields类型，则调用WithFields方法。
//
// 如果参数是一个error则输出error级别日志，否在输出info级别日志。
func NewPrintFunc(log Logout) func(...interface{}) {
	log = log.WithField("depth", 2).WithFields(nil)
	return func(args ...interface{}) {
		fields, ok := args[0].(Fields)
		if ok {
			printLogout(log.WithFields(fields), args[1:])
		} else {
			printLogout(log, args)
		}
	}
}

func printLogout(log Logout, args []interface{}) {
	if len(args) == 1 {
		err, ok := args[0].(error)
		if ok {
			log.Error(err)
			return
		}
	}
	log.Info(args...)
}

// logFormatNameFileLine 函数获得调用的文件位置和函数名称。
//
// 文件位置会从第一个src后开始截取，处理gopath下文件位置。
func logFormatNameFileLine(depth int) (string, string, int) {
	var name string
	ptr, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		// slash := strings.LastIndex(file, "/")
		slash := strings.Index(file, "src")
		if slash >= 0 {
			file = file[slash+4:]
		}
		name = runtime.FuncForPC(ptr).Name()
	}
	return name, file, line
}

// GetPanicStack 函数返回panic栈信息。
func GetPanicStack(depth int) []string {
	pc := make([]uintptr, DefaultRecoverDepth)
	n := runtime.Callers(depth, pc)
	if n == 0 {
		return nil
	}

	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)
	stack := make([]string, 0, DefaultRecoverDepth)

	frame, more := frames.Next()
	for more {
		pos := strings.Index(frame.File, "src")
		if pos >= 0 {
			frame.File = frame.File[pos+4:]
		}
		pos = strings.LastIndex(frame.Function, "/")
		if pos >= 0 {
			frame.Function = frame.Function[pos+1:]
		}
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))

		frame, more = frames.Next()
	}
	return stack
}

func printEmpty(...interface{}) {
	// Do nothing because  not print message.
}
