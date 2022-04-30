# policy based access control

Package policy 实现基于策略访问控制。

pbac 通过策略限制访问权限，每个策略拥有多条描述，按照顺序依次匹配，命中则执行effect。

pbac条件直接使用And关系,允许使用多种多样的方法限制请求，额外条件可以使用policy.RegisterCondition函数注册条件。

如果一个策略Statement的Data属性不为空，则为数据权限，在没有匹配到一个非数据权限时会通过鉴权，保存多数据权限的Data Expression，用户对指定表操作时生产对应的表数据限定sql。

- Policy
	- [Pbac](../_example/policyPbac.go)
	- [Rbac](../_example/policyRbac.go)
	- [数据权限](../_example/policyData.go)
	- [策略限制条件](../_example/policyCondition.go)
	- [策略数据表达式](../_example/policyExpression.go)

## PolicyExample

```json
{
	"policy_id": 1,
	"policy_name": "PolicyExample",
	"description": "show polict all example",
	"statement": [
		{
			"effect": true,
			"action": ["*:*:Get*"],
			"resource": ["users/*",	"groups/*"],
			"conditions": {
				"and": {},
				"or": {},
				"method": ["GET", "POST", "OPTIONS"],
				"sourceip": ["127.0.0.1", "192.168.0.0/24"],
				"time": {"before": "2020-01-01", "after": "2222-01-01"},
				"params": {
					"user_id": ["1", "2"],
					"group_id": ["1"]
				}
			},
			"data": [
				{"kind": "and",	"data": []},
				{"kind": "or",	"data": []},
				{"kind": "value", "name": "user_id", "value": ["value:param:Userid"]},
				{"kind": "value", "name": "user_id", "not": true, "value": ["value:param:Userid"]},
				{"kind": "range", "name": "group_id", "min": "1", "max": "4"},
				{"kind": "sql", "name": "group_id", "sql": "group_id in %s", "value": ["1", "3"]}
			],
			"sql": [
				{
					"schema":"",
					"table":"",
					"disable": ["password","salf"],
					"conditions": [
						{"expr":"creat_time<now()-3d"},
						{"expr":"user_id=?", "values": ["value:param:Userid"] }
					]
				},
				{
					"schema":"",
					"table":"Accopt:UserMenu",
					"conditions": [{"expr": "MenuName='Home'"} ]
				}
			],
			"fitle": [
				{
					"package":"",
					"name":"",
					"rows":[
						{"creat_time": "now-3d", } 
					],
					"cloumns": [
						{"field": "email", "action": "zero", }
					]
				}
			]
		}
	]
}
```
