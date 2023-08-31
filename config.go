package eudore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

/*
Config defines configuration management and uses configuration read-write and analysis functions.

Get/Set read and write data implementation:

	Use custom struct or map as data storage
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

	使用自定义struct或map作为数据存储
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
	Get(string) any
	Set(string, any) error
	ParseOption(...ConfigParseFunc)
	Parse(context.Context) error
}

// ConfigParseFunc 定义配置解析函数。
//
// Config 默认解析函数为eudore.ConfigAllParseFunc。
type ConfigParseFunc func(context.Context, Config) error

// configStd 使用结构体或map保存配置，通过属性或反射来读写属性。
type configStd struct {
	Data  any               `alias:"data" json:"data" xml:"data" yaml:"data" description:"any data"`
	Map   map[string]any    `alias:"map" json:"map" xml:"map" yaml:"map" description:"map data"`
	Funcs []ConfigParseFunc `alias:"funcs" json:"funcs" xml:"funcs" yaml:"funcs" description:"all parse funcs"`
	Err   error             `alias:"err" json:"err" xml:"err" yaml:"err" description:"parsing error"`
	Lock  rwLocker          `alias:"lock" json:"-" xml:"-" yaml:"-"`
}

type rwLocker interface {
	sync.Locker
	RLock()
	RUnlock()
}

type MetadataConfig struct {
	Health bool   `alias:"health" json:"health" xml:"health" yaml:"health"`
	Name   string `alias:"name" json:"name" xml:"name" yaml:"name"`
	Error  error  `alias:"error,omitempty" json:"error,omitempty" xml:"error,omitempty" yaml:"error,omitempty"`
}

// NewConfig function creates a configStd.
// If the incoming parameter is empty, use map[string]any as metadata.
//
// If the metadata type is map[string]any, use map to read and write key values,
// otherwise use eudore.Set and eudore.Get methods to read and write metadata.
//
// If the incoming configuration object implements the same read-write lock method as sync.RLock,
// use the configured read-write lock, otherwise a sync.RWMutex lock will be created.
//
// configStd has implemented the json.Marshaler and json.Unmarshaler interfaces.
//
// NewConfig 函数创建一个configStd，如果传入参数为空，使用map[string]any作为元数据。
//
// 如果元数据类型为map[string]any使用map读写键值，否则其他类型使用eudore.Set和eudore.Get方法去读元写数据。
//
// 如果传入的配置对象实现sync.RLock一样的读写锁方法，则使用配置的读写锁，否则会创建一个sync.RWMutex锁。
//
// configStd已实现json.Marshaler和json.Unmarshaler接口.
func NewConfig(data any) Config {
	if data == nil {
		data = make(map[string]any)
	}
	mu, ok := data.(rwLocker)
	if !ok {
		mu = &sync.RWMutex{}
	}
	m, _ := data.(map[string]any)
	return &configStd{
		Data:  data,
		Map:   m,
		Funcs: DefaultConfigAllParseFunc,
		Lock:  mu,
	}
}

func (conf *configStd) Metadata() any {
	return MetadataConfig{
		Health: conf.Err == nil,
		Name:   "eudore.configStd",
		Error:  conf.Err,
	}
}

// The Get method realizes to read the data attributes, and uses the RLock method to lock the data,
// if key is empty string return metadata.
//
// Get 方法实现读取数据属性，并使用RLock方法锁定数据，如果key为空字符串返回元数据。
func (conf *configStd) Get(key string) any {
	if len(key) == 0 {
		return conf.Data
	}
	conf.Lock.RLock()
	defer conf.Lock.RUnlock()
	if conf.Map != nil {
		return conf.Map[key]
	}
	return GetAnyByPath(conf.Data, key)
}

// The Set method implements setting data, and uses the Lock method to lock the data,
// If key is empty string set metadata.
//
// Set 方法实现设置数据，并使用Lock方法锁定数据，如果key为空字符串设置元数据。
func (conf *configStd) Set(key string, val any) error {
	conf.Lock.Lock()
	defer conf.Lock.Unlock()
	if len(key) == 0 {
		conf.Data = val
		conf.Map, _ = val.(map[string]any)
		return nil
	}
	if conf.Map != nil {
		conf.Map[key] = val
		return nil
	}
	return SetAnyByPath(conf.Data, key, val)
}

// ParseOption executes a configuration parsing function option.
//
// ParseOption 执行一个配置解析函数选项。
func (conf *configStd) ParseOption(fn ...ConfigParseFunc) {
	if fn == nil {
		conf.Funcs = nil
		conf.Err = nil
	} else {
		conf.Funcs = append(conf.Funcs, fn...)
	}
}

// The Parse method executes all configuration parsing functions.
// If the parsing function returns error, it stops parsing and returns error.
//
// Parse 方法执行全部配置解析函数，如果其中解析函数返回error，则停止解析并返回error。
func (conf *configStd) Parse(ctx context.Context) error {
	if conf.Err != nil {
		return conf.Err
	}
	log := NewLoggerWithContext(ctx)
	for _, fn := range conf.Funcs {
		conf.Err = fn(ctx, conf)
		if conf.Err != nil {
			if !errors.Is(conf.Err, context.Canceled) {
				name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
				log.Errorf("config parse func %v error: %v", name, conf.Err)
			}
			return conf.Err
		}
	}
	log.Info("config parse done")
	return nil
}

// MarshalJSON implements the json.Marshaler interface,
// which enables json serialization to directly manipulate the saved data.
//
// MarshalJSON 实现json.Marshaler接口，使json序列化直接操作保存的数据。
func (conf *configStd) MarshalJSON() ([]byte, error) {
	conf.Lock.RLock()
	defer conf.Lock.RUnlock()
	return json.Marshal(conf.Data)
}

// UnmarshalJSON implements the json.Unmarshaler interface,
// which enables json deserialization to directly manipulate the saved data.
//
// UnmarshalJSON 实现json.Unmarshaler接口，使json反序列化直接操作保存的数据。
func (conf *configStd) UnmarshalJSON(data []byte) error {
	conf.Lock.Lock()
	defer conf.Lock.Unlock()
	return json.Unmarshal(data, &conf.Data)
}

/*
The NewConfigParseEnvFile function creates an Env file configuration parsing method.

If a line of the Env file is in env format from the beginning,
it will be loaded to the os as Env.

If the first character of the Env value is "'" as a multi-line value,
until the end of a line also has "'";
Newline characters "\r\n" "\n" in multi-line values are replaced with "\n" and TrimSpace is performed.

If the Env value is an empty string, the Env will be deleted from os.

NewConfigParseEnvFile 函数创建Env文件配置解析方法。

如果Env文件一行从开始为env格式，则作为Env加载到os。

如果Env值为第一个字符为"'"作为多行值，直到一行结尾同样具有"'"；
多行值中的换行符"\r\n" "\n"被替换为"\n"，并执行TrimSpace。

如果Env值为空字符串会从os删除这个Env。

example:

	EUDORE_NAME=eudore
	EUDORE_DEBUG=
	EUDORE_KEY='
	-----BEGIN RSA PRIVATE KEY-----
	-----END RSA PRIVATE KEY-----
	'
*/
func NewConfigParseEnvFile(files ...string) ConfigParseFunc {
	if files == nil {
		files = strings.Split(DefaultConfigEnvFiles, ";")
	}
	reg := regexp.MustCompile(`[a-zA-Z]\w*`)
	return func(ctx context.Context, c Config) error {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			log := NewLoggerWithContext(ctx)
			log.Info("confif load env file", file)
			lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
			keys := make([]string, 0, len(lines))
			char := "'"
			for i := range lines {
				key, val, ok := strings.Cut(lines[i], "=")
				if ok && reg.MatchString(key) {
					if strings.HasPrefix(val, char) {
						if strings.HasSuffix(val, char) {
							val = strings.TrimSpace(strings.TrimSuffix(val[1:], char))
						} else if i+1 < len(lines) {
							lines[i+1] = lines[i] + "\n" + lines[i+1]
							continue
						}
					}

					keys = append(keys, key)
					log.Infof("set file environment: %s=%s", key, val)
					if val != "" {
						os.Setenv(key, val)
					} else {
						os.Unsetenv(key)
					}
				}
			}
			os.Setenv("EUDORE_CONFIG_LOAD_ENVS", strings.Join(keys, ","))
		}
		return nil
	}
}

