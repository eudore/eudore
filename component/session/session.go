package session

import (
	"github.com/eudore/eudore"
)

// SessionProvider 定义session提供者。
type SessionProvider interface {
	SessionInit(eudore.Context) (Session, error)
}

// Session 定义Session操作接口。
type Session interface {
	SetSession(interface{}, interface{}) error
	GetSession(interface{}) interface{}
	DeleteSession(interface{}) error
	SessionID() string
	SessionRelease()
	SessionFlush() error
}

// Context 定义Session Context对象。
type Context interface {
	eudore.Context
	Session
}
