package eudore

/*
保存各种全局函数，用于根据名称获得对应的函数。
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	// etcd "github.com/coreos/etcd/client"
)

// 保存全局函数
var (
	GlobalRouterCheckFunc    = make(map[string]RouterCheckFunc)
	GlobalRouterNewCheckFunc = make(map[string]RouterNewCheckFunc)
	GlobalConfigReadFunc     = make(map[string]ConfigReadFunc)
)

func init() {
	// RouterCheckFunc
	GlobalRouterCheckFunc["isnum"] = RouterCheckFuncIsnum
	// RouterNewCheckFunc
	GlobalRouterNewCheckFunc["min"] = RouterNewCheckFuncMin
	GlobalRouterNewCheckFunc["regexp"] = RouterNewCheckFuncRegexp
	// ConfigReadFunc
	GlobalConfigReadFunc["default"] = ConfigReadFile
	GlobalConfigReadFunc["file"] = ConfigReadFile
	GlobalConfigReadFunc["https"] = ConfigReadHttp
	GlobalConfigReadFunc["http"] = ConfigReadHttp
}

// ConfigParseRead 函数使用'keys.config'读取配置内容，并使用[]byte类型保存到'keys.configdata'。
func ConfigParseRead(c Config) error {
	path := GetString(c.Get("keys.config"))
	if path == "" {
		return nil //fmt.Errorf("config data is null")
	}
	// read protocol
	// get read func
	s := strings.SplitN(path, "://", 2)
	fn := GlobalConfigReadFunc[s[0]]
	if fn == nil {
		// use default read func
		fmt.Println("undefined read config: " + path + ", use default file:// .")
		fn = GlobalConfigReadFunc["default"]
	}
	data, err := fn(path)
	c.Set("keys.configdata", data)
	return err
}

// ConfigParseConfig 函数获得'keys.configdata'的内容解析配置。
func ConfigParseConfig(c Config) error {
	data := c.Get("keys.configdata")
	if data == nil {
		return nil
	}
	err := json.Unmarshal(data.([]byte), c)
	return err
}

// ConfigParseArgs 函数使用参数设置配置，参数使用--为前缀。
func ConfigParseArgs(c Config) (err error) {
	for _, str := range os.Args[1:] {
		if !strings.HasPrefix(str, "--") {
			continue
		}
		c.Set(split2byte(str[2:], '='))
	}
	return
}

// ConfigParseEnvs 函数使用环境变量设置配置，环境变量使用'ENV_'为前缀,'_'下划线相当于'.'的作用。
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

// ConfigParseMods 函数从'enable'项获得使用的模式的数组字符串，从'mods.xxx'加载配置。
func ConfigParseMods(c Config) error {
	mod, ok := c.Get("enable").([]string)
	if !ok {
		modi, ok := c.Get("enable").([]interface{})
		if ok {
			mod = make([]string, len(modi))
			for i, s := range modi {
				mod[i] = fmt.Sprint(s)
			}
		} else {
			return nil
		}
	}
	for _, i := range mod {
		ConvertTo(c.Get("mods."+i), c.Get(""))
	}
	return nil
}

// ConfigParseHelp 函数测试配置内容，如果存在'keys.help'项会使用JSON标准化输出配置到标准输出。
func ConfigParseHelp(c Config) error {
	ok := c.Get("keys.help") != nil
	if ok {
		Json(c)
	}
	return nil
}

// ConfigReadFile Read config file
func ConfigReadFile(path string) ([]byte, error) {
	if strings.HasPrefix(path, "file://") {
		path = path[7:]
	}
	data, err := ioutil.ReadFile(path)
	last := strings.LastIndex(path, ".") + 1
	if last == 0 {
		return nil, fmt.Errorf("read file config, type is null")
	}
	return data, err
}

// ConfigReadHttp Send http request get config info
func ConfigReadHttp(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data, err
}

// example: etcd://127.0.0.1:2379/config
/*func ConfigReadEtcd(path string) (string, error) {
	server, key := split2byte(path[7:], '/')
	cfg := etcd.Config{
		Endpoints:               []string{"http://" + server},
		Transport:               etcd.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := etcd.New(cfg)
	if err != nil {
		return "", err
	}
	kapi := etcd.NewKeysAPI(c)
	resp, err := kapi.Get(context.Background(), key, nil)
	return resp.Node.Value, err
}*/

