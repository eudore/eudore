package denials

import (
	"github.com/eudore/eudore"
	"net/http"
	"time"
)

const (
	denialsPrefix = "m:denials:"
	HeaderDeny    = "Deny"
)

type Denials struct {
	cache eudore.Cache
	t     time.Duration
}

func NewDenials(cache eudore.Cache, t time.Duration) *Denials {
	return &Denials{
		cache: cache,
		t:     t,
	}
}

func (d *Denials) Handle(ctx eudore.Context) {
	clientIP := ctx.RealIP()
	if d.cache.IsExist(denialsPrefix + clientIP) {
		ctx.Info("denials: " + clientIP)
		ctx.WriteHeader(http.StatusTeapot)
		ctx.End()
		return
	}
	ctx.Next()
	if len(ctx.GetHeader(HeaderDeny)) != 0 {
		d.cache.Set(denialsPrefix+clientIP, 0, d.t)
		ctx.Info("denials add " + clientIP)
	}
}
