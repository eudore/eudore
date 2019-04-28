package eudore


import (
	"sync"
)

type (
	SessionStore interface {
		Load(Context) Session
		Save(Session)
	}
	Session interface {
		Set(key string, value interface{}) error //set session value
		Get(key string) interface{}  //get session value
		Del(key string) error     //delete session value
		SessionID() string                //back current sessionID
	}
	SessionStoreMem struct {
		KeyFync	func(Context) string
		mem		sync.Map
	}
	SessionStd struct {
		Id		string
		Data	map[string]interface{}
	}
)

var (
	DefaultSessionStore SessionStore
)

func NewSession(ctx Context) Session {
	return DefaultSessionStore.Load(ctx)
}

func SessionRelease(sess Session) {
	DefaultSessionStore.Save(sess)
}

func NewSessionStoreMem(interface{}) (SessionStore, error) {
	return &SessionStoreMem{
		KeyFync:	func(ctx Context) string {
			return ctx.GetCookie("sessionid")
		},
	}, nil
}


func (store *SessionStoreMem) Load(ctx Context) Session {
	key := store.KeyFync(ctx)
	sess, ok := store.mem.Load(key)
	if ok {
		return sess.(Session)
	}
	return &SessionStd{Id: key}
}

func (store *SessionStoreMem) Save(sess Session) {
	store.mem.Store(sess.SessionID(), sess)
}

func (sess *SessionStd) Set(key string, val interface{}) error {
	sess.Data[key] = val
	return nil
}

func (sess *SessionStd) Get(key string) interface{} {
	return sess.Data[key]
}

func (sess *SessionStd) Del(key string) error {
	delete(sess.Data, key)
	return nil
}

func (sess *SessionStd) SessionID() string {
	return sess.Id
}
