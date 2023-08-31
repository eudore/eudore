package eudore_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/eudore/eudore"
)

func TestLoggerStd(t *testing.T) {
	app := eudore.NewApp()

	app.GetLevel()
	app.SetLevel(eudore.LoggerFatal)
	app.Debug("0")
	app.Debugf("0")
	app.Info("1")
	app.Infof("1")
	app.Warning("2")
	app.Warningf("2")
	app.Error("3")
	app.Errorf("3")
	app.Fatal("4")
	app.Fatalf("4")

	app.SetLevel(eudore.LoggerDebug)
	app.Debug("0")
	app.Debugf("0")
	app.Info("1")
	app.Infof("1")
	app.Warning("2")
	app.Warningf("2")
	app.Error("3")
	app.Errorf("3")
	app.Fatal("4")
	app.Fatalf("4")

	app.WithField("depth", "enable").Info("1")
	app.WithField("depth", "stack").Info("1")
	app.WithField("depth", "disable").Info("1")
	app.WithField("depth", -2).WithField("depth", "enable").Info("1")
	app.WithField("depth", true).Info("1")
	app.WithField("context", app).WithField("context", app.Context).Info("1")
	app.WithField("caller", "logger").WithField("logger", true).Info("1")
	app.WithFields([]string{"key"}, []any{}).Info("1")

	app.Logger.(interface{ Metadata() any }).Metadata()
	eudore.NewLoggerWithContext(app)
	eudore.NewLoggerWithContext(context.Background())

	app.CancelFunc()
	app.Run()
}

func TestLoggerInit1(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())

	app.Debug("0")
	app.Infof("1")
	app.Warning("2")
	app.Error("3")
	app.Fatal("4")
	app.Logger.(interface{ Metadata() any }).Metadata()

	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(nil))
	app.CancelFunc()
	app.Run()
}

func TestLoggerInit2(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.Info("loggerInit to end")
	app.CancelFunc()
	app.Run()
}

func TestLoggerFormatterText(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Caller:    true,
		Stdout:    true,
		Formatter: "text",
	}))
	loggerWriteData(app)

	app.CancelFunc()
	app.Run()
}

func TestLoggerFormatterJSON(t *testing.T) {
	app := eudore.NewApp()
	loggerWriteData(app)
	app.WithField("utf8", "世界\\ \n \r \t \002 \321 \u2028").Debug()
	app.WithField("field", new(marsha1)).Debug("marsha1")
	app.WithField("field", new(marsha2)).Debug("marsha2")
	app.WithField("field", new(marsha3)).Debug("marsha3")
	app.WithField("field", new(marsha4)).Debug("marsha4")
	app.WithField("field", new(marsha5)).Debug("marsha5")
}

type marsha1 struct{}
type marsha2 struct{}
type marsha3 struct{}
type marsha4 struct{}
type marsha5 bool

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
	Duration *eudore.TimeDuration
	Now      *time.Time
	*eudore.LoggerConfig
	eudore.ServerConfig
}

func loggerWriteData(log eudore.Logger) {
	type M map[string]any
	var eptr *time.Time
	var eany any
	var cslice []any
	var cmap = make(map[string]any)
	var cycle = &StructCycle{}
	var echan chan int
	var dura eudore.TimeDuration
	cslice = append(cslice, "slice", 0, cslice)
	cmap["data"] = "map data"
	cmap["this"] = cmap
	cycle.StructCycle = cycle

	log = log.WithField("depth", "disable").WithField("logger", true)
	log.WithField("json", eudore.LoggerDebug).Debug()
	log.WithField("fmt.Stringer", eudore.HandlerFunc(eudore.HandlerEmpty)).Debug()
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
	log.WithField("array", []int{1, 2, 3}).Debug()
	log.WithField("func", eudore.NewApp).Debug()
	log.WithField("bytes", []byte("bytes")).Debug()
	log.WithField("any empty", eany).Debug()
	log.WithField("any empty", []any{eany}).Debug()
	log.WithField("chan empty", echan).Debug()
	log.WithField("depth", "disable").Info("depth")
	log.WithField("depth", "enable").Info("depth")
	log.WithField("depth", "stack").Info("depth")
}

type logConfig struct {
	Level1 eudore.LoggerLevel `alias:"level" json:"level1"`
	Level2 eudore.LoggerLevel `alias:"level2" json:"level2"`
	Level3 eudore.LoggerLevel `alias:"level3" json:"level3"`
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

func TestLoggerHookFatal(t *testing.T) {
	defer func() {
		t.Logf("LoggerHookFatal recover %v", recover())
	}()

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:    true,
		HookFatal: true,
	}))

	app.Fatal("stop app")
	app.Run()

	log := eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:    true,
		HookFatal: true,
	})
	log.Fatal("stop logger")
}

