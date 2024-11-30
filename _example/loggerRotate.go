package main

/*
通过设置LoggerStdConfig的MaxSize属性输出日志滚动，在日志名称中必须包含index关键字用于指定索引
*/

import (
	"time"

	"bou.ke/monkey"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:     false,
		Path:       "logger/logger-yyyy-mm-dd-hh-index.log",
		Link:       "logger/logger.log",
		MaxSize:    1 << 10, // 1k
		MaxCount:   5,
		TimeFormat: "Mon Jan 2 15:04:05 -0700 MST 2006",
	}))

	// This is as unsafe as it sounds and I don't recommend anyone do it outside of a testing environment.
	mytime := time.Now()
	patch := monkey.Patch(time.Now, func() time.Time { return mytime })
	defer patch.Unpatch()

	for i := 0; i < 100; i++ {
		if i%30 == 9 {
			mytime = mytime.Add(time.Hour)
		}
		app.Debug("now is", time.Now().String())
	}

	app.CancelFunc()
	app.Run()
}
