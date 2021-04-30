package eudore

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// ConfigParseFunc 定义配置解析函数。
//
// Config 默认解析函数为eudore.ConfigAllParseFunc
type ConfigParseFunc func(Config) error

/*
Config defines configuration management and uses configuration read-write and analysis functions.

Get/Set read and write data implementation:
	Use custom map or struct as data storage
	Support Lock concurrency safety
	Access attributes based on string path hierarchy

The default analysis function implementation:
	Custom configuration analysis function
	Parse multiple json files
	Parse the length and short parameters of the command line
	Parse Env environment variables
	Configuration differentiation
	Generate help information based on the structure
	Switch working directory

Config 定义配置管理，使用配置读写和解析功能。

Get/Set读写数据实现下列功能:
	使用自定义map或struct作为数据存储
	支持Lock并发安全
	基于字符串路径层次访问属性

默认解析函数实现下列功能:
	自定义配置解析函数
	解析多json文件
	解析命令行长短参数
	解析Env环境变量
	配置差异化
	根据结构体生成帮助信息
	切换工作目录
*/
type Config interface {
	Get(string) interface{}
	Set(string, interface{}) error
	ParseOption([]ConfigParseFunc) []ConfigParseFunc
	Parse() error
}

// configMap 使用map保存配置。
type configMap struct {
	Keys   map[string]interface{} `alias:"keys"`
	Print  func(...interface{})   `alias:"print"`
	funcs  []ConfigParseFunc      `alias:"-"`
	Locker sync.RWMutex           `alias:"-"`
}

