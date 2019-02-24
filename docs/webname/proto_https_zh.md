# pprotocol	https

## golang https

实现思路使用标准库`crypto/tls`将net.Listener对象连接升级成tls连接。

### 标准库定义

```go
// https://golang.org/src/net/http/server.go?s=89338:89413#L2858
func (srv *Server) ServeTLS(l net.Listener, certFile, keyFile string) error {
	// Setup HTTP/2 before srv.Serve, to initialize srv.TLSConfig
	// before we clone it and create the TLS Listener.
	if err := srv.setupHTTP2_ServeTLS(); err != nil {
		return err
	}

	config := cloneTLSConfig(srv.TLSConfig)
	if !strSliceContains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	tlsListener := tls.NewListener(l, config)
	return srv.Serve(tlsListener)
}
```

golang ServeTLS的定义，先检查了http2设置，默认是环境变量开启。

然后配置了ssl信息，使用tls.NewListener将net.Listener转换成tls连接。

net.Listener的定义是一个interface，tls.NewListener函数返回一个net.Listener的实现，封装了一层tls处理。

最后将tls连接给http.Server启动服务。

### 配置https

```go
func ListenAndServeTLS(srv *http.Server, certFile, keyFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}	
	// 配置https
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// 连接ssl连接
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	// 启动监听
	return srv.Serve(tls.NewListener(ln, config))
}
```

### http2

golang开启http2有两种方法

- 使用net/http启动TLS连接就，会默认启动http2，GOEBUG参数控制。

- 使用golang.org/x/net/http2包

```go
func ListenAndServeTLS(srv *http.Server, certFile, keyFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}	
	// 配置http2
	if strings.Contains(os.Getenv("GODEBUG"), "http2server=0") {
		http2.ConfigureServer(srv, &http2.Server{})
	}
	// 配置https
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// 连接ssl连接
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	// 启动监听
	return srv.Serve(tls.NewListener(ln, config))
}
```

### 双向https

golang net/http不支持双向https，需要自己在tls基础上配置。

ClientAuthType声明服务器将遵循TLS客户端身份验证的策略。

```go
\\ ClientAuthType declares the policy the server will follow for TLS Client Authentication.
type ClientAuthType int
const (
        NoClientCert ClientAuthType = iota
        RequestClientCert
        RequireAnyClientCert
        VerifyClientCertIfGiven
        RequireAndVerifyClientCert
)
```

RequireAndVerifyClientCert就是双向https使用的策略。

双向配置在tls基础上加上ClientAuth和ClientCAs配置。

```go
func ListenAndServeMutualTLS(srv *http.Server, certFile, keyFile, trustFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}	
	// 配置http2，可选
	if strings.Contains(os.Getenv("GODEBUG"), "http2server=0") {
		http2.ConfigureServer(srv, &http2.Server{})
	}
	// 配置https
	config := &tls.Config{}
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// 配置双向https
	srv.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	pool := x509.NewCertPool()
	data, err := ioutil.ReadFile(trustFile)
	if err != nil {
		log.Println(err)
		return err
	}
	pool.AppendCertsFromPEM(data)
	srv.TLSConfig.ClientCAs = pool
	// 连接ssl连接
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	// 启动监听
	return srv.Serve(tls.NewListener(ln, config))
}
```


## https优化

https站点在线监测

[www.ssllabs.com][100]

[myssl.com][101]

### spdy and http2

搜索关键字`HTTP,HTTPS,SPDY,HTTP2.0`

http协议加上ssl协议就是https，2012年google提出一个https优化方案叫做spdy，然后最新http2就是参考了spdy，spdy相当于http2的过度方案，现在基本退出了，新的直接使用http2，在部分老的主机空间还能看到spdy协议传输的内容。

http2协议需要强制https，如果在http基础上构建http2就叫h2c，目前五大阅览器内核都不支持h2c，会协议退化成http1.1。

grpc协议是基于http2协议，grpc协议的传输协议是http2，序列化协议pb。

### rsa and ecc

rsa证书兼容性好，ecc证书体积小、速度快、加密强、兼容差。

目前最常用的密钥交换算法有 RSA 和 ECDHE：RSA 历史悠久，支持度好，但不支持 PFS（Perfect Forward Secrecy）；而 ECDHE 是使用了 ECC（椭圆曲线）的 DH（Diffie-Hellman）算法，计算速度快，支持 PFS。


