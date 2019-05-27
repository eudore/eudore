# Session

```golang
type (
	Session interface {
		Component
		SessionLoad(Context) SessionData
		SessionSave(SessionData)
		SessionFlush(string)
	}
	SessionData interface {
		Set(key string, value interface{}) error //set session value
		Get(key string) interface{}  //get session value
		Del(key string) error     //delete session value
		SessionID() string                //back current sessionID
	}
)
```