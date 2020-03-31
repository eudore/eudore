package eudore_test

import (
	"errors"
	"github.com/eudore/eudore"
	"testing"
)

func TestValidaterRegister2(t *testing.T) {
	eudore.RegisterValidations("test", 100)
	eudore.RegisterValidations("test", func(string) func(string) bool {
		return func(string) bool {
			return false
		}
	})
	eudore.RegisterValidations("test", func(string) func(string) string {
		return func(key string) string {
			return key
		}
	})
	eudore.RegisterValidations("test", func(string) func(string, string) {
		return func(string, string) {}
	})
}

type (
	valid1 struct{}
	valid2 struct{}
	valid3 struct {
		V    *valid1
		Age  string `validate:"isnum"`
		Age2 string `validate:"isnum"`
	}
	valid4 struct {
		V *valid2
	}
	valid5 struct {
		Name string
		Age  int `validate:"min:aa"`
	}
	valid6 struct {
		Age string `validate:"min:aa"`
	}
	valid7 struct {
		Name string
		Age  int `validate:"max:aa"`
	}
	valid8 struct {
		Age string `validate:"max:aa"`
	}
)

func (valid1) Validate() error { return nil }
func (valid2) Validate() error { return errors.New("test valid2 error") }

func TestValidaterHandle2(t *testing.T) {
	t.Log(eudore.Validate(struct{}{}))
	t.Log(eudore.Validate(t))
	t.Log(eudore.Validate(t.Log))
	t.Log(eudore.Validate(new(valid1)))
	t.Log(eudore.Validate(new(valid2)))
	t.Log(eudore.Validate(&valid3{
		Age:  "11",
		Age2: "22",
	}))
	t.Log(eudore.Validate(&valid3{
		Age:  "11",
		Age2: "aa",
		V:    new(valid1),
	}))
	t.Log(eudore.Validate(&valid3{
		Age:  "11",
		Age2: "22",
		V:    new(valid1),
	}))
	t.Log(eudore.Validate(&valid4{
		V: new(valid2),
	}))

	t.Log(eudore.Validate(new(valid5)))
	t.Log(eudore.Validate(new(valid6)))
	t.Log(eudore.Validate(new(valid7)))
	t.Log(eudore.Validate(new(valid8)))
}

func TestValidaterHandleVar2(t *testing.T) {
	t.Log(eudore.ValidateVar(0, "nozero"))
	t.Log(eudore.ValidateVar(1, "nozero"))
	t.Log(eudore.ValidateVar("1", "nozero"))
	t.Log(eudore.ValidateVar(new(valid3), "nozero"))
	t.Log(eudore.ValidateVar(2, "min:1"))
	t.Log(eudore.ValidateVar("2", "min:1"))
	t.Log(eudore.ValidateVar("a2", "min:1"))
	t.Log(eudore.ValidateVar("a2", "min:5"))
	t.Log(eudore.ValidateVar(2, "max:1"))
	t.Log(eudore.ValidateVar(2, "max:5"))
	t.Log(eudore.ValidateVar("2aa", "max:1"))
	t.Log(eudore.ValidateVar("2", "max:5"))
	eudore.GetValidateStringFunc("min:5")
	eudore.GetValidateStringFunc("nil")
	eudore.GetValidateStringFunc("nil:0")
	eudore.GetValidateStringFunc("regexp:[[[")
	t.Log(eudore.ValidateVar("2", "len:2"))
	t.Log(eudore.ValidateVar("22", "len:2"))
	t.Log(eudore.ValidateVar("2", "len:=2"))
	t.Log(eudore.ValidateVar("22", "len:=2"))
	t.Log(eudore.ValidateVar("222", "len:>2"))
	t.Log(eudore.ValidateVar("2", "len:>2"))
	t.Log(eudore.ValidateVar("222", "len:<2"))
	t.Log(eudore.ValidateVar("2", "len:<2"))
	t.Log(eudore.ValidateVar("2", "len:"))
	t.Log(eudore.ValidateVar("2", "regexp:.*"))
	eudore.DefaultRouterValidater = nil
	eudore.GetValidateStringFunc("nil")
	eudore.DefaultRouterValidater = eudore.DefaultValidater
}

func TestValidaterString(t *testing.T) {
	eudore.RegisterValidations("hello1", func(interface{}) bool { return true })
	eudore.RegisterValidations("hello2", func(string) func(string) bool {
		return func(string) bool {
			return true
		}
	})
	eudore.RegisterValidations("hello3", func(int) bool { return true })
	t.Log(eudore.GetValidateStringFunc("hello1")("2"))
	t.Log(eudore.GetValidateStringFunc("hello2:")("2"))
	t.Log(eudore.GetValidateStringFunc("hello3") == nil)
}
