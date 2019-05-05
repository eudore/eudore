package eudore


import (
	"fmt"
	"sync"
)

type (
	Session interface {
		Component
		SessionLoad(Context) SessionData
		SessionSave(SessionData)
		SessionRelease(string)
	}
	SessionData interface {
		Set(key string, value interface{}) error //set session value
		Get(key string) interface{}  //get session value
		Del(key string) error     //delete session value
		SessionID() string                //back current sessionID
	}
	SessionMap struct {
		KeyFync	func(Context) string
		mem		sync.Map
	}
	SessionStd struct {
		Id		string
		Data	map[string]interface{}
	}
)

var (
	DefaultSession Session
)

func NewSession(name string, arg interface{}) (Session, error) {
	name = ComponentPrefix(name, "session")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	l, ok := c.(Session)
	if ok {
		return l, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Session type", name)
}


func NewSessionMap(interface{}) (Session, error) {
	return &SessionMap{
		KeyFync:	func(ctx Context) string {
			return ctx.GetCookie("sessionid")
		},
	}, nil
}


func SessionLoad(ctx Context) SessionData {
	return DefaultSession.SessionLoad(ctx)
}

func SessionSave(sess SessionData) {
	DefaultSession.SessionSave(sess)
}

func SessionRelease(id string) {
	DefaultSession.SessionRelease(id)
}



func (store *SessionMap) SessionLoad(ctx Context) SessionData {
	key := store.KeyFync(ctx)
	sess, ok := store.mem.Load(key)
	if ok {
		return sess.(SessionData)
	}
	return &SessionStd{Id: key}
}

func (store *SessionMap) SessionSave(sess SessionData) {
	store.mem.Store(sess.SessionID(), sess)
}

func (store *SessionMap) SessionRelease(id string) {
	store.mem.Delete(id)
}

func (*SessionMap) GetName() string {
	return ComponentSessionMapName
}

func (*SessionMap) Version() string {
	return ComponentSessionMapVersion
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
