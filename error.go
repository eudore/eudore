package eudore

import (
	"errors"
	"fmt"
)

// 定义默认错误
var (
	// ErrApplicationStop 在app正常退出时返回。
	ErrApplicationStop = errors.New("stop application")
	// ErrHandlerInvalidRange 在http使用range分开请求文件出现错误时返回。
	ErrHandlerInvalidRange = errors.New("invalid range")
)

type (
	// Errors 实现多个error组合。
	Errors struct {
		errs []error
	}
	// ErrorCode 实现具有错误信息和错误码的error。
	ErrorCode struct {
		code    int
		message string
	}
)

// NewErrors 创建Errors对象。
func NewErrors() *Errors {
	return &Errors{}
}

// HandleError 实现处理多个错误，如果非空则保存错误。
func (e *Errors) HandleError(errs ...error) {
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
}

// Error 方法实现error接口，返回错误描述。
func (e *Errors) Error() string {
	switch len(e.errs) {
	case 0:
		return ""
	case 1:
		return e.errs[0].Error()
	default:
		return fmt.Sprint(e.errs)
	}
}

// GetError 方法返回错误，如果没有保存的错误则返回空。
func (e *Errors) GetError() error {
	if len(e.errs) == 0 {
		return nil
	}
	return e
}

// NewErrorCode 创建一个ErrorCode对象。
func NewErrorCode(c int, str string) *ErrorCode {
	return &ErrorCode{code: c, message: str}
}

// Error 方法实现error接口，返回错误信息。
func (e *ErrorCode) Error() string {
	return e.message
}

// Code 方法返回错误状态码。
func (e *ErrorCode) Code() int {
	return e.code
}
