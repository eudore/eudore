package test

import (
	"testing"
	"eudore"
)

func TestLogger(*testing.T) {
	ls, _ := eudore.NewLoggerStd(nil)
	ls.SetFromat(eudore.LoggerFormatJsonIndent)
	ls.WithField("action", "ss").Info("info")

	li, _ := eudore.NewLoggerInit(nil)
	li.Info("init info")
	li.WithField("init", true).Info("-------")
	li.(eudore.LoggerInitHandler).NextHandler(ls)
}