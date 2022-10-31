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

func TestLogger(t *testing.T) {
	log := eudore.NewApp()
	log.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())

	log.SetLevel(eudore.LoggerFatal)
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

	log.SetLevel(eudore.LoggerDebug)
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

	log.WithField("key", "field").Debug("0")
	log.WithField("key", "field").Debugf("0")
	log.WithField("key", "field").Info("1")
	log.WithField("key", "field").Infof("1")
	log.WithField("key", "field").Warning("2")
	log.WithField("key", "field").Warningf("2")
	log.WithField("key", "field").Error("3")
	log.WithField("key", "field").Errorf("3")
	log.WithField("key", "field").Fatal("4")
	log.WithField("key", "field").Fatalf("4")

	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Debug("0")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Debugf("0")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Info("1")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Infof("1")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Warning("2")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Warningf("2")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Error("3")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Errorf("3")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Fatal("4")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Fatalf("4")

	log.Sync()
	// 设置logger
	log.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerStd(nil))

	log.SetLevel(eudore.LoggerDebug)
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

	log.WithField("key", "field").Debug("0")
	log.WithField("key", "field").Debugf("0")
	log.WithField("key", "field").Info("1")
	log.WithField("key", "field").Infof("1")
	log.WithField("key", "field").Warning("2")
	log.WithField("key", "field").Warningf("2")
	log.WithField("key", "field").Error("3")
	log.WithField("key", "field").Errorf("3")
	log.WithField("key", "field").Fatal("4")
	log.WithField("key", "field").Fatalf("4")

	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Debug("0")
	log.WithFields([]string{"key", "k2"}, []interface{}{"Fields"}).Debug("0")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Debugf("0")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Info("1")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Infof("1")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Warning("2")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Warningf("2")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Error("3")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Errorf("3")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Fatal("4")
	log.WithFields([]string{"key"}, []interface{}{"Fields"}).Fatalf("4")

	log.WithField("depth", "stack").Info("1")

	log.Sync()
	eudore.DefaultLoggerNull.Sync()

	log.CancelFunc()
	log.Run()
}

func TestLoggerInit(t *testing.T) {
	log := eudore.NewApp()
	log.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	log.Info("loggerInit to end")
	log.CancelFunc()
	log.Run()
}

func TestLoggerOption(t *testing.T) {
	log := eudore.NewLoggerStd(nil)

	// logger depth
	log = log.WithField("depth", "enable").WithField("context", context.Background()).WithField("caller", "log depth").WithField("logger", true)
	log.Info("file line")
	log.WithField("depth", "disable").Info("file line")
	log.WithField("depth", []string{"disable"}).Info("file line")

	log.WithField("context", context.TODO()).Info("logger context")
}

func TestLoggerWriterFile(t *testing.T) {
	defer func() {
		t.Logf("NewLoggerWriterFile recover %v", recover())
	}()

	// file
	logfile := "tmp-loggerStd.log"
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path: logfile,
	})
	defer os.Remove(logfile)

	log.Info("hello")
	log.Sync()
	os.Remove(logfile)

	// file and std
	log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Std:  true,
		Path: logfile,
	})
	log.Info("hello")
	log.Sync()
	os.Remove(logfile)

	// create error
	func() {
		defer func() {
			t.Logf("NewLoggerWriterFile recover %v", recover())
		}()
		log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
			Path: "out-yyyy-MM-dd-HH-index---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------.log",
		})
	}()
	func() {
		defer func() {
			t.Logf("NewLoggerWriterFile recover %v", recover())
		}()
		log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
			Path: "out----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------.log",
		})
	}()
	log.Info("hello")
	log.Sync()
}

func TestNewLoggerWriterRotate(t *testing.T) {
	defer os.RemoveAll("logger")
	{
		// 占用一个索引文件 rotate跳过
		os.Mkdir("logger", 0644)
		file, err := os.OpenFile("logger/logger-out-2.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			str := []byte("eudore logger Writer test.")
			for i := 0; i < 1024; i++ {
				file.Write(str)
			}
			file.Sync()
			file.Close()
		}
	}

	// date
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Std:  true,
		Path: "logger/logger-yyyy-MM-dd-HH-index.log",
	})
	log.Info("hello")

	// no link
	log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Std:     true,
		Path:    "logger/logger-out-index.log",
		MaxSize: 16 << 10,
	})
	log.Info("hello")

	// rotate file
	log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path:    "logger/logger-out-index.log",
		MaxSize: 16 << 10,
		Link:    "logger/app.log",
	})
	log.Info("hello")

	log = log.WithFields([]string{"name", "type"}, []interface{}{"eudore", "logger"}).WithField("logger", true)
	for i := 0; i < 1000; i++ {
		log.Info("test rotate")
	}
	log.Sync()
}

type (
	marsha1 struct{}
	marsha2 struct{}
	marsha3 struct{}
	marsha4 struct{}
	marsha5 struct{ Num []int }
)

func (marsha1) MarshalJSON() ([]byte, error) {
	return []byte("marsha1"), nil
}
func (marsha2) MarshalJSON() ([]byte, error) {
	return []byte("marsha2"), errors.New("test marshal error")
}
func (marsha3) MarshalText() ([]byte, error) {
	return []byte("\\ \n \r \t \002 \321 世界"), nil
}
func (marsha4) MarshalText() ([]byte, error) {
	return []byte("\\ \n \r \t \002 \321 世界"), errors.New("\\ \n \r \t \002 \321 世界")
}

