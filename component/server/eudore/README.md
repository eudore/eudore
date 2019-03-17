# eudore server

## 介绍

基于`github.com/eudore/eudore/protocol`包封装实现的eudore框架服务端，实现eudore.Server接口。

封装`github.com/eudore/eudore/protocol/server`作为服务端连接处理库。

使用`github.com/eudore/eudore/protocol/http`解析http协议。

使用`github.com/eudore/eudore/protocol/http2`解析http2协议。

使用`github.com/eudore/eudore/protocol/fastcgi`解析fastcgi协议。

## 使用

`import _ "github.com/eudore/eudore/component/server/eudore"`

初始化注册eudore.Component

## 例子