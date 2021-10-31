package eudore

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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
	Switch working directory
	Generate help information based on the structure

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
	切换工作目录
	根据结构体生成帮助信息
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

// NewConfigMap creates a ConfigMap, if the input parameter is map[string]interface{}, it will be used as the initialization data.
//
// ConfigMap will use the passed map as configuration storage to Get/Set a key value.
//
// ConfigMap has implemented json.Marshaler and json.Unmarshaler interfaces.
//
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

// The Get method gets an attribute. If the key is an empty string, it returns the map object that holds all the data.
//
// Get 方法获取一个属性，如果键为空字符串，返回保存全部数据的map对象。
func (c *configMap) Get(key string) interface{} {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	if len(key) == 0 {
		return c.Keys
	}
	return c.Keys[key]
}

// The Set method sets an attribute. If the key is an empty string and the value type is map[string]interface{},
// replace the map object that holds all the data.
//
// Set 方法设置一个属性，如果键为空字符串且值类型是map[string]interface{},则替换保存全部数据的map对象。
func (c *configMap) Set(key string, val interface{}) error {
	c.Locker.Lock()
	defer c.Locker.Unlock()
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
	return nil
}

// ParseOption executes a configuration parsing function option.
//
// ParseOption 执行一个配置解析函数选项。
func (c *configMap) ParseOption(fn []ConfigParseFunc) []ConfigParseFunc {
	c.funcs, fn = fn, c.funcs
	return fn
}

// The Parse method executes all configuration parsing functions.
// If the parsing function returns error, it stops parsing and returns error.
//
// Parse 方法执行全部配置解析函数，如果其中解析函数返回error，则停止解析并返回error。
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

// MarshalJSON implements the json.Marshaler interface, which enables json serialization to directly manipulate the saved data.
//
// MarshalJSON 实现json.Marshaler接口，使json序列化直接操作保存的数据。
func (c *configMap) MarshalJSON() ([]byte, error) {
	c.Locker.RLock()
	defer c.Locker.RUnlock()
	return json.Marshal(c.Keys)
}

// UnmarshalJSON implements the json.Unmarshaler interface, which enables json deserialization to directly manipulate the saved data.
//
// UnmarshalJSON 实现json.Unmarshaler接口，使json反序列化直接操作保存的数据。
func (c *configMap) UnmarshalJSON(data []byte) error {
	c.Locker.Lock()
	defer c.Locker.Unlock()
	return json.Unmarshal(data, &c.Keys)
}

// NewConfigEudore creates a ConfigEudore. If the input parameter is empty, use an empty map[string]interface{} as the initialization data.
//
// ConfigEduoew allows to pass in a map or struct as configuration storage, and use eudore.Set and eudore.Get methods to read and write data.
//
// If the incoming configuration object implements the same read-write lock method as sync.RLock,
// the configured read-write lock is used, otherwise a sync.RWMutex lock will be created.
//
// ConfigEduoe has implemented the json.Marshaler and json.Unmarshaler interfaces.
//
// NewConfigEudore 创建一个ConfigEudore，如果传入参数为空，使用空map[string]interface{}作为初始化数据。
//
// ConfigEduoew允许传入一个map或struct作为配置存储，使用eudore.Set和eudore.Get方法去读写数据。
//
// 如果传入的配置对象实现sync.RLock一样的读写锁方法，则使用配置的读写锁，否则会创建一个sync.RWMutex锁。
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

// The Get method realizes to read the data attributes, and uses the RLock method to lock the data.
//
// Get 方法实现读取数据属性，并使用RLock方法锁定数据。
func (c *configEudore) Get(key string) interface{} {
	if len(key) == 0 {
		return c.Keys
	}
	c.RLock()
	defer c.RUnlock()
	return Get(c.Keys, key)
}

// The Set method implements setting data, and uses the Lock method to lock the data.
//
// Set 方法实现设置数据，并使用Lock方法锁定数据。
func (c *configEudore) Set(key string, val interface{}) error {
	c.Lock()
	defer c.Unlock()
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
		return Set(c.Keys, key, val)
	}
	return nil
}

// ParseOption executes a configuration parsing function option.
//
// ParseOption 执行一个配置解析函数选项。
func (c *configEudore) ParseOption(fn []ConfigParseFunc) []ConfigParseFunc {
	c.funcs, fn = fn, c.funcs
	return fn
}

// The Parse method executes all configuration parsing functions.
// If the parsing function returns error, it stops parsing and returns error.
//
// Parse 方法执行全部配置解析函数，如果其中解析函数返回error，则停止解析并返回error。
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

// MarshalJSON implements the json.Marshaler interface, which enables json serialization to directly manipulate the saved data.
//
// MarshalJSON 实现json.Marshaler接口，使json序列化直接操作保存的数据。
func (c *configEudore) MarshalJSON() ([]byte, error) {
	c.RLock()
	defer c.RUnlock()
	return json.Marshal(c.Keys)
}

// UnmarshalJSON implements the json.Unmarshaler interface, which enables json deserialization to directly manipulate the saved data.
//
// UnmarshalJSON 实现json.Unmarshaler接口，使json反序列化直接操作保存的数据。
func (c *configEudore) UnmarshalJSON(data []byte) error {
	c.Lock()
	defer c.Unlock()
	return json.Unmarshal(data, &c.Keys)
}

func configPrint(c Config, args ...interface{}) {
	c.Set("print", fmt.Sprint(args...))
}

