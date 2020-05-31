package eudore

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

// ConfigParseFunc 定义配置解析函数。
type ConfigParseFunc func(Config) error

// ConfigParseOption 定义配置解析选项，用于修改配置解析函数。
type ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc

// Config 定义配置管理，使用配置读写和解析功能。
type Config interface {
	Get(string) interface{}
	Set(string, interface{}) error
	ParseOption(ConfigParseOption)
	Parse() error
}

// ConfigMap 使用map保存配置。
type ConfigMap struct {
	Keys   map[string]interface{} `alias:"keys"`
	Print  func(...interface{})   `alias:"print"`
	funcs  []ConfigParseFunc      `alias:"-"`
	Locker sync.RWMutex           `alias:"-"`
}

// ConfigEudore 使用结构体或map保存配置，通过反射来读写属性。
type ConfigEudore struct {
	Keys          interface{}          `alias:"keys"`
	Print         func(...interface{}) `alias:"print"`
	funcs         []ConfigParseFunc    `alias:"-"`
	configRLocker `alias:"-"`
}

type configRLocker interface {
	sync.Locker
	RLock()
	RUnlock()
}

// NewConfigMap 创建一个ConfigMap，如果传入参数为map[string]interface{},则作为初始化数据。
func NewConfigMap(arg interface{}) Config {
	var keys map[string]interface{}
	if ks, ok := arg.(map[string]interface{}); ok {
		keys = ks
	} else {
		keys = make(map[string]interface{})
	}
	return &ConfigMap{
		Keys:  keys,
		Print: printEmpty,
		funcs: ConfigAllParseFunc,
	}
}

// Get 方法获取一个属性，如果键为空字符串，返回保存全部数据的map对象。
func (c *ConfigMap) Get(key string) interface{} {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	if len(key) == 0 {
		return c.Keys
	}
	return c.Keys[key]
}

// Set 方法设置一个属性，如果键为空字符串且值类型是map[string]interface{},则替换保存全部数据的map对象。
func (c *ConfigMap) Set(key string, val interface{}) error {
	c.Locker.Lock()
	if len(key) == 0 {
		keys, ok := val.(map[string]interface{})
		if ok {
			c.Keys = keys
		}
	} else if key == "print" {
		fn, ok := val.(func(...interface{}))
		if ok {
			c.Print = fn
		} else {
			c.Print(val)
		}
	} else {
		c.Keys[key] = val
	}
	c.Locker.Unlock()
	return nil
}

// ParseOption 执行一个配置解析函数选项。
func (c *ConfigMap) ParseOption(fn ConfigParseOption) {
	c.funcs = fn(c.funcs)
}

// Parse 方法执行全部配置解析函数，如果其中解析函数返回err，则停止解析并返回err。
func (c *ConfigMap) Parse() (err error) {
	for _, fn := range c.funcs {
		err = fn(c)
		if err != nil {
			c.Print(err)
			return
		}
	}
	return nil
}

// MarshalJSON 实现json.Marshaler接口，试json序列化直接操作保存的数据。
func (c *ConfigMap) MarshalJSON() ([]byte, error) {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	return json.Marshal(c.Keys)
}

// UnmarshalJSON 实现json.Unmarshaler接口，试json反序列化直接操作保存的数据。
func (c *ConfigMap) UnmarshalJSON(data []byte) error {
	c.Locker.Lock()
	defer c.Locker.Unlock()
	return json.Unmarshal(data, &c.Keys)
}

// NewConfigEudore 创建一个ConfigEudore，如果传入参数为空，使用空map[string]interface{}作为初始化数据。
func NewConfigEudore(i interface{}) Config {
	if i == nil {
		i = make(map[string]interface{})
	}
	mu, ok := i.(configRLocker)
	if !ok {
		mu = new(sync.RWMutex)
	}
	return &ConfigEudore{
		Keys:          i,
		Print:         printEmpty,
		funcs:         ConfigAllParseFunc,
		configRLocker: mu,
	}
}

// Get 方法实现读取数据属性的一个属性。
func (c *ConfigEudore) Get(key string) (i interface{}) {
	if len(key) == 0 {
		return c.Keys
	}
	c.RLock()
	i = Get(c.Keys, key)
	c.RUnlock()
	return
}

