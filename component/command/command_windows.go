package command

/*
windows平台禁止使用，原因：pid文件锁不支持。
*/

import (
	"context"
	"github.com/eudore/eudore"
)

// Command is a command parser that performs the corresponding behavior based on the current command.
//
// Command 对象是一个命令解析器，根据当前命令执行对应行为。
type Command struct{}

// Init 函数初始化定义程序启动命令。
func Init(*eudore.App) error {
	return nil
}

// NewCommand 函数返回一个命令解析对象，需要当前命令和进程pid文件路径，如果行为是start会执行handler。
func NewCommand(context.Context, string, string) *Command {
	return nil
}

// Run parse the command and execute it.
//
// Run 函数解析命令并执行。
func (c *Command) Run() error {
	return nil
}
