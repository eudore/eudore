package eudore

import (
	"os"
	"fmt"
	"strings"
	"encoding/json"
)


func ParseInitData(c Config) error {
	c.Set("config", "file:///data/web/golang/src/wejass/config/config-eudore.json")
	c.Set("command", "start")
	c.Set("pidfile", "/var/run/eudore.pid")
	return nil
}

func ParseRead(c Config) error {
	path := c.Get("#config").(string)
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
	data, err := fn(path)
	c.Set("configdata", data)
	return err
}

func ParseConfig(c Config) error {
	err := json.Unmarshal([]byte(c.Get("configdata").(string)), c.Get(""))
	//Json(string(c.Config.Data), c)
	return err	
}

func ParseArgs(c Config) (err error) {
	for _, str := range os.Args[1:] {
		if !strings.HasPrefix(str, "--") {
			// fmt.Println("invalid args", str)
			continue
		}
		c.Set(split2byte(str[2:], '='))
	}
	return
}


func ParseEnvs(c Config) error {
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "ENV_") {
			k, v := split2byte(value, '=')
			k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
			c.Set(k, v)
		}
	}
	return nil
}