package eudore

import (
	"bytes"
	"context"
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
	Entrys []*LoggerEntry
}

func (h *loggerHandlerInit) HandlerPriority() int {
	return 100
}

func (h *loggerHandlerInit) HandlerEntry(entry *LoggerEntry) {
	h.Lock()
	defer h.Unlock()
	if h.Entrys == nil {
		panic(ErrLoggerInitUnmounted)
	}
	h.Entrys = append(h.Entrys, &LoggerEntry{
		Level:   entry.Level,
		Time:    entry.Time,
		Message: entry.Message,
		Keys:    append([]string{}, entry.Keys...),
		Vals:    append([]any{}, entry.Vals...),
	})
}

// The Unmount method get [ContextKeyLogger] from [context.Context] and outputs
// the saved entry.
//
// If it cannot be get, use [NewLogger].
func (h *loggerHandlerInit) Unmount(ctx context.Context) {
	h.Lock()
	defer h.Unlock()
	logger, _ := ctx.Value(ContextKeyLogger).(Logger)
	if logger == nil {
		logger = NewLogger(nil)
	}

	logger = logger.WithField("depth", "disable").WithField("logger", true)
	for _, data := range h.Entrys {
		entry := logger.WithField("time", data.Time).
			WithFields(data.Keys, data.Vals)
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
	Count [6]uint64
}

// NewLoggerHookMeta function creates [LoggerHandler] to implement log counting.
//
// Implement the Metadata method and return the count and size of the
// [LoggerEntry].
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

func (h *loggerHookMeta) HandlerPriority() int {
	return DefaultLoggerPriorityHookMeta
}

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

// The NewLoggerHookFilter function creates [LoggerHandler] to implement
// log filtering or modification.
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
			funcs = append(funcs, loggerHookFilterFunc{
				Key:        strs[0],
				FuncRunner: FuncRunner{kind, fn},
			})
		}
		if len(funcs) > 0 {
			h.Funcs = append(h.Funcs, funcs)
		}
	}
}

func (h *loggerHookFilter) HandlerPriority() int {
	return DefaultLoggerPriorityHookFilter
}

func (h *loggerHookFilter) HandlerEntry(entry *LoggerEntry) {
	for i := range h.Funcs {
		h.HandlerRule(entry, h.Funcs[i])
		if entry.Level == LoggerDiscard {
			return
		}
	}
}

