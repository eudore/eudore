package main

/*
RegisterValidations可以注册校验函数和校验构建函数，具体查看文档。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
)

type newUserRequest struct {
	Username string `validate:"regexp:^[a-zA-Z]*$"`
	Name     string `validate:"nozero"`
	Age      int    `validate:"min:21,max:40"`
	Password string `validate:"lenmin:8"`
}

func main() {
	fmt.Println(eudore.DefaultValidater.Validate(newUserRequest{
		Username: "abc",
		Name:     "eudore",
		Age:      21,
		Password: "12345678",
	}))

	var name = "8hs8a"
	fmt.Println(eudore.DefaultValidater.ValidateVar(name, "regexp:^[a-zA-Z]*$"))
}
