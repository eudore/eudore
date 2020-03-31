# Eudore

[![Go Report Card](https://goreportcard.com/badge/github.com/eudore/eudore)](https://goreportcard.com/report/github.com/eudore/eudore)
[![GoDoc](https://godoc.org/github.com/eudore/eudore?status.svg)](https://pkg.go.dev/github.com/eudore/eudore?tab=doc)

eudore是轻而全的http框架，高度解耦而轻，功能丰富而全。

反馈和交流请加群组：[QQ群373278915](//shang.qq.com/wpa/qunwpa?idkey=869ec8f1272b4757771c3e406349f1128cfa3bd9ca668937dda8dfb223261a60)。

## Features

- 易扩展：主要设计目标、核心全部解耦，接口即为逻辑。
- 简单：对象语义明确，框架代码量少复杂度低，无依赖库。
- 易用：支持各种Appcation和Context扩展添加功能。
- 高性能：各部分在同类库中没有明显性能问题。
- 两项创新：[新Radix路由实现](https://github.com/eudore/eudore/wiki/4.5.1-eudore-router-radix)和[处理函数扩展机制](https://github.com/eudore/eudore/wiki/4.7-eudore-handler)

## 安装

eudore基于`go version go1.10.1 linux/amd64`下开发，运行依赖go1.9+版本。

```bash
go get -v -u github.com/eudore/eudore
```

## 文档

- [源码](https://github.com/eudore/eudore)
- [godoc](https://pkg.go.dev/github.com/eudore/eudore?tab=doc)
- [演示例子 90+](_example#example)
- [wiki文档](https://github.com/eudore/eudore/wiki)
- [实践](https://github.com/eudore/website)

## 许可

MIT
