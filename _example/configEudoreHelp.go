package main

/*
If the NewConfigParseHelp function uses the structure configuration to output the'flag' and'description' tags to produce the default parameter description.

By default, only the parameter description is output. For other descriptions, please wrap the NewConfigParseHelp method.

Note that the properties of the configuration structure need to be non-empty, otherwise it will not enter the traversal.

NewConfigParseHelp 函数如果使用结构体配置输出'flag'和'description' tag生产默认参数描述。

默认仅输出参数描述，其他描述内容请包装NewConfigParseHelp方法。

注意配置结构体的属性需要是非空，否则不会进入遍历。
*/
import (
	"github.com/eudore/eudore"
	"os"
	"sync"
)

func main() {
	os.Args = append(os.Args, "-db=localhost", "-h", "eudore", "--help", "--component.server.readtimeout=12s")
	conf := &helpConfig{Iface: &helpDBConfig{}}
	conf.Link = conf
	app := eudore.NewApp(eudore.NewConfigEudore(conf))
	app.Parse()

	app.Infof("config db is %v", conf.Component)
	app.Infof("config readtimeout is %v", conf.Component.Server.ReadTimeout)

	app.CancelFunc()
	app.Run()
}

type helpConfig struct {
	sync.RWMutex
	Command   string                      `json:"command" alias:"command" description:"app start command, start/stop/status/restart" flag:"cmd"`
	Pidfile   string                      `json:"pidfile" alias:"pidfile" description:"pid file localtion"`
	Workdir   string                      `json:"workdir" alias:"workdir" description:"set app working directory"`
	Config    string                      `json:"config" alias:"config" description:"config path" flag:"f"`
	Help      bool                        `json:"help" alias:"help" description:"output help info" flag:"h"`
	Enable    []string                    `json:"enable" alias:"enable" description:"enable config mods"`
	Mods      map[string]*helpConfig      `json:"mods" alias:"mods" description:"config mods"`
	Listeners []eudore.ServerListenConfig `json:"listeners" alias:"listeners"`
	Component *helpComponentConfig        `json:"component" alias:"component"`
	Length    int                         `json:"length" alias:"length" description:"this is int"`
	Num       [3]int                      `json:"num" alias:"num" description:"this is array"`
	Body      []byte                      `json:"body" alias:"body" description:"this is []byte"`
	Float     float64                     `json:"body" alias:"body" description:"this is float"`

	Auth  *helpAuthConfig `json:"auth" alias:"auth"`
	Iface interface{}
	Link  interface{} `json:"-" alias:"link"`
	// Node *Node
}

type Node struct {
	Next *Node
}

// ComponentConfig 定义website使用的组件的配置。
type helpComponentConfig struct {
	DB     helpDBConfig            `json:"db" alias:"db"`
	Logger *eudore.LoggerStdConfig `json:"logger" alias:"logger"`
	Server *eudore.ServerStdConfig `json:"server" alias:"server"`
	Notify map[string]string       `json:"notify" alias:"notify"`
	Pprof  *helpPprofConfig        `json:"pprof" alias:"pprof"`
	Black  map[string]bool         `json:"black" alias:"black"`
}
type helpDBConfig struct {
	Driver string `json:"driver" alias:"driver" description:"database driver type"`
	Config string `json:"config" alias:"config" description:"database config info" flag:"db"`
}
type helpPprofConfig struct {
	Godoc     string            `json:"godoc" alias:"godoc" description:"godoc server"`
	BasicAuth map[string]string `json:"basicauth" alias:"basicauth" description:"basic auth username and password"`
}

type helpAuthConfig struct {
	Secrets  map[string]string    `json:"secrets" alias:"secrets" description:"default auth secrets"`
	IconTemp string               `json:"icontemp" alias:"icontemp" description:"save icon temp dir"`
	Sender   helpMailSenderConfig `json:"sender" alias:"sender" description:""`
}
type helpMailSenderConfig struct {
	Username string `json:"username" alias:"username" description:"email send username"`
	Password string `json:"password" alias:"password" description:"email send password"`
	Addr     string `json:"addr" alias:"addr"`
	Subject  string `json:"subject" alias:"subject"`
}