#### nginx ssl配置

但从 Nginx 1.11.0 开始提供了对 RSA/ECC 双证书的支持。它的实现原理是：分析在 TLS 握手中双方协商得到的 Cipher Suite，如果支持 ECDSA 就返回 ECC 证书，否则返回 RSA 证书。

nginx配置指定多次证书即可。

```nginx
ssl_certificate     example.com.rsa.crt;
ssl_certificate_key example.com.rsa.key;

ssl_certificate     example.com.ecdsa.crt;
ssl_certificate_key example.com.ecdsa.key;
```
#### acme.sh签名

acme.sh是一个开源的ACME（ Automated Certificate Management Environment）协议客户端，可以自动签发Let's Encrypt的证书。

```bash
# first
~/.acme.sh/acme.sh --issue -w /data/web -d example.com -d www.example.com
~/.acme.sh/acme.sh --issue -w /data/web -d example.com -d www.example.com --keylength ec-256
# renew
~/.acme.sh/acme.sh --renew -d example.com -d www.example.com --force
~/.acme.sh/acme.sh --renew -d example.com -d www.example.com --force --ecc
```

-w是验证文件的目录，-d是单域名可以指定多个。

验证域名所有权时，acme.sh工具会在-d指定的目录下建立.well-known这个目录，里面就是验证附件。

然后验证服务器会访问对应的域名的.well-known路径下的数据，内容一致则验证成功，然后给予证书。

### CAA

CAA，全称Certificate Authority Authorization，即证书颁发机构授权。它为了改善PKI（Public Key Infrastructure：公钥基础设施）生态系统强度、减少证书意外错误发布的风险，通过DNS机制创建CAA资源记录，从而限定了特定域名颁发的证书和CA（证书颁发机构）之间的联系。


根据规范（RFC 6844），CAA记录格式由以下元素组成：

`CAA <flags> <tag> <value>`

&#60;flags>定义为0~255无符号整型，取值：

```
Issuer Critical Flag：0
1~7为保留标记
```

&#60;tag>定义为US-ASCII和0~9，取值：

```
CA授权任何类型的域名证书（Authorization Entry by Domain） : issue
CA授权通配符域名证书（Authorization Entry by Wildcard Domain） : issuewild
指定CA可报告策略违规（Report incident by IODEF report） : iodef
auth、path和policy为保留标签	
```

例如：

```
example.com.  CAA 0 issue "letsencrypt.org"
example.com.  CAA 0 issuewild "comodoca.com"
example.com.  CAA 0 iodef "mailto:example@example.com"
```

每一个域名CAA解析可以配置多条，issue是证书签发者，issuewild是通配符域名证书签发者，iodef是接收违规信息的邮箱。

支持CAA的DNS服务商： CloudXNS 、阿里云等

### HPKP

HPKP(HTTP Public Key Pinning)为公钥固定报告,相当于生成一个公钥hash值写入header里面，阅览器自动验证。

HPKP 官方文档见 RFC7469，它的基本格式如下：

Public-Key-Pins: pin-sha256="base64=="; max-age=expireTime [; includeSubdomains][; report-uri="reportURI"]

- pin-sha256 即证书指纹，允许出现多次（实际上最少应该指定两个），证书匹配其中一个即可；

- max-age 和 includeSubdomains 分别是过期时间和是否包含子域，它们在 HSTS（HTTP Strict Transport Security）中也有，格式和含义一致；

- report-uri 用来指定验证失败时的上报地址，格式和含义跟 CSP（Content Security Policy）中的同名字段一致；

- includeSubdomains 和 report-uri 两个参数均为可选；

#### bash生成

```bash
openssl x509 -in example.com.pem -noout -pubkey | openssl asn1parse -noout -inform pem -out public.key
openssl dgst -sha256 -binary public.key | openssl enc -base64
```

#### nginx配置

`add_header      Public-Key-Pins         'pin-sha256="sMU3CCjru4a49HAhlUSFaR1ryqFCVzv/eScJ9sE8jqY="; pin-sha256="1WDPq2eHdQ+RNNmbZCKIxy/0POuXu8Vbd6OfCy1N6aA="; max-age=2592000; includeSubDomains';
`

### CT

证书透明度

### OCSP装订



[1]: https://golang.org/pkg/crypto/tls/#ClientAuthType
[100]: https://www.ssllabs.com/ssltest/
[101]: https://myssl.com/
