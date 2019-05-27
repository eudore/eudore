package eudore

/*
定义组件的基本信息
*/

import (
	"fmt"
	"sort"
	"strings"
)



// 当前组件名称和版本。
const (
	ComponentConfigName			=	"config"
	ComponentConfigMapName		=	"config-map"
	ComponentConfigMapVersion	=	"eudore config map v1.0, use map save all config info."
	ComponentConfigEudoreName		=	"config-eudore"
	ComponentConfigEudoreVersion	=	"eudore config eudore v1.0, use reflect set and get all config info."
	ComponentLoggerName			=	"logger"
	ComponentLoggerInitName		=	"logger-init"
	ComponentLoggerInitVersion	=	"eudore logger init v1.0, save all entry."
	ComponentLoggerStdName		=	"logger-std"
	ComponentLoggerStdVersion	=	"eudore logger std v1.0, output log to /dev/null."
	// ComponentLoggerMultiName	=	"logger-multi"
	// ComponentLoggerMultiVersion	=	"eudore logger multi v1.0, output log to multiple logger."
	ComponentServerName			=	"server"
	ComponentServerStdName		=	"server-std"
	ComponentServerStdVersion	=	"eudore server std v1.0."
	ComponentServerFastcgiName	=	"server-fastcgi"
	ComponentServerEudoreName	=	"server-eudore"
	ComponentServerEudoreVersion	=	"eudore server eudore v0.1."
	ComponentServerFasthttpName		=	"server-fasthttp"
	ComponentServerFasthttpVersion	=	"eudore server fasthttp."
	ComponentServerMultiName	=	"server-multi"
	ComponentServerMultiVersion	=	"eudore server multi v1.0, server multi manage Multiple server."
	ComponentRouterName			=	"router"
	ComponentRouterRadixName	=	"router-radix"
	ComponentRouterRadixVersion	=	"eudore router radix,use radix tree std router."
	ComponentRouterFullName		=	"router-full"
	ComponentRouterFullVersion	=	"eudore router full,use radix tree full router."
	ComponentRouterDebugName	=	"router-debug"
	ComponentRouterDebugVersion	=	"eudore router debug, use router full."
	ComponentRouterHostName		=	"router-host"
	ComponentRouterHostVersion	=	"eudore router host."
	ComponentRouterInitName	=	"router-init"
	ComponentRouterInitVersion	=	"eudore router init."
	ComponentCacheName			=	"cache"
	ComponentCacheMapName		=	"cache-map"
	ComponentCacheMapVersion	=	"eudore cache map v1.0, from sync.Map."
	ComponentCacheGroupName		=	"cache-group"
	ComponentCacheGroupVersion	=	"eudore cache group v1.0."
	ComponentSessionName		=	"session"
	ComponentSessionMapName		=	"session-map"
	ComponentSessionMapVersion	=	"eudore session map v1.0, from sync.Map."
	ComponentSessionCacheName		=	"session-cache"
	ComponentSessionCacheVersion	=	"eudore session cache v1.0, save all data to eudore cache."
	ComponentViewName			=	"view"
	ComponentViewStdName		=	"view-std"
	ComponentViewStdVersion		=	"eudore view std v1.0, golang std library html/template."
	ErrComponentNameNil			=	"Failed to create component, component name is empty."
)

type (
	// New component func
	ComponentFunc func(interface{}) (Component, error)
	//
	// Get/Set Name Used for get component name and clone component.
	ComponentName interface {
		GetName() string
	}
	// All Component Method.
	//
	// Version output component Info.
	Component interface {
		ComponentName
		Version() string
	}
	ComponentConfig struct {
		Name	string		`set:"name"`
		Config	interface{}	`set:"config"`
	}
	Printer interface {
		Print(...interface{})
	}
)

var (
	// save new component func.
	defaultcom map[string]string
	components map[string]ComponentFunc
)


