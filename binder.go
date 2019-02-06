package eudore


import (
	"fmt"
	"encoding/json"
	"encoding/xml"
)

var (
	BinderDefault	=	BindFunc(BinderDefaultFunc)
	BinderForm		=	BindFunc(BindFormFunc)
	BinderUrl		=	BindFunc(BindUrlFunc)
	BinderJSON		=	BindFunc(BindJsonFunc)
	BinderXML		=	BindFunc(BindXmlFunc)
)


type (
	Binder interface {
		Bind(RequestReader, interface{}) error
	}
	BindFunc func(RequestReader, interface{}) error
)

func (fn BindFunc) Bind(r RequestReader, i interface{}) error {
	return fn(r, i)
}

func BinderDefaultFunc(r RequestReader, i interface{}) error {
	switch r.Header().Get("Content-Type") {
	case MimeApplicationJson:
		return BinderJSON.Bind(r, i)
	case MimeTextXml, MimeApplicationXml:
		return BinderXML.Bind(r, i)
	case MimeApplicationForm, MimeMultipartForm:
		return BinderForm.Bind(r, i)
	default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
		return BinderForm.Bind(r, i)
	}
}

func BindFormFunc(r RequestReader, i interface{}) error {
	fmt.Println("未完成: binder.FormBinder.Bind")
	return nil
}

func BindUrlFunc(r RequestReader, i interface{}) error {
	return nil
}

func BindJsonFunc(r RequestReader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

func BindXmlFunc(r RequestReader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}