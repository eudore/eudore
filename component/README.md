# Eudore Component范式

组件由Config和Handler组成，基于golang struct组合语法特性，Config先实现Component接口，然后Handler包含Config部分。

如果没有Config部分，Handler直接实现Component接口。

所有componentname-std组件为默认使用组件，以良好的兼容性为实现标准。

# Component接口

```golang
Component interface {
	GetName() string
	Version() string
}
```

`GetName() string`方法主要获取当前组件类型，可用于组件复制。例如路由组件继续父路由类型。

`Version() string`方法获取当前组件的信息，一般由版本和表述组合。