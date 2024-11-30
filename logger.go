package eudore

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// enum the logger levels.
const (
	LoggerDebug LoggerLevel = iota
	LoggerInfo
	LoggerWarning
	LoggerError
	LoggerFatal
	LoggerDiscard
)

// Logger defines the Logger interface to implement structured logging.
//
// The default implementation uses [NewLogger].
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warning(args ...any)
	Error(args ...any)
	// The Fatal method outputs the [LoggerFatal] log,
	// but does not stop the App.
	//
	// If [NewLoggerHookFatal] is enabled,
	// the App will be stopped when [LoggerFatal] occurs.
	Fatal(args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)

	// The WithField method sets a logging field.
	//
	// If the key is "logger" and "depth",
	// modify the Logger data but do not save the field.
	WithField(key string, val any) Logger
	// The WithFields method sets multiple properties,
	// but key will not modify Logger data.
	WithFields(keys []string, vals []any) Logger

	// The GetLevel method obtains the current Logger output level
	// and determines the level to cancel log generation.
	GetLevel() LoggerLevel
	// The SetLevel method sets the current Logger output level.
	SetLevel(level LoggerLevel)
}

// LoggerLevel defines the [Logger] level.
type LoggerLevel int

// loggerStd defines the default Logger implementation.
type loggerStd struct {
	LoggerEntry
	Handlers []LoggerHandler
	Pool     *sync.Pool
	Logger   bool
	Depth    int32
}

// LoggerEntry defines logger entry data and buffer.
type LoggerEntry struct {
	Level   LoggerLevel
	Time    time.Time
	Message string
	Keys    []string
	Vals    []any
	Buffer  []byte
}

// LoggerHandler defines how to process [LoggerEntry].
type LoggerHandler interface {
	// The HandlerPriority method returns the Handler processing order,
	// with smaller values taking priority.
	HandlerPriority() int
	// The HandlerEntry method processes the Entry data
	// and ends subsequent processing after setting Level=LoggerDiscard.
	HandlerEntry(entry *LoggerEntry)
}

// LoggerConfig defines [NewLogger] configuration,
// initializes [Logger] and creates default [LoggerHandler].
//
// If Formatter is json/text, use [NewLoggerFormatterJSON] or
// [NewLoggerFormatterText].
//
// If AsyncSize is greater than 0, use [NewLoggerWriterAsync].
//
// If Stdout is true and [DefaultLoggerWriterStdout],
// use [NewLoggerWriterStdout]; if DefaultLoggerWriterStdoutColor StdColor
// is true and [DefaultLoggerWriterStdoutColor], Output color Level.
//
// If Path contains the keyword yyyy/mm/dd/hh or MaxSize is non-zero,
// use [NewLoggerWriterRotate].
// Else if Path is not empty, use [NewLoggerWriterFile].
//
// If HookFilter is non-nil, use [NewLoggerHookFilter].
//
// If HookFatal is true, use [NewLoggerHookFatal].
//
// If HookMeta is true and AsyncSize is 0, use [NewLoggerHookMeta].
type LoggerConfig struct {
	// Custom LoggerHandler
	Handlers     []LoggerHandler `alias:"handlers" json:"-" yaml:"-"`
	Level        LoggerLevel     `alias:"level" json:"level" yaml:"level"`
	AsyncSize    int             `alias:"asyncSize" json:"asyncSize" yaml:"asyncSize"`
	AsyncTimeout time.Duration   `alias:"asyncTimeout" json:"asyncTimeout" yaml:"asyncTimeout"`
	Caller       bool            `alias:"caller" json:"caller" yaml:"caller"`
	Stdout       bool            `alias:"stdout" json:"stdout" yaml:"stdout"`
	StdColor     bool            `alias:"stdColor" json:"stdColor" yaml:"stdColor"`
	Formatter    string          `alias:"formater" json:"formater" yaml:"formater"`
	TimeFormat   string          `alias:"timeFormat" json:"timeFormat" yaml:"timeFormat"`
	HookFilter   [][]string      `alias:"hookFilter" json:"hookFilter" yaml:"hookFilter"`
	HookFatal    bool            `alias:"hookFatal" json:"hookFatal" yaml:"hookFatal"`
	HookMeta     bool            `alias:"hookMeta" json:"hookMeta" yaml:"hookMeta"`
	Path         string          `alias:"path" json:"path" yaml:"path"`
	Link         string          `alias:"link" json:"link" yaml:"link"`
	MaxSize      uint64          `alias:"maxSize" json:"maxSize" yaml:"maxSize"`
	MaxAge       int             `alias:"maxAge" json:"maxAge" yaml:"maxAge"`
	MaxCount     int             `alias:"maxCount" json:"maxCount" yaml:"maxCount"`
}

