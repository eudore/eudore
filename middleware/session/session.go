package session

import (
	"time"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/util/cast"
)

const (
	SessionId		=	"sessionid"
)

type (
	SessionProvider struct {
		eudore.Cache
	}
)

func NewSession(c eudore.Cache) *SessionProvider {
	return &SessionProvider{Cache:	c}
}

func (sp *SessionProvider) Handle(ctx eudore.Context) {
	sid := ctx.GetCookie(SessionId)
	sess := sp.Get(sid)
	data, ok := sess.(map[string]interface{})
	if ok {
		ctx.SetValue(eudore.ValueSession, data)
		ctx.Next()
		sp.Set(sid, data, 3600 * time.Second)
	}else {
		ctx.Next()
		sess = ctx.Value(eudore.ValueSession)
		if sess != nil {
			sp.Set(ctx.RequestID(), sess, 3600 * time.Second)
			ctx.SetCookieValue(SessionId, ctx.RequestID(), 3600)
		}
	}
}


func SessionStart(ctx eudore.Context) cast.Map {
	return cast.NewMap(ctx.Value(eudore.ValueSession))
}

func SessionRelease(ctx eudore.Context, sess map[string]interface{}) {
	ctx.SetValue(eudore.ValueSession, sess)
}
