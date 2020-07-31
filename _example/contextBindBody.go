package main

/*
bind根据请求中Content-Type Header来决定bind解析数据的方法，常用json和form两种。

例如存在Request Header Content-Type: application/json，Bind就会使用Json解析。

如果请求方法是Get或Head，使用Uri参数绑定。
*/

import (
	"encoding/xml"
	"errors"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"io"
)

type (
	putFileInfo struct {
		Name         string `json:"name" alias:"name"`
		Type         string `json:"type" alias:"type"`
		Size         int    `json:"size" alias:"size"`
		LastModified int64  `json:"lastModified" alias:"lastModified"`
	}
	Email struct {
		Where string `xml:"where,attr"`
		Addr  string
	}
	xmlResult struct {
		XMLName     xml.Name `xml:"Person"`
		Name        string   `xml:"FullName"`
		Phone       string
		Email       []Email
		Groups      []string `xml:"Group>Value"`
		City, State string
	}
)

func main() {
	app := eudore.NewApp()
	// 上传文件信息
	app.PutFunc("/file/data/:path", func(ctx eudore.Context) {
		var info putFileInfo
		ctx.Bind(&info)
		ctx.RenderWith(&info, eudore.RenderIndentJSON)
	})
	app.PutFunc("/person/:path", func(ctx eudore.Context) {
		var person xmlResult
		ctx.Bind(&person)
		ctx.Debugf("%#v", person)
	})
	app.PutFunc("/body", func(ctx eudore.Context) {
		ctx.Debugf("body: %s", ctx.Body())
		var data map[string]interface{}
		ctx.Bind(&data)
	})
	app.PutFunc("/with", func(ctx eudore.Context) {
		ctx.Debugf("body: %s", ctx.Body())
		var data map[string]interface{}
		ctx.BindWith(&data, func(eudore.Context, io.Reader, interface{}) error {
			return errors.New("test bind with error")
		})
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/file/data/2").WithBodyString(`{"name": "eudore","type": "file", "size":720,"lastModified":1257894000}`).Do()
	client.NewRequest("PUT", "/file/data/2").WithHeaderValue("Content-Type", "application/json").WithBodyString(`{"name": "eudore","type": "file", "size":720,"lastModified":1257894000}`).Do().CheckStatus(200).Out()
	client.NewRequest("PUT", "/person/2").WithHeaderValue("Content-Type", "application/xml").WithBodyString(`
		<Person>
			<FullName>Grace R. Emlin</FullName>
			<Company>Example Inc.</Company>
			<Email where="home">
				<Addr>gre@example.com</Addr>
			</Email>
			<Email where='work'>
				<Addr>gre@work.com</Addr>
			</Email>
			<Group>
				<Value>Friends</Value>
				<Value>Squash</Value>
			</Group>
			<City>Hanga Roa</City>
			<State>Easter Island</State>
		</Person>
	`).Do().CheckStatus(200).Out()
	client.NewRequest("PUT", "/body").WithBodyString(`{"name": "eudore","type": "file", "size":720,"lastModified":1257894000}`).Do()
	client.NewRequest("PUT", "/with").WithBodyString(`{"name": "eudore","type": "file", "size":720,"lastModified":1257894000}`).Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