type MetadataLogger struct {
	Health     bool      `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name       string    `json:"name" protobuf:"2,name=name" yaml:"name"`
	Count      [6]uint64 `json:"count" protobuf:"3,name=count" yaml:"count"`
	Size       uint64    `json:"size" protobuf:"4,name=size" yaml:"size"`
	SizeFormat string    `json:"sizeFormat" protobuf:"5,name=sizeFormat" yaml:"sizeFormat"`
}

// The NewLogger function creates default [Logger] using [LoggerConfig].
func NewLogger(config *LoggerConfig) Logger {
	if config == nil {
		config = &LoggerConfig{
			Stdout:    true,
			StdColor:  true,
			HookFatal: DefaultLoggerHookFatal,
		}
	}

	handlers := config.getHandlers()
	size := DefaultLoggerEntryFieldsLength
	buff := DefaultLoggerEntryBufferLength
	pool := &sync.Pool{}
	pool.New = func() any {
		return &loggerStd{
			Handlers: handlers,
			Pool:     pool,
			LoggerEntry: LoggerEntry{
				Level:  config.Level,
				Keys:   make([]string, 0, size),
				Vals:   make([]any, 0, size),
				Buffer: make([]byte, 0, buff),
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

func (c *LoggerConfig) getHandlers() []LoggerHandler {
	hs := c.Handlers
	hs = append(hs, c.getFormatter()...)
	hs = append(hs, c.getHooks()...)
	hs = append(hs, c.getWriters()...)
	sort.Slice(hs, func(i, j int) bool {
		return hs[i].HandlerPriority() < hs[j].HandlerPriority()
	})
	return hs
}

func (c *LoggerConfig) getFormatter() []LoggerHandler {
	c.TimeFormat = GetAnyByString(c.TimeFormat,
		DefaultLoggerFormatterFormatTime,
		time.RFC3339,
	)
	c.Formatter = GetAnyByString(c.Formatter,
		DefaultLoggerFormatter,
		"json",
	)

	// formatter
	switch strings.ToLower(c.Formatter) {
	case "json":
		return []LoggerHandler{NewLoggerFormatterJSON(c.TimeFormat)}
	case "text":
		return []LoggerHandler{NewLoggerFormatterText(c.TimeFormat)}
	default:
		return []LoggerHandler{}
	}
}

func (c *LoggerConfig) getHooks() []LoggerHandler {
	// hook
	var hooks []LoggerHandler
	if len(c.HookFilter) > 0 {
		hooks = append(hooks, NewLoggerHookFilter(c.HookFilter))
	}
	if c.HookMeta && c.AsyncSize < 1 {
		hooks = append(hooks, NewLoggerHookMeta())
	}
	if c.HookFatal {
		hooks = append(hooks, NewLoggerHookFatal(nil))
	}
	return hooks
}

func (c *LoggerConfig) getWriters() []LoggerHandler {
	c.Stdout = c.Stdout && DefaultLoggerWriterStdout
	c.StdColor = c.StdColor && DefaultLoggerWriterStdoutColor
	c.Path = strings.TrimSpace(c.Path)
	// writer-stdout
	var writers []LoggerHandler
	if c.Stdout {
		writers = append(writers, NewLoggerWriterStdout(c.StdColor))
	}
	// writer-rotate
	if c.Path != "" {
		var hook []func(string, string)
		if c.Link != "" {
			hook = append(hook, hookFileLink(c.Link))
		}
		if c.MaxAge > 0 || c.MaxCount > 1 {
			hook = append(hook, hookFileRecycle(c.MaxAge, c.MaxCount))
		}
		h, err := NewLoggerWriterRotate(c.Path, c.MaxSize, hook...)
		if err != nil {
			panic(err)
		}
		writers = append(writers, h)
	}
	if c.AsyncSize > 0 && writers != nil {
		sort.Slice(writers, func(i, j int) bool {
			return writers[i].HandlerPriority() < writers[j].HandlerPriority()
		})
		if c.AsyncTimeout == 0 {
			c.AsyncTimeout = time.Second
		}
		return []LoggerHandler{NewLoggerWriterAsync(writers,
			c.AsyncSize, DefaultLoggerEntryBufferLength, c.AsyncTimeout,
		)}
	}
	return writers
}

// The NewLoggerInit function creates an initial log processor that only
// records logs.
//
// Used before [LoggerConfig] is parsed.
//
// Get a new [Logger] to process logs when Unmounting,
// and output the recorded logs to the new [Logger].
//
// If you continue to output logs after Unmount,
// it will panic [ErrLoggerInitUnmounted].
func NewLoggerInit() Logger {
	return NewLogger(&LoggerConfig{
		Handlers: []LoggerHandler{&loggerHandlerInit{
			Entrys: make([]*LoggerEntry, 0, 20),
		}},
		Formatter: "disable",
		HookMeta:  true,
	})
}

// NewLoggerNull defines empty log output and discards all logs.
func NewLoggerNull() Logger {
	return NewLogger(&LoggerConfig{
		Level:     LoggerDiscard,
		Formatter: "disable",
	})
}

// The NewLoggerWithContext method gets the Logger from the
// [context.Context] [ContextKeyLogger].
//
// If the Logger cannot be get, the [DefaultLoggerNull] object is returned.
func NewLoggerWithContext(ctx context.Context) Logger {
	log, ok := ctx.Value(ContextKeyLogger).(Logger)
	if ok {
		return log
	}
	return DefaultLoggerNull
}

// The Mount method causes LoggerStd to mount the [context.Context],
// which is passed to [LoggerHandler].
func (log *loggerStd) Mount(ctx context.Context) {
	for i := range log.Handlers {
		anyMount(ctx, log.Handlers[i])
	}
}

// The Unmount method causes LoggerStd to unload the [context.Context],
// which is passed to [LoggerHandler].
func (log *loggerStd) Unmount(ctx context.Context) {
	for i := len(log.Handlers) - 1; i > -1; i-- {
		anyUnmount(ctx, log.Handlers[i])
	}
}

// The Metadata method find the first anyMetadata object from [Handlers] and
// returns meta.
func (log *loggerStd) Metadata() any {
	for i := range log.Handlers {
		meta := anyMetadata(log.Handlers[i])
		if meta != nil {
			return meta
		}
	}
	return nil
}

func (log *loggerStd) GetLevel() LoggerLevel {
	return log.Level
}

func (log *loggerStd) SetLevel(level LoggerLevel) {
	log.Level = level
}

func (log *loggerStd) Debug(args ...any) {
	log.format(LoggerDebug, args...)
}

func (log *loggerStd) Info(args ...any) {
	log.format(LoggerInfo, args...)
}

func (log *loggerStd) Warning(args ...any) {
	log.format(LoggerWarning, args...)
}

func (log *loggerStd) Error(args ...any) {
	log.format(LoggerError, args...)
}

func (log *loggerStd) Fatal(args ...any) {
	log.format(LoggerFatal, args...)
}

func (log *loggerStd) Debugf(format string, args ...any) {
	log.formatf(LoggerDebug, format, args...)
}

func (log *loggerStd) Infof(format string, args ...any) {
	log.formatf(LoggerInfo, format, args...)
}

func (log *loggerStd) Warningf(format string, args ...any) {
	log.formatf(LoggerWarning, format, args...)
}

func (log *loggerStd) Errorf(format string, args ...any) {
	log.formatf(LoggerError, format, args...)
}

func (log *loggerStd) Fatalf(format string, args ...any) {
	log.formatf(LoggerFatal, format, args...)
}

// The WithFields method sets multiple properties, but does not set the
// Field property.
func (log *loggerStd) WithFields(key []string, value []any) Logger {
	if log.Logger {
		log = log.getLogger()
	}
	log.Keys = append(log.Keys, key...)
	log.Vals = append(log.Vals, value...)
	return log
}

// The WithField method sets a logging field.
//
// If the key is "logger" and the value is bool(true), LoggerEntry will be set
// to Logger.
//
// If the key is "depth" and the value type is int,
// set the number of layers to add or delete in the log call stack;
//
// If the key is "depth" and the value type is string value "enable" or
// "disable" or "stack", enable or disable the output of the log call stack;
// And add the key: file/func/stack.
// If the relevant key is used, you need to disable depth first.
//
// If the key is "time" and the value type is time.time,
// set the time attribute of the log output.
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

// The withFieldDepth method handles the withDepth attribute,
// can inline with cost 53.
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
	entry.Keys = entry.Keys[:0]
	entry.Vals = entry.Vals[:0]
	entry.Buffer = entry.Buffer[:0]
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
		log.Keys = append(log.Keys, "error")
		log.Vals = append(log.Vals,
			"Logger: The number of field keys and values are not equal",
		)
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
			if sliceIndex(log.Keys, "stack") == -1 {
				log.Keys = append(log.Keys, "stack")
				log.Vals = append(log.Vals,
					GetCallerStacks(int(log.Depth&0xff)+1),
				)
			}
		}

		for _, h := range log.Handlers {
			if log.Level < LoggerDiscard {
				h.HandlerEntry(&log.LoggerEntry)
			}
		}
	}
}

// The String method implements the [fmt.Stringer] interface and formats level.
func (l LoggerLevel) String() string {
	return DefaultLoggerLevelStrings[l]
}

// The MarshalText method implements [the encoding.TextMarshaler] interface.
func (l LoggerLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// The UnmarshalText method implements the [encoding.TextUnmarshaler] interface.
func (l *LoggerLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	for i, s := range DefaultLoggerLevelStrings {
		if s == str {
			*l = LoggerLevel(i)
			return nil
		}
	}
	n, err := strconv.Atoi(str)
	if err == nil && n < len(DefaultLoggerLevelStrings) && n > -1 {
		*l = LoggerLevel(n)
		return nil
	}
	return fmt.Errorf(ErrLoggerLevelUnmarshalText, text)
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

// The GetCallerFuncFile function obtains the called file location and
// function name.
//
// func name does not retain the package path, file name ignores the
// $GOPATH path.
func GetCallerFuncFile(depth int) (string, string) {
	var pcs [1]uintptr
	runtime.Callers(depth+1, pcs[:])
	fs := runtime.CallersFrames(pcs[:])
	f, _ := fs.Next()

	return trimFuncName(f.Function),
		trimFileName(f.File + ":" + strconv.Itoa(f.Line))
}

// The GetCallerStacks function returns the caller stack information.
//
// func name does not retain the package path, file name ignores the
// $GOPATH path.
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
		stack = append(stack,
			trimFileName(f.File+":"+strconv.Itoa(f.Line))+
				" "+
				trimFuncName(f.Function),
		)
		f, more = fs.Next()
	}
	return stack
}
