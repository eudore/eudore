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
	"sync/atomic"
	"time"
)

const (
	LogDebug LoggerLevel = iota
	LogInfo
	LogWarning
	LogError
	LogFatal
	numSeverity = 5
)

var (
	LogLevelString        = [5]string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}
	_              Logger = (*LoggerInit)(nil)
	_              Logger = (*LoggerStd)(nil)
)

type (
	// LoggerLevel 定义日志级别
	LoggerLevel int32
	Fields      map[string]interface{}
	// 日志输出接口
	Logout interface {
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
	Logger interface {
		Logout
		Sync() error
		SetLevel(LoggerLevel)
	}
	// LoggerInitHandler 定义初始日志处理器必要接口，使用新日志处理器处理当前记录的全部日志。
	LoggerInitHandler interface {
		NextHandler(Logger)
	}
	// LoggerInit the initial log processor only records the log. After setting the log processor,
	// it will forward the log of the current record to the new log processor for processing the log generated before the program is initialized.
	//
	// LoggerInit 初始日志处理器仅记录日志，再设置日志处理器后，
	// 会将当前记录的日志交给新日志处理器处理，用于处理程序初始化之前产生的日志。
	LoggerInit struct {
		level LoggerLevel
		data  []*entryInit
	}
	entryInit struct {
		Level   LoggerLevel `json:"level"`
		Fields  Fields      `json:"fields,omitempty"`
		Time    time.Time   `json:"time"`
		Message string      `json:"message,omitempty"`
	}
)

// NewLoggerInit 函数创建一个初始化日志处理器。
func NewLoggerInit() Logger {
	return &LoggerInit{}
}

func (l *LoggerInit) newEntry() *entryInit {
	entry := &entryInit{}
	entry.Time = time.Now()
	l.data = append(l.data, entry)
	return entry
}

// NextHandler 方法实现LoggerInitHandler接口，设置当然Logger的存储日志的处理者。
func (l *LoggerInit) NextHandler(logger Logger) {
	for _, e := range l.data {
		switch e.Level {
		case LogDebug:
			logger.WithFields(e.Fields).WithField("time", e.Time).Info(e.Message)
		case LogInfo:
			logger.WithFields(e.Fields).WithField("time", e.Time).Info(e.Message)
		case LogWarning:
			logger.WithFields(e.Fields).WithField("time", e.Time).Warning(e.Message)
		case LogError:
			logger.WithFields(e.Fields).WithField("time", e.Time).Error(e.Message)
		case LogFatal:
			logger.WithFields(e.Fields).WithField("time", e.Time).Fatal(e.Message)
		}
	}
	l.data = l.data[0:0]
}

// SetLevel 方法设置日志处理级别。
func (l *LoggerInit) SetLevel(level LoggerLevel) {
	l.level = level
}

// Sync 方法将
func (l *LoggerInit) Sync() error {
	// log, _ := NewLoggerStd(nil)
	// l.NextHandler(log)
	// return log.Sync()
	return nil
}

func (l *LoggerInit) WithField(key string, value interface{}) Logout {
	return l.newEntry().WithField(key, value)
}

func (l *LoggerInit) WithFields(fields Fields) Logout {
	return l.newEntry().WithFields(fields)
}

func (l *LoggerInit) Debug(args ...interface{}) {
	l.newEntry().Debug(args...)
}

func (l *LoggerInit) Info(args ...interface{}) {
	l.newEntry().Info(args...)
}

func (l *LoggerInit) Warning(args ...interface{}) {
	l.newEntry().Warning(args...)
}

func (l *LoggerInit) Error(args ...interface{}) {
	l.newEntry().Error(args...)
}

func (l *LoggerInit) Fatal(args ...interface{}) {
	l.newEntry().Fatal(args...)
}

func (l *LoggerInit) Debugf(format string, args ...interface{}) {
	l.newEntry().Debugf(format, args...)
}

func (l *LoggerInit) Infof(format string, args ...interface{}) {
	l.newEntry().Infof(format, args...)
}

func (l *LoggerInit) Warningf(format string, args ...interface{}) {
	l.newEntry().Warningf(format, args...)
}

func (l *LoggerInit) Errorf(format string, args ...interface{}) {
	l.newEntry().Errorf(format, args...)
}

func (l *LoggerInit) Fatalf(format string, args ...interface{}) {
	l.newEntry().Fatalf(format, args...)
}

func (e *entryInit) Debug(args ...interface{}) {
	e.Level = 0
	e.Message = fmt.Sprintln(args...)
	e.Message = e.Message[:len(e.Message)-1]
}

func (e *entryInit) Info(args ...interface{}) {
	e.Level = 1
	e.Message = fmt.Sprintln(args...)
	e.Message = e.Message[:len(e.Message)-1]
}

func (e *entryInit) Warning(args ...interface{}) {
	e.Level = 2
	e.Message = fmt.Sprintln(args...)
	e.Message = e.Message[:len(e.Message)-1]
}

func (e *entryInit) Error(args ...interface{}) {
	e.Level = 3
	e.Message = fmt.Sprintln(args...)
	e.Message = e.Message[:len(e.Message)-1]
}

func (e *entryInit) Fatal(args ...interface{}) {
	e.Level = 4
	e.Message = fmt.Sprintln(args...)
	e.Message = e.Message[:len(e.Message)-1]
}

func (e *entryInit) Debugf(format string, args ...interface{}) {
	e.Level = 0
	e.Message = fmt.Sprintf(format, args...)
}

func (e *entryInit) Infof(format string, args ...interface{}) {
	e.Level = 1
	e.Message = fmt.Sprintf(format, args...)
}

func (e *entryInit) Warningf(format string, args ...interface{}) {
	e.Level = 2
	e.Message = fmt.Sprintf(format, args...)
}

func (e *entryInit) Errorf(format string, args ...interface{}) {
	e.Level = 3
	e.Message = fmt.Sprintf(format, args...)
}

func (e *entryInit) Fatalf(format string, args ...interface{}) {
	e.Level = 4
	e.Message = fmt.Sprintf(format, args...)
}

func (e *entryInit) WithField(key string, value interface{}) Logout {
	if e.Fields == nil {
		e.Fields = make(Fields)
	}
	e.Fields[key] = value
	return e
}

func (e *entryInit) WithFields(fields Fields) Logout {
	e.Fields = fields
	return e
}

// Level type
func (l *LoggerLevel) String() string {
	return LogLevelString[atomic.LoadInt32((*int32)(l))]
}

func (l *LoggerLevel) MarshalText() (text []byte, err error) {
	text = []byte(l.String())
	return
}

func (l *LoggerLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	for i, s := range LogLevelString {
		if s == str {
			atomic.StoreInt32((*int32)(l), int32(i))
			return nil
		}
	}
	n, err := strconv.Atoi(str)
	if err == nil && n < 5 && n > -1 {
		atomic.StoreInt32((*int32)(l), int32(n))
		return nil
	}
	return fmt.Errorf("level UnmarshalText error")
}

func NewLoggerPrintFunc(log Logger) func(...interface{}) {
	return func(args ...interface{}) {
		if len(args) == 1 {
			err, ok := args[0].(error)
			if ok {
				log.Error(err)
				return
			}
		}
		log.Info(args...)
	}
}

func LogFormatFileLine(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		// slash := strings.LastIndex(file, "/")
		slash := strings.Index(file, "src")
		if slash >= 0 {
			file = file[slash+4:]
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
