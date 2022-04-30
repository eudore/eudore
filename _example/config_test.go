package eudore_test

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/eudore/eudore"
)

func TestConfigMap(t *testing.T) {
	app := eudore.NewApp()
	config := eudore.NewConfigMap(map[string]interface{}{
		"name":   "eudore",
		"type":   "ConfigMap",
		"number": 3,
	})
	app.SetValue(eudore.ContextKeyConfig, config)
	// 获取全部数据
	config.Set("auth.secret", "secret")

	app.Infof("ConfigMap data: %# v", config.Get(""))
	app.Infof("ConfigMap data name: %v", config.Get("name"))

	// 设置基础map
	config.Set("", map[string]interface{}{
		"data": "new data",
	})
	app.Infof("ConfigMap data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigMapParse(t *testing.T) {
	app := eudore.NewApp()
	config := eudore.NewConfigMap(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{func(config eudore.Config) error {
		config.Set("parse", true)
		return nil
	}})
	app.Infof("ConfigMap parse eror: %v", config.Parse())
	app.Infof("ConfigMap data: %# v", config.Get(""))

	config.ParseOption([]eudore.ConfigParseFunc{func(config eudore.Config) error {
		config.Set("error", true)
		return errors.New("ConfigMap parse test error")
	}})
	app.Infof("ConfigMap parse eror: %v", config.Parse())
	app.Infof("ConfigMap data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigMapJSON(t *testing.T) {
	app := eudore.NewApp()
	config := eudore.NewConfigMap(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{func(config eudore.Config) error {
		return json.Unmarshal([]byte(`{"name":"eudore"}`), config)
	}})
	app.Infof("ConfigMap parse eror: %v", config.Parse())
	app.Infof("ConfigMap data: %# v", config.Get(""))

	body, err := json.Marshal(config)
	app.Infof("ConfigMap json data: %s,error: %v", body, err)

	app.CancelFunc()
	app.Run()
}

func TestConfigEudore2(t *testing.T) {
	app := eudore.NewApp()
	// 默认 map[string]interface{}
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.Set("name", "eudore")
	config.Set("type", "ConfigEudore")
	app.Infof("ConfigEudore data: %#v", config.Get(""))
	app.Infof("ConfigEudore name data: %#v", config.Get("name"))

	config.Set("", struct{ Name, Message string }{"eudore", "msg"})
	app.Infof("ConfigEudore data: %#v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigEudoreParse(t *testing.T) {
	app := eudore.NewApp()
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{func(config eudore.Config) error {
		config.Set("parse", true)
		return nil
	}})
	app.Infof("ConfigEudore parse eror: %v", config.Parse())
	app.Infof("ConfigEudore data: %# v", config.Get(""))

	config.ParseOption([]eudore.ConfigParseFunc{func(config eudore.Config) error {
		config.Set("error", true)
		return errors.New("ConfigEudore parse test error")
	}})
	app.Infof("ConfigEudore parse eror: %v", config.Parse())
	app.Infof("ConfigEudore data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigEudoreJSON(t *testing.T) {
	app := eudore.NewApp()
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{func(config eudore.Config) error {
		return json.Unmarshal([]byte(`{"name":"eudore"}`), config)
	}})
	app.Infof("ConfigEudore parse eror: %v", config.Parse())
	app.Infof("ConfigEudore data: %# v", config.Get(""))

	body, err := json.Marshal(config)
	app.Infof("ConfigEudore json data: %s,error: %v", body, err)

	app.CancelFunc()
	app.Run()
}

func TestConfigPrint(t *testing.T) {
	app := eudore.NewApp()

	config1 := eudore.NewConfigMap(nil)
	config1.Set("print", eudore.NewPrintFunc(app))
	config1.Set("print", "ConfigMap print log")

	config2 := eudore.NewConfigStd(nil)
	config2.Set("print", eudore.NewPrintFunc(app))
	config2.Set("print", "ConfigEudore print log")

	app.CancelFunc()
	app.Run()
}

func TestConfigParseJSON(t *testing.T) {
	filepath1 := "tmp-config1.json"
	defer tempConfigFile(filepath1, `{"help":true,"workdir":".","name":"eudore"}`)()
	filepath2 := "tmp-config2.json"
	defer tempConfigFile(filepath2, `name:eudore`)()

	app := eudore.NewApp()
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseJSON("config")})

	app.Infof("NewConfigParseJSON parse empty error %v:", config.Parse())

	config.Set("config", filepath1)
	app.Infof("NewConfigParseJSON parse file error: %v", config.Parse())

	config.Set("config", []string{filepath1})
	app.Infof("NewConfigParseJSON parse mutil file error: %v", config.Parse())

	config.Set("config", "not-"+filepath1)
	app.Infof("NewConfigParseJSON parse not file error: %v", config.Parse())

	config.Set("config", filepath2)
	app.Infof("NewConfigParseJSON parse error: %v", config.Parse())

	app.Infof("Config data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func tempConfigFile(path, content string) func() {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	file.Write([]byte(content))
	file.Close()
	return func() {
		os.Remove(file.Name())
	}
}

func TestConfigParseArgs(t *testing.T) {
	os.Args = append(os.Args, "--name=eudore")
	app := eudore.NewApp()
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseArgs(nil)})

	app.Infof("NewConfigParseArgs parse error: %v", config.Parse())
	app.Infof("Config data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigParseArgsShort(t *testing.T) {
	type configShort struct {
		Help   bool   `alias:"help" json:"help" flag:"h"`
		Config string `alias:"config" json:"config" flag:"c"`
		Name   string `alias:"name" json:"name"`
	}
	shortMapping := map[string][]string{
		"f": {"config"},
	}
	os.Args = append(os.Args, "--name=eudore", "-f=config.json", "-h", "--help")

	app := eudore.NewApp()
	config := eudore.NewConfigStd(&configShort{false, "eudore", "msg"})
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseArgs(shortMapping)})

	app.Infof("NewConfigParseArgs parse error: %v", config.Parse())
	app.Infof("Config data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigParseEnvs(t *testing.T) {
	os.Setenv("ENV_NAME", "eudore")
	defer os.Unsetenv("ENV_NAME")
	// init envs by cmd
	app := eudore.NewApp()
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseEnvs("ENV_")})

	app.Infof("NewConfigParseEnvs parse error: %v", config.Parse())
	app.Infof("Config data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigParseWorkdir(t *testing.T) {
	app := eudore.NewApp()
	config := eudore.NewConfigStd(nil)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseWorkdir("workdir")})

	app.Infof("NewConfigParseWorkdir parse empty dir error: %v", config.Parse())

	config.Set("workdir", ".")
	app.Infof("NewConfigParseWorkdir parse error: %v", config.Parse())
	app.Infof("Config data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

func TestConfigParseHelp(t *testing.T) {
	conf := &helpConfig{Iface: &helpDBConfig{}}
	conf.Link = conf

	app := eudore.NewApp()
	config := eudore.NewConfigStd(conf)
	app.SetValue(eudore.ContextKeyConfig, config)
	config.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseHelp("help")})

	app.Infof("NewConfigParseHelp parse not help error: %v", config.Parse())

	config.Set("help", true)
	app.Infof("NewConfigParseHelp parse error: %v", config.Parse())
	app.Infof("Config data: %# v", config.Get(""))

	app.CancelFunc()
	app.Run()
}

