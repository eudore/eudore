package eudore

/*
实现配置的管理和解析。
*/

import (
	"io"
	"os"
	"fmt"
	// "time"
	"sync"
	"strings"
	"encoding/json"
)

type (
	// 修改参数
	Seter interface {
		Set(string, interface{}) error
	}
	//
	ConfigParseFunc func(Config) error
	ConfigReadFunc func(string) (string, error)
	ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc
	//
	Config interface {
		Component
		Get(string) interface{}
		Set(string, interface{}) error
		// Help(io.Writer) error
		ParseFuncs(ConfigParseOption)
		Parse() error
	}
	ConfigMap struct {
		Keys	map[string]interface{}
		funcs	[]ConfigParseFunc
		mu		sync.RWMutex
	}
	ConfigEudore struct {
		key 	interface{}
		mu 		sync.RWMutex
		funcs	[]ConfigParseFunc
	}
)


func MapToStruct(from, to interface{}) error {
	body, err := json.Marshal(from)
	if err != nil {
		return nil
	}
	return json.Unmarshal(body, to)
}



// new router
func NewConfig(name string, arg interface{}) (Config, error) {
	name = AddComponetPre(name, "config")
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
	}else {
		keys = make(map[string]interface{})
	}
	return &ConfigMap{
		Keys: keys,
		funcs:	[]ConfigParseFunc{
			ParseInitData,
			ParseRead,
			ParseConfig,
			ParseArgs,
			ParseEnvs,
			// ParseKeys,
		},
	}, nil
}

func (c *ConfigMap) Get(key string) interface{} {
	key = c.realkeys(key)
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(key) == 0 {
		return &c.Keys
	}
	// fmt.Println(key)
	return searchMap(c.Keys, strings.Split(key, "."))
}

func searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}
	next, ok := source[path[0]]
	if ok {
		// Fast path
		if len(path) == 1 {
			return next
		}

		if m, ok := next.(map[string]interface{}); ok {
			return searchMap(m, path[1:])
		}
	}
	return nil
}

func (c *ConfigMap) Set(key string, val interface{}) error {
	c.mu.Lock()
	if len(key) == 0 {
		keys, ok := val.(map[string]interface{})
		if ok {
			c.Keys = keys
		}
	}else {
		// if len(key) > 0 && key[0] == 35 {
		// 	key = c.keyofstar(key)
		// }
		// c.Keys[key] = val	
		path := strings.Split(c.realkeys(key), ".")
		m := deepSearch(c.Keys, path[0:len(path)-1])
		m[path[len(path)-1]] = val
	}
	c.mu.Unlock()
	return nil
}


// from github.com/spf13/viper
//
// deepSearch scans deep maps, following the key indexes listed in the
// sequence "path".
// The last value is expected to be another map, and is returned.
//
// In case intermediate keys do not exist, or map to a non-map value,
// a new map is created and inserted, and the search continues from there:
// the initial map "m" may be modified!
func deepSearch(m map[string]interface{}, path []string) map[string]interface{} {
	for _, k := range path {
		m2, ok := m[k]
		if !ok {
			// intermediate key does not exist
			// => create it and continue from there
			m3 := make(map[string]interface{})
			m[k] = m3
			m = m3
			continue
		}
		m3, ok := m2.(map[string]interface{})
		if !ok {
			// intermediate key is a value
			// => replace with a new map
			m3 = make(map[string]interface{})
			m[k] = m3
		}
		// continue search from here
		m = m3
	}
	return m
}

// If the first string of the string is the hash sign "#", the string of the secondary object to the first dot "." is replaced with the corresponding key, which is looped.
//
// 如果字符串首字符串为井号"#"，将次对象到首个点"."的字符串替换成对应的key，会循环处理。
//
// #logerr => #component.logger => omponent.logger
func (c *ConfigMap) realkeys(key string) string {
	keys, ok := c.Keys["keys"].(map[string]interface{})
	if !ok {
		keys = make(map[string]interface{})
	}
	//
	strs := strings.Split(key, ".")
	for i, str := range strs {
		if len(str) > 0 && str[0] == 35 {
			newkey, ok := keys[str]
			if ok {
				str = newkey.(string)
			}else {
				str = str[1:]
			}
			strs[i] = c.realkeys(str)
		}
	}
	return strings.Join(strs, ".")
	// for len(key) > 0 && key[0] == 35 {
	// 	// first "." pos
	// 	index := strings.IndexByte(key, 46)
	// 	if index == -1 {
	// 		index = len(key)
	// 	}
	// 	// get new key
	// 	newkey, ok := keys[key[:index]]
	// 	if ok {
	// 		key = newkey.(string) + key[index:]
	// 	}else {
	// 		key = key[1:]
	// 		// break 
	// 	}
	// }
	// return key
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


func (c *ConfigEudore) Get(key string) (i interface{}) {
	if len(key) == 0 {
		return c.key
	}
	c.mu.Lock()
	i = Get(c.key, key)
	c.mu.Unlock()
	return 
}

func (c *ConfigEudore) Set(key string, val interface{}) (err error) {
	c.mu.RLock()
	if len(key) == 0 {
		c.key = val
	}else {
		_, err = Set(c.key, key, val)		
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
