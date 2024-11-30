package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	filepath "path"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewLoggerFunc(app),
		middleware.NewBodyLimitFunc(1<<20), // 1MB
	)
	up := app.Group("")
	up.AddMiddleware(handleName)
	up.PostFunc("/upload/body/:name", handleUploadBody)
	up.PostFunc("/upload/form/:name", handleUploadForm2)
	up.PostFunc("/upload/multi/:name/:part", handleUploadMultiPart)
	up.PutFunc("/upload/multi/:name", handleUploadMultiDone)

	app.GetFunc("/", handleUi)

	app.Listen(":8088")
	app.Run()
}

// 提前校验一下Name防止攻击。
func handleName(ctx eudore.Context) {
	name := filepath.Clean(ctx.GetParam("name"))
	if name == "." {
		ctx.Fatalf("invliad name '%s'", ctx.GetParam("name"))
		return
	}
	ctx.SetParam("name", name)
}

func handleUploadBody(ctx eudore.Context) error {
	file, err := os.OpenFile(ctx.GetParam("name"), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// 不要使用io.ReadAllz在Write，这样会将全部文件读入到内存中，使用io.Copy流式复制减少内存使用和GC压力。
	_, err = io.Copy(file, ctx)
	return err
}

func handleUploadForm(ctx eudore.Context) error {
	head := ctx.FormFile("file")
	if head == nil {
		return fmt.Errorf("not found multipart file")
	}
	read, err := head.Open()
	if err != nil {
		return err
	}

	name := ctx.GetParam("name")
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, read)
	return err
}

func handleUploadForm2(ctx eudore.Context) error {
	type Up struct {
		Name string                `alias:"name" json:"name"`
		File *multipart.FileHeader `alias:"file" json:"-"`
	}

	var up Up
	err := ctx.Bind(&up)
	if err != nil {
		return err
	}

	read, err := up.File.Open()
	if err != nil {
		return err
	}

	// up.Name 来源于Body没有进行校验
	file, err := os.OpenFile(up.Name, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, read)
	return err
}

func handleUploadMultiPart(ctx eudore.Context) error {
	os.Mkdir("chunk", 0644)
	path := fmt.Sprintf("chunk/%s.%s", ctx.GetParam("name"), ctx.GetParam("part"))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, ctx)
	if err != nil {
		os.Remove(file.Name())
	}
	return nil
}

func handleUploadMultiDone(ctx eudore.Context) error {
	name := ctx.GetParam("name")
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	total := eudore.GetAnyByString[int](ctx.GetQuery("total"))
	// 使用MultiWriter流式计算Hash
	hash := md5.New()
	w := io.MultiWriter(file, hash)
	for i := 0; i < total; i++ {
		chunk, err := os.Open(fmt.Sprintf("chunk/%s.%d", name, i))
		if err != nil {
			os.Remove(file.Name())
			return err
		}
		io.Copy(w, chunk)
		chunk.Close()
	}
	for i := 0; i < total; i++ {
		os.Remove(fmt.Sprintf("chunk/%s.%d", name, i))
	}

	stat, _ := file.Stat()
	return ctx.Render(map[string]any{
		"name":  name,
		"hash":  fmt.Sprintf("md5:%x", hash.Sum(nil)),
		"size":  stat.Size(),
		"total": total,
	})
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
<button onclick="upload('multi')">multi</button>
<script>
	async function upload(t) {
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
		}else if (t=="form") {
			var data = new FormData();
			data.append('name', file.name);
			data.append('file', file);
			fetch('/upload/form/'+file.name, {method: 'post', body: data}).then((data) => {
				console.log(data);
			});			
		}else if (t=="multi"){
			// 分段4M或强制4段。
			const chunksize = file.size> (4<<20) ? (1<<20): Math.ceil(file.size/4)
			const chunktotal = Math.ceil(file.size / chunksize);
			for (let i = 0; i < chunktotal; i++) {
				const start = i * chunksize;
				const end = Math.min(start + chunksize, file.size);
				const chunk = file.slice(start, end); 

				await fetch('/upload/multi/'+file.name+'/'+i, {
					method: 'POST',
					body: chunk,
					headers: {'Content-Type': 'application/octet-stream'}
				});
			}
			await fetch('/upload/multi/'+file.name+'?total='+chunktotal, {method: 'PUT'});
		}
	}
</script>
</body>
</html>`)
}
