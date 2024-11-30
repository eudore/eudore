package eudore_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"bou.ke/monkey"
	. "github.com/eudore/eudore"
)

func TestLoggerStd(t *testing.T) {
	log := NewLogger(nil)

	log.GetLevel()
	log.SetLevel(LoggerFatal)
	log.Debug("0")
	log.Debugf("0")
	log.Info("1")
	log.Infof("1")
	log.Warning("2")
	log.Warningf("2")
	log.Error("3")
	log.Errorf("3")
	log.Fatal("4")
	log.Fatalf("4")

	log.SetLevel(LoggerDebug)
	log.Debug("0")
	log.Debugf("0")
	log.Info("1")
	log.Infof("1")
	log.Warning("2")
	log.Warningf("2")
	log.Error("3")
	log.Errorf("3")
	log.Fatal("4")
	log.Fatalf("4")

	log.WithField("depth", "enable").Info("1")
	log.WithField("depth", "stack").Info("1")
	log.WithField("depth", "disable").Info("1")
	log.WithField("depth", -2).WithField("depth", "enable").Info("1")
	log.WithField("depth", true).Info("1")
	log.WithField("caller", "logger").WithField("logger", true).Info("1")
	log.WithFields([]string{"key"}, []any{}).Info("1")

	log.(interface{ Metadata() any }).Metadata()
	ctx := context.WithValue(context.Background(),
		ContextKeyLogger, log,
	)
	NewLoggerWithContext(ctx)
	NewLoggerWithContext(context.Background())

	NewLogger(&LoggerConfig{
		Handlers: []LoggerHandler{
			NewLoggerWriterStdout(true),
			NewLoggerWriterStdout(false),
		},
	}).Info("Stdout")
}

func TestLoggerInit1(t *testing.T) {
	defer func() {
		recover()
	}()

	ctx := context.WithValue(context.Background(),
		ContextKeyLogger, NewLogger(nil),
	)
	log := NewLoggerInit()
	log.Info("loggerInit to end")
	log.Debug("0")
	log.Infof("1")
	log.Warning("2")
	log.Error("3")
	log.Fatal("4")
	log.(interface{ Unmount(context.Context) }).Unmount(ctx)
	log.(interface{ Unmount(context.Context) }).Unmount(context.Background())
	log.(interface{ Metadata() any }).Metadata()
	log.Debug("test panic")
}

func TestLoggerFormatterText(*testing.T) {
	log := NewLogger(&LoggerConfig{
		Caller:    true,
		Stdout:    true,
		Formatter: "text",
	})
	loggerWriteData(log)
}

func TestLoggerFormatterJSON(*testing.T) {
	log := NewLogger(nil)
	loggerWriteData(log)
	log.WithField("utf8", "世界\\ \n \r \t \002 \321 \u2028").Debug()
	log.WithField("field", new(marsha1)).Debug("marsha1")
	log.WithField("field", new(marsha2)).Debug("marsha2")
	log.WithField("field", new(marsha3)).Debug("marsha3")
	log.WithField("field", new(marsha4)).Debug("marsha4")
	log.WithField("field", new(marsha5)).Debug("marsha5")
}

type (
	marsha1 struct{}
	marsha2 struct{}
	marsha3 struct{}
	marsha4 struct{}
	marsha5 bool
)

func (marsha1) MarshalJSON() ([]byte, error) {
	return []byte("\"marsha1\""), nil
}

func (marsha2) MarshalJSON() ([]byte, error) {
	return []byte("\"marsha2\""), errors.New("test marshal error")
}

func (marsha3) MarshalText() ([]byte, error) {
	return []byte("\\ \n \r \t \002 \321 世界"), nil
}

func (marsha4) MarshalText() ([]byte, error) {
	return []byte("\\ \n \r \t \002 \321 世界"), errors.New("\\ \n \r \t \002 \321 世界")
}
func (marsha5) Method() {}

type StructCycle struct {
	Name string `json:"name,omitempty"`
	Err  error
	*StructCycle
}

type StructAnon struct {
	Duration *TimeDuration
	Now      *time.Time
	*LoggerConfig
	ServerConfig
}