// configEudore 使用结构体或map保存配置，通过属性或反射来读写属性。
type configEudore struct {
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
//
// ConfigMap将使用传入的map作为配置存储去Get/Set一个键值。
//
// ConfigMap已实现json.Marshaler和json.Unmarshaler接口.
func NewConfigMap(arg interface{}) Config {
	var keys map[string]interface{}
	if ks, ok := arg.(map[string]interface{}); ok {
		keys = ks
	} else {
		keys = make(map[string]interface{})
	}
	return &configMap{
		Keys:  keys,
		Print: printEmpty,
		funcs: ConfigAllParseFunc,
	}
}

// Get 方法获取一个属性，如果键为空字符串，返回保存全部数据的map对象。
func (c *configMap) Get(key string) interface{} {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	if len(key) == 0 {
		return c.Keys
	}
	return c.Keys[key]
}

// Set 方法设置一个属性，如果键为空字符串且值类型是map[string]interface{},则替换保存全部数据的map对象。
func (c *configMap) Set(key string, val interface{}) error {
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
func (c *configMap) ParseOption(fn []ConfigParseFunc) []ConfigParseFunc {
	c.funcs, fn = fn, c.funcs
	return fn
}

// Parse 方法执行全部配置解析函数，如果其中解析函数返回err，则停止解析并返回err。
func (c *configMap) Parse() (err error) {
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
func (c *configMap) MarshalJSON() ([]byte, error) {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	return json.Marshal(c.Keys)
}

// UnmarshalJSON 实现json.Unmarshaler接口，试json反序列化直接操作保存的数据。
func (c *configMap) UnmarshalJSON(data []byte) error {
	c.Locker.Lock()
	defer c.Locker.Unlock()
	return json.Unmarshal(data, &c.Keys)
}

// NewConfigEudore 创建一个ConfigEudore，如果传入参数为空，使用空map[string]interface{}作为初始化数据。
//
// ConfigEduoew允许传入一个map或struct作为配置存储，使用eudore.Set和eudore.Get方法去读写数据。
//
// 如果传入的配置对象实现sync.RLock一样的读写锁，则使用配置的读写锁，否则会创建一个sync.RWMutex锁。
//
// ConfigEduoe已实现json.Marshaler和json.Unmarshaler接口.
func NewConfigEudore(i interface{}) Config {
	if i == nil {
		i = make(map[string]interface{})
	}
	mu, ok := i.(configRLocker)
	if !ok {
		mu = new(sync.RWMutex)
	}
	return &configEudore{
		Keys:          i,
		Print:         printEmpty,
		funcs:         ConfigAllParseFunc,
		configRLocker: mu,
	}
}

// Get 方法实现读取数据属性的一个属性。
func (c *configEudore) Get(key string) (i interface{}) {
	if len(key) == 0 {
		return c.Keys
	}
	c.RLock()
	i = Get(c.Keys, key)
	c.RUnlock()
	return
}

// Set 方法实现设置数据的一个属性。
func (c *configEudore) Set(key string, val interface{}) (err error) {
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
func (c *configEudore) ParseOption(fn []ConfigParseFunc) []ConfigParseFunc {
	c.funcs, fn = fn, c.funcs
	return fn
}

// Parse 方法执行全部配置解析函数，如果其中解析函数返回err，则停止解析并返回err。
func (c *configEudore) Parse() (err error) {
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
func (c *configEudore) MarshalJSON() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	return json.Marshal(c.Keys)
}

// UnmarshalJSON 实现json.Unmarshaler接口，试json反序列化直接操作保存的数据。
func (c *configEudore) UnmarshalJSON(data []byte) error {
	c.Lock()
	defer c.Unlock()
	return json.Unmarshal(data, &c.Keys)
}

func configPrint(c Config, args ...interface{}) {
	c.Set("print", fmt.Sprint(args...))
}

// ConfigParseJSON 方法解析json文件配置。
func ConfigParseJSON(c Config) error {
	configPrint(c, "config read paths: ", c.Get("config"))
	for _, path := range GetStrings(c.Get("config")) {
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

// ConfigParseArgs 函数使用参数设置配置，参数使用'--'为前缀。
//
// 如果结构体存在flag tag将作为该路径的缩写，tag长度小于5使用'-'为前缀。
func ConfigParseArgs(c Config) (err error) {
	flag := &eachTags{tag: "flag", Repeat: make(map[uintptr]string)}
	flag.Each("", reflect.ValueOf(c.Get("")))
	short := make(map[string][]string)
	for i, tag := range flag.Tags {
		short[flag.Vals[i]] = append(short[flag.Vals[i]], tag[1:])
	}

	for _, str := range os.Args[1:] {
		key, val := split2byte(str, '=')
		if len(key) > 1 && key[0] == '-' && key[1] != '-' {
			for _, lkey := range short[key[1:]] {
				val := val
				if val == "" && reflect.ValueOf(c.Get(lkey)).Kind() == reflect.Bool {
					val = "true"
				}
				configPrint(c, fmt.Sprintf("config set short arg %s: --%s=%s", key[1:], lkey, val))
				c.Set(lkey, val)
			}
		} else if strings.HasPrefix(key, "--") {
			if val == "" && reflect.ValueOf(c.Get(key[2:])).Kind() == reflect.Bool {
				val = "true"
			}
			configPrint(c, "config set arg: ", str)
			c.Set(key[2:], val)
		}
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
	mod := GetStrings(c.Get("enable"))
	mod = append([]string{getOS()}, mod...)
	for _, i := range mod {
		m := c.Get("mods." + i)
		if m != nil {
			configPrint(c, "config load mod "+i)
			ConvertTo(m, c.Get(""))
		}
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

// ConfigParseHelp 函数测试配置内容，如果存在help项会层叠获取到结构体的的description tag值作为帮助信息输出。
//
// 注意配置结构体的属性需要是非空，否则不会进入遍历。
func ConfigParseHelp(c Config) error {
	if !GetBool(c.Get("help")) {
		return nil
	}

	conf := reflect.ValueOf(c.Get(""))
	flag := &eachTags{tag: "flag", Repeat: make(map[uintptr]string)}
	flag.Each("", conf)
	flagmap := make(map[string]string)
	for i, tag := range flag.Tags {
		flagmap[tag[1:]] = flag.Vals[i]
	}

	desc := &eachTags{tag: "description", Repeat: make(map[uintptr]string)}
	desc.Each("", conf)
	var length int
	for i, tag := range desc.Tags {
		desc.Tags[i] = tag[1:]
		if len(tag) > length {
			length = len(tag)
		}
	}

	for i, tag := range desc.Tags {
		f, ok := flagmap[tag]
		if ok && !strings.Contains(tag, "{") && len(f) < 5 {
			fmt.Printf("  -%s,", f)
		}
		fmt.Printf("\t --%s=%s\t%s\n", tag, strings.Repeat(" ", length-len(tag)), desc.Vals[i])
	}
	return nil
}

type eachTags struct {
	tag     string
	Tags    []string
	Vals    []string
	Repeat  map[uintptr]string
	LastTag string
}

func (each *eachTags) Each(prefix string, iValue reflect.Value) {
	switch iValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		if !iValue.IsNil() {
			_, ok := each.Repeat[iValue.Pointer()]
			if ok {
				return
			}
			each.Repeat[iValue.Pointer()] = prefix
		}
	}

	switch iValue.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !iValue.IsNil() {
			each.Each(prefix, iValue.Elem())
		}
	case reflect.Map:
		if each.LastTag != "" {
			each.Tags = append(each.Tags, fmt.Sprintf("%s.{%s}", prefix, iValue.Type().Key().Name()))
			each.Vals = append(each.Vals, each.LastTag)
		}
	case reflect.Slice, reflect.Array:
		length := "n"
		if iValue.Kind() == reflect.Array {
			length = fmt.Sprint(iValue.Type().Len() - 1)
		}
		last := each.LastTag
		if last != "" {
			each.Tags = append(each.Tags, fmt.Sprintf("%s.{0-%s}", prefix, length))
			each.Vals = append(each.Vals, last)
		}
		each.LastTag = last
		each.Each(fmt.Sprintf("%s.{0-%s}", prefix, length), reflect.New(iValue.Type().Elem()))
	case reflect.Struct:
		each.EachStruct(prefix, iValue)
	}
}

func (each *eachTags) EachStruct(prefix string, iValue reflect.Value) {
	iType := iValue.Type()
	for i := 0; i < iType.NumField(); i++ {
		if iValue.Field(i).CanSet() {
			val := iType.Field(i).Tag.Get(each.tag)
			name := iType.Field(i).Tag.Get("alias")
			if name == "" {
				name = iType.Field(i).Name
			}
			if val != "" && each.getValueKind(iType.Field(i).Type) != "" {
				each.Tags = append(each.Tags, prefix+"."+name)
				each.Vals = append(each.Vals, val)
			}
			each.LastTag = val
			each.Each(prefix+"."+name, iValue.Field(i))
		}
	}
}

func (each *eachTags) getValueKind(iType reflect.Type) string {
	switch iType.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	default:
		if iType.Kind() == reflect.Slice && iType.Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		return ""
	}
}