// InitSignal 函数定义初始化系统信号。
func InitSignal(app *Eudore) error {
	// Register signal
	// signal 9
	app.RegisterSignal(syscall.SIGKILL, func(app *Eudore) error {
		app.WithField("signal", 9).Info("eudore received SIGKILL, eudore stop HTTP server.")
		return app.Close()
	})
	// signal 12
	app.RegisterSignal(syscall.SIGUSR2, func(app *Eudore) error {
		app.WithField("signal", 12).Info("eudore received SIGUSR2, eudore restarting HTTP server.")
		err := app.Restart()
		if err != nil {
			app.Error("eudore reload error: ", err)
		} else {
			app.Info("eudore restart success.")
		}
		return err
	})
	// signal 15
	app.RegisterSignal(syscall.SIGTERM, func(app *Eudore) error {
		app.WithField("signal", 15).Info("eudore received SIGTERM, eudore shutting down HTTP server.")
		err := app.Shutdown()
		if err != nil {
			app.Error("eudore shutdown error: ", err)
		}
		return err
	})

	return nil
}

// InitConfig 函数定义解析配置。
func InitConfig(app *Eudore) error {
	return app.Config.Parse()
}

// InitWorkdir 函数初始化工作空间，从config获取workdir的值为工作空间，然后切换目录。
func InitWorkdir(app *Eudore) error {
	dir := GetString(app.Config.Get("workdir"))
	if dir != "" {
		app.Logger.Info("changes working directory to: " + dir)
		return os.Chdir(dir)
	}
	return nil
}

/*
// InitLogger 初始化日志组件。
func InitLogger(app *Eudore) error {
	key := GetDefaultString(app.Config.Get("keys.logger"), "component.logger")
	c := app.Config.Get(key)
	if c != nil {
		_, err := app.RegisterComponent("", c)
		if err != nil {
			return err
		}
		ComponentSet(app.Router, "print", app.Logger.Debug)
		Set(app.Server, "print", app.Logger.Debug)
	}
	return nil
}

// InitServer 初始化服务组件。
func InitServer(app *Eudore) error {
	key := GetDefaultString(app.Config.Get("keys.server"), "component.server")
	c := app.Config.Get(key)
	if c != nil {
		_, err := app.RegisterComponent("", c)
		if err != nil {
			return err
		}
		Set(app.Server, "print", app.Logger.Debug)
	}
	return nil
}*/

// InitServerStart 函数启动Eudore Server。
func InitServerStart(app *Eudore) error {
	if app.Server == nil {
		err := fmt.Errorf("Eudore can't start the service, the server is empty.")
		app.Error(err)
		return err
	}

	if initlog, ok := app.Logger.(LoggerInitHandler); ok {
		app.Logger, _ = NewLoggerStd(nil)
		initlog.NextHandler(app.Logger)
	}

	lns, err := newServerListens(app.Config.Get("listeners"))
	if err != nil {
		return err
	}
	for i := range lns {
		ln, err := lns[i].Listen()
		if err != nil {
			app.Error(err)
			continue
		}
		app.Logger.Infof("listen %s %s", ln.Addr().Network(), ln.Addr().String())
		app.AddListener(ln)
	}

	if fn, ok := app.Config.Get("keys.context").(PoolGetFunc); ok {
		app.ContextPool.New = fn
	}

	app.Server.AddHandler(app)

	/*	ComponentSet(app.Server, "errfunc", func(err error) {
		fields := make(Fields)
		file, line := LogFormatFileLine(-1)
		fields["component"] = app.Server.GetName()
		fields["file"] = file
		fields["line"] = line
		app.Logger.WithFields(fields).Errorf("server error: %v", err)
	})*/
	go func() {
		app.HandleError(app.Server.Start())
	}()
	return nil
}

// RouterCheckFuncIsnum 检查字符串是否为数字。
func RouterCheckFuncIsnum(arg string) bool {
	_, err := strconv.Atoi(arg)
	return err == nil
}

// RouterNewCheckFuncMin 生成一个检查字符串最小值的RouterCheckFunc函数。
func RouterNewCheckFuncMin(str string) RouterCheckFunc {
	n, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num >= n
	}
}

// RouterNewCheckFuncRegexp 生成一个正则匹配的RouterCheckFunc函数。
func RouterNewCheckFuncRegexp(str string) RouterCheckFunc {
	// 创建正则表达式
	re, err := regexp.Compile(str)
	if err != nil {
		return nil
	}
	// 返回正则匹配校验函数
	return func(arg string) bool {
		return re.MatchString(arg)
	}
}