/*
The NewConfigParseDefault function creates a default variable parsing function
that gets the value from ENV to set the default variable.

NewConfigParseDefault 函数创建一个默认变量解析函数，从ENV获取值设置默认变量。

env to keys:

	EUDORE_CONTEXT_MAX_HANDLER                => DefaultContextMaxHandler
	EUDORE_CONTEXT_MAX_APPLICATION_FORM_SIZE  => DefaultContextMaxApplicationFormSize
	EUDORE_CONTEXT_MAX_MULTIPART_FORM_MEMORY  => DefaultContextMaxMultipartFormMemory
	EUDORE_CONTEXT_FORM_MAX_MEMORY            => DefaultContextFormMaxMemory
	EUDORE_HANDLER_EMBED_CACHE_CONTROL        => DefaultHandlerEmbedCacheControl
	EUDORE_HANDLER_EMBED_TIME                 => DefaultHandlerEmbedTime
	EUDORE_LOGGER_DEPTH_MAX_STACK             => DefaultLoggerDepthMaxStack
	EUDORE_LOGGER_ENABLE_HOOK_FATAL           => DefaultLoggerEnableHookFatal
	EUDORE_LOGGER_ENABLE_HOOK_META            => DefaultLoggerEnableHookMeta
	EUDORE_LOGGER_ENABLE_STD_COLOR            => DefaultLoggerEnableStdColor
	EUDORE_LOGGER_ENTRY_BUFFER_LENGTH         => DefaultLoggerEntryBufferLength
	EUDORE_LOGGER_ENTRY_FIELDS_LENGTH         => DefaultLoggerEntryFieldsLength
	EUDORE_LOGGER_FORMATTER                   => DefaultLoggerFormatter
	EUDORE_LOGGER_FORMATTER_FORMAT_TIME       => DefaultLoggerFormatterFormatTime
	EUDORE_LOGGER_FORMATTER_KEY_LEVEL         => DefaultLoggerFormatterKeyLevel
	EUDORE_LOGGER_FORMATTER_KEY_MESSAGE       => DefaultLoggerFormatterKeyMessage
	EUDORE_LOGGER_FORMATTER_KEY_TIME          => DefaultLoggerFormatterKeyTime
	EUDORE_LOGGER_WRITER_STDOUT_WINDOWS_COLOR => DefaultLoggerWriterStdoutWindowsColor
	EUDORE_ROUTER_LOGGER_KIND                 => DefaultRouterLoggerKind
	EUDORE_SERVER_READ_TIMEOUT                => DefaultServerReadTimeout
	EUDORE_SERVER_READ_HEADER_TIMEOUT         => DefaultServerReadHeaderTimeout
	EUDORE_SERVER_WRITE_TIMEOUT               => DefaultServerWriteTimeout
	EUDORE_SERVER_IDLE_TIMEOUT                => DefaultServerIdleTimeout
	EUDORE_SERVER_SHUTDOWN_WAIT               => DefaultServerShutdownWait
	EUDORE_DAEMON_PIDFILE                     => DefaultDaemonPidfile
	EUDORE_GODOC_SERVER                       => DefaultGodocServer
	EUDORE_TRACE_SERVER                       => DefaultTraceServer
*/
func NewConfigParseDefault() ConfigParseFunc {
	return func(ctx context.Context, c Config) error {
		parseEnvDefault(&DefaultContextMaxHandler, "CONTEXT_MAX_HANDLER")
		parseEnvDefault(&DefaultContextMaxApplicationFormSize, "CONTEXT_MAX_APPLICATION_FORM_SIZE")
		parseEnvDefault(&DefaultContextMaxMultipartFormMemory, "CONTEXT_MAX_MULTIPART_FORM_MEMORY")
		parseEnvDefault(&DefaultHandlerEmbedCacheControl, "HANDLER_EMBED_CACHE_CONTROL")
		parseEnvDefault(&DefaultHandlerEmbedTime, "HANDLER_EMBED_TIME")
		parseEnvDefault(&DefaultLoggerDepthMaxStack, "LOGGER_DEPTH_MAX_STACK")
		parseEnvDefault(&DefaultLoggerEnableHookFatal, "LOGGER_ENABLE_HOOK_FATAL")
		parseEnvDefault(&DefaultLoggerEnableHookMeta, "LOGGER_ENABLE_HOOK_META")
		parseEnvDefault(&DefaultLoggerEnableStdColor, "LOGGER_ENABLE_STD_COLOR")
		parseEnvDefault(&DefaultLoggerEntryBufferLength, "LOGGER_ENTRY_BUFFER_LENGTH")
		parseEnvDefault(&DefaultLoggerEntryFieldsLength, "LOGGER_ENTRY_FIELDS_LENGTH")
		parseEnvDefault(&DefaultLoggerFormatter, "LOGGER_FORMATTER")
		parseEnvDefault(&DefaultLoggerFormatterFormatTime, "LOGGER_FORMATTER_FORMAT_TIME")
		parseEnvDefault(&DefaultLoggerFormatterKeyLevel, "LOGGER_FORMATTER_KEY_LEVEL")
		parseEnvDefault(&DefaultLoggerFormatterKeyMessage, "LOGGER_FORMATTER_KEY_MESSAGE")
		parseEnvDefault(&DefaultLoggerFormatterKeyTime, "LOGGER_FORMATTER_KEY_TIME")
		parseEnvDefault(&DefaultLoggerWriterStdoutWindowsColor, "LOGGER_WRITER_STDOUT_WINDOWS_COLOR")
		parseEnvDefault(&DefaultRouterLoggerKind, "ROUTER_LOGGER_KIND")
		parseEnvDefault(&DefaultServerReadTimeout, "SERVER_READ_TIMEOUT")
		parseEnvDefault(&DefaultServerReadHeaderTimeout, "SERVER_READ_HEADER_TIMEOUT")
		parseEnvDefault(&DefaultServerWriteTimeout, "SERVER_WRITE_TIMEOUT")
		parseEnvDefault(&DefaultServerIdleTimeout, "SERVER_IDLE_TIMEOUT")
		parseEnvDefault(&DefaultServerShutdownWait, "SERVER_SHUTDOWN_WAIT")
		parseEnvDefault(&DefaultDaemonPidfile, "DAEMON_PIDFILE")
		parseEnvDefault(&DefaultGodocServer, "GODOC_SERVER")
		parseEnvDefault(&DefaultTraceServer, "TRACE_SERVER")
		return nil
	}
}

