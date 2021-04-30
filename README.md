# Eudore

[![godoc](https://godoc.org/github.com/eudore/eudore?status.svg)](https://godoc.org/github.com/eudore/eudore)
[![go report card](https://goreportcard.com/badge/github.com/eudore/eudore)](https://goreportcard.com/report/github.com/eudore/eudore)
[![codecov](https://codecov.io/gh/eudore/eudore/branch/master/graph/badge.svg)](https://codecov.io/gh/eudore/eudore)

eudore是一个golang轻量级web框架核心，可以轻松扩展成一个技术栈专用框架，具有完整框架设计体系。

反馈和交流请加群组：[QQ群373278915](//shang.qq.com/wpa/qunwpa?idkey=869ec8f1272b4757771c3e406349f1128cfa3bd9ca668937dda8dfb223261a60)。

## Features

- 易扩展：主要设计目标、核心全部解耦，接口即为逻辑。
- 简单：对象语义明确，框架代码量少[复杂度低](https://goreportcard.com/report/github.com/eudore/eudore#gocyclo)，无依赖库。
- 易用：允许Appcation和Context自由添加功能方法。
- 高性能：各部分实现与同类库相比性能相似。
- 两项创新：[新Radix路由实现](https://github.com/eudore/eudore/wiki/4.5.1-eudore-router-radix)和[处理函数扩展机制](https://github.com/eudore/eudore/wiki/4.7-eudore-handler)

## 安装

eudore基于`go version go1.10.1 linux/amd64`下开发，运行依赖go1.9+版本。

```bash
go get -v -u github.com/eudore/eudore
```

## 文档

- [源码](https://github.com/eudore/eudore)
- [godoc](https://godoc.org/github.com/eudore/eudore)
- [example演示 100+](_example#example)
- [wiki文档](https://github.com/eudore/eudore/wiki)
- [更新说明](CHANGELOG.md)
- [实践](https://github.com/eudore/website)
