# XSS

跨站脚本攻击

用户将恶意代码植入到页面中。

## 防范

### 转义

### CSP

使用CSP减少XSS攻击。

CSP设置禁止内联js和设置js加载来源，js等静态全部写入单独文件。

页面仅保留html，静态资源使用[SRI][2]保证内容正确。

webserver添加header或者html页面添加meta实现。

`<meta http-equiv="Content-Security-Policy" content="script-src 'self'">`

[2]: https://developer.mozilla.org/zh-CN/docs/Web/Security/%E5%AD%90%E8%B5%84%E6%BA%90%E5%AE%8C%E6%95%B4%E6%80%A7
