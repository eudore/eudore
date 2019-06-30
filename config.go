/*
Config

Config实现配置的管理和解析。

文件：config.go
*/
package eudore

/*
keys
component
middleware
handler
*/

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	// "strings"
)

type (
	// 修改参数
	Seter interface {
		Set(string, interface{}) error
	}
	ConfigReadFunc func(string) ([]byte, error)
	//
	ConfigParseFunc   func(Config) error
	ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc
	//
	Config interface {
		Component
		Get(string) interface{}
		Set(string, interface{}) error
		ParseFuncs(ConfigParseOption)
		Parse() error
	}
	ConfigMap struct {
		Keys  map[string]interface{}
		funcs []ConfigParseFunc
		mu    sync.RWMutex
	}
	ConfigEudore struct {
		Keys  interface{}       `set:"key"`
		mu    sync.RWMutex      `set:"-"`
		funcs []ConfigParseFunc `set:"-"`
	}
)

// new router
func NewConfig(name string, arg interface{}) (Config, error) {
	name = ComponentPrefix(name, "config")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	r, ok := c.(Config)
	if ok {
		return r, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Config type", name)
}

func NewConfigMap(arg interface{}) (Config, error) {
	var keys map[string]interface{}
	if ks, ok := arg.(map[string]interface{}); ok {
		keys = ks
	} else {
		keys = make(map[string]interface{})
	}
	return &ConfigMap{
		Keys: keys,
		funcs: []ConfigParseFunc{
			ConfigParseInit,
			ConfigParseRead,
			ConfigParseConfig,
			ConfigParseArgs,
			ConfigParseEnvs,
			ConfigParseMods,
			ConfigParseHelp,
		},
	}, nil
}

func (c *ConfigMap) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(key) == 0 {
		return c.Keys
	}
	return c.Keys[key]
}

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

func (c *ConfigMap) Help(w io.Writer) error {
	if w == nil {
		w = os.Stdout
	}
	h, ok := c.Keys["Help"]
	if ok {
		_, err := fmt.Fprint(w, h)
		return err
	}
	return nil
}

func (c *ConfigMap) ParseFuncs(fn ConfigParseOption) {
	c.funcs = fn(c.funcs)
}

func (c *ConfigMap) Parse() (err error) {
	for _, fn := range c.funcs {
		err = fn(c)
		if err != nil {
			return
		}
	}
	return nil
}

func (c *ConfigMap) GetName() string {
	return ComponentConfigMapName
}

func (c *ConfigMap) Version() string {
	return ComponentConfigMapVersion
}

func (c *ConfigMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Keys)
}

func (c *ConfigMap) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.Keys)
}

func NewConfigEudore(i interface{}) (Config, error) {
	if i == nil {
		i = make(map[string]interface{})
	}
	return &ConfigEudore{
		Keys: i,
		funcs: []ConfigParseFunc{
			ConfigParseInit,
			ConfigParseRead,
			ConfigParseConfig,
			ConfigParseArgs,
			ConfigParseEnvs,
			ConfigParseMods,
			ConfigParseHelp,
		},
	}, nil
}

func (c *ConfigEudore) Get(key string) (i interface{}) {
	if len(key) == 0 {
		return c.Keys
	}
	c.mu.Lock()
	i = Get(c.Keys, key)
	c.mu.Unlock()
	return
}

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

func (c *ConfigEudore) ParseFuncs(fn ConfigParseOption) {
	c.funcs = fn(c.funcs)
}

func (c *ConfigEudore) Parse() (err error) {
	for _, fn := range c.funcs {
		err = fn(c)
		if err != nil {
			return
		}
	}
	return nil
}

func (c *ConfigEudore) GetName() string {
	return ComponentConfigEudoreName
}

func (c *ConfigEudore) Version() string {
	return ComponentConfigEudoreVersion
}

func (c *ConfigEudore) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Keys)
}

func (c *ConfigEudore) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &c.Keys)
}
