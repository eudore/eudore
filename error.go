package eudore

import (
	"log"
	"fmt"
	"errors"
)
const (
	StatueRouter		=	610
	StatusLogger		=	611
	StatusCache			=	612
)

var (
	ErrRouterSetNoSupportType		=	errors.New("router set type is nosupport")
	ErrComponentNoSupportField		=	errors.New("component no support field")
	ErrServerNotSetRuntimeInfo		=	errors.New("server not set runtime info")
	ErrApplicationStop				=	errors.New("stop application")
	ErrHandlerInvalidRange			=	errors.New("invalid range")
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

type ErrorHttp struct {
	code	int
	message		string
}

func NewError(c int,str string) *ErrorHttp {
	return &ErrorHttp{code:	c, message:	str}
}

func (e *ErrorHttp) Error() string {
	return e.message
}

func (e *ErrorHttp) Code() int {
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
func DefaultErrorHandleFunc(e error){
	fmt.Println(e)
}
