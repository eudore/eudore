package eudore

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 枚举使用的日志级别。
const (
	LoggerDebug LoggerLevel = iota
	LoggerInfo
	LoggerWarning
	LoggerError
	LoggerFatal
	LoggerDiscard
)

/*
Logger defines a log output interface to implement the following functions:

	Five-level log format output
	Log entries with Fields attribute
	json/text ordered formatted output
	Custom processing Hook
	Expression filter log
	Initialize log processing
	Standard output stream displays colored Level
	Set file line information output
	Log file soft link
	log file rollover
	log file cleanup

Logger 定义日志输出接口实现下列功能:

	五级日志格式化输出
	日志条目带Fields属性
	json/text有序格式化输出
	自定义处理Hook
	表达式过滤日志
	初始化日志处理
	标准输出流显示彩色Level
	设置文件行信息输出
	日志文件软连接
	日志文件滚动
	日志文件清理
*/
type Logger interface {
	Debug(...any)
	Info(...any)
	Warning(...any)
	Error(...any)
	Fatal(...any)
	Debugf(string, ...any)
	Infof(string, ...any)
	Warningf(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
	WithField(string, any) Logger
	WithFields([]string, []any) Logger
	GetLevel() LoggerLevel
	SetLevel(LoggerLevel)
}

// LoggerLevel 定义日志级别。
type LoggerLevel int32

// loggerStd 定义日志默认实现条目信息。
type loggerStd struct {
	LoggerEntry
	Handlers []LoggerHandler
	Pool     *sync.Pool
	Logger   bool
	Depth    int32
}

// LoggerEntry 定义日志条目数据对象。
type LoggerEntry struct {
	Level   LoggerLevel
	Time    time.Time
	Message string
	Keys    []string
	Vals    []any
	Buffer  []byte
}

// LoggerHandler 定义LoggerEntry处理方法
//
// HandlerPriority 方法返回Handler处理顺序，小值优先。
//
// HandlerEntry 方法处理Entry内容，设置Level=LoggerDiscard后结束后续处理。
type LoggerHandler interface {
	HandlerPriority() int
	HandlerEntry(*LoggerEntry)
}

// LoggerConfig 定义loggerStd配置信息。
type LoggerConfig struct {
	// 设置额外的LoggerHandler，和配置初始化创建的Handlers排序后处理LoggerEntry。
	Handlers []LoggerHandler `alias:"handlers" json:"-" xml:"-" yaml:"-"`
	// 设置日志输出级别。
	Level LoggerLevel `alias:"level" json:"level" xml:"level" yaml:"level"`
	// 是否记录调用者信息。
	Caller bool `alias:"caller" json:"caller" xml:"caller" yaml:"caller"`
	// 设置Entry输出格式，默认值为json，
	// 如果为json/text启用NewLoggerFormatterJSON/NewLoggerFormatterText。
	Formatter string `alias:"formater" json:"formater" xml:"formater" yaml:"formater"`
	// 设置日志时间输出格式，默认值为DefaultLoggerFormatterFormatTime或time.RFC3339。
	TimeFormat string `alias:"timeformat" json:"timeformat" xml:"timeformat" yaml:"timeformat"`
	// 设置Entry过滤规则；如果非空启用NewLoggerHookFilter。
	HookFilter [][]string `alias:"hoolfilter" json:"hoolfilter" xml:"hoolfilter" yaml:"hoolfilter"`
	// 是否处理Fatal级别日志，调用应用结束方法；如果为true启用NewLoggerHookMeta。
	HookFatal bool `alias:"hookfatal" json:"hookfatal" xml:"hookfatal" yaml:"hookfatal"`
	// 是否采集Meta信息，记录日志count、size；如果为true启用NewLoggerHookFatal。
	HookMeta bool `alias:"hookmeta" json:"hookmeta" xml:"hookmeta" yaml:"hookmeta"`
	// 是否输出日志到os.Stdout标准输出流；如果存在Env EnvEudoreDaemonEnable时会强制修改为false；
	// 如果为true启动NewLoggerWriterStdout。
	Stdout bool `alias:"stdout" json:"stdout" xml:"stdout" yaml:"stdout"`
	// 是否输出日志时使用彩色Level，默认在windows系统下禁用。
	StdColor bool `alias:"stdcolor" json:"stdcolor" xml:"stdcolor" yaml:"stdcolor"`
	// 设置日志文件输出路径；如果非空启用NewLoggerWriterFile，
	// 如果Path包含关键字yyyy/mm/dd/hh或MaxSize非0则改为启用NewLoggerWriterRotate。
	Path string `alias:"path" json:"path" xml:"path" yaml:"path" description:"Output file path."`
	// 设置日志文件滚动size，在文件名后缀之前添加索引值。
	MaxSize uint64 `alias:"maxsize" json:"maxsize" xml:"maxsize" yaml:"maxsize" description:"roatte file max size"`
	// 设置日志文件最多保留天数，如果非0使用hookFileRecycle。
	MaxAge int `alias:"maxage" json:"maxage" xml:"maxage" yaml:"maxage"`
	// 设置日志文件最多保留数量，如果非0使用hookFileRecycle。
	MaxCount int `alias:"maxcount" json:"maxcount" xml:"maxcount" yaml:"maxcount"`
	// 设置日志文件软链接名称，如果非空使用hookFileLink。
	Link string `alias:"link" json:"link" xml:"link" yaml:"link" description:"Output file link to path."`
}

type MetadataLogger struct {
	Health     bool      `alias:"health" json:"health" xml:"health" yaml:"health"`
	Name       string    `alias:"name" json:"name" xml:"name" yaml:"name"`
	Count      [5]uint64 `alias:"count" json:"count" xml:"count" yaml:"count"`
	Size       uint64    `alias:"size" json:"size" xml:"size" yaml:"size"`
	SizeFormat string    `alias:"sizeformat" json:"sizeformat" xml:"sizeformat" yaml:"sizeformat"`
}

/*
NewLogger 创建一个标准日志处理器。

默认配置:

	&LoggerConfig{
		Stdout:    true,
		StdColor:  DefaultLoggerEnableStdColor,
		HookFatal: DefaultLoggerEnableHookFatal,
		HookMeta:  DefaultLoggerEnableHookMeta,
	}
*/
func NewLogger(config *LoggerConfig) Logger {
	if config == nil {
		config = &LoggerConfig{
			Stdout:    true,
			StdColor:  DefaultLoggerEnableStdColor,
			HookFatal: DefaultLoggerEnableHookFatal,
			HookMeta:  DefaultLoggerEnableHookMeta,
		}
	}

	handlers := config.getHandlers()
	pool := &sync.Pool{}
	pool.New = func() any {
		return &loggerStd{
			Pool:     pool,
			Handlers: handlers,
			LoggerEntry: LoggerEntry{
				Level:  config.Level,
				Keys:   make([]string, 0, DefaultLoggerEntryFieldsLength),
				Vals:   make([]any, 0, DefaultLoggerEntryFieldsLength),
				Buffer: make([]byte, 0, DefaultLoggerEntryBufferLength),
			},
		}
	}

	log := pool.New().(*loggerStd)
	log.Logger = true
	log.Depth = 4
	if config.Caller {
		log.Depth |= 0x100
	}
	return log
}

func (config *LoggerConfig) getHandlers() []LoggerHandler {
	config.TimeFormat = GetAnyByString(config.TimeFormat, DefaultLoggerFormatterFormatTime, time.RFC3339)
	config.Formatter = GetAnyByString(config.Formatter, DefaultLoggerFormatter, "json")
	config.Stdout = config.Stdout && !GetAnyByString[bool](os.Getenv(EnvEudoreDaemonEnable))
	config.StdColor = config.StdColor && (runtime.GOOS != "windows" || DefaultLoggerWriterStdoutWindowsColor)
	config.Path = strings.TrimSpace(config.Path)

	hs := config.Handlers
	// formatter
	switch strings.ToLower(config.Formatter) {
	case "json":
		hs = append(hs, NewLoggerFormatterJSON(config.TimeFormat))
	case "text":
		hs = append(hs, NewLoggerFormatterText(config.TimeFormat))
	}

	// hook
	if len(config.HookFilter) > 0 {
		hs = append(hs, NewLoggerHookFilter(config.HookFilter))
	}
	if config.HookMeta {
		hs = append(hs, NewLoggerHookMeta())
	}
	if config.HookFatal {
		hs = append(hs, NewLoggerHookFatal(nil))
	}

	// writer-stdout
	if config.Stdout {
		hs = append(hs, NewLoggerWriterStdout(config.StdColor))
	}
	// writer-rotate
	if config.Path != "" {
		var hook []func(string, string)
		if config.Link != "" {
			hook = append(hook, hookFileLink(config.Link))
		}
		if config.MaxAge > 0 || config.MaxCount > 1 {
			hook = append(hook, hookFileRecycle(config.MaxAge, config.MaxCount))
		}
		h, err := NewLoggerWriterRotate(config.Path, config.MaxSize, hook...)
		if err != nil {
			panic(err)
		}
		hs = append(hs, h)
	}

	sort.Slice(hs, func(i, j int) bool {
		return hs[i].HandlerPriority() < hs[j].HandlerPriority()
	})
	return hs
}

// NewLoggerInit The initial log processor only records logs, and gets a new Logger to process logs when Unmount.
//
// If the subsequent Logger is not set after LoggerInit is set, App.Run() must be called to release the log in LoggerInit.
//
// NewLoggerInit 初始日志处理器仅记录日志，在Unmount时获取新Logger处理日志.
//
// 在设置LoggerInit后未设置后续Logger，必须调用App.Run()将LoggerInit内日志释放出来。
func NewLoggerInit() Logger {
	return NewLogger(&LoggerConfig{
		Handlers:  []LoggerHandler{&loggerHandlerInit{}},
		Formatter: "disable",
		HookMeta:  true,
	})
}

// NewLoggerNull 定义空日志输出，丢弃所有日志。
func NewLoggerNull() Logger {
	return NewLogger(&LoggerConfig{
		Level:     LoggerDiscard,
		Formatter: "disable",
	})
}

// NewLoggerWithContext 方法从环境上下文ContextKeyLogger获取Logger，如果无法获取Logger返回DefaultLoggerNull对象。
func NewLoggerWithContext(ctx context.Context) Logger {
	log, ok := ctx.Value(ContextKeyLogger).(Logger)
	if ok {
		return log
	}
	return DefaultLoggerNull
}

// Mount 方法使LoggerStd挂载上下文，上下文传递给LoggerStdData。
func (log *loggerStd) Mount(ctx context.Context) {
	for i := range log.Handlers {
		anyMount(ctx, log.Handlers[i])
	}
}

// Unmount 方法使LoggerStd卸载上下文，上下文传递给LoggerStdData。
func (log *loggerStd) Unmount(ctx context.Context) {
	for i := len(log.Handlers) - 1; i > -1; i-- {
		anyUnmount(ctx, log.Handlers[i])
	}
}

// Metadata 方法从Handlers查找到第一个anyMetadata对象返回meta。
func (log *loggerStd) Metadata() any {
	for i := range log.Handlers {
		meta := anyMetadata(log.Handlers[i])
		if meta != nil {
			return meta
		}
	}
	return nil
}

// GetLevel 方法获取当前日志输出级别，判断级别取消日志生成。
func (log *loggerStd) GetLevel() LoggerLevel {
	return log.Level
}

// SetLevel 方法设置当前日志输出级别。
func (log *loggerStd) SetLevel(level LoggerLevel) {
	log.Level = level
}

// Debug 方法条目输出Debug级别日志。
func (log *loggerStd) Debug(args ...any) {
	log.format(LoggerDebug, args...)
}

// Info 方法条目输出Info级别日志。
func (log *loggerStd) Info(args ...any) {
	log.format(LoggerInfo, args...)
}

// Warning 方法条目输出Warning级别日志。
func (log *loggerStd) Warning(args ...any) {
	log.format(LoggerWarning, args...)
}

// Error 方法条目输出Error级别日志。
func (log *loggerStd) Error(args ...any) {
	log.format(LoggerError, args...)
}

// Fatal 方法条目输出Fatal级别日志。
func (log *loggerStd) Fatal(args ...any) {
	log.format(LoggerFatal, args...)
}

// Debugf 方法格式化写入流Debug级别日志。
func (log *loggerStd) Debugf(format string, args ...any) {
	log.formatf(LoggerDebug, format, args...)
}

// Infof 方法格式写入流出Info级别日志。
func (log *loggerStd) Infof(format string, args ...any) {
	log.formatf(LoggerInfo, format, args...)
}

// Warningf 方法格式化输出写入流Warning级别日志。
func (log *loggerStd) Warningf(format string, args ...any) {
	log.formatf(LoggerWarning, format, args...)
}

// Errorf 方法格式化写入流Error级别日志。
func (log *loggerStd) Errorf(format string, args ...any) {
	log.formatf(LoggerError, format, args...)
}

// Fatalf 方法格式化写入流Fatal级别日志。
func (log *loggerStd) Fatalf(format string, args ...any) {
	log.formatf(LoggerFatal, format, args...)
}

// WithFields 方法一次设置多个属性，但是不会设置Field属性。
func (log *loggerStd) WithFields(key []string, value []any) Logger {
	if log.Logger {
		log = log.getLogger()
	}
	log.Keys = append(log.Keys, key...)
	log.Vals = append(log.Vals, value...)
	return log
}

// WithField 方法设置一个日志属性，指定key时会执行特色行为。
//
// 如果key为"logger"值为bool(true)，将LoggerEntry设置为Logger。
//
// 如果key为"depth"值类型为int，设置日志调用堆栈增删层数;
// 如果key为"depth"值类型为string值"enable"或"disable"，启用或关闭日志调用位置输出;
// 并增加key: file/func/stack，如果使用到相关key需要先禁用depth。
//
// 如果key为"time"值类型为time.time，设置日志输出的时间属性。
func (log *loggerStd) WithField(key string, value any) Logger {
	if log.Logger {
		log = log.getLogger()
	}
	switch key {
	case "logger":
		val, ok := value.(bool)
		if ok && val {
			log.Logger = true
			return log
		}
	case ParamDepth:
		return log.withFieldDepth(key, value)
	case "time":
		val, ok := value.(time.Time)
		if ok {
			log.Time = val
			return log
		}
	}
	log.Keys = append(log.Keys, key)
	log.Vals = append(log.Vals, value)
	return log
}

// withFieldDepth 方法处理withDepth属性，cost 53 可内联。
func (log *loggerStd) withFieldDepth(key string, value any) Logger {
	switch val := value.(type) {
	case int:
		log.Depth += int32(val)
	case string:
		switch val {
		case "enable":
			log.Depth |= 0x100
		case "stack":
			log.Depth |= 0x200
		case "disable":
			log.Depth &^= 0x300
		}
	default:
		log.Keys = append(log.Keys, key)
		log.Vals = append(log.Vals, value)
	}
	return log
}

func (log *loggerStd) getLogger() *loggerStd {
	entry := log.Pool.Get().(*loggerStd)
	entry.Time = time.Now()
	entry.Message = ""
	entry.Keys = entry.Keys[0:0]
	entry.Vals = entry.Vals[0:0]
	entry.Buffer = entry.Buffer[0:0]
	entry.Level = log.Level
	entry.Depth = log.Depth
	if len(log.Keys) > 0 {
		entry.Keys = append(entry.Keys, log.Keys...)
		entry.Vals = append(entry.Vals, log.Vals...)
	}
	return entry
}

func (log *loggerStd) format(level LoggerLevel, args ...any) {
	if log.Level <= level {
		if log.Logger {
			log = log.getLogger()
		}
		log.Level = level
		log.Message = fmt.Sprintln(args...)
		log.Message = log.Message[:len(log.Message)-1]
		log.handler()
		log.Pool.Put(log)
	}
}

func (log *loggerStd) formatf(level LoggerLevel, format string, args ...any) {
	if log.Level <= level {
		if log.Logger {
			log = log.getLogger()
		}
		log.Level = level
		log.Message = fmt.Sprintf(format, args...)
		log.handler()
		log.Pool.Put(log)
	}
}

func (log *loggerStd) handler() {
	if len(log.Keys) > len(log.Vals) {
		log.Keys = log.Keys[0:len(log.Vals)]
		log.Keys = append(log.Keys, "loggererr")
		log.Vals = append(log.Vals, "LoggerStd: The number of field keys and values are not equal")
	}

	if len(log.Message) > 0 || len(log.Keys) > 0 {
		switch log.Depth >> 8 {
		case 1:
			fname, file := GetCallerFuncFile(int(log.Depth) & 0xff)
			if fname != "" {
				log.Keys = append(log.Keys, "func")
				log.Vals = append(log.Vals, fname)
			}
			if file != "" {
				log.Keys = append(log.Keys, "file")
				log.Vals = append(log.Vals, file)
			}
		case 2, 3:
			log.Keys = append(log.Keys, "stack")
			log.Vals = append(log.Vals, GetCallerStacks(int(log.Depth&0xff)+1))
		}
		for _, h := range log.Handlers {
			if log.Level < LoggerDiscard {
				h.HandlerEntry(&log.LoggerEntry)
			}
		}
	}
}

// String 方法实现ftm.Stringer接口，格式化输出日志级别。
func (l LoggerLevel) String() string {
	return DefaultLoggerLevelStrings[l]
}

// MarshalText 方法实现encoding.TextMarshaler接口，用于编码日志级别。
func (l LoggerLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText 方法实现encoding.TextUnmarshaler接口，用于解码日志级别。
func (l *LoggerLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	for i, s := range DefaultLoggerLevelStrings {
		if s == str {
			*l = LoggerLevel(i)
			return nil
		}
	}
	n, err := strconv.Atoi(str)
	if err == nil && n < 5 && n > -1 {
		*l = LoggerLevel(n)
		return nil
	}
	return ErrLoggerLevelUnmarshalText
}

var works = [...]string{"/pkg/mod/", "/src/"}

func trimFileName(name string) string {
	for _, w := range works {
		pos := strings.Index(name, w)
		if pos != -1 {
			name = name[pos+len(w):]
		}
	}
	return name
}

func trimFuncName(name string) string {
	pos := strings.LastIndexByte(name, '/')
	if pos != -1 {
		name = name[pos+1:]
	}
	return name
}

// GetCallerFuncFile 函数获得调用的文件位置和函数名称。
//
// 文件位置会从第一个src后开始截取，处理gopath下文件位置。
func GetCallerFuncFile(depth int) (string, string) {
	var pcs [1]uintptr
	runtime.Callers(depth+1, pcs[:])
	fs := runtime.CallersFrames(pcs[:])
	f, _ := fs.Next()

	return trimFuncName(f.Function), trimFileName(f.File + ":" + strconv.Itoa(f.Line))
}

// GetCallerStacks 函数返回caller栈信息。
func GetCallerStacks(depth int) []string {
	pc := make([]uintptr, DefaultLoggerDepthMaxStack)
	n := runtime.Callers(depth, pc)
	if n == 0 {
		return nil
	}

	stack := make([]string, 0, n)
	fs := runtime.CallersFrames(pc[:n])
	f, more := fs.Next()
	for more {
		stack = append(stack, trimFileName(f.File+":"+strconv.Itoa(f.Line))+" "+trimFuncName(f.Function))
		f, more = fs.Next()
	}
	return stack
}
