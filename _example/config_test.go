package eudore_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/eudore/eudore"
)

func TestConfigStdGetSet(t *testing.T) {
	c := NewConfig(map[string]interface{}{
		"name":   "eudore",
		"type":   "ConfigMap",
		"number": 3,
	})
	c.Parse(context.Background())
	c.Set("auth.secret", "secret")
	t.Logf("data: %# v", c.Get(""))
	t.Logf("data name: %v", c.Get("name"))

	type Config struct {
		Name string `alias:"name"`
		Type string `alias:"type"`
	}
	c.Set("", &Config{Name: "eudore"})
	c.Set("type", "config")
	t.Logf("data name: %v", c.Get("name"))

	c.(interface{ Metadata() any }).Metadata()
}

func TestConfigStdpParse(t *testing.T) {
	c := NewConfig(nil)
	c.ParseOption(func(ctx context.Context, config Config) error {
		config.Set("parse", true)
		return nil
	})
	t.Logf("parse eror: %v", c.Parse(context.Background()))
	t.Logf("data: %# v", c.Get(""))

	c.ParseOption()
	c.ParseOption(func(ctx context.Context, config Config) error {
		config.Set("error", true)
		return errors.New("parse test error")
	})
	t.Logf("parse eror: %v", c.Parse(context.Background()))
	t.Logf("parse eror: %v", c.Parse(context.Background()))
	t.Logf("data: %# v", c.Get(""))
}

func TestConfigParseJSON(t *testing.T) {
	filepath1 := "tmp-config1.json"
	defer tempConfigFile(filepath1, `{"help":true,"workdir":".","name":"eudore"}`)()
	filepath2 := "tmp-config2.json"
	defer tempConfigFile(filepath2, `name:eudore`)()

	c := NewConfig(nil)
	c.ParseOption()
	c.ParseOption(
		NewConfigParseJSON("config"),
		func(ctx context.Context, config Config) error {
			return json.Unmarshal([]byte(`{"name":"eudore"}`), config)
		},
	)
	body, err := json.Marshal(c)
	t.Logf("ConfigMap json data: %s,error: %v", body, err)

	t.Logf("NewConfigParseJSON parse empty error %v:", c.Parse(context.Background()))

	c.Set("config", filepath1)
	t.Logf("NewConfigParseJSON parse file error: %v", c.Parse(context.Background()))

	c.Set("config", []string{filepath1})
	t.Logf("NewConfigParseJSON parse mutil file error: %v", c.Parse(context.Background()))

	c.Set("config", "not-"+filepath1)
	t.Logf("NewConfigParseJSON parse not file error: %v", c.Parse(context.Background()))

	os.Setenv("ENV_CONFIG", "path")
	t.Logf("NewConfigParseJSON parse env error: %v", c.Parse(context.Background()))
	os.Unsetenv("ENV_CONFIG")

	os.Args = append(os.Args, "--config=path")
	t.Logf("NewConfigParseJSON parse args error: %v", c.Parse(context.Background()))
	os.Args = os.Args[:len(os.Args)-1]

	c.ParseOption()
	c.ParseOption(NewConfigParseDecoder("config", "json-custom",
		func(reader io.Reader, data any) error {
			return json.NewDecoder(reader).Decode(data)
		},
	))
	c.Set("config", filepath2)
	t.Logf("NewConfigParseJSON parse decoder error: %v", c.Parse(context.Background()))

	path := "config" + strings.Repeat("-", 256) + ".json"
	c.ParseOption()
	c.ParseOption(NewConfigParseJSON("config"))
	c.Set("config", path)
	t.Logf("NewConfigParseJSON parse open error: %v", c.Parse(context.Background()))
}

func tempConfigFile(path, content string) func() {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	file.Write([]byte(content))
	file.Close()
	return func() {
		os.Remove(file.Name())
	}
}

type Config020 struct {
	Workdir   string
	Name      string  `alias:"name"`
	Namespace *string `alias:"namespace" flag:"n"`
	Body      []byte  `alias:"body" flag:"b"`
	Slices    []int   `alias:"slices" flag:"s"`
	Any       any
}

func TestConfigParseArgs(t *testing.T) {
	os.Args = append(os.Args, "start", "--name=eudore")
	defer func() {
		os.Args = os.Args[:len(os.Args)-2]
	}()
	conf := &Config020{}
	conf.Any = conf

	c := NewConfig(conf)
	c.ParseOption()
	c.ParseOption(NewConfigParseArgs())

	t.Logf("NewConfigParseArgs parse error: %v", c.Parse(context.Background()))
	t.Logf("Config data: %# v", c.Get(""))
}

func TestConfigParseArgsShort(t *testing.T) {
	type configShort struct {
		Help   bool   `alias:"help" json:"help" flag:"h"`
		Config string `alias:"config" json:"config" flag:"c"`
		Name   string `alias:"name" json:"name"`
	}
	os.Args = append(os.Args, "--name=eudore", "-f=config.json", "-h", "--help")

	c := NewConfig(&configShort{false, "eudore", "msg"})
	c.ParseOption()
	c.ParseOption(NewConfigParseArgs())

	t.Logf("NewConfigParseArgs parse error: %v", c.Parse(context.Background()))
	t.Logf("Config data: %# v", c.Get(""))
}

func TestConfigParseEnvs(t *testing.T) {
	os.Setenv("ENV_NAME", "eudore")
	defer os.Unsetenv("ENV_NAME")
	// init envs by cmd
	c := NewConfig(nil)
	c.ParseOption()
	c.ParseOption(NewConfigParseEnvs("ENV_"))

	t.Logf("NewConfigParseEnvs parse error: %v", c.Parse(context.Background()))
	t.Logf("Config data: %# v", c.Get(""))
}

func TestConfigParseEnvsFile(t *testing.T) {
	defer tempConfigFile(".env", "A=2\r\nB=\r\nC='2\r\n2\r\n2'\r\nD='2\r\n2\r\n2'2\r\n")()
	defer os.Unsetenv("ENV_NAME")
	// init envs by cmd

	path := "config" + strings.Repeat("-", 256) + ".json"
	c := NewConfig(nil)

	c.ParseOption()
	c.ParseOption(NewConfigParseEnvFile(".env", path))

	t.Logf("NewConfigParseEnvFile parse error: %v", c.Parse(context.Background()))
	t.Logf("Config data: %# v", c.Get(""))

	c.ParseOption()
	c.ParseOption(NewConfigParseEnvFile())
	c.Parse(context.Background())
}

func TestConfigParseWorkdir(t *testing.T) {
	c := NewConfig(nil)
	c.ParseOption()
	c.ParseOption(NewConfigParseWorkdir("workdir"))

	t.Logf("NewConfigParseWorkdir parse empty dir error: %v", c.Parse(context.Background()))

	c.Set("workdir", ".")
	t.Logf("NewConfigParseWorkdir parse error: %v", c.Parse(context.Background()))
	t.Logf("Config data: %# v", c.Get(""))
}
