package jwt

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/eudore/eudore"
)

const (
	BearerStar		=	"Bearer "
)

type (
	VerifyFunc func([]byte) string	
)

func NewVerifyHS256(secret []byte) VerifyFunc {
	return func(b []byte) string {
		h := hmac.New(sha256.New, secret)
		h.Write(b)
		return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	}
}

func NewJwt(fn VerifyFunc) eudore.HandlerFunc {
	if fn == nil {
		fn = NewVerifyHS256([]byte("secret"))
	}
	return func(ctx eudore.Context) {
		jwtstr := ctx.GetHeader(eudore.HeaderAuthorization)
		if len(jwtstr) == 0 {
			return
		}
		if strings.HasPrefix(jwtstr, BearerStar) {
			jwt, err := fn.ParseToken(jwtstr[7:])
			if err != nil {
				ctx.WithField("error", "jwt invalid").Warning(err)
				return
			}
			if int64(jwt["exp"].(float64)) < time.Now().Unix() {
				ctx.Warning("jwt expirese")
				return
			}
			ctx.SetValue(eudore.ValueJwt, jwt)
		}else {
			ctx.WithField("error", "bearer invalid").Warning("")	
		}
	}
}

func (fn VerifyFunc) SignedToken(claims map[string]interface{}) string {
	payload, _ := json.Marshal(claims)
	var unsigned string = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.` + base64.RawURLEncoding.EncodeToString(payload)
	return fmt.Sprintf("%s.%s",  unsigned, fn([]byte(unsigned)))
}

func (fn VerifyFunc) ParseToken(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("Error: incorrect # of results from string parsing.")
	}

	if fn([]byte(parts[0] + "." + parts[1])) != parts[2] {
		return nil, errors.New("Error：jwt validation error.")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	//
	dst := make(map[string]interface{})
	err = json.Unmarshal(payload, &dst)
	if err != nil {
		return nil, err
	}
	return dst, nil
}