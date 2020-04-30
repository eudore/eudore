package beego

import (
	beegosession "github.com/astaxie/beego/session"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/session"
)

// Provider 定义beego session提供者
type Provider struct {
	*beegosession.Manager
}

type _Context = eudore.Context

// Session 定义beego session
type Session struct {
	_Context
	beegosession.Store
}

// NewExtendContextSession 函数返回beego/session库处理session。Context对象的扩展函数。
func NewExtendContextSession(manager *beegosession.Manager) func(func(session.Context)) eudore.HandlerFunc {
	return func(fn func(session.Context)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			sess, err := manager.SessionStart(ctx.Response(), ctx.Request())
			if err != nil {
				ctx.Fatal(err)
				return
			}

			fn(&Session{
				_Context: ctx,
				Store:    sess,
			})
		}
	}
}

// SetSession 方法设置一个会话数据。
func (sess *Session) SetSession(key, value interface{}) error {
	return sess.Set(key, value)
}

// GetSession 方法获取一个会话数据。
func (sess *Session) GetSession(key interface{}) interface{} {
	return sess.Get(key)
}

// DeleteSession 方法删除一个会话数据。
func (sess *Session) DeleteSession(key interface{}) error {
	return sess.Delete(key)
}

// SessionID 方法返回会话id。
func (sess *Session) SessionID() string {
	return sess.SessionID()
}

// SessionRelease 方法释放本次会话对象。
func (sess *Session) SessionRelease() {
	sess.Store.SessionRelease(sess.Response())
}

// SessionFlush 方法清空本会话数据，需要Release后保存。
func (sess *Session) SessionFlush() error {
	return sess.Flush()
}

// SessionInit 方法让提供者初始化一个会话对象。
func (p Provider) SessionInit(ctx eudore.Context) (session.Session, error) {
	sess, err := p.SessionStart(ctx.Response(), ctx.Request())
	if err != nil {
		return nil, err
	}
	return &Session{
		_Context: ctx,
		Store:    sess,
	}, nil
}
