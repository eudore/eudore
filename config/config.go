package config

import (
	"os"
	"fmt"
	"sync"
	"errors"
	"strings"
)

type (
	ConfigFunc func(*Config) error
	ReadFunc func(string) (string, error)
	// Global is the current operation is a data object.
	// The muc type is sync.RWMutex, which is used to guarantee concurrent security when getting/set data.
	// Parses is a parsing function queue that parses in the current order when parsing.
	// keys saves the configuration data and data as the mapping starting with '#'.
	// 
	// Global是当前操作是数据对象。
	// muc类型是sync.RWMutex，用于保证在get/set数据时的并发安全。
	// Parses是一个解析函数队列，解析时按当前顺序解析。
	// keys保存配置数据和data为'#'开头的映射。
	Config struct {
		Global		interface{}
		muc			sync.RWMutex
		Parses		[]ConfigFunc	`json:"-"`
		mu			sync.Mutex
		keys		sync.Map
	}
)

var (
	errArg 			=	errors.New("undefined args")
	errType 		= 	errors.New("undefined type")
	errValue 		= 	errors.New("undefined value type")
	errUndefined	=	errors.New("setdata use undefined interface")
)

// Returns a default eudore.config.Config object.
// The list of default parsing functions is: ParseStart, ParseRead, ParseConfig, ParseModes, ParseArgs, ParseEnvs, ParseEnd.
// The parsing function uses three detailed data, #config, #workdir, #help, #test.
// eudore uses two pieces of data, #command, #pidfile.
//
// 返回一个默认eudore.config.Config对象。
// 默认解析函数列表为：ParseStart、ParseRead、ParseConfig、ParseModes、ParseArgs、ParseEnvs、ParseEnd。
// 解析函数使用三详细数据，#config、#workdir、#help、#test。
// eudore使用两项数据，#command、#pidfile。
func NewConfig(i interface{}) *Config {
	return &Config{
		Global: i,
		Parses: []ConfigFunc{
			ParseStart,
			ParseRead,
			ParseConfig,
			ParseModes,
			ParseArgs,
			ParseEnvs,
			ParseEnd,
		},
	}
}



func (c *Config) SetKey(key , val string) {
	c.keys.Store(key, val)
}

func (c *Config) GetKey(key string) string {
	val, ok := c.keys.Load(key)
	if ok {
		return val.(string)
	}
	return ""
}

// Output Config.Global the Help().
// 
// 输出Config.Global对象的H帮助信息。
func (c *Config) Help() error {
	return Help(c.Global, "  --", os.Stdout)
}

// Use the key to set the data for Config.Gloabl.
// value is unmarshal a interface{} data.
// If the first character of the key is '#', then the new key mapped by keys is used.
// 
// 使用key设置Config.Gloabl的数据。
// value解析成一个interface{}的数据。
// 如果key的首字符是'#',则使用keys的映射出的新key。
func (c *Config) SetData(key, value string) error {
	c.muc.Lock()
	defer c.muc.Unlock()
	// if key first is '#', set key is
	if len(key) > 0 && key[0] == 35 {
		key = c.keyofstar(key)
	}
	return SetData(c.Global, key, value)
}

// Use the key to read the data of Config.Gloabl.
// If the first character of the key is '#', then the new key mapped by keys is used.
// 
// 使用key读取Config.Gloabl的数据。
// 如果key的首字符是'#',则使用keys的映射出的新key。
func (c *Config) GetData(key string) (interface{}, error) {
	c.muc.RLock()
	defer c.muc.RUnlock()
	fmt.Println("------- get:", key)
	// if key first is '#', set key is
	if len(key) > 0 && key[0] == 35 {
		key = c.keyofstar(key)	
	}
	return GetData(c.Global, key)
}

// If the first string of the string is the hash sign "#", the string of the secondary object to the first dot "." is replaced with the corresponding key, which is looped.
//
// 如果字符串首字符串为井号"#"，将次对象到首个点"."的字符串替换成对应的key，会循环处理。
//
// #logerr => #component.logger => omponent.logger
func (c *Config) keyofstar(key string) string {
	for len(key) > 0 && key[0] == 35 {
		// first "." pos
		index := strings.IndexByte(key, 46)
		if index == -1 {
			index = len(key)
		}
		// get new key
		newkey, ok := c.keys.Load(key[:index])
		if ok {
			key = newkey.(string) + key[index:]
		}else {
			break
		}
	}
	return key
}

// use key GetData and convert the value to int.
// If the value cannot be read, the second argument is returned.
// 
// 使用key GetData，将值转换成int。
// 如果无法读取值，则返回第二个参数。
func (c *Config) GetInt(key string, num ...int) int {
	i, err := c.GetData(key)
	if err == nil {
		n, ok := i.(int)
		if ok {
			return n
		}
	}
	for _, n := range num {
		return n
	}
	return 0
}


// use key GetData and convert the value to string.
// If the value cannot be read, the second argument is returned.
// 
// 使用key GetData，将值转换成string。
// 如果无法读取值，则返回第二个参数。
func (c *Config) GetString(key string, strs ...string) string {
	i, err := c.GetData(key)
	if err == nil {
		str, ok := i.(string)
		if ok {
			return str
		}
	}
	for _, str := range strs {
		return str
	}
	return ""
}

// use key GetData and convert the value to bool.
// If the value cannot be read, the second argument is returned.
// 
// 使用key GetData，将值转换成bool。
// 如果无法读取值，则返回第二个参数。
func (c *Config) GetBool(key string, bs ...bool) bool {
	i, err := c.GetData(key)
	if err == nil {
		b, ok := i.(bool)
		if ok {
			return b
		}
	}
	for _, b := range bs {
		return b
	}
	return false
}


// Execute the parsing object, if an parsing function returns an error, the end parsing returns an error.
//
// 执行解析对象，如果一个解析函数返回错误，则结束解析返回错误。
func (c *Config) Parse() (err error) {
	for _, fn := range c.Parses {
		err = fn(c)
		if err != nil {
			return
		}
	}
	return
}

// Append multiple new analytic functions.
//
// 追加多个新的解析函数。
func (c *Config) AddParseFunc(fns ...ConfigFunc) {
	c.Parses = append(c.Parses, fns...)
}

// Clear the current parsing function queue.
//
// 清空当前解析函数队列。
func (c *Config) CleanParseFunc() {
	c.Parses = c.Parses[0:0]
}
