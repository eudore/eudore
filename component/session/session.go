package session

import (
	"crypto/md5"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
}

// ErrDataNotFound 定义Store未获得数据的error。
var ErrDataNotFound = errors.New("data not found")

type (
	// Session 定义会话管理对象。
	Session interface {
		GetSessionID(eudore.Context) string
		SessionLoad(eudore.Context) (map[string]interface{}, error)
		SessionSave(eudore.Context, map[string]interface{}) error
		SessionFlush(eudore.Context) error
	}
	// Store 定义存储对象接口。
	Store interface {
		Insert(string) error
		Delete(string) error
		Update(string, map[string]interface{}) error
		Select(string) (map[string]interface{}, error)
		Clean(time.Time) error
	}
	// sessionStd 定义默认使用的Session。
	sessionStd struct {
		Store
		Maxage  int
		KeyFunc func(eudore.Context) string `alias:"keyfunc"`
		SetFunc func(eudore.Context, string, int)
	}
	// StoreMap 使用sync.Map实现的Store。
	StoreMap struct {
		data sync.Map
	}
	storeMapKey struct {
		data map[string]interface{}
		time time.Time
	}
	// ContextSession 是使用Session实现的Context扩展。
	ContextSession struct {
		eudore.Context
		Session
	}
)

// NewSessionMap 创建一个SessionMap，使用sync.Map保存数据。
func NewSessionMap() Session {
	return NewsessionStd(NewStoreMap())
}

// NewStoreMap 函数创建一个map存储。
func NewStoreMap() Store {
	return &StoreMap{}
}

// NewsessionStd 创建一个使用Store为存储的Session对象。
func NewsessionStd(store Store) Session {
	if store == nil {
		return nil
	}
	session := &sessionStd{
		Store:  store,
		Maxage: 3600,
		KeyFunc: func(ctx eudore.Context) string {
			return ctx.GetCookie("sessionid")
		},
		SetFunc: func(ctx eudore.Context, key string, age int) {
			ctx.SetCookieValue("sessionid", key, age)
		},
	}
	go session.Clean()
	return session
}

// Clean 方法执行清理存储过期内容。
func (session *sessionStd) Clean() {
	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		session.Store.Clean(time.Now().Add(time.Duration(-1*session.Maxage) * time.Second))
	}
}

// GetSessionID 方法获取请求上下文的sessionid。
func (session *sessionStd) GetSessionID(ctx eudore.Context) string {
	key := session.KeyFunc(ctx)
	if key == "" {
		return newSessionID()
	}
	return key
}

// SessionLoad 方法实现加载一个会话数据，
func (session *sessionStd) SessionLoad(ctx eudore.Context) (map[string]interface{}, error) {
	key := session.GetSessionID(ctx)
	data, err := session.Store.Select(key)
	if err == nil {
		return data, nil
	}
	if err != ErrDataNotFound {
		return nil, err
	}
	err = session.Store.Insert(key)
	if err != nil {
		return nil, err
	}
	session.SetFunc(ctx, key, session.Maxage)
	return make(map[string]interface{}), nil
}

// SessionSave 方法实现将一个会话数据保存。
func (session *sessionStd) SessionSave(ctx eudore.Context, data map[string]interface{}) error {
	session.SetFunc(ctx, session.GetSessionID(ctx), session.Maxage)
	return session.Store.Update(session.GetSessionID(ctx), data)
}

// SessionFlush 方法实现使用一个sessionid删除一个会话数据。
func (session *sessionStd) SessionFlush(ctx eudore.Context) error {
	session.SetFunc(ctx, session.GetSessionID(ctx), -1)
	return session.Store.Delete(session.GetSessionID(ctx))
}

func newSessionID() string {
	nano := time.Now().UnixNano()
	rand.Seed(nano)
	rndNum := rand.Int63()
	return md5String(strconv.FormatInt(nano, 10) + md5String(strconv.FormatInt(rndNum, 10)))
}

