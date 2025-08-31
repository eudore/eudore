package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

type bearer struct {
	Signing     signingMethod
	GetKeyFunc  func(ctx eudore.Context) string
	PayloadFunc func(ctx eudore.Context, str []byte)
	head        string
	key         any
}

// SigningMethod can be used add new methods for signing or verifying tokens. It
// takes a decoded signature as an input in the Verify function and produces a
// signature in Sign. The signature is then usually base64 encoded as part of a
// JWT.
//
// From: https://pkg.go.dev/github.com/golang-jwt/jwt/v5#SigningMethod
type signingMethod interface {
	Verify(signingString string, sig []byte, key any) error // Returns nil if signature is valid
	Alg() string                                            // returns the alg identifier for this method (example: 'HS256')
}

type signatureUser struct {
	Userid     json.Number `json:"userid"`
	Username   string      `json:"username,omitempty"`
	NotBefore  int64       `json:"nbf,omitempty"`
	Expiration int64       `json:"exp,omitempty"`
}

// NewBearerAuthFunc function creates middleware to implement parsing
// JWT Bearer Tokens, supporting custom signing methods and parsing logic.
//
// This middleware extracts the Bearer Token from the [eudore.HeaderAuthorization]
// field in the HTTP request header, verifies its signature, parses the payload,
// and sets the user information to [eudore.ParamUserid] and [eudore.ParamUsername] .
//
// If the token does not exist or its format does not conform to the Bearer std,
// the process is skipped. Other middleware can be used to implement
// Bearer data format checks. If parsing fails, a [eudore.StatusUnauthorized] response is returned.
//
// The parameter key is the JWT key, with built-in HS256 parsing by default;
// it is compatible with [github.com/golang-jwt/jwt/v5.SigningMethod] as the parsing method.
//
//	type signatureUser struct {
//		Userid     json.Number `json:"userid"`
//		Username   string      `json:"username,omitempty"`
//		NotBefore  int64       `json:"nbf,omitempty"`
//		Expiration int64       `json:"exp,omitempty"`
//	}
//
// options: [NewOptionKeyFunc]
// [NewOptionBearerSignaturer] [NewOptionBearerPayload].
//
// RFC 6750: The OAuth 2.0 Authorization Framework: Bearer Token Usage
//
// RFC 7519: JSON Web Token (JWT).
func NewBearerAuthFunc(key any, options ...Option) Middleware {
	b := &bearer{
		Signing: newSigningMethodHS256(key),
		GetKeyFunc: func(ctx eudore.Context) string {
			return ctx.GetHeader(eudore.HeaderAuthorization)
		},
		key: key,
	}
	applyOption(b, options)
	b.head = base64Encoding.EncodeToString([]byte(`{"alg":"` + b.Signing.Alg() + `","typ":"JWT"}`))

	return func(ctx eudore.Context) {
		token := b.GetKeyFunc(ctx)
		if len(token) < 8 || token[:7] != valueBearerAuth {
			return
		}

		user, err := b.Parse(ctx, token[7:])
		if err != nil {
			msg := "Bearer error=\"invalid_token\", error_description=\"" + err.Error() + "\""
			ctx.SetHeader(eudore.HeaderWWWAuthenticate, msg)
			writePage(ctx, eudore.StatusUnauthorized, DefaultPageBearerAuth, err.Error())
			return
		}
		ctx.SetParam(eudore.ParamUserid, user.Userid.String())
		ctx.SetParam(eudore.ParamUsername, user.Username)
	}
}

func (b *bearer) Parse(ctx eudore.Context, token string) (*signatureUser, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != b.head {
		return nil, ErrBearerTokenInvalid
	}

	payload, err := base64Encoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	sig, err := base64Encoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	err = b.Signing.Verify(token[:len(token)-len(parts[2])-1], sig, b.key)
	if err != nil {
		return nil, err
	}

	var user signatureUser
	err = json.Unmarshal(payload, &user)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if user.NotBefore != 0 && user.NotBefore > now {
		return nil, fmt.Errorf(ErrBearerTokenNotValid, user.NotBefore)
	}
	if user.Expiration != 0 && user.Expiration < now {
		return nil, fmt.Errorf(ErrBearerTokenExpired, user.Expiration)
	}
	if b.PayloadFunc != nil {
		b.PayloadFunc(ctx, payload)
	}
	return &user, nil
}

// The NewOptionBearerSignaturer function creates BearerAuth custom signing method [Option].
//
// The signaturer parameter must implement the [SigningMethod] interface.
//
// You can directly use the signing method defined in [github.com/golang-jwt/jwt], for example:
//
//	NewOptionBearerSignaturer(jwt.SigningMethodHS256)
//
// From: https://pkg.go.dev/github.com/golang-jwt/jwt/v5#SigningMethod
//
//	type SigningMethod interface {
//		Verify(signingString string, sig []byte, key any) error // Returns nil if signature is valid
//		Alg() string                                            // returns the alg identifier for this method (example: 'HS256')
//	}
func NewOptionBearerSignaturer(signaturer any) Option {
	return func(data any) {
		v1, ok1 := data.(*bearer)
		v2, ok2 := signaturer.(signingMethod)
		if ok1 && ok2 {
			v1.Signing = v2
		}
	}
}

// The NewOptionBearerPayload function creates BearerAuth payload and parses [Option].
//
// If payload parsing succeeds, parse the payload into type *T and save it to
// [eudore.Context.SetValue(key, *T)].
func NewOptionBearerPayload[T any](key any) Option {
	return func(data any) {
		b, ok := data.(*bearer)
		if ok {
			b.PayloadFunc = func(ctx eudore.Context, payload []byte) {
				val := new(T)
				err := json.Unmarshal(payload, val)
				if err != nil {
					return
				}

				ctx.SetValue(key, val)
			}
		}
	}
}

func newSigningMethodHS256(data any) signingMethod {
	var key []byte
	switch val := data.(type) {
	case []byte:
		key = val
	case string:
		key = []byte(val)
	}

	return signingHmac{
		pool: &sync.Pool{
			New: func() any {
				return hmac.New(sha256.New, key)
			},
		},
	}
}

type signingHmac struct {
	pool *sync.Pool
}

func (fn signingHmac) Verify(signingString string, sig []byte, _ any) error {
	h := fn.pool.Get().(hash.Hash)
	defer fn.pool.Put(h)

	h.Reset()
	h.Write([]byte(signingString))
	if !hmac.Equal(sig, h.Sum(nil)) {
		return ErrBearerSignatureInvalid
	}
	return nil
}

func (fn signingHmac) Alg() string {
	return "HS256"
}
