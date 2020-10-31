package main

/*
ListenTLS方法一般均默认开启了h2，如果需要仅开启https，需要手动listen监听然后使用app.Serve启动服务。
*/

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	createssl()
	defer os.Remove("ca.cer")
	defer os.Remove("server.key")
	defer os.Remove("server.cer")
	defer os.Remove("client.key")
	defer os.Remove("client.cer")
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug("istls:", ctx.Istls())
	})

	ln, err := (&eudore.ServerListenConfig{
		Addr:     ":8088",
		HTTPS:    true,
		HTTP2:    true,
		Mutual:   true,
		Keyfile:  "server.key",
		Certfile: "server.cer",
	}).Listen()
	if err == nil {
		app.Serve(ln)
	} else {
		app.Error(err)
	}

	ln, err = (&eudore.ServerListenConfig{
		Addr:      ":8088",
		HTTPS:     true,
		HTTP2:     true,
		Mutual:    true,
		Keyfile:   "server.key",
		Certfile:  "server.cer",
		Trustfile: "ca.cer",
	}).Listen()
	if err == nil {
		app.Serve(ln)
	} else {
		app.Error(err)
	}

	client := httptest.NewClient(app)
	tp, err := createtp()
	if err == nil {
		client.Client.Transport = tp
	} else {
		app.Options(err)
	}
	client.NewRequest("GET", "https://localhost:8088/").Do().CheckStatus(200).Out()

	// app.CancelFunc()
	app.Run()
}

func createtp() (*http.Transport, error) {
	pool := x509.NewCertPool()
	data, err := ioutil.ReadFile("/tmp/mca/ca.cer")
	pool.AppendCertsFromPEM(data)
	if err != nil {
		return nil, err
	}

	clientCrt, err := tls.LoadX509KeyPair("client.cer", "client.key")
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{clientCrt},
		},
	}, nil
}

/*
# 成服务端私钥
openssl genrsa -out server.key 2048
#生成服务端证书请求文件
openssl req -new -key server.key -out server.csr -subj "/C=CN/ST=BJ/L=beijing/O=eudore/OU=eudore/OU=eudore/CN=localhost"

# 生成客户端私钥
openssl genrsa -out client.key 2048
# 生成客户证书请求文件
openssl req -new -key client.key -out client.csr -subj "/C=CN/ST=BJ/L=beijing/O=eudore/OU=eudore/OU=eudore/CN=localhost"

# 生成根证书私钥
openssl genrsa -out ca.key 2048
# 生成根证书请求文件
openssl req -new -key ca.key -out ca.csr -subj "/C=CN/ST=BJ/L=beijing/O=eudore/OU=eudore/OU=eudore/CN=localhost"
# 生成自签名的根证书文件
openssl x509 -req -days 3650 -sha1 -extensions v3_ca -signkey ca.key -in ca.csr -out ca.cer
# 利用已签名根证书生成服务端证书和客户端证书
# 生成服务端证书
openssl x509 -req -days 3650 -sha1 -extensions v3_req -CA ca.cer -CAkey ca.key -CAcreateserial -in server.csr -out server.cer
# 生成客户端证书
openssl x509 -req -days 3650 -sha1 -extensions v3_req -CA ca.cer -CAkey ca.key -CAcreateserial -in client.csr -out client.cer

openssl x509 -in server.cer -text -noout 2>&1| head -n 15
*/

