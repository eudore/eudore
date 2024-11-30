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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/daemon"
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
	*eudore.App
	*ConfigNotify
	sync.Mutex
	Watcher       *fsnotify.Watcher
	cancelBuild   context.CancelFunc
	cancelProcess context.CancelFunc
}

type ConfigNotify struct {
	Workdir string `json:"workdir" alias:"workdir"`
	Command string `json:"command" alias:"command"`
	Pidfile string `json:"pidfile" alias:"pidfile"`
	Main    string `json:"main" alias:"main"`
	Build   string `json:"build" alias:"build"`
	Start   string `json:"start" alias:"start"`
	Watch   string `json:"watch" alias:"watch"`
}

func main() {
	app := NewApp()
	app.Parse()
	go app.RunNotify()
	app.Run()
}

// NewApp 函数创建一个Notify对象。
func NewApp() *App {
	app := &App{
		App: eudore.NewApp(),
		ConfigNotify: &ConfigNotify{
			Build: "go build -o app .",
			Start: "./app",
			Watch: ".",
		},
	}
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.ConfigNotify))
	app.ParseOption()
	app.ParseOption(
		eudore.NewConfigParseEnvFile(),
		eudore.NewConfigParseEnvs(""),
		eudore.NewConfigParseArgs(),
		eudore.NewConfigParseWorkdir("workdir"),
		daemon.NewParseSignal(),
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
		for _, path := range strings.Split(app.Watch, ";") {
			app.WatchAll(strings.TrimSpace(path))
		}

		app.SetValue(eudore.NewContextKey("watch-close"), eudore.Unmounter(func(context.Context) {
			watcher.Close()
		}))
		return nil
	}
}

// Run 方法启动Notify。
//
// 调用App.Logger
func (app *App) RunNotify() {
	if app.Main != "" {
		app.Build = "go build -o app " + app.Main
	}
	app.Info("notify buildCmd", app.Build)
	app.Info("notify startCmd", app.Start)
	app.buildAndRestart()

	timer := time.AfterFunc(1000*time.Hour, app.buildAndRestart)
	defer func() {
		timer.Stop()
		if app.cancelBuild != nil {
			app.cancelBuild()
		}
		if app.cancelProcess != nil {
			app.cancelProcess()
		}
	}()

	for {
		select {
		case event, ok := <-app.Watcher.Events:
			if !ok {
				return
			}

			// 监听go文件写入
			if event.Name[len(event.Name)-3:] == ".go" && event.Op&fsnotify.Write == fsnotify.Write {
				app.Debug("modified file:", event.Name)

				// 等待0.1秒执行更新，防止短时间大量触发
				timer.Reset(100 * time.Millisecond)
			}
		case err, ok := <-app.Watcher.Errors:
			if !ok {
				return
			}
			app.Error("notify watcher error:", err)
		case <-app.Done():
			return
		}
	}
}

func (app *App) buildAndRestart() {
	// 取消上传编译
	app.Lock()
	if app.cancelBuild != nil {
		app.cancelBuild()
	}
	ctx, cancel := context.WithCancel(app.Context)
	app.cancelBuild = cancel
	app.Unlock()
	// 执行编译命令
	cmd := exec.CommandContext(ctx, startcmd, "-c", app.Build)
	cmd.Env = os.Environ()
	body, err := cmd.CombinedOutput()
	if err != nil {
		app.Errorf("notify build error: %s", body)
		return
	}

	app.Info("notify build success, restart process...")
	time.Sleep(10 * time.Millisecond)
	// 重启子进程
	app.restart()
}

func (app *App) restart() {
	// 关闭旧进程
	app.Lock()
	if app.cancelProcess != nil {
		app.cancelProcess()
	}
	ctx, cancel := context.WithCancel(app.Context)
	app.cancelProcess = cancel
	app.Unlock()
	// 启动新进程
	cmd := exec.CommandContext(ctx, startcmd, "-c", app.Start)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	err := cmd.Start()
	if err != nil {
		app.Error("notify start error:", err)
	}
}

// WatchAll 方法添加一个文件或目录，如果/结尾的目录会递归监听子目录。
func (app *App) WatchAll(path string) {
	if path != "" {
		// 递归目录处理
		listDir(path, app.watch)
		app.watch(path)
	}
}

func (app *App) watch(path string) {
	app.Debug("notify add watch dir " + path)
	err := app.Watcher.Add(path)
	if err != nil {
		app.Error(err)
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
