package main

import (
	"os"
	"time"

	"bou.ke/monkey"
	"github.com/eudore/eudore"
)

func main() {
	defer os.RemoveAll("logger")
	app := eudore.NewApp(eudore.NewLoggerStd(&eudore.LoggerStdConfig{
		Path:       "logger/logger-yyyy-MM-dd-HH-index.log",
		Link:       "logger/logger.log",
		MaxSize:    1 << 10, // 1k
		Std:        false,
		Level:      eudore.LogDebug,
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
	app.Sync()

	app.CancelFunc()
	app.Run()
}