func createssl() {
	ioutil.WriteFile("ca.cer", []byte(`-----BEGIN CERTIFICATE-----
MIIDYjCCAkoCCQDkcCR+EmTT1jANBgkqhkiG9w0BAQUFADBzMQswCQYDVQQGEwJD
TjELMAkGA1UECAwCQkoxEDAOBgNVBAcMB2JlaWppbmcxDzANBgNVBAoMBmV1ZG9y
ZTEPMA0GA1UECwwGZXVkb3JlMQ8wDQYDVQQLDAZldWRvcmUxEjAQBgNVBAMMCWxv
Y2FsaG9zdDAeFw0yMDA0MjgxNTUzMjdaFw0zMDA0MjYxNTUzMjdaMHMxCzAJBgNV
BAYTAkNOMQswCQYDVQQIDAJCSjEQMA4GA1UEBwwHYmVpamluZzEPMA0GA1UECgwG
ZXVkb3JlMQ8wDQYDVQQLDAZldWRvcmUxDzANBgNVBAsMBmV1ZG9yZTESMBAGA1UE
AwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvG1a
zMYrF/suaHIy+vDqFMuv8hrw7Xv3NL29U2zpdB8CnORejTGzX09LZ4L/EK/bHHeo
TKj+2u6ZfOptVcy+mgPbRXw6oZbcNDSZKKfS8sz7VqdzFfyLkTnk5sjAmSqw2ang
yi5dio6wTLwi4JIJKwLRH7RNUaXL7RqW5WMssoU+J9gGgQX/wXzDSgW7HJ+BlKk7
9Tff6hGJqY/5wUXdbMPL1OxdxfBa6Eil4qxRkarDgR0i0ucg3g3citKmECwcSU78
J2Oo/evWs0ePfpXsM3mTbKcUSy7yfw8TS/NF+mCi5ZLtnjJJhLDHeD2XTXnAJu91
GFffbvLW3UjGtONTuQIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQAPLRIRaSFffYcz
Y0NCUky4MbI8egw3KtY43Kj+QN92yQMA+b6+NlLLmdgie0t2n7mw6cXJ2eMGPaJK
wINOv7CZ8ocfYp7ZtghcreCdf6TDsZEZ+7Nme3jnMM3J2t6b7ft4rbDK13m8HheZ
X+ko5g9XG+ilqVGupekWdCYamLlxbuYyZfeon5Gb+kWfFvVjXgF4oknfdXeJD7da
l7oIQ5UdiIUoAYUbdsP6+pCtW8oz2YDRVJbFcnNhZ1FvrZNhhKPy7K0fvTFf91nT
mjXWmb2mT7j+oZ4P84UFWZsTLgqZaCnY3x+9f7yh00cPGcLwdaC+UjmEsluGwtMG
+8lmw2Fo
-----END CERTIFICATE-----
`), 0644)
	ioutil.WriteFile("server.key", []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0zQ4iUpcInR2++dwEdTH/VGtFDXZEMKqZjdFuMQmqkB8+l7f
2zaTDAweVkaGW7dAWyGQFl/p5GYEFyzdjj5lKhP9bSClxGAFnDjQbP9MI4zKKhtW
qWXs1QASUhcM3NpaswUQlu3GtHuxkzZsJv+Y4wvx+6oEwrvqIfh6y+8IsAFHSDHw
c05TwYjE7l5sbc6tGCXplCosdTSGizRH5+rfh3pHyWbPV++RkId7PgDvGgodM73R
QPhOSF6U9e/dIDWpuB6pcmUV1KX/y3YDdaJAEEWDnlCVyUzIh8+iV/BbOb2fb7re
idKaQXp2XxGcyIAJuUamB0ORzNNyPbcy0CKYeQIDAQABAoIBAQC+R3n4FspTMGJS
KPzK080p6H/qiWj6tKcYWAu0nuPG1zrBxuAfe1eXrwNV70v1LXAJqn9J6Ter0k01
I/KnyIcUFdZojtVJysjDKlx9FrTeAmXQ8bht/aoVbG8VDjdEcmTsjE+Z6rbuu9IM
MyfVKsnvJD/q4A5R80LJQDhBqyVEiv8mg6+jDa+4tggWbt0q4ICozZVs26Y6rtFn
EsOAyTRPYYlLpv2WG7j5m61bDkBdaXpJP1JIK79eDAQrutXNM+34Lly/zeW51hOv
qKKQnmtqPQCSgAmQhPLa/9SWdPaTp4Holkuf/LRlkEpdSedKXRh2BKhD1w/N1AF5
nKi0bOYBAoGBAO+H7Ln73AU0d+2smd2VWBcZcPiVkkQpz9qd6c3VAT7PYyzsQcZi
YOEdxw+o78aKTjCuR0bRZJ5IvSmF5ItfQ4xUYPqvkh+VRsf6mMwL3+hNCC4wFmEd
AKJWfmuKyvu5hLncX9Y/H3ANFToZTjUCac75a5PQceFprIjbOZWdYDnJAoGBAOG5
s9Er7eprZTp5ZC9HMOdc1boyh/A+rSFZIQQS9il7w2eZ+Ve1gQblty5Qoh0d1amH
5sT0DSALPFXHAZ4zk1JR6TE0puvgqaXHwfYB2lzb0BP2gYmHqHxsGS49yDO1Bcwq
7rl1PaWkbaE4IWtT821cMrIHGBX2Hg1qnYPfb8ExAoGAZ01PosYkFXqTXkVZ9l46
J3wpZIvdENiXc8k21DZQ2y3Fr9IUa+JxtaSJ/Q72mcF8BzKiOsCDjGACdK3x6smi
8BpT2MlvU3+ljwlcbGOSpTTTmlfSzv8bDugOjYLGF9nii+Wmz1dZz5FU3kGboPDx
gPnAk3cKJhTU/BDPvN6qaUECgYA2I9tsVTQIYN/zyX/tEw84vvyIX2xZhD70W7Ne
jcm7I3M32yeCEQe0hs6L7k0j3K8NrYn9PWgUgn1jOYs6zbYNLZZX9f//XXBzUdlE
zyb31MUwtJRXT1FrHmZfv/PP6yBL2xRNKUCzBSBCZfsmCgm99jo2lxsA0Xpdz2+e
XK4qUQKBgHnBMaPFvIph2KLFgadYTKYs6b4maH4rCc0rrdRrsqCld+h2iN0TTiKf
d7I+u8v/415SnnwAp4HxpMCe1WgmoMotFPHrA3j9FqmQ74A3TzXV0BvasZK9Pd36
LBPlsSEa0bH2ZjPim2xir8m92MKT3npoe9Qu9U49ozFSNFbx9DeK
-----END RSA PRIVATE KEY-----
`), 0644)
	ioutil.WriteFile("server.cer", []byte(`-----BEGIN CERTIFICATE-----
MIIDYjCCAkoCCQDKHAIPMuQDNDANBgkqhkiG9w0BAQUFADBzMQswCQYDVQQGEwJD
TjELMAkGA1UECAwCQkoxEDAOBgNVBAcMB2JlaWppbmcxDzANBgNVBAoMBmV1ZG9y
ZTEPMA0GA1UECwwGZXVkb3JlMQ8wDQYDVQQLDAZldWRvcmUxEjAQBgNVBAMMCWxv
Y2FsaG9zdDAeFw0yMDA0MjgxNTUzMjdaFw0zMDA0MjYxNTUzMjdaMHMxCzAJBgNV
BAYTAkNOMQswCQYDVQQIDAJCSjEQMA4GA1UEBwwHYmVpamluZzEPMA0GA1UECgwG
ZXVkb3JlMQ8wDQYDVQQLDAZldWRvcmUxDzANBgNVBAsMBmV1ZG9yZTESMBAGA1UE
AwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0zQ4
iUpcInR2++dwEdTH/VGtFDXZEMKqZjdFuMQmqkB8+l7f2zaTDAweVkaGW7dAWyGQ
Fl/p5GYEFyzdjj5lKhP9bSClxGAFnDjQbP9MI4zKKhtWqWXs1QASUhcM3NpaswUQ
lu3GtHuxkzZsJv+Y4wvx+6oEwrvqIfh6y+8IsAFHSDHwc05TwYjE7l5sbc6tGCXp
lCosdTSGizRH5+rfh3pHyWbPV++RkId7PgDvGgodM73RQPhOSF6U9e/dIDWpuB6p
cmUV1KX/y3YDdaJAEEWDnlCVyUzIh8+iV/BbOb2fb7reidKaQXp2XxGcyIAJuUam
B0ORzNNyPbcy0CKYeQIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQCkXjuTGNtZH35s
Sp78q9Zzx9qRI6Qid/m+NRVrgiE5arzQSiAc+Stt0ucK8XPNYQnw5IfmTrG32TAZ
SdeKPgtMTY1JYkDAO+Z7m79DHgOKy+qBZwB6+aP3xw2jGZgYggPtC3w4k0cjOZsc
+t0o9du3/U1fmMPdrMJAihns23cPiPFd+8pPC9fnCRFfy2roUUFKoic+6T4tWBbM
wL5oxRH7f4om4vsY1Uhe6VUuXh1R05rlCnX9l/HPazaUj2zEGHh/Drxj6BcwSlOr
1lvoiFWZZMjGfKJ1/Fid9q+duwbrVZrQF1rg+pdjx+mk9ncpNW7WwPypxyaC8b0A
uZwwKuNX
-----END CERTIFICATE-----
`), 0644)
	ioutil.WriteFile("client.key", []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApd+aNQYV4DuWG9Sb9iA6wdEMvRG7tIjBv78Nm9BxZwHaVgwd
BEucdX1ieiyDSVJOVxBKjyGmehtE5VH1kAUHzln2TMKYj8cKb78BNmuU+VqV/C2G
VbuCtB3wT3IXcI2L4ZvM/9w3INZbxo1oTiTfj2wiyo1ZSXygPuWIk+ardb2/u/ZU
Dt6TqGD+EW/FTnoVSUo2qFGvAmrr4T22DUZVgFg2WcAR7+0x1XScIzjqR1gg5a79
pwa3wiiBd+mh+V4dPIH3YkAQtM+5QnOFDWOsfPMMXB5LRZN36Mqz77w8sXAqD10R
KY/Q66L0Br9xfiDs/WbJFcif3Y/btLCKMRXILQIDAQABAoIBADs3C/IJ7h1SqE/f
Ip5G+zLd0lJc1kmo2KH/LniFfTZsrukxAdras0wuKs26vlOakmT6Z+OY+7lzqrDD
BYsYgKTl8MuOXLBXOh6SbXhkB5bNA+Y2ylIo0oxCc9uouz1vCpTL7e8ZSoTqgXDs
YmQjPbwRuonc1Bcr6nkJsCw8mNE7DOpmZL72cuO0kKR7NFh2qhKZI856JBxRc5TZ
VjyHgO7iRLHQmFhigAkDYt/nf/us3rTa2igPoxtoP8NFoEyY8JMyuG8WchAU3+ZV
YP8/K21FTrvWkdBI/NyVKDpfUtgW498+mdD4YCu/FLR06Ydof1y9n5pw0DyTIsmq
SjTjCHECgYEAzn2GLiYCwOYehXPzYKcfevAsJz3XbB/RKBV/yQ/9FBAkIV6dQpbI
zEHei5yagYm+D/oOzzvdsb4IA/Dzj4z6L4sfK2zivucSz7DTmLvnczwSNRlZi5N2
8DiRdr5CickAEHvTEcgJEkI2ygAst9Re375W3qNVgC5/1aiUTrozymcCgYEAzaT+
lDv7tcK/QBlZ+3uCURD/AJUsi2HRTsdEeqWDDq9TbLEFuWjeGbtKTZcX6SpbO5NU
CRijO7ro6fI4x4MBQyB6LEW8MkNPlGeX4Imuhxy3vhTin9hVjCKy8w4dmOqFELY4
W7yKRwrFZwsmrwYhZGtB2f1ST+ay5FZJGJcVJEsCgYAixlb7nKEoFVkchnt9Uofl
r17wOOT3q6AQzRYZKV0orNM433NCjJxCcfFlt7j5idX9YNJvqhha37L/3utVyJs1
uItGR+8j0UyEt7Xa6gI/kOVMFfnTnMESEaTFx6LzC2u8Wu4f9303mvkZKdBeISDd
M3PzyLQUg0A6Hkrju04PjwKBgQC9Y/hGAtw1sG68lNyHPF9vU4zWN4x3rZW7zM9n
eOkzbAsT7hCMimUKI7AxtzaBOc4eFvhtDDDBQMljM/5Q2HkgHlgGUA8b51vyHFoG
pCaFLtCWEdwJRI686fQO3vApNcto8bkD26cp+GSHGwD8blPwjMtv/NqC1b/phQH6
0KHa8wKBgA6tSny2w/V4KuMmOeLFRh+NrdxLrtQl7zasaGcXYg3bgKeNeST565u1
IMoUnpMmCrDHeGZcfAkmZ4XDEXWn4gGIybQuqsJ1znfL6i0S68X/IlLyAuuz8D5Z
q9GrgvzRgNgBZxSZagF0Xjk+wjvDiqe4N3kPYu2qxcpg555kaJgX
-----END RSA PRIVATE KEY-----
`), 0644)
	ioutil.WriteFile("client.cer", []byte(`-----BEGIN CERTIFICATE-----
MIIDYjCCAkoCCQDKHAIPMuQDNTANBgkqhkiG9w0BAQUFADBzMQswCQYDVQQGEwJD
TjELMAkGA1UECAwCQkoxEDAOBgNVBAcMB2JlaWppbmcxDzANBgNVBAoMBmV1ZG9y
ZTEPMA0GA1UECwwGZXVkb3JlMQ8wDQYDVQQLDAZldWRvcmUxEjAQBgNVBAMMCWxv
Y2FsaG9zdDAeFw0yMDA0MjgxNTUzMjdaFw0zMDA0MjYxNTUzMjdaMHMxCzAJBgNV
BAYTAkNOMQswCQYDVQQIDAJCSjEQMA4GA1UEBwwHYmVpamluZzEPMA0GA1UECgwG
ZXVkb3JlMQ8wDQYDVQQLDAZldWRvcmUxDzANBgNVBAsMBmV1ZG9yZTESMBAGA1UE
AwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApd+a
NQYV4DuWG9Sb9iA6wdEMvRG7tIjBv78Nm9BxZwHaVgwdBEucdX1ieiyDSVJOVxBK
jyGmehtE5VH1kAUHzln2TMKYj8cKb78BNmuU+VqV/C2GVbuCtB3wT3IXcI2L4ZvM
/9w3INZbxo1oTiTfj2wiyo1ZSXygPuWIk+ardb2/u/ZUDt6TqGD+EW/FTnoVSUo2
qFGvAmrr4T22DUZVgFg2WcAR7+0x1XScIzjqR1gg5a79pwa3wiiBd+mh+V4dPIH3
YkAQtM+5QnOFDWOsfPMMXB5LRZN36Mqz77w8sXAqD10RKY/Q66L0Br9xfiDs/WbJ
Fcif3Y/btLCKMRXILQIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQClTFnVBwsr+xma
HERTdCd6DsF94H/mEhOexa7zpf9jMCCKoxFHJB3GNoPJazH+4hdYQG09pLyah5bE
tofp/+rI4XeECnQgadM/BeUsWwF6Qhl62o2uhUK/LeqZ8bfwPGlxQA+EshVOLrUN
XyowhwY4QgGVD11EhbenDhzlTvBPoWT2aTofIEJAsoSRMJyxsL19GwWzoqoTk7et
SgGnwIAFBueyDD/6nI79D78s9cjzFSAO8EWZV4NZU2G4wBZm2Mwe8Lp1g5Kn/KDr
/SvgvHfzw5J0T3Jas0fkI+adtFa0+7AoHe9zFcmI/NnOFy0J1MyViCOQhrJstMGl
C5e20aHd
-----END CERTIFICATE-----
`), 0644)
}
