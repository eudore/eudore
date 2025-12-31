package eudore_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/eudore/eudore"
)

type loggerWriterDiscard struct {
	Entrys []string
}

func (*loggerWriterDiscard) HandlerPriority() int {
	return DefaultLoggerPriorityWriterStdout
}
func (w *loggerWriterDiscard) HandlerEntry(entry *LoggerEntry) {
	if len(w.Entrys) > 0 {
		msg := string(entry.Buffer)
		if !strings.Contains(msg, w.Entrys[0]) {
			panic(msg)
		}
		w.Entrys = w.Entrys[1:]
	}
}

type loggerAsyncWait struct{}

func (w *loggerAsyncWait) HandlerPriority() int {
	return DefaultLoggerPriorityWriterAsync + 1
}
func (w *loggerAsyncWait) HandlerEntry(*LoggerEntry) {
	time.Sleep(time.Second * 10)
}

func TestLoggerStd(t *testing.T) {
	entrys := []string{
		`{"time":"none","level":"FATAL","message":"4"}`,
		`{"time":"none","level":"FATAL","message":"4"}`,
		`{"time":"none","level":"DEBUG","message":"0"}`,
		`{"time":"none","level":"DEBUG","message":"0"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"WARNING","message":"2"}`,
		`{"time":"none","level":"WARNING","message":"2"}`,
		`{"time":"none","level":"ERROR","message":"3"}`,
		`{"time":"none","level":"ERROR","message":"3"}`,
		`{"time":"none","level":"FATAL","message":"4"}`,
		`{"time":"none","level":"FATAL","message":"4"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"INFO","depth":true,"message":"1"}`,
		`{"time":"none","level":"INFO","caller":"logger","message":"1"}`,
		`{"time":"none","level":"INFO","error":"Logger: The number of field keys and values are not equal","message":"1"}`,
	}
	log := NewLogger(&LoggerConfig{
		Stdout:     false,
		TimeFormat: "none",
		Handlers:   []LoggerHandler{&loggerWriterDiscard{entrys}},
	})

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
}

