package test

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/kr/pretty"
	"testing"
)

type (
	Config struct {
		Keys    map[string]string      `set:"keys"`
		Modules map[string]interface{} `set:"modules"`
		Handler map[string]interface{} `set:"handler"`
	}

	File struct {
		Dir      string   `set:"dir"`
		Secury   string   `set:"secury"`
		Endpoint []string `set:"endpoint"`
	}
)

func TestKeys(*testing.T) {
	c, _ := eudore.NewConfigEudore(&Config{
		Keys:    make(map[string]string),
		Modules: make(map[string]interface{}),
		Handler: make(map[string]interface{}),
	})
	fmt.Println(c.Set("keys.1", 1))
	fmt.Println(c.Set("keys.2", "2"))
	f := &File{}
	fmt.Println(c.Set("handler.file", f))
	fmt.Println(c.Set("handler.file.dir", "/tmp"))
	fmt.Println(c.Set("handler.file.endpoint", []string{"hk", "sh"}), f)
	fmt.Printf("struct: %# v\n", pretty.Formatter(c))
}
