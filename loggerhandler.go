package eudore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type loggerHandlerInit struct {
	sync.Mutex
	Entrys []LoggerEntry
}

func (h *loggerHandlerInit) HandlerPriority() int {
	return 100
}

func (h *loggerHandlerInit) HandlerEntry(entry *LoggerEntry) {
	entry.Time = time.Now()
	h.Lock()
	h.Entrys = append(h.Entrys, *entry)
	h.Unlock()
}

// Unmount 方法获取ContextKeyLogger.(Logger)接受Init存储的日志。
func (h *loggerHandlerInit) Unmount(ctx context.Context) {
	h.Lock()
	defer h.Unlock()
	logger, _ := ctx.Value(ContextKeyLogger).(Logger)
	if logger == nil {
		logger = NewLogger(nil)
	}

	logger = logger.WithField("depth", "disable").WithField("logger", true)
	for _, data := range h.Entrys {
		entry := logger.WithField("time", data.Time).WithFields(data.Keys, data.Vals)
		switch data.Level {
		case LoggerDebug:
			entry.Debug(data.Message)
		case LoggerInfo:
			entry.Info(data.Message)
		case LoggerWarning:
			entry.Warning(data.Message)
		case LoggerError:
			entry.Error(data.Message)
		case LoggerFatal:
			entry.Fatal(data.Message)
		}
	}
	h.Entrys = nil
}

type loggerHookMeta struct {
	Size  uint64
	Count [5]uint64
}

// NewLoggerHookMeta 函数创建日志Meta处理，记录日志数量和写入量。
func NewLoggerHookMeta() LoggerHandler {
	return &loggerHookMeta{}
}

func (h *loggerHookMeta) Metadata() any {
	return MetadataLogger{
		Health:     true,
		Name:       "eudore.loggerStd",
		Count:      h.Count,
		Size:       h.Size,
		SizeFormat: formatSize(int64(h.Size)),
	}
}
func (h *loggerHookMeta) HandlerPriority() int { return 60 }
func (h *loggerHookMeta) HandlerEntry(entry *LoggerEntry) {
	atomic.AddUint64(&h.Size, uint64(len(entry.Buffer)))
	atomic.AddUint64(&h.Count[entry.Level], 1)
}

type loggerHookFilter struct {
	Rules [][]string
	Funcs [][]loggerHookFilterFunc
}

type loggerHookFilterFunc struct {
	Key string
	FuncRunner
}

// NewLoggerHookFilter 函数创建日志过滤处理器。
//
// 在Mount时如果规则初始化失败，查看FuncCreator的Metadata。
func NewLoggerHookFilter(rules [][]string) LoggerHandler {
	for i := range rules {
		rules[i] = sliceFilter(rules[i], func(t string) bool {
			return len(strings.SplitN(t, " ", 3)) == 3
		})
	}
	return &loggerHookFilter{
		Rules: rules,
		Funcs: make([][]loggerHookFilterFunc, 0, len(rules)),
	}
}

// Mount 方法使LoggerStd挂载上下文，上下文传递给LoggerStdData。
func (h *loggerHookFilter) Mount(ctx context.Context) {
	fc := NewFuncCreatorWithContext(ctx)
	for i := range h.Rules {
		funcs := make([]loggerHookFilterFunc, 0, len(h.Rules[i]))
		for j := range h.Rules[i] {
			strs := strings.SplitN(h.Rules[i][j], " ", 3)
			kind := NewFuncCreateKind(strs[1])
			fn, err := fc.CreateFunc(kind, strs[2])
			if err != nil {
				continue
			}
			funcs = append(funcs, loggerHookFilterFunc{strs[0], FuncRunner{kind, fn}})
		}
		if len(funcs) > 0 {
			h.Funcs = append(h.Funcs, funcs)
		}
	}
}

func (h *loggerHookFilter) HandlerPriority() int { return 10 }
func (h *loggerHookFilter) HandlerEntry(entry *LoggerEntry) {
	for i := range h.Funcs {
		h.HandlerRule(entry, h.Funcs[i])
		if entry.Level == LoggerDiscard {
			return
		}
	}
}

func (h *loggerHookFilter) HandlerRule(entry *LoggerEntry, funcs []loggerHookFilterFunc) {
	for i := range funcs {
		pos := sliceIndex(entry.Keys, funcs[i].Key)
		if pos == -1 {
			return
		}

		kind := NewFuncCreateKindWithType(reflect.TypeOf(entry.Vals[pos]))
		if kind != funcs[i].Kind && kind+FuncCreateNumber != funcs[i].Kind {
			return
		}

		if funcs[i].Kind > FuncCreateAny {
			entry.Vals[pos] = funcRunAny(funcs[i].Kind, funcs[i].Func, entry.Vals[pos])
		} else if !funcs[i].RunPtr(reflect.ValueOf(entry.Vals[pos])) {
			return
		}
	}
	if len(entry.Keys) > 0 && funcs[len(funcs)-1].Kind < FuncCreateSetString {
		entry.Level = LoggerDiscard
	}
}