func TestLoggerHookFilter(t *testing.T) {
	app := eudore.NewApp()
	fc := eudore.NewFuncCreator()
	app.SetValue(eudore.ContextKeyFuncCreator, fc)
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout: true,
		HookFilter: [][]string{
			{"path string prefix=/static/", "status int equal:200"},
			{"mail setstring hidemail", "password setstring hide"},
			{"ti setint default", "tu setuint default", "tf setfloat default", "tb setbool default", "ta setany default"},
			{"1 1 1"},
			{"strs setstring default", "t setany add=240h"},
		},
	}))

	app.WithFields([]string{"path", "status"}, []any{"/index", 200}).Info()
	app.WithFields([]string{"path", "status"}, []any{"/static/index", 200}).Info()
	app.WithFields([]string{"path"}, []any{true}).Info()
	app.WithField("mail", "postmaster@eudore.cn").WithField("password", "123456").Info()
	app.WithFields([]string{"ti", "tu", "tf", "tb", "ta"}, []any{1, uint(1), 1.0, true, time.Now()}).Info()
	app.WithFields([]string{"path"}, []any{nil}).Info()
	app.WithFields([]string{"strs", "t"}, []any{[]string{"", "2"}, time.Now()}).Debug()

	meta, ok := fc.(interface{ Metadata() any }).Metadata().(eudore.MetadataFuncCreator)
	if ok {
		for _, err := range meta.Errors {
			app.Debug("err:", err)
		}
	}

	time.Sleep(1000 * time.Millisecond)
	app.CancelFunc()
	app.Run()
}

func TestLoggerWriterStdout(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:   true,
		StdColor: true,
	}))
	app.Info("color")

	eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:     true,
		StdColor:   true,
		TimeFormat: time.RFC3339Nano + " " + time.RFC3339Nano,
	}).Debug("disable color")

	app.CancelFunc()
	app.Run()
}

func TestLoggerWriterFile(t *testing.T) {
	defer func() {
		t.Logf("NewLoggerWriterFile recover %v", recover())
	}()

	// file
	logfile := "tmp-loggerStd.log"
	log := eudore.NewLogger(&eudore.LoggerConfig{
		Path: logfile,
	})
	defer os.Remove(logfile)
	log.Info("hello")

	// create error
	func() {
		defer func() {
			t.Logf("NewLoggerWriterFile recover %v", recover())
		}()
		log = eudore.NewLogger(&eudore.LoggerConfig{
			Path: "out----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------/1.log",
		})
	}()
	func() {
		defer func() {
			t.Logf("NewLoggerWriterFile recover %v", recover())
		}()
		log = eudore.NewLogger(&eudore.LoggerConfig{
			Path: "out----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------.log",
		})
	}()
}

func TestNewLoggerWriterRotate(t *testing.T) {
	defer os.RemoveAll("logger")
	{
		// 占用一个索引文件 rotate跳过
		os.Mkdir("logger", 0o755)
		file, err := os.OpenFile("logger/app-2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			str := []byte("eudore logger Writer test.")
			for i := 0; i < 1024; i++ {
				file.Write(str)
			}
			file.Close()
		}
	}

	// date
	log := eudore.NewLogger(&eudore.LoggerConfig{
		Stdout: true,
		Path:   "logger/app-yyyy-mm-dd-hh.log",
	})
	log.Info("hello")

	// size
	log = eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:  true,
		Path:    "logger/app-size.log",
		MaxSize: 16 << 10,
	})
	log.Info("hello")

	// link
	log = eudore.NewLogger(&eudore.LoggerConfig{
		Path:     "logger/app.log",
		Link:     "logger/app.log",
		MaxSize:  16 << 10,
		MaxCount: 3,
	})
	log.Info("hello")

	log = log.WithFields([]string{"name", "type"}, []any{"eudore", "logger"}).WithField("logger", true)
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

	eudore.GetCallerStacks(0)
	eudore.NewLogger(&eudore.LoggerConfig{
		Path: "app-yyyy-mm-dd.log",
	})
}

func TestLoggerMonkeyTime(t *testing.T) {
	defer os.RemoveAll("logger")
	log := eudore.NewLogger(&eudore.LoggerConfig{
		Path:     "logger/app-yyyy-mm-dd-hh.log",
		Link:     "logger/app.log",
		MaxSize:  1 << 10, // 1k
		MaxCount: 10,
	})

	// This is as unsafe as it sounds and I don't recommend anyone do it outside of a testing environment.
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
	log := eudore.NewLogger(&eudore.LoggerConfig{
		Path: "t2.log",
	})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		log.WithFields([]string{"animal", "number", "size"}, []any{"walrus", 1, 10}).Info("A walrus appears")
		log.WithField("a", 1).WithField("b", true).Info(data)
	}
	os.Remove("t2.log")
}
