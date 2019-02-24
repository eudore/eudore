package config


import (
	"os"
	"fmt"
	"strings"
	"encoding/json"
)



// Read parameters in advance
var prereaddata = []string{
	"--config.path=",
	"--config.data=",
	"--config.type=",
	"--workdir=",
}

func ParseStart(c *Config) (err error) {
	fmt.Println("-----------------end")
	c.SetKey("#eudore", "eudore")
	c.SetKey("#workdir", "#eudore.workdir")
	c.SetKey("#command", "#eudore.command")
	c.SetKey("#pidfile", "#eudore.pidfile")

	c.SetKey("#config", "#eudore.config")
	c.SetKey("#help", "#eudore.help")
	c.SetKey("#test", "#eudore.test")

	c.SetKey("#component", "component")
	c.SetKey("#logger", "#component.logger")

	c.SetKey("#keys", "keys")
	return nil
}
/*
// Read Config and workdir in advance
	filter := func(str string) string {
		for _, p := range prereaddata {
			if strings.HasPrefix(str, p) {
				return str
			}
		}
		return ""
	}
	for _, arg := range eachstring(c.Args, filter) {
		err = ConfigSetData(c, arg)
		if err != nil {
			return
		}
	}
	for _, arg := range eachstring(c.Envs, filter) {
		err = ConfigSetData(c, arg)
		if err != nil {
			return
		}
	}
	os.Chdir(c.WorkDir)
	return
*/

// Parse config data to Config
func ParseConfig(c *Config) error {
	err := json.Unmarshal([]byte(c.GetKey("config")), c.Global)
	//Json(string(c.Config.Data), c)
	return err	
}

// Load Config use modes
func ParseModes(*Config) error {
	return nil
}

// Use Args Config
func ParseArgs(c *Config) (err error) {
	for _, v := range os.Args[1:] {
		if !strings.HasPrefix(v, "--") {
			fmt.Println("invalid args", v)
			continue
		}
		err = c.SetData(split2(v[2:], "="))
		if err != nil {
			fmt.Println("error:",err,v)
			return
		}
	}
	return
}

// Use Envs Config
func ParseEnvs(c *Config) (err error) {
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "ENV_") {
			k, v := split2(value, "=")
			k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
			err = c.SetData(k, v)
			if err != nil {
				return
			}
		}
	}
	return nil
}


func ParseEnd(c *Config) (err error) {
	if c.GetBool("#help", false) {
		c.Help()
	}
	if c.GetBool("#test", false) {
		Json(c.Global)
		os.Exit(0)
	}
	return nil
}
