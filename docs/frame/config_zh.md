# Config

实现config主要目标为了集成配置库，然后实现框架一体化，完美将配置融合到框架中，Config作为eudore.App的一部分然后全局传递。

Config主要方法未Get/Set来读写数据，和ParseFuncs和Parse来解析数据。

```golang
type (
	ConfigParseFunc func(Config) error
	ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc
	Config interface {
		Component
		Get(string) interface{}
		Set(string, interface{}) error
		ParseFuncs(ConfigParseOption)
		Parse() error
	}
)
```

##  ConfigMap

configmap使用map[string]interface{}作为配置存储，

## ConfigEuodre

configeudore使用自定义结构体来存储配置，如果Get/Set的key路径中含有'.'就会使用eudore.Get/eudore.Set方法来按路径一层层的选择对象属性。

## ConfigParse

`type ConfigParseFunc func(Config) error`定义了配置解析函数，通过加载配置解析函数来解析配置，在Config接口中，ParseFuncs方法来修改Config的解析函数，最后调用Parse来解析全部配置。


例如ParseFuncs追加解析函数，ParseFuncs参数是当前的解析函数，返回新的解析函数,最后安顺序执行解析行为。

```golang
var c, _ := NewConfig("", nil)
// 追加一个解析函数
c.ParseFuncs(func(fn []ConfigParseFunc) []ConfigParseFunc{
	return append(fn, func(c Config) error {
		// 解析配置
		return nil
	})
})
c.Parse()
```

目前Config默认内置了七个解析函数，后续按顺序介绍。

### ConfigParseInit

作用未知，待删除。

### ConfigParseRead

```golang
func ConfigParseRead(c Config) error {
	path := GetString(c.Get("keys.config"))
	if path == "" {
		return nil //fmt.Errorf("config data is null")
	}
	// read protocol
	// get read func
	s := strings.SplitN(path, "://", 2)
	fn := ConfigLoadConfigReadFunc(s[0])
	if fn == nil {
		// use default read func
		fmt.Println("undefined read config: " + path + ", use default file:// .")
		fn = ConfigLoadConfigReadFunc("default")
	}
	data, err := fn(path)
	c.Set("keys.configdata", data)
	return err
}
```

### ConfigParseConfig

```golang
func ConfigParseConfig(c Config) error {
	data := c.Get("keys.configdata")
	if data == nil {
		return nil
	}

	err := json.Unmarshal(data.([]byte), c)
	return err	
}

```

### ConfigParseArgs

ConfigParseArgs函数解析命令行参数，必要是"--"开头，例如`--keys.help=1`,对应的执行函数就是`c.Set("keys.help", "1")`，来设置配置数据。

```golang
func ConfigParseArgs(c Config) (err error) {
	for _, str := range os.Args[1:] {
		if !strings.HasPrefix(str, "--") {
			continue
		}
		c.Set(split2byte(str[2:], '='))
	}
	return
}

```

### ConfigParseEnvs

ConfigParseEnvs解析环境变量配置，配置必须要"ENV_"开头，例如`ENV_KEYS_HELP=1`,前缀ENV_删除，后续配置路径转换成小写，'-'替换成'.',对于命令行是`--keys.help=1`,对应的执行函数就是`c.Set("keys.help", "1")`。

```golang
func ConfigParseEnvs(c Config) error {
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "ENV_") {
			k, v := split2byte(value, '=')
			k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
			c.Set(k, v)
		}
	}
	return nil
}
```

### ConfigParseMods

ConfigParseMods用于差异化配置，先获取enable的字符串数组的数组，表示启用那些模式，处理[]interface{}用于兼容。

然后读取对于的mods.xxx的数据，使用ConvertTo函数来加载对应的配置数据。

```golang
func ConfigParseMods(c Config) error {
	mod, ok  := c.Get("enable").([]string)
	if !ok {
		modi, ok := c.Get("enable").([]interface{})
		if ok {
			mod = make([]string, len(modi))
			for i, s := range modi {
				mod[i] = fmt.Sprint(s)
			}
		}else {
			return nil
		}
	}

	for _, i := range mod {
		ConvertTo(c.Get("mods." + i), c.Get(""))
	}
	return nil
}
```

### ConfigParseHelp

ConfigParseHelp的作用是输出当前配置信息，当前实现是检查`keys.help`key是否存在，存在则使用json输出到标准输出。

```golang
func ConfigParseHelp(c Config) error {
	ok := c.Get("keys.help") != nil
	if ok {
		Json(c)
	}
	return nil
}
```