type helpConfig struct {
	sync.RWMutex
	Command   string                      `json:"command" alias:"command" description:"app start command, start/stop/status/restart" flag:"cmd"`
	Pidfile   string                      `json:"pidfile" alias:"pidfile" description:"pid file localtion"`
	Workdir   string                      `json:"workdir" alias:"workdir" description:"set app working directory"`
	Config    string                      `json:"config" alias:"config" description:"config path" flag:"f"`
	Help      bool                        `json:"help" alias:"help" description:"output help info" flag:"h"`
	Enable    []string                    `json:"enable" alias:"enable" description:"enable config mods"`
	Mods      map[string]*helpConfig      `json:"mods" alias:"mods" description:"config mods"`
	Listeners []eudore.ServerListenConfig `json:"listeners" alias:"listeners"`
	Component *helpComponentConfig        `json:"component" alias:"component"`
	Length    int                         `json:"length" alias:"length" description:"this is int"`
	Num       [3]int                      `json:"num" alias:"num" description:"this is array"`
	Body      []byte                      `json:"body" alias:"body" description:"this is []byte"`
	Float     float64                     `json:"body" alias:"body" description:"this is float"`
	Time      time.Time                   `json:"time" alias:"time" description:"this is time"`
	Map       map[string]interface{}      `json:"map" alias:"map" description:"this is map"`

	Auth  *helpAuthConfig `json:"auth" alias:"auth"`
	Iface interface{}
	Link  interface{} `json:"-" alias:"link"`
	// Node *Node
}

type Node struct {
	Next *Node
}

// ComponentConfig 定义website使用的组件的配置。
type helpComponentConfig struct {
	DB     helpDBConfig            `json:"db" alias:"db"`
	Logger *eudore.LoggerStdConfig `json:"logger" alias:"logger"`
	Server *eudore.ServerStdConfig `json:"server" alias:"server"`
	Notify map[string]string       `json:"notify" alias:"notify"`
	Pprof  *helpPprofConfig        `json:"pprof" alias:"pprof"`
	Black  map[string]bool         `json:"black" alias:"black"`
}
type helpDBConfig struct {
	Driver string `json:"driver" alias:"driver" description:"database driver type"`
	Config string `json:"config" alias:"config" description:"database config info" flag:"db"`
}
type helpPprofConfig struct {
	Godoc     string            `json:"godoc" alias:"godoc" description:"godoc server"`
	BasicAuth map[string]string `json:"basicauth" alias:"basicauth" description:"basic auth username and password"`
}

type helpAuthConfig struct {
	Secrets  map[string]string    `json:"secrets" alias:"secrets" description:"default auth secrets"`
	IconTemp string               `json:"icontemp" alias:"icontemp" description:"save icon temp dir"`
	Sender   helpMailSenderConfig `json:"sender" alias:"sender" description:""`
}
type helpMailSenderConfig struct {
	Username string `json:"username" alias:"username" description:"email send username"`
	Password string `json:"password" alias:"password" description:"email send password"`
	Addr     string `json:"addr" alias:"addr"`
	Subject  string `json:"subject" alias:"subject"`
}

func TestConfigMetadata(t *testing.T) {
	conf := eudore.NewConfigMap(nil)
	t.Log(conf.(interface{ Metadata() interface{} }).Metadata())
	conf = eudore.NewConfigStd(nil)
	t.Log(conf.(interface{ Metadata() interface{} }).Metadata())

	confh := &helpConfig{
		Map: map[string]interface{}{
			"1":    1,
			"chan": make(chan int, 4),
			"func": TestConfigMetadata,
		},
		Time:  time.Now(),
		Iface: &helpDBConfig{},
	}
	confh.Link = confh
	conf = eudore.NewConfigStd(confh)
	t.Log(conf.(interface{ Metadata() interface{} }).Metadata())
}