func parseEnvDefault[T string | bool | TypeNumber | time.Time | time.Duration](val *T, key string) {
	*val = GetAnyByString(os.Getenv("EUDORE_"+key), *val)
}

// NewConfigParseJSON method parses the json file configuration, usually the key is "config".
//
// The configuration item value is string(';' divided into multiple paths) or []string,
// if the loaded file does not exist, the file will be ignored.
//
// NewConfigParseJSON 方法解析json文件配置，通常使用key为"config"。
//
// 配置项值为string(';'分割为多路径)或[]string，如果加载文件不存在将忽略文件。
func NewConfigParseJSON(key string) ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		var paths []string
		switch val := conf.Get(key).(type) {
		case string:
			paths = strings.Split(val, ";")
		case []string:
			paths = val
		default:
			return nil
		}
		log := NewLoggerWithContext(ctx)
		log.Infof("config read json file by key: %s", key)
		for _, path := range paths {
			path = strings.TrimSpace(path)
			file, err := os.Open(path)
			if err != nil {
				log.Warningf("config ignored file: %s", err)
				continue
			}
			defer file.Close()
			err = json.NewDecoder(file).Decode(conf)
			if err != nil {
				err = fmt.Errorf("config parse json file '%s' error: %w", path, err)
				log.Info(err)
				return err
			}
			log.Infof("config load json file: %s", path)
		}
		return nil
	}
}

