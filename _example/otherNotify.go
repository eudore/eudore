package main

/*
go build -o ~/go/bin/gonotify otherNotify.go

先app.Config设置notify配置，然后启动notify。
如果是notify的程序可以通过环境变量eudore.EnvEudoreIsNotify检测。
当程序启动时会如果eudore.EnvEudoreIsNotify不存在，则使用notify开始监听阻塞app后续初始化，否在就忽略notify然后进行正常app启动。

实现原理基于fsnotify检测目录内go文件变化，然后执行编译命令，如果编译成功就kill原进程并执行启动命令。

其他类似工具：air
*/

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/fsnotify/fsnotify"
)

var startcmd string

func init() {
	if runtime.GOOS == "windows" {
		startcmd = "powershell"
	} else {
		startcmd = "bash"
	}
}

// Notify 定义监听重启对象。
type App struct {
	sync.Mutex
	*NotifyConfig
	*eudore.App
	Watcher     *fsnotify.Watcher
	lastBuild   context.CancelFunc
	lastProcess context.CancelFunc
}

type NotifyConfig struct {
	Workdir string `json:"workdir" alias:"workdir"`
	Command string `json:"command" alias:"command"`
	Pidfile string `json:"pidfile" alias:"pidfile"`
	Build   string `json:"build" alias:"build"`
	Start   string `json:"start" alias:"start"`
	Watch   string `json:"watch" alias:"watch"`
}

func main() {
	app := NewApp()
	app.Parse()
	go app.Run()
	app.App.Run()
}

// NewApp 函数创建一个Notify对象。
func NewApp() *App {
	conf := &NotifyConfig{
		Build: "",
		Start: "go run .",
		Watch: ".",
	}
	app := &App{
		NotifyConfig: conf,
		App:          eudore.NewApp(),
	}
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(conf))
	app.ParseOption(
		eudore.DefaultConfigAllParseFunc,
		app.NewParseWatcherFunc(),
	)
	return app
}

func (app *App) NewParseWatcherFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		app.Watcher = watcher
		return nil
	}
}

// Run 方法启动Notify。
//
// 调用App.Logger
func (n *App) Run() {
	n.App.Info("notify buildCmd", n.Build)
	n.App.Info("notify startCmd", n.Start)
	for _, path := range strings.Split(n.Watch, ";") {
		n.WatchAll(strings.TrimSpace(path))
	}

	n.buildAndRestart()

	var timer = time.AfterFunc(1000*time.Hour, n.buildAndRestart)
	defer func() {
		timer.Stop()
		if n.lastBuild != nil {
			n.lastBuild()
		}
		if n.lastProcess != nil {
			n.lastProcess()
		}
	}()

	for {
		select {
		case event, ok := <-n.Watcher.Events:
			if !ok {
				return
			}

			// 监听go文件写入
			if event.Name[len(event.Name)-3:] == ".go" && event.Op&fsnotify.Write == fsnotify.Write {
				n.App.Debug("modified file:", event.Name)

				// 等待0.1秒执行更新，防止短时间大量触发
				timer.Reset(100 * time.Millisecond)
			}
		case err, ok := <-n.Watcher.Errors:
			if !ok {
				return
			}
			n.App.Error("notify watcher error:", err)
		case <-n.App.Done():
			return
		}
	}
}

func (n *App) buildAndRestart() {
	if n.Build != "" {
		// 取消上传编译
		n.Lock()
		if n.lastBuild != nil {
			n.lastBuild()
		}
		ctx, cannel := context.WithCancel(n.App.Context)
		n.lastBuild = cannel
		n.Unlock()
		// 执行编译命令
		cmd := exec.CommandContext(ctx, startcmd, "-c", n.Build)
		cmd.Env = os.Environ()
		body, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("notify build error: \n%s", body)
			n.App.Errorf("notify build error: %s", body)
			return
		}
	}
	n.App.Info("notify build success, restart process...")
	time.Sleep(10 * time.Millisecond)
	// 重启子进程
	n.restart()
}

func (n *App) restart() {
	// 关闭旧进程
	n.Lock()
	if n.lastProcess != nil {
		n.lastProcess()
	}
	ctx, cannel := context.WithCancel(n.App.Context)
	n.lastProcess = cannel
	n.Unlock()
	// 启动新进程
	cmd := exec.CommandContext(ctx, startcmd, "-c", n.Start)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	err := cmd.Start()
	if err != nil {
		n.App.Error("notify start error:", err)
	}
}

// WatchAll 方法添加一个文件或目录，如果/结尾的目录会递归监听子目录。
func (n *App) WatchAll(path string) {
	if path != "" {
		// 递归目录处理
		listDir(path, n.watch)
		n.watch(path)
	}
}

func (n *App) watch(path string) {
	n.App.Debug("notify add watch dir " + path)
	err := n.Watcher.Add(path)
	if err != nil {
		n.App.Error(err)
	}
}

func listDir(path string, fn func(string)) {
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		// 忽略隐藏目录，例如: .git
		if f.IsDir() && f.Name()[0] != '.' {
			path := filepath.Join(path, f.Name())
			fn(path)
			listDir(path, fn)
		}
	}
}
