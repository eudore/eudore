package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
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

func TestLogger(*testing.T) {
	log, _ := NewLoggerStd()
	log.WithField("bench", true).WithField("test", "jgu").WithField("test", true).WithField("test", true).WithField("test", true).Info("this message")
	log.Info("logger entryStd 2.")
	log.(*LoggerStd).writer.Flush()
}

func BenchmarkLogger(b *testing.B) {
	b.ReportAllocs()
	log, _ := NewLoggerStd()
	for i := 0; i < b.N; i++ {
		// log.Info("Info")
		log.WithField("bench", true).WithField("test", "jgu").WithField("test", true).WithField("test", true).WithField("test", true).Info("Info")
	}
}

type (
	LoggerLevel int
	LogOut      interface {
		// Debug(...interface{})
		Info(...interface{})
		// Warning(...interface{})
		// Error(...interface{})
		// Fatal(...interface{})
		WithField(string, interface{}) LogOut
		WithFields(Fields) LogOut
	}
	Logger interface {
		LogOut
		io.Writer
		HandlerEntry(interface{})
	}
	Fields interface {
		Add(string, interface{}) Fields
		Range(func(string, interface{}))
		Init() Fields
	}
	LoggerStd struct {
		pool   sync.Pool
		writer *bufio.Writer
		json   *json.Encoder
		tmpl   *template.Template
	}
	entryStd struct {
		logger  *LoggerStd
		Level   LoggerLevel `json:"level"`
		Fields  Fields      `json:"fields,omitempty"`
		Time    string      `json:"timestamp"`
		Message string      `json:"message,omitempty"`
	}
	FieldsMap map[string]interface{}
)

func (level LoggerLevel) String() string {
	return "Info"
}

func NewLoggerStd() (Logger, error) {
	// file, _ := os.OpenFile("/tmp/01.log", os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
	tmpl := template.New("test").Delims("{", "}").Option("missingkey=zero")
	tmpl.Parse("[{.Time} {.Fields.file}:{.Fields.line}] {.Level}: {.Message}\n")
	l := &LoggerStd{
		pool: sync.Pool{},
		// writer:	bufio.NewWriter(file),
		writer: bufio.NewWriter(os.Stdout),
	}
	l.json = json.NewEncoder(l.writer)
	l.tmpl = tmpl
	l.pool.New = func() interface{} {
		return &entryStd{
			logger: l,
			Fields: make(FieldsMap, 5),
		}
	}
	return l, nil
}

func (l *LoggerStd) newEntry() *entryStd {
	entry := l.pool.Get().(*entryStd)
	file, line := LogFormatFileLine(0)
	entry.Fields.Add("file", file)
	entry.Fields.Add("line", line)
	return entry
}

func (l *LoggerStd) Info(args ...interface{}) {
	l.newEntry().Info(args...)
}

func (l *LoggerStd) WithField(key string, val interface{}) LogOut {
	return l.newEntry().WithField(key, val)
}

func (l *LoggerStd) WithFields(f Fields) LogOut {
	return l.newEntry().WithFields(f)
}

func (l *LoggerStd) Write(p []byte) (int, error) {
	return l.writer.Write(p)
}

func (l *LoggerStd) HandlerEntry(i interface{}) {
	// l.json.Encode(i)
	l.tmpl.Execute(l.writer, i)
}

func (e *entryStd) Info(args ...interface{}) {
	e.Level = LogInfo
	e.Time = time.Now().Format("2006-01-02 15:04:05")
	e.Message = fmt.Sprint(args...)
	// fmt.Fprint(e.logger, e)
	// e.json.Encode(e)
	// fmt.Println(e)
	e.logger.HandlerEntry(e)
	e.Fields = e.Fields.Init()
	e.logger.pool.Put(e)
}

func (e *entryStd) WithField(key string, val interface{}) LogOut {
	e.Fields = e.Fields.Add(key, val)
	return e
}

func (e *entryStd) WithFields(f Fields) LogOut {
	e.Fields = f
	return e
}

func (f FieldsMap) Add(key string, val interface{}) Fields {
	f[key] = val
	return f
}

func (f FieldsMap) Range(fn func(key string, val interface{})) {
	for k, v := range f {
		fn(k, v)
	}
}

func (f FieldsMap) Init() Fields {
	if len(f) > 0 {
		return make(FieldsMap)
	}
	return f
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
