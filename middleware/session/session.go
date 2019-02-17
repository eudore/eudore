package session

import (
	"time"
	"github.com/eudore/eudore"
)

const (
	SessionId		=	"sessionid"
)

func SessionFunc(ctx eudore.Context) {
	sid := ctx.GetCookie(SessionId)
	sess := ctx.App().Cache.Get(sid)
	data, ok := sess.(map[string]interface{})
	if ok {
		ctx.SetValue(eudore.ValueSession, data)
		ctx.Next()
		ctx.App().Cache.Set(sid, data, 3600 * time.Second)
	}else {
		ctx.Next()
		sess = ctx.Value(eudore.ValueSession)
		if sess != nil {
			ctx.App().Cache.Set(ctx.RequestID(), sess, 3600 * time.Second)
			ctx.SetCookieValue(SessionId, ctx.RequestID(), 3600)
		}
	}
}