func TestLoggerInit(t *testing.T) {
	entrys := []string{
		`{"time":"none","level":"INFO","message":"loggerInit to end"}`,
		`{"time":"none","level":"DEBUG","message":"0"}`,
		`{"time":"none","level":"INFO","message":"1"}`,
		`{"time":"none","level":"WARNING","message":"2"}`,
		`{"time":"none","level":"ERROR","message":"3"}`,
		`{"time":"none","level":"FATAL","message":"4"}`,
	}
	ctx := context.WithValue(context.Background(),
		ContextKeyLogger, NewLogger(&LoggerConfig{
			Stdout:     false,
			TimeFormat: "none",
			Handlers:   []LoggerHandler{&loggerWriterDiscard{entrys}},
		}),
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

	defer func() { recover() }()
	log.Debug("test panic")
}

func TestLoggerFormatterText(*testing.T) {
	entrys := []string{
		`none DEBUG json="DEBUG"`,
		`none DEBUG fmt.Stringer="="`,
		`none DEBUG fmt.Stringer="0s"`,
		`none DEBUG error="logger wirte error"`,
		`none DEBUG bool=true`,
		`none DEBUG int=1`,
		`none DEBUG uint=2`,
		`none DEBUG float=3.3`,
		`none DEBUG complex="4.1+4.2i"`,
		`none DEBUG map=map`,
		`none DEBUG map alias=eudore_test.M`,
		`none DEBUG map empty=map{}`,
		`none DEBUG map cycle=map`,
		`none DEBUG struct={Name:"name"}`,
		`none DEBUG struct empty={}`,
		`none DEBUG struct cycle=&{Name:"" Err:null loggerStructCycle:`,
		`none DEBUG struct anonymous=&{Duration:null Now:null LoggerConfig:null ServerConfig:{Handler:null ReadTimeout:"0s" WriteTimeout:"0s" ReadHeaderTimeout:"0s" IdleTimeout:"0s" MaxHeaderBytes:0 ErrorLog:null BaseContext:null ConnContext:null}}`,
		`none DEBUG ptr=&{Name:"name"}`,
		`none DEBUG ptr empty=null`,
		`none DEBUG slice empty=[]`,
		`none DEBUG slice cycle=["slice",0,null]`,
		`none DEBUG array=[1,2,3]`,
		`none DEBUG func=`,
		`none DEBUG bytes=[98,121,116,101,115]`,
		`none DEBUG any empty=null`,
		`none DEBUG any empty=[null]`,
		`none DEBUG chan empty=null`,
		`none INFO depth`,
		`none INFO depth`,
		`none INFO depth`,
	}
	log := NewLogger(&LoggerConfig{
		Stdout:     false,
		Formatter:  "text",
		TimeFormat: "none",
		Handlers:   []LoggerHandler{&loggerWriterDiscard{entrys}},
	})
	loggerWriteData(log)
}

func TestLoggerFormatterJSON(*testing.T) {
	entrys := []string{
		`{"time":"none","level":"DEBUG","json":"DEBUG"}`,
		`{"time":"none","level":"DEBUG","fmt.Stringer":"="}`,
		`{"time":"none","level":"DEBUG","fmt.Stringer":"0s"}`,
		`{"time":"none","level":"DEBUG","error":"logger wirte error"}`,
		`{"time":"none","level":"DEBUG","bool":true}`,
		`{"time":"none","level":"DEBUG","int":1}`,
		`{"time":"none","level":"DEBUG","uint":2}`,
		`{"time":"none","level":"DEBUG","float":3.3}`,
		`{"time":"none","level":"DEBUG","complex":"4.1+4.2i"}`,
		`{"time":"none","level":"DEBUG","map":`,
		`{"time":"none","level":"DEBUG","map alias":`,
		`{"time":"none","level":"DEBUG","map empty":{}}`,
		`{"time":"none","level":"DEBUG","map cycle":`,
		`{"time":"none","level":"DEBUG","struct":{"Name":"name"}}`,
		`{"time":"none","level":"DEBUG","struct empty":{}}`,
		`{"time":"none","level":"DEBUG","struct cycle":{"Err":null}}`,
		`{"time":"none","level":"DEBUG","struct anonymous":{"Duration":null,"Now":null,"LoggerConfig":null,"readTimeout":"0s","writeTimeout":"0s","readHeaderTimeout":"0s","idleTimeout":"0s","maxHeaderBytes":0}}`,
		`{"time":"none","level":"DEBUG","ptr":{"Name":"name"}}`,
		`{"time":"none","level":"DEBUG","ptr empty":null}`,
		`{"time":"none","level":"DEBUG","slice empty":[]}`,
		`{"time":"none","level":"DEBUG","slice cycle":["slice",0,null]}`,
		`{"time":"none","level":"DEBUG","array":[1,2,3]}`,
		`{"time":"none","level":"DEBUG","func":`,
		`{"time":"none","level":"DEBUG","bytes":[98,121,116,101,115]}`,
		`{"time":"none","level":"DEBUG","any empty":null}`,
		`{"time":"none","level":"DEBUG","any empty":[null]}`,
		`{"time":"none","level":"DEBUG","chan empty":null}`,
		`{"time":"none","level":"INFO","message":"depth"}`,
		`{"time":"none","level":"INFO","message":"depth"}`,
		`{"time":"none","level":"INFO","message":"depth"}`,
		`{"time":"none","level":"DEBUG","utf8":"世界\\ \n \r \t \u0002 \ufffd \u2028"}`,
		`{"time":"none","level":"DEBUG","field":"marsha1","message":"marsha1"}`,
		`{"time":"none","level":"DEBUG","field":"test marshal error","message":"marsha2"}`,
		`{"time":"none","level":"DEBUG","field":"\\ \n \r \t \u0002 \ufffd 世界","message":"marsha3"}`,
		`{"time":"none","level":"DEBUG","field":"\\ \n \r \t \u0002 \ufffd 世界","message":"marsha4"}`,
		`{"time":"none","level":"DEBUG","field":false,"message":"marsha5"}`,
	}
	log := NewLogger(&LoggerConfig{
		Stdout:     false,
		Formatter:  "json",
		TimeFormat: "none",
		Handlers:   []LoggerHandler{&loggerWriterDiscard{entrys}},
	})
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

type loggerStructCycle struct {
	Name string `json:"name,omitempty"`
	Err  error
	*loggerStructCycle
}

type loggerStructAnon struct {
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
	cycle := &loggerStructCycle{}
	var echan chan int
	var dura TimeDuration
	cslice = append(cslice, "slice", 0, cslice)
	cmap["data"] = "map data"
	cmap["this"] = cmap
	cycle.loggerStructCycle = cycle

	log = log.WithField("depth", "disable").WithField("logger", true)
	log.WithField("json", LoggerDebug).Debug()
	log.WithField("fmt.Stringer", Cookie{}).Debug()
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
	log.WithField("struct anonymous", &loggerStructAnon{}).Debug()
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
	if err != nil {
		t.Error(err)
	}
	json.Marshal(conf)
	fmt.Sprint(conf.Level1)

	err = json.Unmarshal([]byte(`{"level3": "33"}`), conf)
}

type appFatal struct{}

func (appFatal) SetValue(any, any) {}

func TestLoggerHook(t *testing.T) {
	// std
	NewLogger(&LoggerConfig{Handlers: []LoggerHandler{
		NewLoggerWriterStdout(true),
		NewLoggerWriterStdout(false),
	}}).Info("Stdout")
	NewLogger(&LoggerConfig{TimeFormat: strings.Repeat("none", 16),
		Handlers: []LoggerHandler{NewLoggerWriterStdout(true)},
	}).Info("Stdout")

	// caller
	{
		log := NewLogger(&LoggerConfig{
			Caller:    true,
			Stdout:    false,
			Formatter: "text",
		})
		log.Info("callers")
		log.WithField(FieldDepth, DefaultLoggerDepthKindStack).Info("stacks")
		GetCallerStacks(0)
		GetCallerStacks(64)
	}

	// fatal
	{
		log := NewLogger(&LoggerConfig{
			Stdout:    false,
			HookFatal: true,
		})
		log.(interface{ Mount(context.Context) }).Mount(context.WithValue(
			context.Background(),
			ContextKeyAppCancel, context.CancelFunc(func() {}),
		))
		log.Fatal("stop app")
		func() {
			defer func() { recover() }()
			NewLogger(&LoggerConfig{
				Stdout:    false,
				HookFatal: true,
			}).Fatal("stop logger")
		}()
	}
}

func TestLoggerHookFilter(t *testing.T) {
	fc := NewFuncCreator()
	ctx := context.WithValue(context.Background(), ContextKeyFuncCreator, fc)

	entrys := []string{
		`{"time":"none","level":"INFO","path":"/index","status":200}`,
		`{"time":"none","level":"INFO","path":true}`,
		`{"time":"none","level":"INFO","mail":"pos****@eudore.cn","password":"***"}`,
		`{"time":"none","level":"INFO","ti":0,"tu":0,"tf":0,"tb":false,"ta":"0001-01-01T00:00:00Z"}`,
		`{"time":"none","level":"INFO","path":null}`,
		`{"time":"none","level":"DEBUG","strs":["",""],"t":"2024-12-24T16:00:00Z"}`,
		`{"time":"none","level":"DEBUG","message":"err: funcCreator create kind invalid func 1 err: invalid func kind 0"}`,
	}
	log := NewLogger(&LoggerConfig{
		Stdout:     false,
		TimeFormat: "none",
		Handlers:   []LoggerHandler{&loggerWriterDiscard{entrys}},
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
	log.WithFields([]string{"strs", "t"}, []any{[]string{"", "2"}, time.Unix(1734192000, 0).UTC()}).Debug()

	meta, ok := fc.(interface{ Metadata() any }).Metadata().(MetadataFuncCreator)
	if ok {
		for _, err := range meta.Errors {
			log.Debug("err:", err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	log.(interface{ Unmount(context.Context) }).Unmount(context.Background())
}

func loggerRelease(data any) {
	closer, ok := data.(interface{ Unmount(ctx context.Context) })
	if ok {
		closer.Unmount(context.Background())
	}
}

func TestLoggerWriterAsync(t *testing.T) {
	logfile := "tmp-loggerStd.log"
	loggerRelease(NewLogger(&LoggerConfig{Stdout: true, Path: logfile, AsyncSize: 64}))
	defer os.Remove(logfile)
	log := NewLogger(&LoggerConfig{
		Handlers: []LoggerHandler{
			NewLoggerWriterAsync([]LoggerHandler{
				&loggerAsyncWait{},
			}, 5, 2048, time.Millisecond*50),
			&loggerWriterDiscard{},
			NewLoggerWriterAsync([]LoggerHandler{
				&loggerAsyncWait{},
			}, 5, 2048, time.Millisecond),
			&loggerWriterDiscard{},
		},
		Stdout:     false,
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
	// file
	logfile := "tmp-loggerStd.log"
	defer os.Remove(logfile)
	log := NewLogger(&LoggerConfig{
		Path: logfile,
	})
	log.Info("hello")
	loggerRelease(log)

	// create error
	path := strings.Repeat("-", 256)
	func() {
		defer func() { recover() }()
		log = NewLogger(&LoggerConfig{
			Path: "out1" + path + "/1.log",
		})
	}()
	func() {
		defer func() { recover() }()
		log = NewLogger(&LoggerConfig{
			Path: "out2" + path + ".log",
		})
	}()
	func() {
		defer func() { recover() }()
		log = NewLogger(&LoggerConfig{
			Path: "out3-yyyy-mm-dd" + path + ".log",
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
	{
		log := NewLogger(&LoggerConfig{
			Path: "logger/app-yyyy-mm-dd-hh.log",
		})
		log.Info("hello")
		loggerRelease(log)
	}

	// size
	{
		log := NewLogger(&LoggerConfig{
			Path:    "logger/app-size.log",
			MaxSize: 16 << 10,
		})
		log.Info("hello")
		loggerRelease(log)
	}

	// link
	{
		log := NewLogger(&LoggerConfig{
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
		loggerRelease(log)
	}

	// rotate
	{
		log := NewLogger(&LoggerConfig{
			Path:    "logger/app-yyyy-mm-dd-hh.log",
			MaxSize: 1 << 10, // 1k
		})
		now := time.Now()
		for i := 0; i < 100; i++ {
			if i%30 == 9 {
				now = now.Add(time.Hour)
			}
			log.WithField(FieldTime, now).Debug("now is", now.String())
		}
		loggerRelease(log)
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
