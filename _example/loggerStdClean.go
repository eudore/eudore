package main

/*
NewLoggerWriterRotate方法创建一个日志切割写入流，可以使用第四个参数...func(string)指定创建文件名称的操作。
例如：创建软连接、文件保留指定数量、文件保存指定时间
*/

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

func main() {
	defer os.RemoveAll("logger")

	lw, err := eudore.NewLoggerWriterRotate("logger/logger-index.log", false, 1<<10, newCleanSizeFile(6))
	app := eudore.NewApp(eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Writer:     lw,
		Level:      eudore.LogDebug,
		TimeFormat: "Mon Jan 2 15:04:05 -0700 MST 2006",
	}))
	app.Options(err)

	for i := 0; i < 100; i++ {
		app.Debug("now is", time.Now().String())
	}
	app.Sync()

	app.CancelFunc()
	app.Run()
}

// 保留n个文件
func newCleanSizeFile(n int) func(string) {
	names := make([]string, n)
	index := 0
	return func(name string) {
		fmt.Println("open file:", name)
		if names[index] != "" {
			fmt.Println("delete file:", names[index])
			os.Remove(names[index])
		}
		names[index] = name
		index++
		if index == n {
			index = 0
		}

	}
}

// 定时定义过期文件，未测试
func newCleanTimeFile(t time.Duration) func(string) {
	var mu sync.Mutex
	var names []string
	var times []time.Time

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			// 清理一个文件，保留写入中文件，判断文件结束时间。
			if len(names) > 1 && times[1].After(time.Now()) {
				os.Remove(names[0])
				names = names[1:]
				times = times[1:]
			}
			mu.Unlock()
		}
	}()

	return func(name string) {
		mu.Lock()
		names = append(names, name)
		times = append(times, time.Now().Add(t))
		mu.Unlock()
	}
}
