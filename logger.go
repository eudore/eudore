/*
Logger

Logger定义通用日志处理接口

文件: logger.go
*/
package eudore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
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
	poolEntryStd          = sync.Pool{}
	_              Logger = (*LoggerInit)(nil)
	_              Logger = (*LoggerStd)(nil)
)

type (
	// 日志级别
	LoggerLevel int32
	LoggerTime  struct {
		Time   time.Time
		Format string
	}
	Fields map[string]interface{}
	// LoggerHandleFunc		func(io.Writer, Entry)
	// 日志输出接口
	LogOut interface {
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
		WithField(key string, value interface{}) LogOut
		WithFields(fields Fields) LogOut
	}
	// 日志处理器定义
	Logger interface {
		Component
		LogOut
		// HandleEntry(interface{})
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
		data []*entryInit
	}
	entryInit struct {
		Level   LoggerLevel `json:"level"`
		Fields  Fields      `json:"fields,omitempty"`
		Time    time.Time   `json:"time"`
		Message string      `json:"message,omitempty"`
	}

	// 标准日志处理实现，将日志输出到标准输出或者文件。
	//
	// 日志格式默认json，可以指定为模板格式。
	LoggerStd struct {
		Config *LoggerStdConfig
		out    *bufio.Writer
		pool   sync.Pool
		ticker *time.Ticker
		handle func(interface{})
	}
	LoggerStdConfig struct {
		Std        bool        `set:"std"`
		Path       string      `set:"path"`
		Level      LoggerLevel `set:"level"`
		Format     string      `set:"format" default:"json"`
		TimeFormat string      `set:"timeformat" default:"2006-01-02 15:04:05"`
	}
	// 标准日志条目
	entryStd struct {
		pool       *sync.Pool
		logger     *LoggerStd
		checklevel LoggerLevel
		Time       *LoggerTime `json:"time"`
		Level      LoggerLevel `json:"level"`
		Fields     Fields      `json:"fields,omitempty"`
		Message    string      `json:"message,omitempty"`
	}
)

func init() {
	poolEntryStd = sync.Pool{
		New: func() interface{} {
			return &entryStd{}
		},
	}
}