func (h *loggerHookFilter) HandlerRule(entry *LoggerEntry,
	funcs []loggerHookFilterFunc,
) {
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
			entry.Vals[pos] = funcRunAny(
				funcs[i].Kind, funcs[i].Func, entry.Vals[pos],
			)
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

// The NewLoggerHookFatal function creates [LoggerHandler] to implement handle
// [LoggerFatal] logs.
//
// In Mount, get [ContextKeyAppCancel] as [context.CancelFunc].
//
// If you can't get the Fatal handler, use panic.
func NewLoggerHookFatal(fn func(*LoggerEntry)) LoggerHandler {
	return &loggerHookFatal{fn}
}

func (h *loggerHookFatal) Mount(ctx context.Context) {
	if h.Callback == nil {
		cancel, ok := ctx.Value(ContextKeyAppCancel).(context.CancelFunc)
		if ok {
			h.Callback = func(_ *LoggerEntry) {
				cancel()
			}
		}
	}
}

func (h *loggerHookFatal) HandlerPriority() int {
	return DefaultLoggerPriorityHookFatal
}

func (h *loggerHookFatal) HandlerEntry(entry *LoggerEntry) {
	if entry.Level == LoggerFatal {
		if h.Callback == nil {
			panic(entry.Message)
		}
		h.Callback(entry)
	}
}

type loggerWriterAsync struct {
	loggerHookMeta
	Handlers []LoggerHandler
	pool     sync.Pool
	timeout  time.Duration
	async    chan *LoggerEntry
	done     chan struct{}
}

// The NewLoggerWriterAsync function creates [LoggerHandler] to implement
// async Writer for log processing.
//
// size specifies the asynchronous buffer size, after the timeout, the overflow
// log will be discarded; buff specifies the length of the multiplexed []byte.
//
// This [LoggerHandler] implements the Metadata method to
// record the number of discarded logs.
//
// The [LoggerEntry] used by handlers only has Level and Buffer field data.
//
// If you continue to use it after Unmount,
// it will panic send on closed channel.
func NewLoggerWriterAsync(handlers []LoggerHandler, size, buff int,
	timeout time.Duration,
) LoggerHandler {
	w := &loggerWriterAsync{
		pool: sync.Pool{
			New: func() any {
				buf := make([]byte, buff)
				return &buf
			},
		},
		timeout: timeout,
		async:   make(chan *LoggerEntry, size),
		done:    make(chan struct{}),
	}
	w.Handlers = append([]LoggerHandler{&w.loggerHookMeta}, handlers...)
	return w
}

func (w *loggerWriterAsync) Mount(ctx context.Context) {
	go func() {
		for {
			select {
			case log := <-w.async:
				for _, h := range w.Handlers {
					h.HandlerEntry(log)
				}
				w.pool.Put(&log.Buffer)
			case <-w.done:
				return
			}
		}
	}()
	for _, h := range w.Handlers {
		anyMount(ctx, h)
	}
}

func (w *loggerWriterAsync) Unmount(ctx context.Context) {
	for _, h := range w.Handlers {
		anyUnmount(ctx, h)
	}
	close(w.done)
}

func (w *loggerWriterAsync) HandlerPriority() int {
	return DefaultLoggerPriorityWriterAsync
}

func (w *loggerWriterAsync) HandlerEntry(entry *LoggerEntry) {
	buf := w.pool.Get().(*[]byte)
	log := &LoggerEntry{
		Level:  entry.Level,
		Buffer: append((*buf)[:0], entry.Buffer...),
	}

	// try write
	select {
	case w.async <- log:
		return
	default:
	}

	select {
	case w.async <- log:
	case <-time.After(w.timeout):
		atomic.AddUint64(&w.loggerHookMeta.Count[LoggerDiscard], 1)
		w.pool.Put(buf)
	}
}

type loggerWriterStdout struct {
	sync.Mutex
}
type loggerWriterStdoutColor struct {
	sync.Mutex
}

// The NewLoggerWriterStdout function creates [LoggerHandler] to output logs to
// [os.Stdout].
func NewLoggerWriterStdout(color bool) LoggerHandler {
	if color {
		return &loggerWriterStdoutColor{}
	}
	return &loggerWriterStdout{}
}

func (w *loggerWriterStdout) HandlerPriority() int {
	return DefaultLoggerPriorityWriterStdout
}

func (w *loggerWriterStdout) HandlerEntry(entry *LoggerEntry) {
	w.Lock()
	_, _ = os.Stdout.Write(entry.Buffer)
	w.Unlock()
}

func (w *loggerWriterStdoutColor) HandlerPriority() int {
	return DefaultLoggerPriorityWriterStdout
}

func (w *loggerWriterStdoutColor) HandlerEntry(entry *LoggerEntry) {
	// Search for level in the first 64 char
	pos := bytes.Index(entry.Buffer[:64], loggerLevelDefaultBytes[entry.Level])
	w.Lock()
	if pos != -1 {
		_, _ = os.Stdout.Write(entry.Buffer[:pos])
		_, _ = os.Stdout.Write(loggerLevelColorBytes[entry.Level])
		_, _ = os.Stdout.Write(
			entry.Buffer[pos+loggerLevelDefaultLen[entry.Level]:],
		)
	} else {
		os.Stdout.Write(entry.Buffer)
	}
	w.Unlock()
}

type loggerWriterFile struct {
	sync.Mutex
	File *os.File
}

// The NewLoggerWriterFile function creates [LoggerHandler] to write logs to
// [os.File].
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

func (w *loggerWriterFile) HandlerPriority() int {
	return DefaultLoggerPriorityWriterFile
}

func (w *loggerWriterFile) HandlerEntry(entry *LoggerEntry) {
	w.Lock()
	_, _ = w.File.Write(entry.Buffer)
	w.Unlock()
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

// The NewLoggerWriterRotate function creates [LoggerHandler] to write logs to
// [os.File].
//
// If maxsize is set or name contains the string yyyy/yy/mm/dd/hh,
// the log file will be rotated according to size or time.
//
// hooks is a callback function used when opening a new file name to implement
// log file cleaning and linking.
func NewLoggerWriterRotate(name string, maxsize uint64,
	hooks ...func(string, string),
) (LoggerHandler, error) {
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
		openhooks: hooks,
	}
	h.pattern = h.getFilePattern()
	return h, h.rotateFile()
}

func (w *loggerWriterRotate) HandlerEntry(entry *LoggerEntry) {
	w.Lock()
	defer w.Unlock()
	if w.writeSize+uint64(len(entry.Buffer)) >= w.maxSize {
		_ = w.rotateFile()
	} else if entry.Time.After(w.nextTime) {
		w.nextIndex = getNextIndex(w.name, w.maxSize)
		w.nextTime = getNextTime(w.name)
		_ = w.rotateFile()
	}

	n, _ := w.File.Write(entry.Buffer)
	w.writeSize += uint64(n)
}

func (w *loggerWriterRotate) rotateFile() error {
	for {
		name := w.getRotateName()
		_ = os.MkdirAll(filepath.Dir(name), 0o755)
		file, err := os.OpenFile(
			name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644,
		)
		if err != nil {
			return err
		}
		w.nextIndex++

		stat, _ := file.Stat()
		w.writeSize = uint64(stat.Size())
		if w.writeSize < w.maxSize {
			_ = w.File.Sync()
			_ = w.File.Close()
			w.File = file
			for _, fn := range w.openhooks {
				fn(name, w.pattern)
			}
			return nil
		}
		file.Close()
	}
}

func (w *loggerWriterRotate) getRotateName() string {
	name := w.name
	if w.nextTime.Unix() != roatteMaxTime {
		name = fileFormatTime(name)
	}
	if w.maxSize != roatteMaxSize {
		ext := path.Ext(w.name)
		name = name[:len(name)-len(ext)] +
			"-" +
			strconv.Itoa(w.nextIndex) +
			ext
	}
	return name
}

func (w *loggerWriterRotate) getFilePattern() string {
	name := w.name
	if w.nextTime.Unix() != roatteMaxTime {
		k := DefaultLoggerWriterRotateDataKeys
		v := [...]int{2, 2, 2, 4}
		for i := range k {
			name = strings.ReplaceAll(name, k[i], strings.Repeat("[0-9]", v[i]))
		}
	}
	if w.maxSize != roatteMaxSize {
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
			datas := [...]int{
				now.Hour(), now.Day(),
				int(now.Month()), now.Year(),
			}
			datas[i]++
			return time.Date(datas[3], time.Month(datas[2]), datas[1],
				datas[0], 0, 0, 0, now.Location(),
			)
		}
	}
	return time.Unix(roatteMaxTime, 0)
}

func hookFileLink(link string) func(string, string) {
	_ = os.MkdirAll(filepath.Dir(link), 0o755)
	return func(name, _ string) {
		if !filepath.IsAbs(name) {
			pwd, _ := os.Getwd()
			name = filepath.Join(pwd, name)
		}
		_ = os.Remove(link)
		_ = os.Symlink(name, link)
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
