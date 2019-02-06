# data storage

通常web用户数据存储有三种方式cookies、session、jwt

|  方案  |  cookies  |  session  |  jwt  |
| ------------ | ------------ | ------------ | ------------ |
|  传输体积 |  大 会传输全部cookie |  小 仅需sid |  中等 数据消息体 |
|  中心化 |  无 客户端存储 |  有 服务端存储 |  无 请求消息中 |
|  数据伪造 |  可以 发送自己想要的cookie |  不行 无法修改 | 几 乎不可能 需要破解hash |
|  敏感信息 |  泄露 暴露全部数据 |  不会泄露 |  泄露 |
|  通用 |  通用 一般阅览器不禁用cookie |  依赖cookie或url重写 |  同session |
|  服务端主动删除 |  不行 需要客户端下一次请求时删除 |  直接删除对应存储数据 |  不行 |
|  安全性 |  不安全 CSRF |  相对安全  |  相对安全 |

## cookies

Cookie是Web服务器利用set-cookie响应报头发送给客户端的一段消息。客户端在随后的请求中返回给服务器,服务器可以读取(而不可以改变)该消息。  

Cookie不会以任何方式得到解释和执行,以名-值对的显示保存消息,服务器通过再次发送修改后的cookie来改变cookie。 

访问对应的站点会自动发送cookie信息

浏览器对Cookie大小和数量有限制。


## Session

seesion是一种服务端数据存储方案,

### golang session

一组核心的session接口定义。[beego cache][2]

```go
type (
	Session interface {
		ID() string							// back current sessionID
		Set(key, value interface{}) error	// set session value
		Get(key interface{}) interface{}	// get session value
		Del(key interface{}) error			// delete session value
		Release(w http.ResponseWriter)		// release value, save seesion to store
	}
	type Provider interface {
		SessionRead(sid string) (Session, error)
	}
)
```

Provider.SessionRead	用sid来从Provider读取一个Session返回,sid就是sessionid一般存储与cookie中,也可以使用url参数值,Session对象会更具sid从对应存储中读取数据,然后将数据反序列化来初始化Session对象。

Session.Release			从名称上是释放这个Seesion,但是一般实际作用是将对应Session对象序列化,然后存储到对应的存储实现中,如果只是读取Session可以不Release。

## JWT

jwt全称Json Web Token

它定义了一种紧凑且独立的方式，可以将各方之间的信息作为JSON对象进行安全传输。该信息可以验证和信任，因为是经过数字签名的。

jwt构成由Header、Playload、Signature三部分组成,三部分数据拼接起来,中间以“.”连接的字符串就是一个JWT。

### 特点

jwt是无状态的,不需要服务端存储。

jwt是固定格式,客户端也可以解析出jwt数据,所以不能存储敏感信息。

由于客户端不清楚服务端使用的私钥或者独特的Hash哈希函数,客户端无法篡改JWT数据。

与ssl机制、ak机制相似,都使用公私钥签名数据结果,保证无法篡改数据。

### Header
声明加密的算法 通常直接使用 HMAC SHA256

### Playload

Playload就是自己数据json序列化后的base64编码的字符串

### Signature

Cookie: JWT-SESSION=


## HTML5 Web Storage

localStorage and sessionStorage是html5标准的存储方法,需要客户端对h5的支持。

sessionStorage: 仅在当前会话下有效，关闭页面或浏览器后被清除

localstorage: 除非被清除，否则永久保存

### sessionStorage数据跨标签共享

原理利用localStorage可以跨标签，触发一个localStorage事件，将sessionStorage数据临时写入到localStorage，其他标签从localStorage恢复sessionStorage数据，然后清理临时数据。

```js
"use strict";
void function(){
	if (!sessionStorage.length) {
		// 这个调用能触发目标事件，从而达到共享数据的目的
		localStorage.setItem('getSessionStorage', Date.now());
	};
	// 检查是否支持h5 storage
    if(window.localStorage){
        // 监听本地回话同步
		// 监听storage事件
        window.addEventListener("storage",function(event){
			// 触发getsessionStorage获取
            if(event.key =="getsessionStorage"){
				// 将本页面sessionStorage序列化存储到localStorage，并会触发其他页面的同步。
                localStorage.setItem("sessionStorage",JSON.stringify(sessionStorage));
				// 删除临时同步数据
                localStorage.removeItem("sessionStorage");
            }else if(event.key=="sessionStorage"&& !sessionStorage.length){
				// 将新的sessionStorage取出，并设置到本页面
                var data=JSON.parse(event.newValue);
                for(var key in data)
                    sessionStorage.setItem(key,data[key]);
            }
        });
    }
}();
```

[1]: https://golang.org/pkg/net/http/#Cookie

[2]: https://github.com/astaxie/beego/blob/master/session/session.go
