# ACL

ACL 是 Access Control List 的缩写，称为访问控制列表，包含了对一个对象或一条记录可进行何种操作的权限定义。

例如：

```json
{
	"*": {
		"Read": true,
		"Any": false,
	},
	"admin": {
		"Any": true
	},
	"eudore": {
		"Read": true,
		"Create": true,
		"Update": true,
		"Delete": false,
		"Grant": false
	}
}
```