func loggerWriteData(log Logger) {
	type M map[string]any
	var eptr *time.Time
	var eany any
	var cslice []any
	cmap := make(map[string]any)
	cycle := &StructCycle{}
	var echan chan int
	var dura TimeDuration
	cslice = append(cslice, "slice", 0, cslice)
	cmap["data"] = "map data"
	cmap["this"] = cmap
	cycle.StructCycle = cycle

	log = log.WithField("depth", "disable").WithField("logger", true)
	log.WithField("json", LoggerDebug).Debug()
	log.WithField("fmt.Stringer", HandlerFunc(HandlerEmpty)).Debug()
	log.WithField("fmt.Stringer", &dura).Debug()
	log.WithField("error", fmt.Errorf("logger wirte error")).Debug()
	log.WithField("bool", true).Debug()
	log.WithField("int", 1).Debug()
	log.WithField("uint", uint(2)).Debug()
	log.WithField("float", 3.3).Debug()
	log.WithField("complex", complex(4.1, 4.2)).Debug()
	log.WithField("map", map[string]int{"a": 1, "b": 2}).Debug()
	log.WithField("map alias", M{"a": 1, "b": 2}).Debug()
	log.WithField("map empty", map[string]int{}).Debug()
	log.WithField("map cycle", cmap).Debug()
	log.WithField("struct", struct{ Name string }{"name"}).Debug()
	log.WithField("struct empty", struct{}{}).Debug()
	log.WithField("struct cycle", cycle).Debug()
	log.WithField("struct anonymous", &StructAnon{}).Debug()
	log.WithField("ptr", &struct{ Name string }{"name"}).Debug()
	log.WithField("ptr empty", eptr).Debug()
	log.WithField("slice empty", cslice[0:0]).Debug()
	log.WithField("slice cycle", cslice).Debug()
	log.WithField("array", [...]int{1, 2, 3}).Debug()
	log.WithField("func", HandlerEmpty).Debug()
	log.WithField("bytes", []byte("bytes")).Debug()
	log.WithField("any empty", eany).Debug()
	log.WithField("any empty", []any{eany}).Debug()
	log.WithField("chan empty", echan).Debug()
	log.WithField("depth", "disable").Info("depth")
	log.WithField("depth", "enable").Info("depth")
	log.WithField("depth", "stack").Info("depth")
}

type logConfig struct {
	Level1 LoggerLevel `alias:"level" json:"level1"`
	Level2 LoggerLevel `alias:"level2" json:"level2"`
	Level3 LoggerLevel `alias:"level3" json:"level3"`
}

func TestLoggerLevel(t *testing.T) {
	conf := &logConfig{}
	jsonBlob := []byte(`{"level1":"1","level2":"info","level3":"3"}`)
	err := json.Unmarshal(jsonBlob, conf)
	t.Logf("%v\t%#v\n", err, conf)
	jsonBlob = []byte(`{"level3": "33"}`)
	err = json.Unmarshal(jsonBlob, conf)
	t.Logf("%v\t%#v\n", err, conf)
	json.Marshal(conf)
	t.Log(conf.Level1)
}

type appFatal struct{}

func (appFatal) SetValue(any, any) {}

func TestLoggerHookFatal(t *testing.T) {
	defer func() {
		t.Logf("LoggerHookFatal recover %v", recover())
	}()
	log := NewLogger(&LoggerConfig{
		Stdout:    true,
		HookFatal: true,
	})
	log.(interface{ Mount(context.Context) }).Mount(context.WithValue(
		context.Background(),
		ContextKeyAppCancel, context.CancelFunc(func() {}),
	))
	log.Fatal("stop app")

	log = NewLogger(&LoggerConfig{
		Stdout:    true,
		HookFatal: true,
	})
	log.Fatal("stop logger")
}

