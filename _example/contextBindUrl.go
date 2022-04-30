package main

/*
默认只有Get和Head请求时，使用url参数绑定参数。

BinderURLWithBinder函数可以时其他方法也使用url参数绑定。

当body是使用url编码时，也可以直接绑定数据。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	urlPutFileInfo struct {
		Name         string   `json:"name" alias:"name"`
		Type         string   `json:"type" alias:"type"`
		Size         int      `json:"size" alias:"size"`
		LastModified int64    `json:"lastModified" alias:"lastModified"`
		Tags         []string `json:"tags" alias:"tags"`
	}
)

func main() {
	app := eudore.NewApp()
	// 附加非GET和HEAD方法下使用url参数绑定。
	app.SetValue(eudore.ContextKeyBind, eudore.NewBindWithURL(eudore.NewBinds(nil)))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AnyFunc("/file/data/:path", func(ctx eudore.Context) {
		var info urlPutFileInfo
		ctx.Bind(&info)
		ctx.Debugf("%#v", info)
	})
	app.GetFunc("/binderr", func(ctx eudore.Context) {
		// 设置测试数据
		ctx.Request().URL.RawQuery = "%gh&%ij"

		var info urlPutFileInfo
		ctx.Bind(&info)
		ctx.Debugf("%#v", info)
	})

	client := httptest.NewClient(app)
	// get方法使用url参数绑定
	client.NewRequest("GET", "/file/data/2?name=eudore&type=2&type=3&size=722&tags=t1&tags=t2").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()
	// put方法使用url参数绑定，需要BinderURLWithBinder函数支持。
	client.NewRequest("PUT", "/file/data/2?name=eudore&type=2&size=722").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()
	// url error
	client.NewRequest("PUT", "/file/data/2?%gh&%ij").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()
	client.NewRequest("PUT", "/file/data/2").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).WithBodyString("%gh&%ij").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/binderr").Do().CheckStatus(200).Out()
	// url body绑定
	client.NewRequest("PUT", "/file/data/2").WithBodyString("name=eudore&type=2&size=722").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