func NewLogger(name string, arg interface{}) (Logger, error) {
	name = ComponentPrefix(name, "logger")
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

func NewLoggerStd(arg interface{}) (Logger, error) {
	// 解析配置
	config := &LoggerStdConfig{
		Format:     "json",
		TimeFormat: "2006-01-02 15:04:05",
	}
	ConvertTo(arg, config)

	// 创建并初始化日志处理器
	l := &LoggerStd{
		Config: config,
		pool:   sync.Pool{},
	}
	l.initPool()
	if err := l.initOut(); err != nil {
		return nil, err
	}
	if err := l.initHandle(); err != nil {
		return nil, err
	}

	// 定时写入日志
	go func() {
		l.ticker = time.NewTicker(time.Millisecond * 50)
		for range l.ticker.C {
			l.out.Flush()
		}
	}()
	return l, nil
}

func (l *LoggerStd) initPool() {
	l.pool.New = func() interface{} {
		return &entryStd{
			pool:       &l.pool,
			logger:     l,
			checklevel: l.Config.Level,
			Time: &LoggerTime{
				Format: l.Config.TimeFormat,
			},
		}
	}
}

func (l *LoggerStd) initOut() error {
	if len(l.Config.Path) == 0 {
		l.out = bufio.NewWriter(os.Stdout)
	} else {
		file, err := os.OpenFile(l.Config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		if l.Config.Std {
			l.out = bufio.NewWriter(io.MultiWriter(os.Stdout, file))
		} else {
			l.out = bufio.NewWriter(file)
		}
	}
	return nil
}

func (l *LoggerStd) initHandle() error {
	if l.out == nil {
		return fmt.Errorf("logger out is nil")
	}
	if l.Config.Format == "json" {
		handle := json.NewEncoder(l.out)
		l.handle = func(i interface{}) {
			handle.Encode(i)
		}
	} else {
		tmpl, err := template.New("").Parse(l.Config.Format)
		if err != nil {
			return err
		}
		l.handle = func(i interface{}) {
			tmpl.Execute(l.out, i)
		}
	}
	return nil
}

func (l *LoggerStd) Flush() error {
	return l.out.Flush()
}

func (l *LoggerStd) Set(key string, val interface{}) error {
	switch i := val.(type) {
	// case LoggerHandleFunc:
	// l.LoggerHandleFunc = i
	case LoggerLevel:
		l.Config.Level = i
	default:
		return ErrComponentNoSupportField
	}
	return nil
}

func (l *LoggerStd) HandleEntry(e interface{}) {
	l.handle(e)
}

func (l *LoggerStd) newEntry() (entry *entryStd) {
	entry = l.pool.Get().(*entryStd)
	entry.Time.Time = time.Now()
	entry.Fields = nil
	return
}

func (l *LoggerStd) WithField(key string, value interface{}) LogOut {
	return l.newEntry().WithField(key, value)
}

func (l *LoggerStd) WithFields(fields Fields) LogOut {
	return l.newEntry().WithFields(fields)
}

func (l *LoggerStd) Debug(args ...interface{}) {
	l.newEntry().Debug(args...)
}

func (l *LoggerStd) Info(args ...interface{}) {
	l.newEntry().Info(args...)
}

func (l *LoggerStd) Warning(args ...interface{}) {
	l.newEntry().Warning(args...)
}

func (l *LoggerStd) Error(args ...interface{}) {
	l.newEntry().Error(args...)
}

func (l *LoggerStd) Fatal(args ...interface{}) {
	l.newEntry().Fatal(args...)
}

func (l *LoggerStd) Debugf(format string, args ...interface{}) {
	l.newEntry().Debugf(format, args...)
}

func (l *LoggerStd) Infof(format string, args ...interface{}) {
	l.newEntry().Infof(format, args...)
}

func (l *LoggerStd) Warningf(format string, args ...interface{}) {
	l.newEntry().Warningf(format, args...)
}

func (l *LoggerStd) Errorf(format string, args ...interface{}) {
	l.newEntry().Errorf(format, args...)
}

func (l *LoggerStd) Fatalf(format string, args ...interface{}) {
	l.newEntry().Fatalf(format, args...)
}

func (l *LoggerStd) GetName() string {
	return ComponentLoggerStdName
}

func (l *LoggerStd) Version() string {
	return ComponentLoggerStdVersion
}

func (l *LoggerStdConfig) GetName() string {
	return ComponentLoggerStdName
}

func (e *entryStd) Debug(args ...interface{}) {
	if e.checklevel < 1 {
		e.Level = 0
		e.Message = fmt.Sprintln(args...)
		e.Message = e.Message[:len(e.Message)-1]
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Info(args ...interface{}) {
	if e.checklevel < 2 {
		e.Level = 1
		e.Message = fmt.Sprintln(args...)
		e.Message = e.Message[:len(e.Message)-1]
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Warning(args ...interface{}) {
	if e.checklevel < 3 {
		e.Level = 2
		e.Message = fmt.Sprint(args...)
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Error(args ...interface{}) {
	if e.checklevel < 4 {
		e.Level = 3
		e.Message = fmt.Sprint(args...)
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Fatal(args ...interface{}) {
	e.Level = 4
	e.Message = fmt.Sprint(args...)
	e.logger.HandleEntry(e)
	e.pool.Put(e)
	panic(args)
}

func (e *entryStd) Debugf(format string, args ...interface{}) {
	if e.checklevel < 1 {
		e.Level = 0
		e.Message = fmt.Sprintf(format, args...)
		e.Message = e.Message[:len(e.Message)-1]
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Infof(format string, args ...interface{}) {
	if e.checklevel < 2 {
		e.Level = 1
		e.Message = fmt.Sprintf(format, args...)
		e.Message = e.Message[:len(e.Message)-1]
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Warningf(format string, args ...interface{}) {
	if e.checklevel < 3 {
		e.Level = 2
		e.Message = fmt.Sprintf(format, args...)
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Errorf(format string, args ...interface{}) {
	if e.checklevel < 4 {
		e.Level = 3
		e.Message = fmt.Sprintf(format, args...)
		e.logger.HandleEntry(e)
	}
	e.pool.Put(e)
}

func (e *entryStd) Fatalf(format string, args ...interface{}) {
	e.Level = 4
	e.Message = fmt.Sprintf(format, args...)
	e.logger.HandleEntry(e)
	e.pool.Put(e)
	panic(args)
}

func (e *entryStd) WithField(key string, value interface{}) LogOut {
	if e.Fields == nil {
		e.Fields = make(Fields, 3)
	}
	if key == "time" {
		var ok bool
		e.Time.Time, ok = value.(time.Time)
		if ok {
			return e
		}
	}
	e.Fields[key] = value
	return e
}

func (e *entryStd) WithFields(fields Fields) LogOut {
	e.Fields = fields
	return e
}

func NewLoggerInit(interface{}) (Logger, error) {
	return &LoggerInit{}, nil
}

func (l *LoggerInit) newEntry() *entryInit {
	entry := &entryInit{}
	entry.Time = time.Now()
	l.data = append(l.data, entry)
	return entry
}

func (l *LoggerInit) HandleEntry(e interface{}) {
	// Do nothing because of LoggerInit not handler entry.
}

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

func (l *LoggerInit) WithField(key string, value interface{}) LogOut {
	return l.newEntry().WithField(key, value)
}

func (l *LoggerInit) WithFields(fields Fields) LogOut {
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

func (l *LoggerInit) GetName() string {
	return ComponentLoggerInitName
}

func (l *LoggerInit) Version() string {
	return ComponentLoggerInitVersion
}

func (e *entryInit) Debug(args ...interface{}) {
	e.Level = 0
	e.Message = fmt.Sprint(args...)
}

func (e *entryInit) Info(args ...interface{}) {
	e.Level = 1
	e.Message = fmt.Sprint(args...)
}

func (e *entryInit) Warning(args ...interface{}) {
	e.Level = 2
	e.Message = fmt.Sprint(args...)
}

func (e *entryInit) Error(args ...interface{}) {
	e.Level = 3
	e.Message = fmt.Sprint(args...)
}

func (e *entryInit) Fatal(args ...interface{}) {
	e.Level = 4
	e.Message = fmt.Sprint(args...)
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

func (e *entryInit) WithField(key string, value interface{}) LogOut {
	if e.Fields == nil {
		e.Fields = make(Fields)
	}
	e.Fields[key] = value
	return e
}

func (e *entryInit) WithFields(fields Fields) LogOut {
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

func (t *LoggerTime) String() string {
	return t.Time.Format(t.Format)
}

func (t *LoggerTime) MarshalText() (text []byte, err error) {
	text = []byte(t.String())
	return
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