// NewConfigParseArgs function uses the eudore.Set method to set the command line parameter data,
// and the command line parameter uses the format of'--{key}.{sub}={value}'.
//
// Shortsmap is mapped as a short parameter. If the structure has a'flag' tag,
// it will be used as the abbreviation of the path.
// The tag length must be less than 5, the command line format is'-{short}={value},
// and the short parameter will automatically be long parameter.
//
// NewConfigParseArgs 函数使用eudore.Set方法设置命令行参数数据，命令行参数使用'--{key}.{sub}={value}'格式。
//
// shortsmap作为短参数映射，如果结构体存在'flag' tag将作为该路径的缩写，
// tag长度需要小于5，命令行格式为'-{short}={value},短参数将会自动为长参数。
func NewConfigParseArgs(shortsmap map[string][]string) ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		// 使用结构体tag初始化shorts
		shorts := make(map[string][]string)
		flag := &eachTags{tag: "flag", Repeat: make(map[uintptr]string)}
		flag.Each("", reflect.ValueOf(conf.Get("")))
		for i, tag := range flag.Tags {
			shorts[flag.Vals[i]] = append(shorts[flag.Vals[i]], tag[1:])
		}
		for k, v := range shortsmap {
			shorts[k] = append(shorts[k], v...)
		}

		args := []string{}
		log := NewLoggerWithContext(ctx)
		for _, str := range os.Args[1:] {
			key, val, _ := strings.Cut(str, "=")
			switch {
			case strings.HasPrefix(key, "--"): // 长参数
				log.Info("set os argument: " + str)
				conf.Set(key[2:], val)
			case len(key) > 1 && key[0] == '-' && key[1] != '-': // 短参数
				for _, lkey := range shorts[key[1:]] {
					log.Infof("set os short argument '%s': --%s=%s", key[1:], lkey, val)
					conf.Set(lkey, val)
				}
			default:
				args = append(args, str)
			}
		}
		conf.Set("args", args)
		return nil
	}
}