func funcRunAny(kind FuncCreateKind, fn, i any) any {
	v := reflect.Indirect(reflect.ValueOf(i))
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		r := FuncRunner{kind, fn}
		for i := 0; i < v.Len(); i++ {
			r.Run(v.Index(i))
		}
		return i
	}

	switch kind {
	case FuncCreateSetString:
		return fn.(func(string) string)(v.String())
	case FuncCreateSetInt:
		return fn.(func(int) int)(int(v.Int()))
	case FuncCreateSetUint:
		return fn.(func(uint) uint)(uint(v.Uint()))
	case FuncCreateSetFloat:
		return fn.(func(float64) float64)(v.Float())
	case FuncCreateSetBool:
		return fn.(func(bool) bool)(v.Bool())
	default:
		return fn.(func(any) any)(i)
	}
}

type loggerHookFatal struct {
	Callback func(*LoggerEntry)
}

// NewLoggerHookFatal 函数创建Fatal级别日志处理Hook。
func NewLoggerHookFatal(fn func(*LoggerEntry)) LoggerHandler {
	return &loggerHookFatal{fn}
}

func (h *loggerHookFatal) Mount(ctx context.Context) {
	if h.Callback == nil {
		app, ok := ctx.Value(ContextKeyApp).(*App)
		if ok {
			h.Callback = func(entry *LoggerEntry) {
				app.SetValue(ContextKeyError, fmt.Errorf(entry.Message))
			}
		}
	}
}

func (h *loggerHookFatal) HandlerPriority() int { return 101 }
func (h *loggerHookFatal) HandlerEntry(entry *LoggerEntry) {
	if entry.Level == LoggerFatal {
		if h.Callback == nil {
			panic(entry.Message)
		}
		h.Callback(entry)
	}
}

type loggerWriterStdout struct {
	sync.Mutex
}
type loggerWriterStdoutColor struct {
	sync.Mutex
}

func NewLoggerWriterStdout(color bool) LoggerHandler {
	if color {
		return &loggerWriterStdoutColor{}
	}
	return &loggerWriterStdout{}
}

func (h *loggerWriterStdout) HandlerPriority() int {
	return 90
}

func (h *loggerWriterStdout) HandlerEntry(entry *LoggerEntry) {
	h.Lock()
	_, _ = os.Stdout.Write(entry.Buffer)
	h.Unlock()
}

func (h *loggerWriterStdoutColor) HandlerPriority() int {
	return 90
}

func (h *loggerWriterStdoutColor) HandlerEntry(entry *LoggerEntry) {
	pos := bytes.Index(entry.Buffer[:64], loggerLevelDefaultBytes[entry.Level])
	h.Lock()
	if pos != -1 {
		_, _ = os.Stdout.Write(entry.Buffer[:pos])
		_, _ = os.Stdout.Write(loggerLevelColorBytes[entry.Level])
		_, _ = os.Stdout.Write(entry.Buffer[pos+loggerLevelDefaultLen[entry.Level]:])
	} else {
		os.Stdout.Write(entry.Buffer)
	}
	h.Unlock()
}

type loggerWriterFile struct {
	sync.Mutex
	File *os.File
}

// NewLoggerWriterFile 函数创建一个文件输出的日志写入流。
func NewLoggerWriterFile(name string) (LoggerHandler, error) {
	err := os.MkdirAll(filepath.Dir(name), 0o755)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	return &loggerWriterFile{File: file}, nil
}

func (h *loggerWriterFile) HandlerPriority() int {
	return 100
}

func (h *loggerWriterFile) HandlerEntry(entry *LoggerEntry) {
	h.Lock()
	_, _ = h.File.Write(entry.Buffer)
	h.Unlock()
}

type loggerWriterRotate struct {
	loggerWriterFile
	name      string
	pattern   string
	writeSize uint64
	maxSize   uint64
	nextIndex int
	nextTime  time.Time
	openhooks []func(string, string)
}

// max uint64, 9999-12-31 23:59:59 +0000 UTC.
const roatteMaxSize, roatteMaxTime = 0xffffffffffffffff, 253402300799

