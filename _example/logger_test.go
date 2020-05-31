package eudore_test

import (
	"bou.ke/monkey"
	"encoding/json"
	"errors"
	"github.com/eudore/eudore"
	"os"
	"runtime"
	"testing"
)

type loggerInitHandler2 interface {
	NextHandler(eudore.Logger)
}

func TestLoggerInit2(t *testing.T) {
	log := eudore.NewLoggerInit()
	log.SetLevel(eudore.LogInfo)
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

	log.WithFields(eudore.Fields{"key": "Fields"}).Debug("0")
	log.WithFields(eudore.Fields{"key": "Fields"}).Debugf("0")
	log.WithFields(eudore.Fields{"key": "Fields"}).Info("1")
	log.WithFields(eudore.Fields{"key": "Fields"}).Infof("1")
	log.WithFields(eudore.Fields{"key": "Fields"}).Warning("2")
	log.WithFields(eudore.Fields{"key": "Fields"}).Warningf("2")
	log.WithFields(eudore.Fields{"key": "Fields"}).Error("3")
	log.WithFields(eudore.Fields{"key": "Fields"}).Errorf("3")
	log.WithFields(eudore.Fields{"key": "Fields"}).Fatal("4")
	log.WithFields(eudore.Fields{"key": "Fields"}).Fatalf("4")

	// 判断是LoggerInit
	if initlog, ok := log.(loggerInitHandler2); ok {
		// 创建日志
		log2 := eudore.NewLoggerStd(nil)
		// 新日志处理LoggerInit保存的日志。
		initlog.NextHandler(log2)
		log = log2
	}

	log.SetLevel(eudore.LogDebug)
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

	log.WithFields(eudore.Fields{"key": "Fields"}).Debug("0")
	log.WithFields(eudore.Fields{"key": "Fields"}).Debugf("0")
	log.WithFields(eudore.Fields{"key": "Fields"}).Info("1")
	log.WithFields(eudore.Fields{"key": "Fields"}).Infof("1")
	log.WithFields(eudore.Fields{"key": "Fields"}).Warning("2")
	log.WithFields(eudore.Fields{"key": "Fields"}).Warningf("2")
	log.WithFields(eudore.Fields{"key": "Fields"}).Error("3")
	log.WithFields(eudore.Fields{"key": "Fields"}).Errorf("3")
	log.WithFields(eudore.Fields{"key": "Fields"}).Fatal("4")
	log.WithFields(eudore.Fields{"key": "Fields"}).Fatalf("4")

	log.Sync()
}

type (
	marsha1 struct{}
	marsha2 struct{}
	marsha3 struct{}
	marsha4 struct{}
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

func TestLoggerStd2(t *testing.T) {
	log := eudore.NewLoggerStd(nil)
	log.WithField("json", eudore.LogDebug).Debug()
	log.WithField("stringer", eudore.HandlerFunc(eudore.HandlerEmpty)).Debug("2")
	log.WithField("bool", true).Debug("2")
	log.WithField("int", 1).Debug("2")
	log.WithField("uint", uint(2)).Debug("2")
	log.WithField("float", 3.3).Debug("2")
	log.WithField("complex", complex(4.1, 4.2)).Debug("2")
	log.WithField("array", []int{1, 2, 3}).Debug("2")
	log.WithField("map", map[string]int{"a": 1, "b": 2}).Debug("2")
	log.WithField("struct", struct{ Name string }{"name"}).Debug("2")
	log.WithField("ptr", &struct{ Name string }{"name"}).Debug("2")
	log.WithField("func", TestLoggerStd2).Debug("2")
	log.WithField("bytes", []byte("bytes")).Debug("2")
	var i interface{}
	log.WithField("nil", i).Debug("2")

	log.WithField("utf8 string", "\\ \n \r \t \002 \321 世界").Debug("2")
	log.WithField("utf8 bytes", []byte("\\ \n \r \t \002 \321 世界")).Debug("2")

	log.WithField("nil", new(marsha1)).Debug("marsha1")
	log.WithField("nil", new(marsha2)).Debug("marsha2")
	log.WithField("nil", new(marsha3)).Debug("marsha3")
	log.WithField("nil", new(marsha4)).Debug("marsha4")

	log.Sync()
}

type logConfig struct {
	Level1 eudore.LoggerLevel `alias:"level" json:"level1"`
	Level2 eudore.LoggerLevel `alias:"level2" json:"level2"`
	Level3 eudore.LoggerLevel `alias:"level3" json:"level3"`
}

func TestLoggerLevel2(t *testing.T) {
	conf := &logConfig{}
	var jsonBlob = []byte(`{
  "level1": "1",
  "level2": "info",
  "level3": "3"
}`)
	err := json.Unmarshal(jsonBlob, conf)
	t.Logf("%v\t%#v\n", err, conf)
	jsonBlob = []byte(`{"level3": "33"}`)
	err = json.Unmarshal(jsonBlob, conf)
	t.Logf("%v\t%#v\n", err, conf)
}

func TestLoggerStdOut2(t *testing.T) {
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path: "/tmp/1/2/3.log",
	})
	log.Info("hello")
	log.Sync()

	logfile := "tmp-loggerStd.log"
	log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path: logfile,
	})
	log.Info("hello")
	log.Sync()
	os.Remove(logfile)

	log = eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Std:  true,
		Path: logfile,
	})
	log.Info("hello")
	log.Sync()
	os.Remove(logfile)
}

func TestLoggerStdOut3(t *testing.T) {
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Writer: eudore.NewLoggerWriterStd(),
	})
	log.Info("hello")
	log.Sync()
}

func TestLoggerStdOut4(t *testing.T) {
	defer func() {
		os.RemoveAll("logger")
		t.Log(recover())
	}()
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path: "logger/",
	})
	log.Info("hello")
	log.Sync()
}

func TestLoggerStdOut5(t *testing.T) {
	defer os.RemoveAll("logger")
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Std:     true,
		Path:    "logger/logger-out-index.log",
		MaxSize: 16 << 10,
	})
	log.Info("hello")
	log.Sync()
}

func TestLoggerStdOut6(t *testing.T) {
	defer func() {
		os.RemoveAll("logger")
		t.Log(recover())
	}()
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Std:     true,
		Path:    "logger/log-index/",
		MaxSize: 16 << 10,
	})
	log.Info("hello")
	log.Sync()
}

func TestLoggerCaller5(t *testing.T) {
	patch1 := monkey.Patch(runtime.Caller, func(int) (uintptr, string, int, bool) { return 0, "", 0, false })
	patch2 := monkey.Patch(runtime.Callers, func(int, []uintptr) int { return 0 })
	defer patch1.Unpatch()
	defer patch2.Unpatch()

	app := eudore.NewApp(eudore.NewRouterFull())
	app.AddMiddleware("8888")
	app.AnyFunc("/:path|panic", eudore.HandlerEmpty)
	app.AnyFunc("/*", eudore.HandlerRouter404)

	app.CancelFunc()
	app.Run()
}

func BenchmarkLogerStd(b *testing.B) {
	data := map[string]interface{}{
		"a": 1,
		"b": 2,
	}
	log := eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path: "t2.log",
	})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		log.WithFields(eudore.Fields{
			"animal": "walrus",
			"number": 1,
			"size":   10,
		}).Info("A walrus appears")
		log.WithField("a", 1).WithField("b", true).Info(data)
	}
	log.Sync()
	os.Remove("t2.log")
}
