# post data

通常post表单使用url和form两种编码格式。

## application/x-www-form-urlencoded

这种就键值对数据，使用url编码

## multipart/form-data

form表单使用`mime/multipart/`库来解析

```golang
	_, params, err := mime.ParseMediaType(r.Header.Get(HeaderContentType))
	if err != nil {
		return err
	}

	form, err := multipart.NewReader(r, params["boundary"]).ReadForm(32 << 20)
	if err != nil {
		return err
	}

	form...
```

form对象就是解析到的数据。

form对象定义在[`mime/multipart`库](https://golang.org/pkg/mime/multipart/#Form),value就是键值对数据，File是上form中上传的临时文件。

form会把上传的文件存到一个临时目录，最后删除掉。

```golang
type Form struct {
	Value map[string][]string
	File  map[string][]*FileHeader
}
```