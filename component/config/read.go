package config


import (
	"fmt"
	"time"
	"strings"
	"net/http"
	"io/ioutil"
	"context"
	etcd "github.com/coreos/etcd/client"

)

var configreads		map[string]ReadFunc

func init() {
	configreads = make(map[string]ReadFunc)
	AddReadFunc("default", ReadFile)
	AddReadFunc("file", ReadFile)
	AddReadFunc("https", ReadHttp)
	AddReadFunc("http", ReadHttp)
	AddReadFunc("etcd", ReadEtcd)
}

func AddReadFunc(name string, fn ReadFunc) {
	configreads[name] = fn
}


// Read config data and type to Config
func ParseRead(c *Config) error {
	path := c.GetString("#config")
	if path == "" {
		return fmt.Errorf("config data is null")
	}
	// read protocol
	// get read func
	s := strings.SplitN(path, "://", 2)
	fn, ok := configreads[s[0]]
	if !ok {
		// use default read func
		fmt.Println("undefined read config: " + path + ", use default.")
		fn = configreads["default"]
	}
	data, err := fn( path)
	c.SetKey("config", data)
	return err
}

// Read config file
func ReadFile(path string) (string, error) {
	if strings.HasPrefix(path, "file://") {
		path = path[7:]
	}
	data, err := ioutil.ReadFile(path)
	
	last := strings.LastIndex(path, ".") + 1
	if last == 0 {
		return "", fmt.Errorf("read file config, type is null")
	}
	return string(data), err
}

// Send http request get config info
func ReadHttp(path string) (string, error) {
	resp, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}




//
// example: etcd://127.0.0.1:2379/config
func ReadEtcd(path string) (string, error) {
	server, key := split2(path[7:], "/")
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
}