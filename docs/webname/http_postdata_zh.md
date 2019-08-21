# post data

使用Post方法发送一个请求，通过有三种传递数据的方法，不算Header，uri、url表单、form表单，后两种就是指一般表单请求。

## uri

**uri方式的数据在请的uri中**，例如下面这个请求的数据就是a=1&b=2，在请求行中，是**url编码**。

```
[root@izj6cffbpd9lzl3tcm2csxz ~]# curl -v -XPOST 'www.example.com?a=1&b=2'
* About to connect() to www.example.com port 80 (#0)
*   Trying 93.184.216.34...
* Connected to www.example.com (93.184.216.34) port 80 (#0)
> POST /?a=1&b=2 HTTP/1.1
> User-Agent: curl/7.29.0
> Host: www.example.com
> Accept: */*
> 
< HTTP/1.1 411 Length Required
< Content-Type: text/html
< Content-Length: 357
< Connection: close
< Date: Tue, 16 Jul 2019 09:31:13 GMT
< Server: ECSF (sjc/4E8D)
```

## application/x-www-form-urlencoded

这种就键值对数据，**使用url编码**,使用curl发送请求，可以看见显示`Content-Length`的长度是7，**`Content-Type`显示是url编码**，这种是url编码表单，数据在在body里面，内容没有显示，但是可以根据数据猜出请求body内容就是`a=1&b=2`，长度恰好是7。

```
[root@izj6cffbpd9lzl3tcm2csxz ~]# curl -v -d a=1 -d b=2 www.example.com
* About to connect() to www.example.com port 80 (#0)
*   Trying 93.184.216.34...
* Connected to www.example.com (93.184.216.34) port 80 (#0)
> POST / HTTP/1.1
> User-Agent: curl/7.29.0
> Host: www.example.com
> Accept: */*
> Content-Length: 7
> Content-Type: application/x-www-form-urlencoded
> 
* upload completely sent off: 7 out of 7 bytes
< HTTP/1.1 200 OK
< Accept-Ranges: bytes
< Cache-Control: max-age=604800
< Content-Type: text/html; charset=UTF-8
< Date: Tue, 16 Jul 2019 09:25:44 GMT
< Etag: "1541025663"
< Expires: Tue, 23 Jul 2019 09:25:44 GMT
< Last-Modified: Fri, 09 Aug 2013 23:54:35 GMT
< Server: EOS (vny006/044F)
< Content-Length: 1270
```

## multipart/form-data

这个是curl发送一个form表单的内容，可以看到`Content-Type`的值是`multipart/form-data; boundary=----------------------------e55eb170f828`,里面boundary的值相当于表单的分隔符，用于解析body。

```
[root@izj6cffbpd9lzl3tcm2csxz ~]# curl -v --form a=1 --form b=2  www.example.com
* About to connect() to www.example.com port 80 (#0)
*   Trying 93.184.216.34...
* Connected to www.example.com (93.184.216.34) port 80 (#0)
> POST / HTTP/1.1
> User-Agent: curl/7.29.0
> Host: www.example.com
> Accept: */*
> Content-Length: 228
> Expect: 100-continue
> Content-Type: multipart/form-data; boundary=----------------------------e55eb170f828
> 
< HTTP/1.1 100 Continue
< HTTP/1.1 200 OK
< Accept-Ranges: bytes
< Cache-Control: max-age=604800
< Content-Type: text/html; charset=UTF-8
< Date: Tue, 16 Jul 2019 09:24:46 GMT
< Etag: "1541025663"
< Expires: Tue, 23 Jul 2019 09:24:46 GMT
< Last-Modified: Fri, 09 Aug 2013 23:54:35 GMT
< Server: EOS (vny006/044E)
< Content-Length: 1270
```

### golang解析form表单

通常net/http中Request会解析数据。

form表单也可以使用`mime/multipart`库来解析。

```golang
	_, params, err := mime.ParseMediaType(r.Header.Get(HeaderContentType))
	if err != nil {
		return err
	}

	form, err := multipart.NewReader(r, params["boundary"]).ReadForm(32 << 20)
	if err != nil {
		return err
	}
	defer form.RemoveAll()

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