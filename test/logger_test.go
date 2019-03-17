package test

import (
	"testing"
	"github.com/eudore/eudore"
)

func TestLogger(*testing.T) {
	ls, _ := eudore.NewLoggerStd(nil)
	ls.WithField("action", "ss").Info("info")

	li, _ := eudore.NewLoggerInit(nil)
	li.Info("init info")
	li.WithField("init", true).Info("-------")
	li.(eudore.LoggerInitHandler).NextHandler(ls)
}

func BenchmarkStdLogger(b *testing.B) {
	b.ReportAllocs()
	log, _ := eudore.NewLoggerStd(nil)
	for i := 0; i < b.N; i++ {
		log.WithField("bench", true).WithField("test", true).Info("Info")
	}
}