package eudore

import (
	"os"
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	
	// etcd "github.com/coreos/etcd/client"
)


func ConfigParseInit(c Config) error {
	return nil
}

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

func ConfigParseConfig(c Config) error {
	data := c.Get("keys.configdata")
	if data == nil {
		return nil
	}

	err := json.Unmarshal(data.([]byte), c)
	return err	
}

func ConfigParseArgs(c Config) (err error) {
	for _, str := range os.Args[1:] {
		if !strings.HasPrefix(str, "--") {
			continue
		}
		c.Set(split2byte(str[2:], '='))
	}
	return
}

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

func ConfigParseHelp(c Config) error {
	ok := c.Get("keys.help") != nil
	if ok {
		Json(c)
	}
	return nil
}

// Read config file
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
// Send http request get config info
func ConfigReadHttp(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data, err
}
//
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