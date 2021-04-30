package main

/*
ConfigEudore的配置存储读写如果组合一个sync.RWMutex对象，实现sync读写锁的四个方法

那么ConfigEudore的Get/Se方法会使用该锁，这样使用属性操作和Get/Set操作使用同一个锁保证并发安全。

// configEudore 使用结构体或map保存配置，通过反射来读写属性。
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
*/

import (
	"sync"

	"github.com/eudore/eudore"
)

type (
	eudoreLockerConfig struct {
		Bool   bool        `alias:"bool"`
		Int    int         `alias:"int"`
		String string      `alias:"string"`
		User   user2       `alias:"user"`
		Struct interface{} `alias:"struct"`
		sync.RWMutex
	}
	user2 struct {
		Name string `alias:"name"`
		Mail string `alias:"mail"`
	}
)

func main() {
	conf := &eudoreLockerConfig{}
	app := eudore.NewApp(eudore.NewConfigEudore(conf))

	// 设属性
	conf.Lock()
	conf.Int = 20
	conf.String = "app set string"
	conf.Unlock()
	app.Set("bool", true)
	app.Set("user.name", "EudoreName")
	app.Set("struct", struct {
		Name string
		Age  int
	}{"eudore", 2020})
	app.Set("field", "not found")

	// 读取部分配置
	app.Debugf("%#v", app.GetInt("int"))
	app.Debugf("%#v", app.GetInt("string"))
	app.Debugf("%#v", app.GetString("string"))
	app.Debugf("%#v", app.GetBool("bool"))
	app.Debugf("%#v", app.Get("struct"))
	app.Debugf("%#v", app.Get("field"))

	// 输出全部配置信息
	app.Debugf("%#v", conf)
	app.Debugf("%#v", app.Get(""))

	app.CancelFunc()
	app.Run()
}
