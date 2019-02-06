package denials

import (
	"time"
	"net/http"
	"eudore"
)

const (
	denialsPrefix 	=	"m:denials:"
	HeaderDeny		=	"Deny"
)

type Denials struct {
	cache eudore.Cache
	t time.Duration
}

func NewDenials(cache eudore.Cache, t time.Duration) *Denials{
	return &Denials{
		cache:	cache,
		t:		t,
	}
}

func (d *Denials) Handle(ctx eudore.Context) {
	if d.cache.IsExist(denialsPrefix + ctx.RemoteAddr()) {
		ctx.Info("denials: " + ctx.RemoteAddr())
		ctx.WriteHeader(http.StatusTeapot)
		ctx.End()
		return
	}
	ctx.Next()
	if len(ctx.GetHeader(HeaderDeny)) != 0 {
		d.cache.Set(denialsPrefix + ctx.RemoteAddr(), 0, d.t)
		ctx.Info("denials add " + ctx.RemoteAddr())
	}
}