// NewLoggerWriterRotate 函数创建一个支持文件切割的的日志写入流。
//
// 如果设置maxsize或name包含字符串yyyy/yy/mm/dd/hh，将可以滚动日志文件。
func NewLoggerWriterRotate(name string, maxsize uint64, fn ...func(string, string)) (LoggerHandler, error) {
	if maxsize == 0 && getNextTime(name).Unix() == roatteMaxTime {
		return NewLoggerWriterFile(name)
	}
	if maxsize == 0 {
		maxsize = roatteMaxSize
	}
	h := &loggerWriterRotate{
		name:      name,
		maxSize:   maxsize,
		nextIndex: getNextIndex(name, maxsize),
		nextTime:  getNextTime(name),
		openhooks: fn,
	}
	h.pattern = h.getFilePattern()
	return h, h.rotateFile()
}

func (h *loggerWriterRotate) HandlerEntry(entry *LoggerEntry) {
	h.Lock()
	defer h.Unlock()
	if h.writeSize+uint64(len(entry.Buffer)) >= h.maxSize {
		h.rotateFile()
	} else if entry.Time.After(h.nextTime) {
		h.nextIndex = getNextIndex(h.name, h.maxSize)
		h.nextTime = getNextTime(h.name)
		h.rotateFile()
	}

	n, _ := h.File.Write(entry.Buffer)
	h.writeSize += uint64(n)
}

func (h *loggerWriterRotate) rotateFile() error {
	for {
		name := h.getRotateName()
		_ = os.MkdirAll(filepath.Dir(name), 0o755)
		file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		h.nextIndex++

		stat, _ := file.Stat()
		h.writeSize = uint64(stat.Size())
		if h.writeSize < h.maxSize {
			h.File.Sync()
			h.File.Close()
			h.File = file
			for _, fn := range h.openhooks {
				fn(name, h.pattern)
			}
			return nil
		}
		file.Close()
	}
}

func (h *loggerWriterRotate) getRotateName() string {
	name := h.name
	if h.nextTime.Unix() != roatteMaxTime {
		name = fileFormatTime(name)
	}
	if h.maxSize != roatteMaxSize {
		ext := path.Ext(h.name)
		name = name[:len(name)-len(ext)] + "-" + strconv.Itoa(h.nextIndex) + ext
	}
	return name
}

func (h *loggerWriterRotate) getFilePattern() string {
	name := h.name
	if h.nextTime.Unix() != roatteMaxTime {
		k := DefaultLoggerWriterRotateDataKeys
		v := [...]int{2, 2, 2, 4}
		for i := range k {
			name = strings.ReplaceAll(name, k[i], strings.Repeat("[0-9]", v[i]))
		}
	}
	if h.maxSize != roatteMaxSize {
		ext := path.Ext(name)
		name = name[:len(name)-len(ext)] + "-*" + ext
	}
	return name
}

func fileFormatTime(n string) string {
	now := time.Now()
	k := DefaultLoggerWriterRotateDataKeys
	v := [...]string{"15", "02", "01", "2006"}
	for i := range k {
		n = strings.ReplaceAll(n, k[i], now.Format(v[i]))
	}
	return n
}

func getNextIndex(name string, size uint64) int {
	index := 0
	if size != roatteMaxSize {
		ext := path.Ext(name)
		name := fileFormatTime(name[:len(name)-len(ext)] + "-")
		list, _ := filepath.Glob(name + "*" + ext)
		for i := range list {
			n, _ := strconv.Atoi(list[i][len(name) : len(list[i])-len(ext)])
			if n > index {
				index = n
			}
		}
	}
	return index
}

func getNextTime(name string) time.Time {
	for i, str := range DefaultLoggerWriterRotateDataKeys {
		if strings.Contains(name, str) {
			now := time.Now()
			datas := [...]int{now.Hour(), now.Day(), int(now.Month()), now.Year()}
			datas[i]++
			return time.Date(datas[3], time.Month(datas[2]), datas[1], datas[0], 0, 0, 0, now.Location())
		}
	}
	return time.Unix(roatteMaxTime, 0)
}

func hookFileLink(link string) func(string, string) {
	os.MkdirAll(filepath.Dir(link), 0o755)
	return func(name, _ string) {
		if !filepath.IsAbs(name) {
			pwd, _ := os.Getwd()
			name = filepath.Join(pwd, name)
		}
		os.Remove(link)
		os.Symlink(name, link)
	}
}

func hookFileRecycle(age, count int) func(string, string) {
	type fileTime struct {
		Name    string
		ModTime time.Time
	}
	return func(_, pattern string) {
		list, _ := filepath.Glob(pattern)
		files := make([]fileTime, 0, len(list))
		for i := range list {
			stat, _ := os.Stat(list[i])
			files = append(files, fileTime{list[i], stat.ModTime()})
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime.Before(files[j].ModTime)
		})

		if count < len(files) {
			files = files[:len(files)-count]
			expr := time.Now().Add(time.Hour * time.Duration(-age))
			for i := range files {
				if files[i].ModTime.Before(expr) {
					os.Remove(files[i].Name)
				}
			}
		}
	}
}