// Set 方法实现设置数据的一个属性。
func (c *ConfigEudore) Set(key string, val interface{}) (err error) {
	c.Lock()
	if len(key) == 0 {
		c.Keys = val
	} else if key == "print" {
		fn, ok := val.(func(...interface{}))
		if ok {
			c.Print = fn
		} else {
			c.Print(val)
		}
	} else {
		err = Set(c.Keys, key, val)
	}
	c.Unlock()
	return
}

// ParseOption 执行一个配置解析函数选项。
func (c *ConfigEudore) ParseOption(fn ConfigParseOption) {
	c.funcs = fn(c.funcs)
}

// Parse 方法执行全部配置解析函数，如果其中解析函数返回err，则停止解析并返回err。
func (c *ConfigEudore) Parse() (err error) {
	for _, fn := range c.funcs {
		err = fn(c)
		if err != nil {
			c.Print(err)
			return
		}
	}
	return nil
}

// MarshalJSON 实现json.Marshaler接口，试json序列化直接操作保存的数据。
func (c *ConfigEudore) MarshalJSON() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	return json.Marshal(c.Keys)
}

// UnmarshalJSON 实现json.Unmarshaler接口，试json反序列化直接操作保存的数据。
func (c *ConfigEudore) UnmarshalJSON(data []byte) error {
	c.Lock()
	defer c.Unlock()
	return json.Unmarshal(data, &c.Keys)
}

func configPrint(c Config, args ...interface{}) {
	c.Set("print", fmt.Sprint(args...))
}

// ConfigParseJSON 方法解析json文件配置。
func ConfigParseJSON(c Config) error {
	configPrint(c, "config read paths: ", c.Get("keys.config"))
	for _, path := range GetArrayString(c.Get("keys.config")) {
		file, err := os.Open(path)
		if err == nil {
			err = json.NewDecoder(file).Decode(c)
			file.Close()
		}
		if err == nil {
			configPrint(c, "config load path: ", path)
			return nil
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("config load %s error: %s", path, err.Error())
		}
	}
	return nil
}

// ConfigParseArgs 函数使用参数设置配置，参数使用--为前缀。
func ConfigParseArgs(c Config) (err error) {
	for _, str := range os.Args[1:] {
		if !strings.HasPrefix(str, "--") {
			continue
		}
		configPrint(c, "config set arg: ", str)
		c.Set(split2byte(str[2:], '='))
	}
	return
}

// ConfigParseEnvs 函数使用环境变量设置配置，环境变量使用'ENV_'为前缀,'_'下划线相当于'.'的作用。
func ConfigParseEnvs(c Config) error {
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "ENV_") {
			configPrint(c, "config set env: ", value)
			k, v := split2byte(value, '=')
			k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
			c.Set(k, v)
		}
	}
	return nil
}

// ConfigParseMods 函数从'enable'项获得使用的模式的数组字符串，从'mods.xxx'加载配置。
//
// 默认会加载OS mod,如果是docker环境下使用docker模式。
func ConfigParseMods(c Config) error {
	mod := GetArrayString(c.Get("enable"))
	mod = append([]string{getOS()}, mod...)
	configPrint(c, "config load mods: ", mod)
	for _, i := range mod {
		ConvertTo(c.Get("mods."+i), c.Get(""))
	}
	return nil
}

func getOS() string {
	// check docker
	_, err := os.Stat("/.dockerenv")
	if err == nil || !os.IsNotExist(err) {
		return "docker"
	}
	// 返回默认OS
	return runtime.GOOS
}

// ConfigParseWorkdir 函数初始化工作空间，从config获取workdir的值为工作空间，然后切换目录。
func ConfigParseWorkdir(c Config) error {
	dir := GetString(c.Get("workdir"))
	if dir != "" {
		configPrint(c, "changes working directory to: "+dir)
		return os.Chdir(dir)
	}
	return nil
}

// ConfigParseHelp 函数测试配置内容，如果存在'keys.help'项会使用JSON标准化输出配置到标准输出。
func ConfigParseHelp(c Config) error {
	ok := c.Get("keys.help") != nil
	if ok {
		indent, err := json.MarshalIndent(&c, "", "\t")
		fmt.Println(string(indent), err)
	}
	return nil
}