func init() {
	defaultcom = map[string]string{
		ComponentConfigName:	ComponentConfigMapName,
		ComponentLoggerName:	ComponentLoggerStdName,
		ComponentServerName:	ComponentServerStdName,
		ComponentRouterName:	ComponentRouterRadixName,
		ComponentCacheName:		ComponentCacheMapName,
		ComponentSessionName:	ComponentSessionMapName,
		ComponentViewName:		ComponentViewStdName,
	}
	components = make(map[string]ComponentFunc)
	RegisterComponent(ComponentConfigMapName, func(arg interface{}) (Component, error) {
		return NewConfigMap(arg)
	})
	RegisterComponent(ComponentConfigEudoreName, func(arg interface{}) (Component, error) {
		return NewConfigEudore(arg)
	})
	RegisterComponent(ComponentLoggerInitName, func(arg interface{}) (Component, error) {
		return NewLoggerInit(arg)
	})
	RegisterComponent(ComponentLoggerStdName, func(arg interface{}) (Component, error) {
		return NewLoggerStd(arg)
	})
	// RegisterComponent(ComponentLoggerMultiName, func(arg interface{}) (Component, error) {
	// 	return NewLoggerMulti(arg)
	// })
	RegisterComponent(ComponentServerStdName, func(arg interface{}) (Component, error) {
		return NewServerStd(arg)
	})
	RegisterComponent(ComponentServerMultiName, func(arg interface{}) (Component, error) {
		return NewServerMulti(arg)
	})
	// router
	RegisterComponent(ComponentRouterRadixName, func(arg interface{}) (Component, error) {
		return NewRouterRadix(arg)
	})
	RegisterComponent(ComponentRouterFullName, func(arg interface{}) (Component, error) {
		return NewRouterFull(arg)
	})
	RegisterComponent(ComponentCacheMapName, func(interface{}) (Component, error) {
		return NewCacheMap()
	})
	RegisterComponent(ComponentCacheGroupName, func(i interface{}) (Component, error) {
		return NewCacheGroup(i)
	})
	RegisterComponent(ComponentSessionMapName, func(i interface{}) (Component, error) {
		return NewSessionMap(i)
	})
	RegisterComponent(ComponentSessionCacheName, func(i interface{}) (Component, error) {
		return NewSessionCache(i)
	})
	RegisterComponent(ComponentViewStdName, func(i interface{}) (Component, error) {
		return NewViewStd(i)
	})
}

// Create a new Component.
// If name has default name,use default.
func NewComponent(name string, arg interface{}) (Component, error) {
	if len(name) == 0 {
		name = ComponentGetName(arg)
	}
	if len(name) == 0 {
		return nil, fmt.Errorf(ErrComponentNameNil)
	}
	// load defalut name
	if dfname, ok := defaultcom[name]; ok {
		name = dfname
	}
	// find new func
	fn, ok := components[name]
	if ok {
		return fn(arg)
	}
	return nil, fmt.Errorf("Unregistered component: %s", name)
}

// Register a component with the name name and fn of type ComponentFunc,
// which is the registered new constructor.
//
// 注册一个组件，名称name，fn类型为ComponentFunc，是注册的新组建的构造函数。
func RegisterComponent(name string, fn ComponentFunc) {
	components[name] = fn
}

// List all registered component names.
//
// 列出所有已注册的组件名称。
func ComponentList() []string {
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}



// Handle the component name prefix,
// name nil returns pre,
// The name prefix is not pre, then add the prefix and "-".
//
// 处理组件名称前缀，
// 名称为nil则返回pre，
// 名称前缀不是pre，则增加前缀和“ - ”。
func ComponentPrefix(name ,pre string) string {
	if len(name) == 0 || name == pre {
		return pre
	}
	if !strings.HasPrefix(name, pre + "-") {
		name = pre + "-" + name
	}
	return name
}

func ComponentGetName(i interface{}) string {
	if c, ok := i.(ComponentName); ok {
		return c.GetName()
	}
	if m, ok := i.(map[string]interface{}); ok{
		val, ok := m["name"]
		if ok {
			return val.(string)
		}
	}
	return ""
}


func ComponentSet(c Component, key string, val interface{}) (err error) {
	s, ok := c.(Seter)
	if ok {
		err = s.Set(key, val)
		if err != ErrComponentNoSupportField {
			return
		}
	}
	_, err = Set(c, key, val)
	return
}

