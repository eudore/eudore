package gorilla

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/session"
	gorillasession "github.com/gorilla/sessions"
)

// SessionName 定义会话cooke id
const SessionName = "sessionid"

// Provider 定义gorilla session提供者
type Provider struct {
	gorillasession.Store
}

type _Context = eudore.Context

// Session 定义gorilla session
type Session struct {
	_Context
	*gorillasession.Session
}

// NewExtendContextSession 函数返回gorilla/session库处理session。Context对象的扩展函数。
//
// 需要提供一个gorilla。Store对象。
func NewExtendContextSession(store gorillasession.Store) func(func(session.Context)) eudore.HandlerFunc {
	return func(fn func(session.Context)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			session, err := store.Get(ctx.Request(), SessionName)
			if err != nil {
				ctx.Fatal(err)
				return
			}

			fn(&Session{
				_Context: ctx,
				Session:  session,
			})
		}
	}
}

// SetSession 方法设置一个会话数据。
func (sess *Session) SetSession(key, value interface{}) error {
	sess.Values[key] = value
	return nil
}

// GetSession 方法获取一个会话数据。
func (sess *Session) GetSession(key interface{}) interface{} {
	return sess.Values[key]
}

// DeleteSession 方法删除一个会话数据。
func (sess *Session) DeleteSession(key interface{}) error {
	delete(sess.Values, key)
	return nil
}

// SessionID 方法返回会话id。
func (sess *Session) SessionID() string {
	return sess.ID
}

// SessionRelease 方法释放本次会话对象。
func (sess *Session) SessionRelease() {
	sess.Save(sess.Request(), sess.Response())
}

// SessionFlush 方法清空本会话数据，需要Release后保存。
func (sess *Session) SessionFlush() error {
	sess.Values = make(map[interface{}]interface{})
	return nil
}

// SessionInit 方法让提供者初始化一个会话对象。
func (p Provider) SessionInit(ctx eudore.Context) (session.Session, error) {
	session, err := p.Get(ctx.Request(), SessionName)
	if err != nil {
		return nil, err
	}
	return &Session{
		_Context: ctx,
		Session:  session,
	}, nil
}
