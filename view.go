package eudore

import (
	"io"
	"fmt"
	"html/template"
)


type (
	View interface {
		Component
		ExecuteTemplate(wr io.Writer, name string, data interface{}) error
	}
	ViewStdConfig struct {
		Basetemp	[]string
		Tempdir		string
	}
	ViewStd struct {
		*ViewStdConfig
		templates	map[string]*template.Template
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
			Basetemp:	[]string{"", "/data/web/templates/base.html"},
			Tempdir:	"/data/web/templates/",	
		},
		templates:	make(map[string]*template.Template),
	}, nil
}

func (v *ViewStd) ExecuteTemplate(wr io.Writer, name string, data interface{}) (err error) {
	tmp, ok := v.templates[name]
	if !ok {
		v.Basetemp[0] = v.Tempdir + name
		tmp, err = template.ParseFiles(v.Basetemp...)
		if err != nil {
			return
		}
		// No cache, save new template
		// templates[name] = tmp
	}
	return tmp.Execute(wr, data)
}

func (*ViewStdConfig) GetName() string {
	return ComponentViewStdName
}

func (*ViewStdConfig) Version() string {
	return ComponentViewStdVersion
}