func md5String(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

// Insert 方法创建一个新的会话数据。
func (store *StoreMap) Insert(key string) error {
	store.data.Store(key, storeMapKey{data: make(map[string]interface{}), time: time.Now()})
	return nil
}

// Delete 方法删除一个数据。
func (store *StoreMap) Delete(key string) error {
	store.data.Delete(key)
	return nil
}

// Update 方法保存数据。
func (store *StoreMap) Update(key string, val map[string]interface{}) error {
	store.data.Store(key, storeMapKey{data: val, time: time.Now()})
	return nil
}

// Select 方法加载数据。
func (store *StoreMap) Select(key string) (map[string]interface{}, error) {
	data, ok := store.data.Load(key)
	if ok {
		return data.(storeMapKey).data, nil
	}
	return nil, ErrDataNotFound
}

// Clean 方法清理map过期数据。
func (store *StoreMap) Clean(expires time.Time) error {
	store.data.Range(func(key, value interface{}) bool {
		val, ok := value.(storeMapKey)
		if ok && expires.After(val.time) {
			store.data.Delete(key)
		}
		return true
	})
	return nil
}

// NewExtendContextSession 转换ContextSession处理函数为Context处理函数。
func NewExtendContextSession(fn func(ContextSession)) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		fn(ContextSession{Context: ctx})
	}
}

// DeleteSession 方法删除当前会话数据
func (ctx ContextSession) DeleteSession() error {
	err := ctx.Session.SessionFlush(ctx.Context)
	if err != nil {
		ctx.Error(err)
	}
	return err
}

// GetSession 获取会话数据。
func (ctx ContextSession) GetSession() map[string]interface{} {
	data, err := ctx.Session.SessionLoad(ctx.Context)
	if err != nil {
		ctx.Error(err)
	}
	return data
}

// SetSession 方法设置当前会话的数据
func (ctx ContextSession) SetSession(data map[string]interface{}) error {
	err := ctx.Session.SessionSave(ctx.Context, data)
	if err != nil {
		ctx.Error(err)
	}
	return err
}

// GetSessionBool 方法获取会话数据转换成bool。
func (ctx ContextSession) GetSessionBool(key string) bool {
	return eudore.GetDefaultBool(ctx.GetSession()[key], false)
}

// GetSessionDefaultBool 方法获取会话数据转换成bool，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultBool(key string, b bool) bool {
	return eudore.GetDefaultBool(ctx.GetSession()[key], b)
}

// GetSessionInt 方法获取会话数据转换成int。
func (ctx ContextSession) GetSessionInt(key string) int {
	return eudore.GetDefaultInt(ctx.GetSession()[key], 0)
}

// GetSessionDefaultInt 方法获取会话数据转换成int，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultInt(key string, i int) int {
	return eudore.GetDefaultInt(ctx.GetSession()[key], i)
}

// GetSessionFloat32 方法获取会话数据转换成float32。
func (ctx ContextSession) GetSessionFloat32(key string) float32 {
	return eudore.GetDefaultFloat32(ctx.GetSession()[key], 0)
}

// GetSessionDefaultFloat32 方法获取会话数据转换成float32，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultFloat32(key string, f float32) float32 {
	return eudore.GetDefaultFloat32(ctx.GetSession()[key], f)
}

// GetSessionFloat64 方法获取会话数据转换成float64。
func (ctx ContextSession) GetSessionFloat64(key string) float64 {
	return eudore.GetDefaultFloat64(ctx.GetSession()[key], 0)
}

// GetSessionDefaultFloat64 方法获取会话数据转换成float64，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultFloat64(key string, f float64) float64 {
	return eudore.GetDefaultFloat64(ctx.GetSession()[key], f)
}

// GetSessionString 方法获取会话数据转换成string。
func (ctx ContextSession) GetSessionString(key string) string {
	return eudore.GetDefaultString(ctx.GetSession()[key], "")
}

// GetSessionDefaultString 方法获取会话数据转换成string，转换失败返回默认值。
func (ctx ContextSession) GetSessionDefaultString(key, str string) string {
	return eudore.GetDefaultString(ctx.GetSession()[key], str)
}
