// +build windows

package notify

/*
windows平台禁止使用，原因：忘记了,导致程序无法启动。
*/

import (
	"github.com/eudore/eudore"
)

// Init 函数是eudpre.ReloadFunc, Eudore初始化内容。
//
// 	app.RegisterInit("eudore-notify", 0x00e, notify.Init)
func Init(app *eudore.Eudore) error {
	return NewNotify(app.App).Run()
}

// Notify 定义监听重启对象。
type Notify struct{}

// NewNotify 函数创建一个Notify对象。
func NewNotify(app *eudore.App) *Notify {
	return nil
}

// Run 方法启动Notify。
//
// 调用App.Logger
func (n *Notify) Run() error {
	return nil
}
