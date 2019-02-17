package jwt_test

import (
	"time"
	"eudore/middleware/jwt"
	"github.com/eudore/eudore"
)

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.Nv24hvNy238QMrpHvYw-BxyCp00jbsTqjVgzk81PiYA

func TestJwt(t *testing.T) {
	fn := jwt.NewVerifyHS256([]byte("secret"))
	t.Log(fn.SignedToken(map[string]interface{}{
		"uid": "1",
		"exp": time.Date(2022, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	}))
	t.Log(fn([]byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NjU0MDMyMDAsInVpZCI6IjEifQ")))
}