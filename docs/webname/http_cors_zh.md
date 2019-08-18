# CORS

跨域资源共享(Cross-Origin Resource Sharing)

阅览器发送跨域请求时，先回检查是否为简单请求，是简单请求就直接访问；

简单请求要求请求方法为HEAD、GET、POST之一，请求headers只有Accept、Accept-Language、Content-Language、Last-Event-ID、Content-Type，没有额外的header，且Content-Type的值为application/x-www-form-urlencoded、multipart/form-data、text/plain三种之一。

如果不是简单请求，就会自动发送一个同路径的Options请求，**必须包含Origin和Access-Control-Request-Headers两个个header**；

在Response中**必须包含Access-Control-Allow-Origin和Access-Control-Allow-Methods两个header**。

如果需要发送请求不满足响应允许请求限制，就不会发送实际请求并阅览器报错。

如果满足了会发送跨越请求，请求包好Origin Header，服务器会检查Origin是否允许并返回Access-Control-Allow-Origin header。

# Header

|  header |  类型 |  描述 |
| ------------ | ------------ | ------------ |
| Access-Control-Allow-Credentials | Response header | 允许浏览器读取response的认证内容|
| Access-Control-Allow-Headers | Response header | 实际请求中允许携带的首部字段|
| Access-Control-Allow-Methods | Response header | 实际请求所允许使用的 HTTP 方法|
| Access-Control-Allow-Origin | Response header | 指定允许请求的源站 |
| Access-Control-Expose-Headers | Response header | 允许客户端读取额外的headers|
| Access-Control-Max-Age | Response header | preflight请求的结果能够被缓存多久|
| Origin | Request header | 表明预检请求或实际请求的源站 |
| Access-Control-Request-Headers | Request header | 实际请求需要附加的header |
| Access-Control-Request-Method | Request header | 实际请求的方法 |

## Access-Control-Allow-Credentials

可选，表示是否允许发送Cookie。

## Access-Control-Allow-Headers

可选，为了响应Access-Control-Request-Headers

## Access-Control-Allow-Methods

**必要**，返回跨域整个站允许的方法，不是单个请求允许的方法。

## Access-Control-Allow-Origin

**必要**，返回允许的源站，一般值为'*'或者Origin的值。

## Access-Control-Expose-Headers

可选，跨域请求完成后，如果客户端读取客户端的header，除了Cache-Control、Content-Language、Content-Type、Expires、Last-Modified、Pragma六个基本Header，其他Header无法读取到，需要在Access-Control-Expose-Headers中指定才能读取到。

## Access-Control-Max-Age

可选，指定跨域验证缓存时间，不用每次跨域都要options验证，减少服务端压力。

在Firefox中，上限是24小时 （即86400秒），而在Chromium 中则是10分钟（即600秒）。Chromium 同时规定了一个默认值 5 秒。如果值为 -1，则表示禁用缓存，每一次请求都需要提供预检请求，即用OPTIONS请求进行检测。

## Origin

**必要**，每次跨域请求都会附带，指定跨域源站。

## Access-Control-Request-Headers

可选，如果发送跨域请求时，添加了其他header，再发送options时，就会发送Access-Control-Request-Headers，值为其他header，服务端会验证其中允许的header，并使用Access-Control-Allow-Headers返回

## Access-Control-Request-Method

**必要**指定options验证的请求的请求方法。