func TestLoggerHookFilter(t *testing.T) {
	fc := NewFuncCreator()
	ctx := context.WithValue(context.Background(),
		ContextKeyFuncCreator, fc,
	)
	log := NewLogger(&LoggerConfig{
		Stdout: true,
		HookFilter: [][]string{
			{"path string prefix=/static/", "status int equal:200"},
			{"mail setstring hidemail", "password setstring hide"},
			{"ti setint default", "tu setuint default", "tf setfloat default", "tb setbool default", "ta setany default"},
			{"1 1 1"},
			{"strs setstring default", "t setany add=240h"},
		},
	})
	log.(interface{ Mount(context.Context) }).Mount(ctx)

	log.WithFields([]string{"path", "status"}, []any{"/index", 200}).Info()
	log.WithFields([]string{"path", "status"}, []any{"/static/index", 200}).Info()
	log.WithFields([]string{"path"}, []any{true}).Info()
	log.WithField("mail", "postmaster@eudore.cn").WithField("password", "123456").Info()
	log.WithFields([]string{"ti", "tu", "tf", "tb", "ta"}, []any{1, uint(1), 1.0, true, time.Now()}).Info()
	log.WithFields([]string{"path"}, []any{nil}).Info()
	log.WithFields([]string{"strs", "t"}, []any{[]string{"", "2"}, time.Now()}).Debug()

	meta, ok := fc.(interface{ Metadata() any }).Metadata().(MetadataFuncCreator)
	if ok {
		for _, err := range meta.Errors {
			log.Debug("err:", err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	log.(interface{ Unmount(context.Context) }).Unmount(context.Background())
}

type loggerAsyncWait struct{}

func (w *loggerAsyncWait) HandlerPriority() int {
	return DefaultLoggerPriorityWriterAsync + 1
}

func (w *loggerAsyncWait) HandlerEntry(*LoggerEntry) {
	time.Sleep(time.Second * 10)
}

func TestLoggerWriterAsync(t *testing.T) {
	logfile := "tmp-loggerStd.log"
	defer os.Remove(logfile)
	log := NewLogger(&LoggerConfig{
		Handlers: []LoggerHandler{
			NewLoggerWriterAsync([]LoggerHandler{
				&loggerAsyncWait{},
			}, 5, 2048, time.Millisecond*50),
			NewLoggerWriterAsync([]LoggerHandler{
				&loggerAsyncWait{},
			}, 5, 2048, time.Millisecond),
			NewLoggerWriterStdout(true),
			NewLoggerWriterStdout(false),
		},
		Stdout:     true,
		TimeFormat: time.RFC3339Nano + " " + time.RFC3339Nano,
		Path:       logfile,
		AsyncSize:  5,
	})
	log.(interface{ Mount(context.Context) }).Mount(context.Background())
	for i := 0; i < 10; i++ {
		log.Info(i)
	}

	time.Sleep(time.Millisecond * 20)
	log.(interface{ Unmount(context.Context) }).Unmount(context.Background())
}

func TestLoggerWriterFile(t *testing.T) {
	defer func() {
		t.Logf("NewLoggerWriterFile recover %v", recover())
	}()

	// file
	logfile := "tmp-loggerStd.log"
	defer os.Remove(logfile)
	log := NewLogger(&LoggerConfig{
		Path: logfile,
	})
	log.Info("hello")

	// create error
	path := strings.Repeat("-", 256)
	func() {
		defer func() {
			t.Logf("NewLoggerWriterFile recover %v", recover())
		}()
		log = NewLogger(&LoggerConfig{
			Path: "out1" + path + "/1.log",
		})
	}()
	func() {
		defer func() {
			t.Logf("NewLoggerWriterFile recover %v", recover())
		}()
		log = NewLogger(&LoggerConfig{
			Path: "out2" + path + ".log",
		})
	}()
}

func TestNewLoggerWriterRotate(t *testing.T) {
	defer os.RemoveAll("logger")
	{
		// 占用一个索引文件 rotate跳过
		os.Mkdir("logger", 0o755)
		file, err := os.OpenFile("logger/app-2.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644,
		)
		if err == nil {
			str := []byte("eudore logger Writer test.")
			for i := 0; i < 1024; i++ {
				file.Write(str)
			}
			file.Close()
		}
	}

	// date
	log := NewLogger(&LoggerConfig{
		Stdout: true,
		Path:   "logger/app-yyyy-mm-dd-hh.log",
	})
	log.Info("hello")

	// size
	log = NewLogger(&LoggerConfig{
		Stdout:  true,
		Path:    "logger/app-size.log",
		MaxSize: 16 << 10,
	})
	log.Info("hello")

	// link
	log = NewLogger(&LoggerConfig{
		Path:     "logger/app.log",
		Link:     "logger/app.log",
		MaxSize:  16 << 10,
		MaxCount: 3,
	})
	log.Info("hello")

	log = log.WithFields(
		[]string{"name", "type"},
		[]any{"eudore", "logger"},
	).WithField("logger", true)
	for i := 0; i < 1000; i++ {
		log.Info("test rotate")
	}
}

func TestLoggerMonkeyErr(t *testing.T) {
	defer func() {
		t.Logf("MonkeyErr recover %v", recover())
	}()

	patchCallers := monkey.Patch(runtime.Callers, func(int, []uintptr) int { return 0 })
	defer patchCallers.Unpatch()
	patchOpen := monkey.Patch(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, fmt.Errorf("monkey no open")
	})
	defer patchOpen.Unpatch()

	GetCallerStacks(0)
	NewLogger(&LoggerConfig{
		Path: "app-yyyy-mm-dd.log",
	})
}

func TestLoggerMonkeyTime(t *testing.T) {
	defer os.RemoveAll("logger")
	log := NewLogger(&LoggerConfig{
		Path:     "logger/app-yyyy-mm-dd-hh.log",
		Link:     "logger/app.log",
		MaxSize:  1 << 10, // 1k
		MaxCount: 10,
	})

	// This is as unsafe as it sounds and
	// I don't recommend anyone do it outside of a testing environment.
	mytime := time.Now()
	var mu sync.RWMutex
	patch := monkey.Patch(time.Now, func() time.Time {
		mu.RLock()
		defer mu.RUnlock()
		return mytime
	})
	defer patch.Unpatch()

	for i := 0; i < 100; i++ {
		if i%30 == 9 {
			mu.Lock()
			mytime = mytime.Add(time.Hour)
			mu.Unlock()
		}
		log.Debug("now is", time.Now().String())
	}
}

func BenchmarkLoggerStd(b *testing.B) {
	data := map[string]any{
		"a": 1,
		"b": 2,
	}
	log := NewLogger(&LoggerConfig{
		Path: "t2.log",
	})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		log.WithFields(
			[]string{"animal", "number", "size"},
			[]any{"walrus", 1, 10},
		).Info("A walrus appears")
		log.WithField("a", 1).WithField("b", true).Info(data)
	}
	os.Remove("t2.log")
}
