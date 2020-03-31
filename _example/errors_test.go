package eudore_test

import (
	"errors"
	"testing"

	"github.com/eudore/eudore"
)

func TestErrors2(t *testing.T) {
	errs := eudore.NewErrors()
	t.Log(errs.GetError(), errs.Error())
	errs.HandleError(errors.New("test error 1"))
	t.Log(errs.GetError(), errs.Error())
	errs.HandleError(errors.New("test error 2"))
	t.Log(errs.GetError(), errs.Error())
}
