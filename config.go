package eudore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// The Config interface defines config read-write and parsing functions.
//
// Use [ConfigParseFunc] to implement custom parsing.
type Config interface {
	// The Get method implements getting data,
	// and uses the RLock method to lock the data,
	//
	// if key is empty string return metadata.
	Get(key string) any

	// The Set method implements setting data,
	// and uses the Lock method to lock the data,
	//
	// If key is empty string set metadata.
	Set(key string, val any) error

	// ParseOption method adds [ConfigParseFunc],
	// If it is empty, clear the current func list
	ParseOption(fn ...ConfigParseFunc)

	// The Parse method executes all [ConfigParseFunc].
	// If the parsing funcs returns error, it stops parsing and returns error.
	Parse(ctx context.Context) error
}

// ConfigParseFunc defines the [Config] parsing function.
//
// [context.Context] will trigger [DefaultConfigParseTimeout] when parsing.
//
// You can use [Config] to modify the configuration when parsing.
type ConfigParseFunc func(context.Context, Config) error

// configStd uses a structure or map to save configurations,
// and reads and writes properties through attributes or [reflect].
type configStd struct {
	Data  any               `alias:"data"`
	Map   map[string]any    `alias:"map"`
	Funcs []ConfigParseFunc `alias:"funcs"`
	Err   error             `alias:"err"`
	Lock  rwLocker          `alias:"lock"`
}

type rwLocker interface {
	sync.Locker
	RLock()
	RUnlock()
}

