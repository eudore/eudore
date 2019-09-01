# Eudore

[![Go Report Card](https://goreportcard.com/badge/github.com/eudore/eudore)](https://goreportcard.com/report/github.com/eudore/eudore)
[![GoDoc](https://godoc.org/github.com/eudore/eudore?status.svg)](https://godoc.org/github.com/eudore/eudore)

eudore是一个高扩展、高效的http框架及[http文档库](docs)。

反馈和交流[q群373278915](//shang.qq.com/wpa/qunwpa?idkey=869ec8f1272b4757771c3e406349f1128cfa3bd9ca668937dda8dfb223261a60)。

## Features

- 易扩展：主要设计目标，核心全部解耦，接口即可逻辑。
- 简单：对象语义明确，框架代码量少复杂度低，无依赖库。
- 易用：支持各种Appcation和Context扩展添加功能。
- 高性能：各部分在同类库中没有明显性能问题。
- 两项创新：[新Radix路由实现](https://github.com/eudore/erouter)和处理函数扩展机制

## 安装

eudore基于`go version go1.10.1 linux/amd64`下开发，运行依赖go1.9+版本。

```bash
go get -v -u github.com/eudore/eudore
```

## 文档

- [源码](https://github.com/eudore/eudore)
- [godoc](https://godoc.org/github.com/eudore/eudore)
- [例子](docs/example)
- [框架文档](docs/frame)

## 许可

MIT

框架使用无限制且不负责，文档转载需声明出处。