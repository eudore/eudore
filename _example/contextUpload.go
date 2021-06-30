package main

import (
	"fmt"
	"io"
	"os"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.PostFunc("/upload/body/:name", handleUploadBody)
	app.PostFunc("/upload/form/:name", handleUploadForm)
	app.PostFunc("/upload/multi/:name", handleUploadMulti)
	app.GetFunc("/", handleUi)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func handleUploadBody(ctx eudore.Context) error {
	file, err := os.OpenFile(ctx.GetParam("name"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, ctx)
	return err
}

func handleUploadForm(ctx eudore.Context) error {
	file, err := os.OpenFile(ctx.GetParam("name"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	head := ctx.FormFile("file")
	if head == nil {
		return fmt.Errorf("not found multipart file")
	}
	read, err := head.Open()
	if err != nil {
		return err
	}

	_, err = io.Copy(file, read)
	return err
}

func handleUploadMulti(ctx eudore.Context) error {
	return nil
}

func handleUi(ctx eudore.Context) {
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8")
	ctx.WriteString(`<!doctype html>
<html>
<head>
	<meta charset="utf-8">
	<title>eudore文件上传</title>
</head>
<body>
<input type="file" id="files"/>
<button onclick="upload('body')">body</button>
<button onclick="upload('form')">form</button>
<script>
	function upload(t) {
		var file = document.getElementById('files').files[0]
		if(file == undefined){
			return
		}
		if (t=="body") {
			fetch('/upload/body/'+file.name, {method: 'post', body: file, headers: {
				'Content-Type': 'application/octet-stream'
			}}).then((data) => {
				console.log(data);
			});	
		}else if (t=="form"){
			var data = new FormData();
			data.append('file', file);
			fetch('/upload/form/'+file.name, {method: 'post', body: data}).then((data) => {
				console.log(data);
			});			
		}
	}
</script>
</body>
</html>`)
}