// NewConfigParseJSON method parses the json file configuration, usually the key is "config".
//
// The configuration item value is string(';' divided into multiple paths) or []string, if the loaded file does not exist, the file will be ignored.
//
// NewConfigParseJSON 方法解析json文件配置，通常使用key为"config"。
//
// 配置项值为string(';'分割为多路径)或[]string，如果加载文件不存在将忽略文件。
func NewConfigParseJSON(key string) ConfigParseFunc {
	return func(c Config) error {
		var paths []string
		switch val := c.Get(key).(type) {
		case string:
			paths = strings.Split(val, ";")
		case []string:
			paths = val
		default:
			return nil
		}
		configPrint(c, "config read json file by key: ", key)
		for _, path := range paths {
			path = strings.TrimSpace(path)
			file, err := os.Open(path)
			if err != nil {
				configPrint(c, "config ignored file: ", err)
				continue
			}
			defer file.Close()
			err = json.NewDecoder(file).Decode(c)
			if err != nil {
				err = fmt.Errorf("config parse json file '%s' error: %v", path, err)
				configPrint(c, err)
				return err
			}
			configPrint(c, "config load json file: ", path)
		}
		return nil
	}
}

// NewConfigParseArgs function uses the eudore.Set method to set the command line parameter data,
// and the command line parameter uses the format of'--{key}.{sub}={value}'.
//
// Shortsmap is mapped as a short parameter. If the structure has a'flag' tag, it will be used as the abbreviation of the path.
// The tag length must be less than 5, the command line format is'-{short}={value}, and the short parameter will automatically be long parameter.
//
// NewConfigParseArgs 函数使用eudore.Set方法设置命令行参数数据，命令行参数使用'--{key}.{sub}={value}'格式。
//
// shortsmap作为短参数映射，如果结构体存在'flag' tag将作为该路径的缩写，tag长度需要小于5，命令行格式为'-{short}={value},短参数将会自动为长参数。
func NewConfigParseArgs(shortsmap map[string][]string) ConfigParseFunc {
	return func(c Config) error {
		// 使用结构体tag初始化shorts
		shorts := make(map[string][]string)
		flag := &eachTags{tag: "flag", Repeat: make(map[uintptr]string)}
		flag.Each("", reflect.ValueOf(c.Get("")))
		for i, tag := range flag.Tags {
			shorts[flag.Vals[i]] = append(shorts[flag.Vals[i]], tag[1:])
		}
		for k, v := range shortsmap {
			shorts[k] = append(shorts[k], v...)
		}

		for _, str := range os.Args[1:] {
			key, val := split2byte(str, '=')
			if strings.HasPrefix(key, "--") { // 长参数
				if val == "" && reflect.ValueOf(c.Get(key[2:])).Kind() == reflect.Bool {
					val = "true"
				}
				configPrint(c, "config set arg: ", str)
				c.Set(key[2:], val)
			} else if len(key) > 1 && key[0] == '-' && key[1] != '-' { // 短参数
				for _, lkey := range shorts[key[1:]] {
					val := val
					if val == "" && reflect.ValueOf(c.Get(lkey)).Kind() == reflect.Bool {
						val = "true"
					}
					configPrint(c, fmt.Sprintf("config set short arg '%s': --%s=%s", key[1:], lkey, val))
					c.Set(lkey, val)
				}
			}
		}
		return nil
	}
}

// NewConfigParseEnvs function uses the eudore.Set method to set the environment variable data, usually the environment variable prefix uses'ENV_'.
//
// Environment variables will be converted to lowercase paths, and the underscore of'_' is equivalent to the function of'.'.
//
// NewConfigParseEnvs 函数使用eudore.Set方法设置环境变量数据，通常环境变量前缀使用'ENV_'。
//
// 环境变量将转换成小写路径，'_'下划线相当于'.'的作用
//
// exmapel: 'ENV_EUDORE_NAME=eudore' => 'eudore.name=eudore'。
func NewConfigParseEnvs(key string) ConfigParseFunc {
	return func(c Config) error {
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
}

// NewConfigParseWorkdir function initializes the workspace, usually using the key as string("workdir") to obtain the workspace directory and switch.
//
// NewConfigParseWorkdir 函数初始化工作空间，通常使用key为string("workdir"),获取工作空间目录并切换。
func NewConfigParseWorkdir(key string) ConfigParseFunc {
	return func(c Config) error {
		dir, ok := c.Get(key).(string)
		if ok && dir != "" {
			configPrint(c, "changes working directory to: "+dir)
			return os.Chdir(dir)
		}
		return nil
	}
}

// NewConfigParseHelp function if uses the structure configuration to output the'flag' and'description' tags to produce the default parameter description.
//
// By default, only the parameter description is output. For other descriptions, please wrap the NewConfigParseHelp method.
//
// Note that the properties of the configuration structure need to be non-empty, otherwise it will not enter the traversal.
//
// NewConfigParseHelp 函数如果使用结构体配置输出'flag'和'description' tag生产默认参数描述。
//
// 默认仅输出参数描述，其他描述内容请包装NewConfigParseHelp方法。
//
// 注意配置结构体的属性需要是非空，否则不会进入遍历。
func NewConfigParseHelp(key string) ConfigParseFunc {
	return func(c Config) error {
		help, ok := c.Get(key).(bool)
		if !ok || !help {
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
