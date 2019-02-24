# Protocol

- tcp
tcp是协议相关的基础，以下协议都是基于tcp协议实现的
- [http][http]
http是基于tcp实现的文本协议
- [https][https]
https是安全的hhtp协议，基于TLS/SSl协议加密
- http2
http2以google的spdy协议为草案，实现更加高效http协议
- websocket
[websocket][ws]是基于http协议Upgrade机制升级的一种其他协议，可以从http协议Upgrade成websocket协议
- cgi
cgi是用于数据交换，其扩展版本fastcgi、uwsgi等

# Storage

web通常的数据存储方式，相互对比[在此][storage]。

- [cookie][cookie]
- seesion
- jwt
- h5store

# Sso

单点登录方案。

- oauth
- openid
- saml

# Security

web安全相关漏洞和技术

- [xss][xss]
- [csrf][csrf]
- [cors][cors]
- sci
- csp

# Auth

权限认证技术。

- [acl][acl]
- [rbac][rbac]
- pbac
- abac

# Other

https
- rsa and ecc
- caa
- ct
- hpkp
- oscp

http2
push

[http]: proto_http_zh.md
[https]: proto_https_zh.md
[ws]: proto_websocket_zh.md
[storage]: storage_zh.md
[cookie]: http_cookie_zh.md

[xss]: http_xss_zh.md
[csrf]:http_csrf_zh.md
[cors]: http_cors_zh.md

[acl]: ram_acl_zh.md
[rbac]: ram_rbac_zh.md
