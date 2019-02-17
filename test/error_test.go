package test

import (
	"fmt"
	"testing"
	"github.com/eudore/eudore"
)

func TestErros(t *testing.T) {
	var errs = &eudore.Errors{}
	t.Log(errs.GetError())
	errs.HandleError(fmt.Errorf("err1"))
	errs.HandleError(fmt.Errorf("err2"))
	errs.HandleError(fmt.Errorf("err3"))
	t.Log(errs.GetError())
}