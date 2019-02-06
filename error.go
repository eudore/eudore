package eudore

import (
	"log"
	"fmt"
)
const (
	StatueRouter		=	610
	StatusLogger		=	611
	StatusCache			=	612
)

type (
	ErrorFunc func(error)
	ErrorHandler interface {
		HandleError(err error)
	}
	Errors struct{
		errs []error
	}

)

type Error struct {
	code	int
	msg		string
}

func NewError(c int,str string) *Error {
	return &Error{code:	c, msg:	str}
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Code() int {
	return e.code
}


type HttpError struct {
	handle		ErrorFunc
	log 		*log.Logger
}

func NewHttpError(fn ErrorFunc) *HttpError {
	e := &HttpError{
		handle:		fn,
	}
	e.log = log.New(e, "", 0)
	return e
}

func (e *HttpError) Write(p []byte) (n int, err error) {
	e.handle(fmt.Errorf(string(p)))
	return 0, nil
}

func (e *HttpError) Logger() *log.Logger {
	return e.log
}


func NewErrors() *Errors {
	return &Errors{}
}

func (e *Errors) HandleError(errs ...error) {
	for _, err := range errs {		
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
}

func (e Errors) Error() string {
	if len(e.errs) == 0 {
		return ""
	}
	return fmt.Sprint(e.errs)
}

func (e *Errors) GetError() error {
	if len(e.errs) == 0 {
		return nil
	}
	return e
}

// The default error handler, outputting an error to std.out.
//
// 默认错误处理函数，输出错误到std.out。
func ErrorDefaultHandleFunc(e error){
	fmt.Println(e)
}