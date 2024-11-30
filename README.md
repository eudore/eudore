# Eudore

[![godoc](https://godoc.org/github.com/eudore/eudore?status.svg)](https://godoc.org/github.com/eudore/eudore)
[![Build Status](https://github.com/eudore/eudore/actions/workflows/github-action.yml/badge.svg)](https://github.com/eudore/eudore/actions/workflows/github-action.yml)
[![codecov](https://codecov.io/gh/eudore/eudore/branch/master/graph/badge.svg)](https://codecov.io/gh/eudore/eudore)

eudore is the core of a composite web framework, which can replace any content through composition;
For simple apps, it can be used directly, and for complex applications, the framework can be customized by composition.

The framework uses a three-layer structure of App, Controller, and Context.
- eudore.App combines Logger, Config, Router, Client, Server, and Values.
- Custom App combines eudore.App, database/sql, Prometheus, and other custom components.
- Controller uses methods to create automatic routing and
 copies required dependent components from App.
- Context uses HandlerExtender to create custom processing functions and
 uses HandlerDataFunc to implement data binding, verification, filtering, and rendering processes.

Each built-in component does not use any third-party dependencies,
and its performance is similar to that of similar libraries,
which can be achieved by combining and replacing part of the component content.

This project is usually updated at the end of the month, and no API compatibility is guaranteed,
but most of the content has been fixed after years of maintenance.
Developed using go1.20 and GOPATH mode.

See the documentation for more details:
- [wiki](https://github.com/eudore/eudore/wiki)
- [godoc](https://godoc.org/github.com/eudore/eudore)
- [example](_example#example)
- [change log](CHANGELOG.md)