type MetadataConfig struct {
	Health bool   `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name   string `json:"name" protobuf:"2,name=name" yaml:"name"`
	Data   any    `json:"data,omitempty" protobuf:"3,name=data,omitempty" yaml:"data,omitempty"`
	Error  error  `json:"error,omitempty" protobuf:"4,name=error,omitempty" yaml:"error,omitempty"`
}

// The NewConfig function creates a [Config] implementation.
//
// The data variable is the configuration data and
// is read and written using [GetAnyByPath]/[SetAnyByPath].
//
// If the data type is map[string]any, it is read and written using map key.
//
// If the data variable implements the [rwLocker] read-write lock interface,
// it will be used during [Config.Get]/[Config.Set].
//
// The default [ConfigParseFunc] is [DefaultConfigAllParseFunc].
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
		Funcs: append([]ConfigParseFunc{}, DefaultConfigAllParseFunc...),
		Lock:  mu,
	}
}

func (c *configStd) Metadata() any {
	return MetadataConfig{
		Health: c.Err == nil,
		Name:   "eudore.configStd",
		Data:   anyMetadata(c.Data),
		Error:  c.Err,
	}
}

// The Get method implements getting data,
// and uses the RLock method to lock the data,
//
// if key is empty string return metadata.
func (c *configStd) Get(key string) any {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	if len(key) == 0 {
		if c.Map != nil {
			return &c.Map
		}
		return c.Data
	}
	if c.Map != nil {
		return c.Map[key]
	}
	return GetAnyByPath(c.Data, key)
}

// The Set method implements setting data,
// and uses the Lock method to lock the data,
//
// If key is empty string set metadata.
func (c *configStd) Set(key string, val any) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if len(key) == 0 {
		c.Data = val
		c.Map, _ = val.(map[string]any)
		return nil
	}
	if c.Map != nil {
		c.Map[key] = val
		return nil
	}
	return SetAnyByPath(c.Data, key, val)
}

// ParseOption method adds [ConfigParseFunc],
// If it is empty, clear the current func list
//
// The default [ConfigParseFunc] is [eudore.ConfigAllParseFunc].
func (c *configStd) ParseOption(fn ...ConfigParseFunc) {
	if fn == nil {
		c.Funcs = nil
		c.Err = nil
	} else {
		c.Funcs = append(c.Funcs, fn...)
	}
}

// The Parse method executes all [ConfigParseFunc].
// If the parsing function returns error, it stops parsing and returns error.
func (c *configStd) Parse(ctx context.Context) error {
	if c.Err != nil {
		return c.Err
	}
	ctx, cancel := context.WithTimeout(ctx, DefaultConfigParseTimeout)
	defer cancel()
	for _, fn := range c.Funcs {
		c.Err = fn(ctx, c)
		if c.Err != nil {
			if !errors.Is(c.Err, context.Canceled) {
				// replace logger in parsing
				name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
				c.Err = fmt.Errorf(ErrConfigParseError, name, c.Err)
				NewLoggerWithContext(ctx).Error(c.Err.Error())
			}
			return c.Err
		}
	}
	NewLoggerWithContext(ctx).Info("config parse done")
	return nil
}

// MarshalJSON implements the [json.Marshaler] interface.
func (c *configStd) MarshalJSON() ([]byte, error) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return json.Marshal(c.Data)
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (c *configStd) UnmarshalJSON(data []byte) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	return json.Unmarshal(data, &c.Data)
}

// The NewConfigParseJSON function creates [ConfigParseFunc] to parse the json
// configuration file.
//
// refer: [NewConfigParseDecoder].
func NewConfigParseJSON(key string) ConfigParseFunc {
	fn := NewConfigParseDecoder(key, "json",
		func(reader io.Reader, data any) error {
			return json.NewDecoder(reader).Decode(data)
		},
	)
	// update func name
	return func(ctx context.Context, conf Config) error {
		return fn(ctx, conf)
	}
}

// The NewConfigParseDecoder function creates [ConfigParseFunc]
// to parse the specified decoder configuration file.
//
// Get the configuration file path from the command line --{key}=,
// environment variable EUDORE_{KEY}, and conf.Get(key) in order.
//
// The allowed type is string or []string.
//
// Will try to load {key} and workdir from [os.Args] and [os.Getenv],
// and the env prefix uses [DefaultConfigEnvPrefix].
func NewConfigParseDecoder(key, format string,
	decode func(io.Reader, any) error,
) ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		log := NewLoggerWithContext(ctx)
		log.Infof("config read %s file by key: %s", format, key)
		for _, path := range getConfigPath(conf, key) {
			path = strings.TrimSpace(path)
			file, err := os.Open(path)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Warningf("config ignored file: %s", err)
				}
				continue
			}
			defer file.Close()

			if format == "json" {
				err = decode(file, conf)
			} else {
				err = decode(file, conf.Get(""))
			}
			if err != nil {
				err = fmt.Errorf(ErrConfigParseDecoder, format, path, err)
				log.Info(err)
				return err
			}
			log.Infof("config load %s file: %s", format, path)
		}
		return nil
	}
}

func getConfigPath(conf Config, key string) []string {
	config := getValueFromArgsAndEnv(key)
	if config != "" {
		workdir := getValueFromArgsAndEnv("workdir")
		workabs, _ := filepath.Abs(workdir)
		wd, _ := os.Getwd()
		if strings.EqualFold(wd, workdir) || strings.EqualFold(wd, workabs) {
			workdir = ""
		}

		if !filepath.IsAbs(config) {
			config = filepath.Join(workdir, config)
		}
		return strings.Split(config, ";")
	}

	switch val := conf.Get(key).(type) {
	case string:
		return strings.Split(val, ";")
	case []string:
		return val
	default:
		return nil
	}
}

func getValueFromArgsAndEnv(key string) string {
	keyarg := fmt.Sprintf("--%s=", key)
	for _, str := range os.Args {
		if strings.HasPrefix(str, keyarg) {
			return str[3+len(key):]
		}
	}

	return os.Getenv(DefaultConfigEnvPrefix + strings.ToUpper(key))
}

// The NewConfigParseEnvs function creates [ConfigParseFunc] to parse
// [os.Environ] into [Config].
//
// Environment variables will remove prefixes and convert to lowercase paths,
// converting '_' to '.'.
//
// exmapel: 'ENV_EUDORE_NAME=eudore' => 'eudore.name=eudore'.
//
// Note: The path after env conversion is all lowercase, and the structure needs
// to specify an 'alias' tag.
func NewConfigParseEnvs(prefix string) ConfigParseFunc {
	if prefix == "" {
		prefix = DefaultConfigEnvPrefix
	}
	l := len(prefix)
	return func(ctx context.Context, conf Config) error {
		log := NewLoggerWithContext(ctx)
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, prefix) {
				log.Infof("set os environment: %s", env)
				k, v, _ := strings.Cut(env, "=")
				if k != "" {
					_ = conf.Set(strings.ToLower(strings.ReplaceAll(k[l:], "_", ".")), v)
				}
			}
		}
		return nil
	}
}

// The NewConfigParseArgs function creates [ConfigParseFunc] to parse [os.Args]
// into [Config].
//
// Command line parameters use '--{key}.{sub}={value}' format,
// short parameters use '-{key}={value}'.
//
// if the struct has 'flag' tag, will be used as an abbreviation for the path.
func NewConfigParseArgs() ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		log := NewLoggerWithContext(ctx)
		args := []string{}
		// Initialize shorts using struct flag tag
		shorts := newStructShorts(conf.Get(""))
		for _, str := range os.Args[1:] {
			key, val, _ := strings.Cut(str, "=")
			switch {
			case strings.HasPrefix(key, "--"): // full param
				log.Info("set os argument: " + str)
				_ = conf.Set(key[2:], val)
			case len(key) > 1 && key[0] == '-' && key[1] != '-': // short param
				for _, lkey := range shorts[key[1:]] {
					log.Infof("set os short argument: %s --%s=%s",
						key[1:], lkey, val,
					)
					_ = conf.Set(lkey, val)
				}
			default:
				args = append(args, str)
			}
		}
		_ = conf.Set("args", args)
		return nil
	}
}

// The NewConfigParseWorkdir function creates [ConfigParseFunc] to initializes
// the workspace,
// usually using the key as string("workdir") to get the workspace directory and
// changes.
func NewConfigParseWorkdir(key string) ConfigParseFunc {
	return func(ctx context.Context, conf Config) error {
		dir, ok := conf.Get(key).(string)
		if ok && dir != "" {
			NewLoggerWithContext(ctx).Info(
				"changes working directory to: " + dir,
			)
			return os.Chdir(dir)
		}
		return nil
	}
}

/*
The NewConfigParseEnvFile function creates the [ConfigParseFunc] parsing
environment files.

The default ENV file is [DefaultConfigEnvFiles], using ; as a separator.

If the line Env is in the env format, use [os.Setenv] to set it to the process.

If the Env value is an empty string, use [os.Unsetenv] to delete this Env.

If the first character of the Env value is ', it is a multi-line value until the
end of a line also has ',

Multi-line values replace \r\n to \n and exec TrimSpace.

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
	return func(ctx context.Context, _ Config) error {
		log := NewLoggerWithContext(ctx)
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			log.Info("confif load env file", file)
			parseEnvFile(log, data)
		}
		return nil
	}
}

var regEnvLine = regexp.MustCompile(`[a-zA-Z]\w*`)

func parseEnvFile(log Logger, data []byte) {
	lines := strings.Split(
		strings.ReplaceAll(string(data), "\r\n", "\n"),
		"\n",
	)
	char := "'"
	for i := range lines {
		key, val, ok := strings.Cut(lines[i], "=")
		if ok && regEnvLine.MatchString(key) {
			if strings.HasPrefix(val, char) {
				if strings.HasSuffix(val, char) {
					val = strings.TrimSuffix(val[1:], char)
					val = strings.TrimSpace(val)
				} else if i+1 < len(lines) {
					lines[i+1] = lines[i] + "\n" + lines[i+1]
					continue
				}
			}

			log.Infof("set file environment: %s=%s", key, val)
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

/*
The NewConfigParseJSON function creates the [ConfigParseFunc] parsing
environment and changes it to the global default configuration.

env to keys:

	ENV_CONFIG_PARSE_TIMEOUT              => DefaultConfigParseTimeout
	ENV_CONTEXT_MAX_APPLICATION_FORM_SIZE => DefaultContextMaxApplicationFormSize
	ENV_CONTEXT_MAX_MULTIPART_FORM_MEMORY => DefaultContextMaxMultipartFormMemory
	ENV_HANDLER_DATA_TEMPLATE_RELOAD      => DefaultHandlerDataTemplateReload
	ENV_HANDLER_EMBED_CACHE_CONTROL       => DefaultHandlerEmbedCacheControl
	ENV_HANDLER_EMBED_TIME                => DefaultHandlerEmbedTime
	ENV_HANDLER_EXTENDER_SHOW_NAME        => DefaultHandlerExtenderShowName
	ENV_LOGGER_ENTRY_BUFFER_LENGTH        => DefaultLoggerEntryBufferLength
	ENV_LOGGER_ENTRY_FIELDS_LENGTH        => DefaultLoggerEntryFieldsLength
	ENV_LOGGER_FORMATTER                  => DefaultLoggerFormatter
	ENV_LOGGER_FORMATTER_FORMAT_TIME      => DefaultLoggerFormatterFormatTime
	ENV_LOGGER_FORMATTER_KEY_LEVEL        => DefaultLoggerFormatterKeyLevel
	ENV_LOGGER_FORMATTER_KEY_MESSAGE      => DefaultLoggerFormatterKeyMessage
	ENV_LOGGER_FORMATTER_KEY_TIME         => DefaultLoggerFormatterKeyTime
	ENV_LOGGER_HOOK_FATAL                 => DefaultLoggerHookFatal
	ENV_LOGGER_WRITER_STDOUT              => DefaultLoggerWriterStdout
	ENV_LOGGER_WRITER_STDOUT_COLOR        => DefaultLoggerWriterStdoutColor
	ENV_ROUTER_LOGGER_KIND                => DefaultRouterLoggerKind
	ENV_SERVER_READ_TIMEOUT               => DefaultServerReadTimeout
	ENV_SERVER_READ_HEADER_TIMEOUT        => DefaultServerReadHeaderTimeout
	ENV_SERVER_WRITE_TIMEOUT              => DefaultServerWriteTimeout
	ENV_SERVER_IDLE_TIMEOUT               => DefaultServerIdleTimeout
	ENV_SERVER_SHUTDOWN_WAIT              => DefaultServerShutdownWait
	ENV_DAEMON_PIDFILE                    => DefaultDaemonPidfile
	ENV_GODOC_SERVER                      => DefaultGodocServer
*/
func NewConfigParseDefault() ConfigParseFunc {
	return func(_ context.Context, _ Config) error {
		parseEnvDefault(&DefaultConfigParseTimeout, "CONFIG_PARSE_TIMEOUT")
		parseEnvDefault(&DefaultContextMaxApplicationFormSize, "CONTEXT_MAX_APPLICATION_FORM_SIZE")
		parseEnvDefault(&DefaultContextMaxMultipartFormMemory, "CONTEXT_MAX_MULTIPART_FORM_MEMORY")
		parseEnvDefault(&DefaultHandlerDataTemplateReload, "HANDLER_DATA_TEMPLATE_RELOAD")
		parseEnvDefault(&DefaultHandlerEmbedCacheControl, "HANDLER_EMBED_CACHE_CONTROL")
		parseEnvDefault(&DefaultHandlerEmbedTime, "HANDLER_EMBED_TIME")
		parseEnvDefault(&DefaultHandlerExtenderShowName, "HANDLER_EXTENDER_SHOW_NAME")
		parseEnvDefault(&DefaultLoggerEntryBufferLength, "LOGGER_ENTRY_BUFFER_LENGTH")
		parseEnvDefault(&DefaultLoggerEntryFieldsLength, "LOGGER_ENTRY_FIELDS_LENGTH")
		parseEnvDefault(&DefaultLoggerFormatter, "LOGGER_FORMATTER")
		parseEnvDefault(&DefaultLoggerFormatterFormatTime, "LOGGER_FORMATTER_FORMAT_TIME")
		parseEnvDefault(&DefaultLoggerFormatterKeyLevel, "LOGGER_FORMATTER_KEY_LEVEL")
		parseEnvDefault(&DefaultLoggerFormatterKeyMessage, "LOGGER_FORMATTER_KEY_MESSAGE")
		parseEnvDefault(&DefaultLoggerFormatterKeyTime, "LOGGER_FORMATTER_KEY_TIME")
		parseEnvDefault(&DefaultLoggerHookFatal, "LOGGER_HOOK_FATAL")
		parseEnvDefault(&DefaultLoggerWriterStdout, "LOGGER_WRITER_STDOUT")
		parseEnvDefault(&DefaultLoggerWriterStdoutColor, "LOGGER_WRITER_STDOUT_COLOR")
		parseEnvDefault(&DefaultRouterLoggerKind, "ROUTER_LOGGER_KIND")
		parseEnvDefault(&DefaultServerReadTimeout, "SERVER_READ_TIMEOUT")
		parseEnvDefault(&DefaultServerReadHeaderTimeout, "SERVER_READ_HEADER_TIMEOUT")
		parseEnvDefault(&DefaultServerWriteTimeout, "SERVER_WRITE_TIMEOUT")
		parseEnvDefault(&DefaultServerIdleTimeout, "SERVER_IDLE_TIMEOUT")
		parseEnvDefault(&DefaultServerShutdownWait, "SERVER_SHUTDOWN_WAIT")
		parseEnvDefault(&DefaultDaemonPidfile, "DAEMON_PIDFILE")
		parseEnvDefault(&DefaultGodocServer, "GODOC_SERVER")
		return nil
	}
}

func parseEnvDefault[T string | bool | time.Time | time.Duration |
	typeNumber](val *T, key string) {
	*val = GetAnyByString(os.Getenv(DefaultConfigEnvPrefix+key), *val)
}

type eachTags struct {
	shorts map[string][]string
	repeat map[uintptr]string
}

func newStructShorts(data any) map[string][]string {
	each := &eachTags{
		shorts: make(map[string][]string),
		repeat: make(map[uintptr]string),
	}
	each.Each("", reflect.ValueOf(data))
	return each.shorts
}

func (each *eachTags) Each(prefix string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		if !v.IsNil() {
			_, ok := each.repeat[v.Pointer()]
			if ok {
				return
			}
			each.repeat[v.Pointer()] = prefix
		}
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			each.Each(prefix, v.Elem())
		}
	case reflect.Struct:
		iType := v.Type()
		for i := 0; i < iType.NumField(); i++ {
			if v.Field(i).CanSet() {
				flag := iType.Field(i).Tag.Get("flag")
				name := iType.Field(i).Tag.Get("alias")
				if name == "" {
					name = iType.Field(i).Name
				}

				if flag != "" && getEachValueKind(iType.Field(i).Type) {
					val := strings.TrimPrefix(prefix+"."+name, ".")
					each.shorts[flag] = append(each.shorts[flag], val)
				} else {
					each.Each(prefix+"."+name, v.Field(i))
				}
			}
		}
	}
}

func getEachValueKind(iType reflect.Type) bool {
	switch iType.Kind() {
	case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Uint,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Ptr:
		return getEachValueKind(iType.Elem())
	default:
		if iType.Kind() == reflect.Slice &&
			iType.Elem().Kind() == reflect.Uint8 {
			return true
		}
		return false
	}
}
