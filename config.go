package eudore

import (
	"encoding/json"
	"sync"
)

type (
	// ConfigReadFunc 定义配置数据读取参数。
	ConfigReadFunc func(string) ([]byte, error)
	// ConfigParseFunc 定义配置解析函数。
	ConfigParseFunc func(Config) error
	// ConfigParseOption 定义配置解析选项，用于修改配置解析函数。
	ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc
	// Config 定义配置管理，使用配置读写和解析功能。
	Config interface {
		Get(string) interface{}
		Set(string, interface{}) error
		ParseOption(ConfigParseOption)
		Parse() error
	}
	// ConfigMap 使用map保存配置。
	ConfigMap struct {
		Keys  map[string]interface{}
		funcs []ConfigParseFunc
		mu    sync.RWMutex
	}
	// ConfigEudore 使用结构体或map保存配置，通过反射来读写属性。
	ConfigEudore struct {
		Keys  interface{}       `set:"key"`
		mu    sync.RWMutex      `set:"-"`
		funcs []ConfigParseFunc `set:"-"`
	}
)

// NewConfigMap 创建一个ConfigMap，如果传入参数为map[string]interface{},则作为初始化数据。
func NewConfigMap(arg interface{}) Config {
	var keys map[string]interface{}
	if ks, ok := arg.(map[string]interface{}); ok {
		keys = ks
	} else {
		keys = make(map[string]interface{})
	}
	return &ConfigMap{
		Keys: keys,
		funcs: []ConfigParseFunc{
			ConfigParseRead,
			ConfigParseConfig,
			ConfigParseArgs,
			ConfigParseEnvs,
			ConfigParseMods,
			ConfigParseHelp,
		},
	}
}

// Get 方法获取一个属性，如果键为空字符串，返回保存全部数据的map对象。
func (c *ConfigMap) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(key) == 0 {
		return c.Keys
	}
	return c.Keys[key]
}

// Set 方法设置一个属性，如果键为空字符串且值类型是map[string]interface{},则替换保存全部数据的map对象。
func (c *ConfigMap) Set(key string, val interface{}) error {
	c.mu.Lock()
	if len(key) == 0 {
		keys, ok := val.(map[string]interface{})
		if ok {
			c.Keys = keys
		}
	} else {
		c.Keys[key] = val
	}
	c.mu.Unlock()
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
			return
		}
	}
	return nil
}

// MarshalJSON 实现json.Marshaler接口，试json序列化直接操作保存的数据。
func (c *ConfigMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Keys)
}

// UnmarshalJSON 实现json.Unmarshaler接口，试json反序列化直接操作保存的数据。
func (c *ConfigMap) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.Keys)
}

// NewConfigEudore 创建一个ConfigEudore，如果传入参数为空，使用空map[string]interface{}作为初始化数据。
func NewConfigEudore(i interface{}) Config {
	if i == nil {
		i = make(map[string]interface{})
	}
	return &ConfigEudore{
		Keys: i,
		funcs: []ConfigParseFunc{
			ConfigParseRead,
			ConfigParseConfig,
			ConfigParseArgs,
			ConfigParseEnvs,
			ConfigParseMods,
			ConfigParseHelp,
		},
	}
}

// Get 方法实现读取数据属性的一个属性。
func (c *ConfigEudore) Get(key string) (i interface{}) {
	if len(key) == 0 {
		return c.Keys
	}
	c.mu.Lock()
	i = Get(c.Keys, key)
	c.mu.Unlock()
	return
}

// Set 方法实现设置数据的一个属性。
func (c *ConfigEudore) Set(key string, val interface{}) (err error) {
	c.mu.RLock()
	if len(key) == 0 {
		c.Keys = val
	} else {
		c.Keys, err = Set(c.Keys, key, val)
	}
	c.mu.RUnlock()
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
			return
		}
	}
	return nil
}

// MarshalJSON 实现json.Marshaler接口，试json序列化直接操作保存的数据。
func (c *ConfigEudore) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Keys)
}

// UnmarshalJSON 实现json.Unmarshaler接口，试json反序列化直接操作保存的数据。
func (c *ConfigEudore) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.Keys)
}
