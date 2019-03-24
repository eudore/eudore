package eudore

import (
	"io"
	"fmt"
	"html/template"
)


type (
	View interface {
		Component
		AddFunc(string, interface{}) View
		ExecuteTemplate(wr io.Writer, name string, data interface{}) error
	}
	ViewStdConfig struct {
		// Basetemp	[]string
		Tempdir		string
	}
	// 基于html/template封装，目前未测试。
	ViewStd struct {
		*ViewStdConfig
		root		*template.Template
		funcs		template.FuncMap
		// templates	map[string]*template.Template
	}
)

func NewView(name string, arg interface{}) (View, error) {
	name = AddComponetPre(name, "view")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	l, ok := c.(View)
	if ok {
		return l, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to View type", name)
}

func NewViewStd(interface{}) (View, error) {
	return &ViewStd{
		ViewStdConfig: &ViewStdConfig{
			// Basetemp:	[]string{"", "/data/web/templates/base.html"},
			Tempdir:	"/data/web/templates/",	
		},
		root:		template.New(""),
		funcs:		make(template.FuncMap),
		// templates:	make(map[string]*template.Template),
	}, nil
}

// 注册一个模板函数给视图，需要在第一次渲染模板选注册。
func (v *ViewStd) AddFunc(name string, fn interface{}) View {
	v.funcs[name] = fn
	return v
}

// 对指定模板进行渲染。
//
// 同一模板第一次渲染时会加载全部模板函数。
//
// 所有模板可以引用根模板的内容。
func (v *ViewStd) ExecuteTemplate(wr io.Writer, path string, data interface{}) (err error) {
	t := v.root.Lookup(path)
	if t == nil {
		t, err = v.loadTemplate(path)
		if err != nil {
			return
		}
	}
	return t.Execute(wr, data)
}

// 给根模板加载一个子模板。
func (v *ViewStd) loadTemplate(path string) (*template.Template, error) {
	t, err := template.New(path).Funcs(v.funcs).ParseFiles(path)
	if err != nil {
		return nil, err
	}
	return v.root.AddParseTree(path, t.Tree)
}

func (v *ViewStd) Set(key string, val interface{}) error {
	return nil
}

func (*ViewStdConfig) GetName() string {
	return ComponentViewStdName
}

func (*ViewStdConfig) Version() string {
	return ComponentViewStdVersion
}