// NewConfigParseEnvs function uses the eudore.Set method to set the environment variable data,
// usually the environment variable prefix uses'ENV_'.
//
// Environment variables will be converted to lowercase paths,
// and the underscore of'_' is equivalent to the function of'.'.
//
// NewConfigParseEnvs 函数使用eudore.Set方法设置环境变量数据，环境变量默认前缀使用'ENV_'。
//
// 环境变量将移除前缀转换成小写路径，'_'下划线相当于'.'的作用
//
// exmapel: 'ENV_EUDORE_NAME=eudore' => 'eudore.name=eudore'。
func NewConfigParseEnvs(prefix string) ConfigParseFunc {
	l := len(prefix)
	return func(ctx context.Context, conf Config) error {
		log := NewLoggerWithContext(ctx)
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, prefix) {
				log.Infof("set os environment: %s", env)
				k, v, _ := strings.Cut(env, "=")
				if k != "" {
					conf.Set(strings.ToLower(strings.ReplaceAll(k[l:], "_", ".")), v)
				}
			}
		}
		return nil
	}
}

// NewConfigParseWorkdir function initializes the workspace,
// usually using the key as string("workdir") to obtain the workspace directory and switch.
//
// NewConfigParseWorkdir 函数初始化工作空间，通常使用key为string("workdir"),获取工作空间目录并切换。
func NewConfigParseWorkdir(key string) ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		dir, ok := conf.Get(key).(string)
		if ok && dir != "" {
			NewLoggerWithContext(ctx).Info("changes working directory to: " + dir)
			return os.Chdir(dir)
		}
		return nil
	}
}

// NewConfigParseHelp function if uses the structure configuration to output the'flag'
// and'description' tags to produce the default parameter description.
//
// By default, only the parameter description is output. For other descriptions,
// please wrap the NewConfigParseHelp method.
//
// Note that the properties of the configuration structure need to be non-empty,
// otherwise it will not enter the traversal.
//
// NewConfigParseHelp 函数如果使用结构体配置输出'flag'和'description' tag生产默认参数描述。
//
// 默认仅输出参数描述，其他描述内容请包装NewConfigParseHelp方法。
//
// 注意配置结构体的属性需要是非空，否则不会进入遍历。
func NewConfigParseHelp(key string) ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		if !GetAny[bool](conf.Get(key)) {
			return nil
		}

		data := reflect.ValueOf(conf.Get(""))
		flag := &eachTags{tag: "flag", Repeat: make(map[uintptr]string)}
		flag.Each("", data)
		flagmap := make(map[string]string)
		for i, tag := range flag.Tags {
			flagmap[tag[1:]] = flag.Vals[i]
		}

		desc := &eachTags{tag: "description", Repeat: make(map[uintptr]string)}
		desc.Each("", data)
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
				fmt.Printf("  -%s,", f) //nolint:forbidigo
			}
			fmt.Printf("\t --%s=%s\t%s\r\n", tag, strings.Repeat(" ", length-len(tag)), desc.Vals[i]) //nolint:forbidigo
		}
		return context.Canceled
	}
}

type eachTags struct {
	tag     string
	Tags    []string
	Vals    []string
	Repeat  map[uintptr]string
	LastTag string
}

func (each *eachTags) Each(prefix string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		if !v.IsNil() {
			_, ok := each.Repeat[v.Pointer()]
			if ok {
				return
			}
			each.Repeat[v.Pointer()] = prefix
		}
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			each.Each(prefix, v.Elem())
		}
	case reflect.Map:
		if each.LastTag != "" {
			each.Tags = append(each.Tags, fmt.Sprintf("%s.{%s}", prefix, v.Type().Key().Name()))
			each.Vals = append(each.Vals, each.LastTag)
		}
	case reflect.Slice, reflect.Array:
		length := "n"
		if v.Kind() == reflect.Array {
			length = fmt.Sprint(v.Type().Len() - 1)
		}
		last := each.LastTag
		if last != "" {
			each.Tags = append(each.Tags, fmt.Sprintf("%s.{0-%s}", prefix, length))
			each.Vals = append(each.Vals, last)
		}
		each.LastTag = last
		each.Each(fmt.Sprintf("%s.{0-%s}", prefix, length), reflect.New(v.Type().Elem()))
	case reflect.Struct:
		each.EachStruct(prefix, v)
	}
}

func (each *eachTags) EachStruct(prefix string, v reflect.Value) {
	iType := v.Type()
	for i := 0; i < iType.NumField(); i++ {
		if v.Field(i).CanSet() {
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
			each.Each(prefix+"."+name, v.Field(i))
		}
	}
}

func (each *eachTags) getValueKind(iType reflect.Type) string {
	switch iType.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
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
