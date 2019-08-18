package notify

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/eudore/eudore"
	"github.com/fsnotify/fsnotify"
)

const (
	// NOTIFY_ENVIRON_KEY 是检查notify启动程序的环境变量。
	NOTIFY_ENVIRON_KEY = "EUDORE_IS_NOTIFY"
)

// Init 函数是eudpre.ReloadFunc, Eudore初始化内容。
func Init(app *eudore.Eudore) error {
	NewNotify(app.App).Run()
	return nil
}

// Notify 定义监听重启对象。
type Notify struct {
	eudore.Logger
	cmd      *exec.Cmd
	buildCmd []string
	startCmd []string
	watchDir []string
}

// NewNotify 函数创建一个Notify对象。
func NewNotify(app *eudore.App) *Notify {
	if app.Config.Get("component.notify.disable") != nil {
		app.Info("notify is disable")
		return nil
	}
	var (
		buildCmd = getArgs(app.Config.Get("component.notify.buildcmd"))
		startCmd = getArgs(app.Config.Get("component.notify.startcmd"))
		watchDir = getArgs(app.Config.Get("component.notify.watchdir"))
	)

	if len(buildCmd) == 0 || len(watchDir) == 0 {
		app.Info("notify config is empty.")
		return nil
	}

	if len(startCmd) == 0 {
		startCmd = os.Args
	}
	return &Notify{
		buildCmd: buildCmd,
		startCmd: startCmd,
		watchDir: watchDir,
		Logger:   app.Logger,
	}
}

// Run 方法启动Notify。
//
// 调用App.Logger
func (n *Notify) Run() {
	if os.Getenv(NOTIFY_ENVIRON_KEY) != "" || n == nil {
		return
	}
	n.buildAndStart()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		n.Error(err)
	}
	defer watcher.Close()

	// 添加函数
	addfn := func(path string) {
		n.Debug("notify add watch dir " + path)
		err = watcher.Add(path)
		if err != nil {
			n.Error(err)
		}
	}
	for _, i := range n.watchDir {
		// 递归目录处理
		if i[len(i)-1] == '/' {
			listDir(i, addfn)
		}
		addfn(i)
	}

	var timer = time.AfterFunc(1000*time.Hour, n.buildAndStart)
	defer timer.Stop()

	defer os.Exit(0)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				break
			}

			// 监听go文件写入
			if event.Name[len(event.Name)-3:] == ".go" && event.Op&fsnotify.Write == fsnotify.Write {
				n.Debug("modified file:", event.Name)

				// 等待0.1秒执行更新，防止短时间大量触发
				timer.Reset(100 * time.Millisecond)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			n.Error("notify watcher error:", err)
		}
	}
}

func (n *Notify) buildAndStart() {
	// 执行编译命令
	body, err := exec.Command(n.buildCmd[0], n.buildCmd[1:]...).CombinedOutput()
	if err != nil {
		fmt.Printf("notify build error: \n%s", body)
		n.Errorf("notify build error: %s", body)
	} else {
		n.Info("notify build success, restart process...")
		time.Sleep(10 * time.Millisecond)
		// 重启子进程
		n.start()
	}
}

func (n *Notify) start() {
	// 关闭旧进程
	if n.cmd != nil {
		n.cmd.Process.Kill()
		n.cmd.Process.Wait()
	}

	// 启动新进程
	n.cmd = exec.Command(n.startCmd[0], n.startCmd[1:]...)
	n.cmd.Stdin = os.Stdin
	n.cmd.Stdout = os.Stdout
	n.cmd.Stderr = os.Stderr
	n.cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", NOTIFY_ENVIRON_KEY, 1), "EUDORE_NOTPID=1")
	err := n.cmd.Start()
	if err != nil {
		n.Error("notify start error:", err)
	}
}

// 转换配置成数组类型
func getArgs(i interface{}) []string {
	if strs, ok := i.([]string); ok {
		return strs
	}
	if strs, ok := i.(string); ok {
		return strings.Split(strs, " ")
	}
	return nil
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
