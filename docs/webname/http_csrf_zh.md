# CSRF

跨站请求伪造(Cross-site request forgery)

挟制用户在当前已登录的Web应用程序上执行非本意的操作的攻击方法。

## 实现

1、用户在A站登陆，浏览器记录Cookie等信息。

2、用户在B站触发了一个A站的请求，请求发送时附带了A站Cookie，访问控制通过，成功请求未知请求。

## 防范 

CSRF两点必要情况，跨域和使用Cookie。

两者缺一不可，可以通过这两点来防范。

### 同源检测

http request header中，可以使用Origin Header和Referer Header来知道发送请求的来源。

在服务端检查请求来源页面，将非法来源的请求直接拒绝，但是对外链这样的无法处理。

http请求发送Referer Header，websocket建立连接请求发送Origin Header。

https默认不发送Referer Header，需要在html head中添加meta标签来发送Referer，如下meta。

`<meta name="referrer" content="always">`

但在某些情况下可以隐藏或者丢失Referer，并不可靠

### CSRF Token

利用第三方网站无法跨越读取cookie。

在请求中添加csrdid，cookie中也有csrfid，在服务端检查csrdid是否一致。

### 双重Cookie验证

类似CSRFToken，会将一个cookie值将到请求中。

### SameSite

SameSite是跨域cookie发送规则。

在set-cookie时，添加SameSite=Strict，静止跨域发送cookie。