func TestLoggerStdJSON(t *testing.T) {
	var ptr *time.Time
	var slice = []int{1}
	log := eudore.NewLoggerStd(nil)
	log.Debug("debug")
	log.Info("info")
	log.WithField("json", eudore.LoggerDebug).Debug()
	log.WithField("stringer", eudore.HandlerFunc(eudore.HandlerEmpty)).Debug("2")
	log.WithField("error", fmt.Errorf("logger wirte error")).Debug()
	log.WithField("bool", true).Debug("2")
	log.WithField("int", 1).Debug("2")
	log.WithField("uint", uint(2)).Debug("2")
	log.WithField("float", 3.3).Debug("2")
	log.WithField("complex", complex(4.1, 4.2)).Debug("2")
	log.WithField("array", []int{1, 2, 3}).Debug("2")
	log.WithField("map", map[string]int{"a": 1, "b": 2}).Debug("2")
	log.WithField("struct", struct{ Name string }{"name"}).Debug("2")
	log.WithField("struct", struct{}{}).Debug("2")
	log.WithField("ptr", &struct{ Name string }{"name"}).Debug("2")
	log.WithField("ptr", ptr).Debug("2")
	log.WithField("slice", slice[0:0]).Debug("2")
	log.WithField("emptry face", []interface{}{ptr}).Debug("2")
	log.WithField("func", TestLoggerStdJSON).Debug("2")
	log.WithField("bytes", []byte("bytes")).Debug("2")
	var i interface{}
	log.WithField("nil", i).Debug("2")

	log.WithField("utf8 string", "\\ \n \r \t \002 \321 世界").Debug("2")
	log.WithField("utf8 bytes", []byte("\\ \n \r \t \002 \321 世界")).Debug("2")

	log.WithField("nil", new(marsha1)).Debug("marsha1")
	log.WithField("nil", new(marsha2)).Debug("marsha2")
	log.WithField("nil", new(marsha3)).Debug("marsha3")
	log.WithField("nil", new(marsha4)).Debug("marsha4")
	log.WithField("nil", new(marsha5)).Debug("marsha5")

	log.Sync()
}

type logConfig struct {
	Level1 eudore.LoggerLevel `alias:"level" json:"level1"`
	Level2 eudore.LoggerLevel `alias:"level2" json:"level2"`
	Level3 eudore.LoggerLevel `alias:"level3" json:"level3"`
}

func TestLoggerLevel(t *testing.T) {
	conf := &logConfig{}
	var jsonBlob = []byte(`{"level1":"1","level2":"info","level3":"3"}`)
	err := json.Unmarshal(jsonBlob, conf)
	t.Logf("%v\t%#v\n", err, conf)
	jsonBlob = []byte(`{"level3": "33"}`)
	err = json.Unmarshal(jsonBlob, conf)
	t.Logf("%v\t%#v\n", err, conf)
}

type loggerStdData016 struct {
	eudore.LoggerStdData
	meta *loggerStdData016Meta
}

type loggerStdData016Meta struct {
	debug   int64
	info    int64
	warning int64
	error   int64
	fatal   int64
}

func (log loggerStdData016) GetLogger() *eudore.LoggerStd {
	entry := log.LoggerStdData.GetLogger()
	_, ok := entry.LoggerStdData.(loggerStdData016)
	if !ok {
		entry.LoggerStdData = loggerStdData016{entry.LoggerStdData, log.meta}
	}
	return entry
}

func (log loggerStdData016) PutLogger(entry *eudore.LoggerStd) {
	switch entry.Level {
	case eudore.LoggerDebug:
		log.meta.debug++
	case eudore.LoggerInfo:
		log.meta.info++
	case eudore.LoggerWarning:
		log.meta.warning++
	case eudore.LoggerError:
		log.meta.error++
	case eudore.LoggerFatal:
		log.meta.fatal++
	}
	log.LoggerStdData.PutLogger(entry)
}

func (log loggerStdData016) Metadata() interface{} {
	return map[string]interface{}{
		"name":    "loggerStdData016",
		"debug":   log.meta.debug,
		"info":    log.meta.info,
		"warning": log.meta.warning,
		"ererror": log.meta.error,
		"fatal":   log.meta.fatal,
	}
}

func TestMetadata(t *testing.T) {
	app := eudore.NewApp()
	meta, ok := app.Logger.(interface{ Metadata() interface{} })
	if ok {
		app.Infof("%#v", meta.Metadata())
	}

	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerStd(&loggerStdData016{eudore.NewLoggerStdDataJSON(nil), &loggerStdData016Meta{}}))
	meta, ok = app.Logger.(interface{ Metadata() interface{} })
	if ok {
		app.Infof("%#v", meta.Metadata())
	}

	app.CancelFunc()
	app.Run()
}

func TestLoggerMonkey(t *testing.T) {
	patch1 := monkey.Patch(runtime.Caller, func(int) (uintptr, string, int, bool) { return 0, "", 0, false })
	patch2 := monkey.Patch(runtime.Callers, func(int, []uintptr) int { return 0 })
	defer patch1.Unpatch()
	defer patch2.Unpatch()

	log := eudore.NewLoggerStd(nil)
	log.WithField("depth", "enable").Error(eudore.GetPanicStack(0))

	defer os.RemoveAll("logger")
	log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path:       "logger/logger-yyyy-MM-dd-HH-index.log",
		Link:       "logger/logger.log",
		MaxSize:    1 << 10, // 1k
		Std:        false,
		Level:      eudore.LoggerDebug,
		TimeFormat: "Mon Jan 2 15:04:05 -0700 MST 2006",
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
	data := map[string]interface{}{
		"a": 1,
		"b": 2,
	}
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path: "t2.log",
	})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		log.WithFields([]string{"animal", "number", "size"}, []interface{}{"walrus", 1, 10}).Info("A walrus appears")
		log.WithField("a", 1).WithField("b", true).Info(data)
	}
	log.Sync()
	os.Remove("t2.log")
}
