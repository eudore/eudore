# CORS

跨域资源共享(CORS)

阅览器发送跨域请求时，先回检查是否为简单请求，是简单请求就直接访问；

如果不是简单请求，就会自动发送一个同路径的Options请求，包含Origin、Access-Control-Request-Headers和Access-Control-Request-Method三个header；

在Response header中Access-Control-Allow-Headers指定了允许发送额外的header，Access-Control-Allow-Methods指定运行的方法，Access-Control-Allow-Origin指定允许的源站；

如果需要发送请求不满足响应允许请求限制，就不会发送实际请求并阅览器报错。


简单请求要求请求方法为HEAD、GET、POST之一，请求headers只有Accept、Accept-Language、Content-Language、Last-Event-ID、Content-Type，没有额外的header，且Content-Type的值为application/x-www-form-urlencoded、multipart/form-data、text/plain三种之一。

|  header |  类型 |  描述 |
| ------------ | ------------ | ------------ |
| Access-Control-Allow-Credentials | Response header | 否允许浏览器读取response的内容|
| Access-Control-Allow-Headers | Response header | 实际请求中允许携带的首部字段|
| Access-Control-Allow-Methods | Response header | 实际请求所允许使用的 HTTP 方法|
| Access-Control-Allow-Origin | Response header | 指定允许请求的源站 |
| Access-Control-Expose-Headers | Response header | 允许客户端读取额外的headers|
| Access-Control-Max-Age | Response header | preflight请求的结果能够被缓存多久|
| Origin | Request header | 表明预检请求或实际请求的源站 |
| Access-Control-Request-Headers | Request header | 实际请求需要附加的header |
| Access-Control-Request-Method | Request header | 实际请求的方法 |
