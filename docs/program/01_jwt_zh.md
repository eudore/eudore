# eudore使用jwt

具有方案两种，中间件加入数据和处理函数扩展。

使用[github.com/dgrijalva/jwt-go](https://github.com/dgrijalva/jwt-go)库。

## 中间件方式解析jwt

例如使用jwt加载认证信息。

首先MiddUidHandler函数初始化一个解析中间件，使用config拿到`*sql.DB`用来初始化`*sql.Stmt`对象，再获取jwt解析使用的secret，在该演示中国直接设置了secret的值，实际应该由config解析时设置进去。

在函数处理时，先拿Authorization Header,判断前缀是Bearer，然后再使用jwt库解析字符串，判断返回数据是否有效，有效就获得userid和name设置到Params里面。

如果没有Authorization Header,如果uri参数存在token，就拿该token字符串到数据库查询到userid和name保存给Params，再后续处理中直接使用参数读取数据即可。

该代码修改了jwt解析库未测试，ak解析和token解析类似,如果不使用token请删除其中代码。

```golang
func MiddUidHandler(app *eudore.App) (eudore.HandlerFunc, error) {
	db, ok := app.Config.Get("keys.db").(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("keys.db not find database.")
	}
	stmtQueryAccessKey, err := db.Prepare("SELECT userid,(SELECT name FROM tb_auth_user_info WHERE id = userid) FROM tb_auth_ram_access_key WHERE key=$1 and expires > now()")
	if err != nil {
		return nil, err
	}

	app.Config.Set("keys.secret", []byte("secret"))
	hmacSampleSecret, ok := app.Config.Get("keys.secret").([]byte)
	if !ok {
		return nil, fmt.Errorf("keys.secret not find.")
	}
	return func(ctx eudore.Context) {
		tokenString := ctx.GetHeader(eudore.HeaderAuthorization)
		if strings.HasPrefix(tokenString, "Bearer ") {
			token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
				return hmacSampleSecret, nil
			})
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx.SetParam("UID", eudore.GetString(claims["userid"]))
				ctx.SetParam("NAME", eudore.GetString(claims["name"]))
				return
			}
			ctx.Error("bearer authorization Header invalid")
		}

		tokenString = ctx.GetQuery("token")
		if tokenString != "" {
			var userId string
			var username string
			err := stmtQueryAccessKey.QueryRow(tokenString).Scan(&userId, &username)
			if err != nil {
				ctx.Error(err)
				return
			}
			ctx.SetParam("UID", userId)
			ctx.SetParam("NAME", username)
		}
	}, err
}

func maim() {
	app := eudore.NewCore()

	uid, err := MiddUidHandler(app.App)
	if err != nil {
		return err
	}
	app.AddMiddleware(eudore.MethodAny, "/api/v1/", uid)

	app.Get("/api/v1/*", func(ctx eudore) {
		fmt.Println(ctx.Params())
	})
	...
}

```


## 处理函数扩展1

原理给Context对象附加一个新的方法。

```golang
type ContextJwt eudore.Context 

func init() {
	eudore.RegisterHandlerFunc(func(fn func(ContextJwt)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(ContextJwt(jwt))
		}
	})
}

hmacSampleSecret := []byte("secret")

func (ctx ContextJwt) ParseJwt() map[string]interface{} {
	tokenString := ctx.GetHeader(eudore.HeaderAuthorization)
	token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})
	if claims, ok := token.Claims.(map[string]interface{}); ok && token.Valid {
		return claims
	}
	return nil
}

func main() {
	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx ContextJwt) {
		fmt.Println(ctx.ParseJwt())
	})
}
```


## 处理函数扩展2

**不推荐的方法，不可持续扩展，如果同时扩展sql.DB就不方法，请使用扩展1**

先注册一个扩展函数,扩展类型为`func(eudore.Context, jwt.MapClaims)`，先拿Authorization Header并解析成jwt，如果解析成功就调用注册的函数。

```golang
func init() {
	hmacSampleSecret := []byte("secret")
	eudore.RegisterHandlerFunc(func(fn func(eudore.Context, jwt.MapClaims)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			tokenString := ctx.GetHeader(eudore.HeaderAuthorization)
			token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
				return hmacSampleSecret, nil
			})
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				fn(ctx, claims)
			}else {
				ctx.Error("bearer authorization Header invalid")
			}
		}
	})
}
func main() {
	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx eudore.Context, data jwt.MapClaims) {
		fmt.Println(data)
	})
	...
}
```