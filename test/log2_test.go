package test

import (
	"fmt"
	"sync"
	"time"
	"testing"
)
const (
	LogDebug	LoggerLevel = iota
	LogInfo
	LogWarning
	LogError
	LogFatal
	numSeverity = 5
)


func BenchmarkLogger(b *testing.B) {
	b.ReportAllocs()
	log := NewLoggerStd()
	for i := 0; i < b.N; i++ {
		// log.Info("Info")
		log.WithField("bench", true).WithField("test", true).WithField("test", true).WithField("test", true).WithField("test", true).Info("Info")
	}
}

type (
	LoggerLevel int
	LogHandler interface {
		HandleEntry(*Entry)
	}
	LogOut interface {
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
		LogHandler
	}
	Fields interface {
		Add(string, interface{}) Fields
		Range(func(string, interface{}))
		Clean() Fields
	}
	Entry struct {
		LogHandler		`json:"-" xml:"-" yaml:"-"`
		Level		LoggerLevel		`json:"level"`
		Fields		Fields		`json:"fields,omitempty"`
		Timestamp	time.Time	`json:"timestamp"`
		Message		string		`json:"message,omitempty"`
	}
	FieldsMap map[string]interface{}
	fa struct {
		key string
		val interface{}
	}
	FieldsArray []fa
	LoggerStd struct {
		pool	sync.Pool
	}
)

func NewLoggerStd() Logger {
	l := &LoggerStd{}
	l.pool.New = func() interface{} {
		return &Entry{
			Fields: make(FieldsMap, 5),
			LogHandler: l,
		}
	}
	return l
}

func (l *LoggerStd) NewEntry() *Entry {
	return l.pool.Get().(*Entry)
}

func (l *LoggerStd) Info(args ...interface{}) {
	e := l.NewEntry()
	e.Message = fmt.Sprint(args...)
	l.HandleEntry(e)
}

func (l *LoggerStd) WithField(key string, val interface{}) LogOut {
	return l.NewEntry().WithField(key, val)
}

func (l *LoggerStd) WithFields(f Fields) LogOut {
	return l.NewEntry().WithFields(f)
}

func (l *LoggerStd) HandleEntry(e *Entry) {
	// fmt.Println(e)
	e.Fields = e.Fields.Clean()
	l.pool.Put(e)
}








func (e *Entry) Info(args ...interface{}) {
	e.Level = LogInfo
	e.Message = fmt.Sprint(args...)
	e.LogHandler.HandleEntry(e)
}

func (e *Entry) WithField(key string, val interface{}) LogOut {
	e.Fields = e.Fields.Add(key, val)
	return e
}

func (e *Entry) WithFields(f Fields) LogOut {
	e.Fields = f
	return e
}


func (e *Entry) write() {
	e.LogHandler.HandleEntry(e)
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

func (f FieldsMap) Clean() Fields {
	if len(f) > 0 {
		return make(FieldsMap)	
	}
	return f
}



func (f FieldsArray) Add(key string, val interface{}) Fields{
	return append(f, fa{key, val})
}

func (f FieldsArray) Range(fn func(key string, val interface{})) {
	for _, i := range f {
		fn(i.key, i.val)
	}
}

func (f FieldsArray) Clean() Fields {
	return f[0:0]
}

