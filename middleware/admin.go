package middleware

import "time"

var now = time.Now()

// AdminStatic 定义admin.html内容
var AdminStatic = `
<!DOCTYPE html>
<html>
<head>
	<title>Eudore Admin</title>
	<meta charset="utf-8">
	<meta name="author" content="eudore">
	<meta name="referrer" content="always">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="Eudore admin manage all eudore web ui">
	<style type="text/css">
	*{margin: 0px;padding: 0px;} 
	html,body,iframe{width: 100%; height: 100%;}
	#nav{display: block;width: 100%;height: 40px;background-color: #000;}
	#nav>ul{max-width: 960px;margin: auto;}
	#nav li{height: 100%;line-height: 40px;float: left;list-style: none;padding: 0 10px;}
	#nav li a:link{color: #ccc;text-decoration:none;}
	#eudore-app{height: calc(100% - 40px);max-width:960px;margin:auto;}
	#eudore-app>div{margin:5px}
	input{height: 24px; padding: 6px 16px; margin: 4px; font-size: 14px; line-height: 1.42857143; color: #2e2e2e; background-color: #fff; background-image: none; border: 1px solid #e5e5e5;}
	button{display: inline-block; vertical-align: baseline; height: 32px; line-height: 24px; font-size: 16px; padding: 0 16px; border: 2px solid #000; border-radius: 2px; box-sizing: border-box; background: transparent; text-align: center; text-decoration: none; cursor: pointer; outline: none; transition: border-color .2s ease; }

	#dump-search input {display: inline-block; width: 40%;  }
	/*#dump-search input:focus {border-color: #6cbbf7;box-shadow: 0 0 0 4px #6cbbf7;}*/
	#dump-search button {height: 36px;}
	.dump-message{border:2px solid #49cc90 ;border-radius:7px;background-color: #ecfaf4;}
	.dump-message>div{width:100%;}
	.dump-message>span:first-child{padding-left:10px;height:40px;line-height:40px;display:inline-block;width:10%}
	.dump-message>span:nth-of-type(2){display:inline-block;width:70%}
	.dump-info{background-color: #bfe1ac;}
	.dump-info>ul{display: block;height: 30px;}
	.dump-info>ul>li{float: left;list-style: none;width: 100px;height: 100%;}
	.dump-info, dump-info-basic,.dump-info-request,.dump-info-response{display: flex;flex-direction: column;}
	.dump-info-request>div,.dump-info-response>div {word-break:break-all; overflow: hidden;}

	.black-node{border:2px solid;border-radius:7px}
	.black-node span:first-child{padding-left:10px;height:40px;line-height:40px;display:inline-block;width:65%}
	.black-node span {display:inline-block;width:10%}
	.black-node > button {display: inline-block;width: 10%}
	.black-white{background-color:#ecfaf4;border-color:#49cc90}
	.black-black{background-color:#feebeb;border-color:#f93e3e}

	.breaker-line{width:100%}
	.breaker-line span:first-child{padding-left:10px;height:40px;line-height:40px;display:inline-block;width:75%}
	.breaker-line span:nth-of-type(2){display:inline-block;width:15%}
	.breaker-line svg{width:25px;height:25px;vertical-align:middle;fill:currentColor;overflow:hidden}
	.breaker-line{border:2px solid;border-radius:7px}
	.breaker .closed{background-color:#ecfaf4;border-color:#49cc90}
	.breaker .half-open{background-color:#fff5ea;border-color:#fca130}
	.breaker .open{background-color:#feebeb;border-color:#f93e3e}

	#request-ui fieldset {border: 0;}
	#request-ui legend {display: block; width: 100%; padding: 0; margin-bottom: 20px; font-size: 21px; line-height: inherit; color: #2e2e2e; border: 0; border-bottom: 1px solid #e5e5e5; }
	#request-ui input {width: 40%;}
	#request-ui li {list-style: none; display: block; height: 24px; line-height: 24px; border-bottom: 4px solid transparent; box-sizing: border-box; font-size: 12px; letter-spacing: .05em; white-space: nowrap; } 
	#request-ui {width: 960px; margin-left: auto; margin-right: auto; }
	#request-ui > * {margin-top: 10px }
	#request-line {display: flex; }
	#request-line > * {height: 40px;}
	#request-line > div {width: 70%; }
	#request-line input {width: 90%;margin: 0 8px; }
	#request-line > button {display: inline-block;}
	#request-uri {width: 100%; }
	#request-body-input-text > textarea {width: 100%; }
	#response-header > div > span:first-child {display: inline-block; line-height: 30px; width: 200px; }

	#request-prompt {position: absolute; z-index: 1000; height: 200px; width: 40%; border: 1px; border-style: solid; overflow-y: auto; overflow-x: all; background-color: #444; }
	.prompt-select {border: 2px; border-style: solid; background-color: #aaa; }
	#request-prompt::-webkit-scrollbar-track { border-radius: 10px; background-color: #F5F5F5; }
	#request-prompt::-webkit-scrollbar {width: 12px; background-color: #F5F5F5; }
	#request-prompt::-webkit-scrollbar-thumb {border-radius: 10px; background-color: #555; }

 	#pprof .profile-name{display:inline-block; width:6rem; }
	#look-value pre {margin-left: 40px;}
	#setting-input{width: 100%; min-height: 400px; margin-top: 20px;}
	</style>
</head>
<body>
<div id='nav'>
	<ul>
		<li><a href="#/dump">dump</a></li>
		<li><a href="#/black">black</a></li>
		<li><a href="#/breaker">breaker</a></li>
		<li><a href="#/request">request</a></li>
		<li><a href="#/pprof/">pprof</a></li>
		<li><a href="#/look/">look</a></li>
		<li><a href="#/expvar">expvar</a></li>
		<li><a href="#/setting">setting</a></li>
	</ul>
</div>
<div id="eudore-message"></div>
<div id='eudore-app'></div>
</body>
<script type="text/javascript">
"use strict";
String.prototype.trimPrefix = function(str){if(this.indexOf(str)==0) {return this.slice(str.length) } return this }
String.prototype.trimSuffix = function(str){let end=this.length-str.length;if(this.lastIndexOf(str)==end) {return this.slice(0,end) } return this }
String.prototype.format = function() {let args = arguments; return this.replace(/\${(\d+)}/g, function(match, number) {return typeof args[number] != 'undefined'? args[number] : match ; }); };
function NewApp(name, config) {
	function NewRouter() { // 路由匹配
		function getSplitPath(key) {
			if (key.length < 2) {return ["/"] }
			let strs = []; let num = -1; let isall = 0; let isconst = false;
			for (let i = 0; i < key.length; i++) {
				if (isall > 0) {
					if (key[i] == '{') {isall++; } else if (key[i] == '}') {isall--; }
					if (isall > 0) {strs[num] = strs[num] + key.slice(i, i + 1); } continue;
				}
				if (key[i] == '/') {if (!isconst) {num++; strs.push(""); isconst = true; }
				} else if (key[i] == ':' || key[i] == '*') {isconst = false; num++;strs.push(""); }
				else if (key[i] == '{') {isall++; continue; }
				strs[num] = strs[num] + key.slice(i, i + 1);
			}
			return strs
		}
		function getSubsetPrefix(str1, str2) {
			let findSubset = false;
			for(let i = 0; i < str1.length && i < str2.length; i++) {if(str1[i] != str2[i]) {return {path: str1.slice(0, i), find: findSubset} } findSubset = true; }
		 	if(str1.length > str2.length){return {path: str2, find: findSubset} } else if(str1.length == str2.length){return {path: str1, find: str1 == str2} }
			return {path: str1, find: findSubset}
		}
		const stdNodeKindConst = 1; const stdNodeKindParam = 2; const stdNodeKindWildcard = 3;
		function newRouterNode(path) {
			let node = {
				Kind: stdNodeKindConst, Name: "", Params: {},Cchildren: [], Pchildren: [], Wchildren: null,
				insertNode(path, node) {
					if(path.length==0){return this } node.Path = path
					if(node.Kind==stdNodeKindConst){return this.insertNodeConst(path, node)
					}else if(node.Kind==stdNodeKindParam){for(let i of this.Pchildren) {if(i.Path == path){return i } } this.Pchildren.push(node) ;
					}else if(node.Kind==stdNodeKindWildcard){if(this.Wchildren == null){this.Wchildren = node; } else {this.Wchildren.Name = node.Name; } return this.Wchildren }
					return node
				},
				insertNodeConst(path, node) {
					// 变量添加常量node
					for(let i in this.Cchildren) {
						let result = getSubsetPrefix(path, this.Cchildren[i].Path) 
						if(result.find){
							let subStr = result.path; 
							if(subStr != this.Cchildren[i].Path){this.Cchildren[i].Path = this.Cchildren[i].Path.trimPrefix(subStr); let newnode = newRouterNode(); newnode.Kind = stdNodeKindConst; newnode.Path = subStr; newnode.Cchildren = [this.Cchildren[i]]; this.Cchildren[i] = newnode; }
							return this.Cchildren[i].insertNode(path.trimPrefix(subStr), node)
						}
					}
					this.Cchildren.push(node);
					for (let i = this.Cchildren.length - 1; i > 0; i--) {if(this.Cchildren[i].Path[0] < this.Cchildren[i-1].Path[0]){this.Cchildren[i], this.Cchildren[i-1] = this.Cchildren[i-1], this.Cchildren[i]; } }
					return node
				},
				lookNode(searchKey, params) { 
					if(searchKey.length == 0 && this.Handler != null){return this }
					if(searchKey.length > 0) {
						for(let children of this.Cchildren) {if(children.Path[0] >= searchKey[0]){if(searchKey.startsWith(children.Path)){let ctx = children.lookNode(searchKey.trimPrefix(children.Path), params); if(ctx != null){return ctx } } } }
						if(this.Pchildren.length != 0){let pos = searchKey.indexOf('/'); if(pos == -1){pos = searchKey.length; } let currentKey = searchKey.slice(0, pos); let nextSearchKey = searchKey.slice(pos, 0); for(let children of this.Pchildren){let ctx = children.lookNode(nextSearchKey, params); if(ctx!=null){params[children.Name] = currentKey; return ctx } }
					} }
					if(this.Wchildren != null){params[this.Wchildren.Name]= searchKey; return this.Wchildren; }
					return null
				}
			}
			if (path==undefined) {return node }
			if (path.startsWith("*")) {node.Kind = stdNodeKindWildcard; if(path.length == 1){node.Name = "*"; }else {node.Name = path.slice(1); }
			}else if (path.startsWith(":")) {node.Kind = stdNodeKindParam; node.Name = path.slice(1); }
			return node
		}

		return {
			Root: newRouterNode(),
			Handler404: {Init() {return true;}, View(ctx) {return ["404 Page not found "+ctx.Path] }, Close() {} },
			Add(path, handler){
				let node = this.Root
				for(let i of getSplitPath(path)) {node = node.insertNode(i, newRouterNode(i)) }
				if(typeof handler=="function"){handler = {View: handler} }
				if(typeof handler.Init!="function"){handler.Init=()=>{return true} }
				if(typeof handler.Close!="function"){handler.Close=()=>{} }
				node.Params.Route = path
				node.Handler = handler
			},
			Match(path) {
				let ctx = {Path: path||'/', Params: {}, Querys: {}};
				let pos = path.indexOf('?')
				if(pos!=-1){
					ctx.Path = path.slice(0, pos)
					for(let pair of (new URLSearchParams(path.slice(pos))).entries()) {
						ctx.Querys[pair[0]]=pair[1]
					}
				}
				let node = this.Root.lookNode(ctx.Path, ctx.Params);
				if(node==null){ctx.Params.Route = "404"; ctx.Handler = this.Handler404; return ctx}
				ctx.Params = {...node.Params,...ctx.Params}; ctx.Handler = node.Handler; return ctx
			},			
		}
	}
	function NewRender(name, config) { // 虚拟dom
		let langs = {
			en: {},
			'zh-CN': {
				'delete': '删除',
				'reset': '重置',
				'submit': '提交',
				'show all': '显示全部'
			},
		}
		langs = langs[config.RenderLanguage]||{}
		let local = localStorage.getItem(config.RenderStore) 
		if(local) {
			local = JSON.parse(local);
			langs = {
				...langs,
				...(local[config.RenderLanguage]||{}),
			}
		}

		function i18n(i) {
			return langs[i]||i
		}
		function getHref(i){
			if(i.indexOf('http://')==0||i.indexOf('https://')==0) {
				return i
			}
			if(config.RouterPrefix.indexOf('#')==0){
				return location.pathname+location.search+config.RouterPrefix+i
			}else {
				return config.RouterPrefix+i+location.search+location.hash
			}
		}
		function isArray(o){return Object.prototype.toString.call(o)=='[object Array]'; }
		function isJson(o){return typeof(o) == "object" && Object.prototype.toString.call(o).toLowerCase() == "[object object]" && !o.length; }
		function isString(o) {return Object.prototype.toString.call(o) === '[object String]'}
		function renderElement(node, oldTree, newTree) {
			if(!newTree) {node.parentNode.removeChild(node);
			}else if(isString(oldTree) && isString(newTree)) {if(oldTree !== newTree) {node.textContent = i18n(newTree); }
			} else if((oldTree.type) === newTree.type) {let oldAttrs = oldTree.props; let newAttrs = newTree.props;
				for(let key in oldAttrs) {if(!newAttrs.hasOwnProperty(key)) {node.removeAttribute(key); } }
				for(let key in newAttrs) {if(!oldAttrs.hasOwnProperty(key) || oldAttrs[key] !== newAttrs[key]) {elemSetAttr(node, key, newAttrs[key], oldAttrs[key]); } }
				renderChildren(node, oldTree.children, newTree.children);
			}else {node.parentNode.replaceChild(createElement(newTree), node); }
		}
		function renderChildren(node, oldChildren, newChildren) {
			if(!oldChildren){oldChildren=[]; } if(!newChildren){newChildren=[]; }
			if(oldChildren.length >=newChildren.length){for(let i=oldChildren.length-1;i>=newChildren.length;i--) {node.removeChild(node.childNodes[i]); } for(let i in newChildren) {renderElement(node.childNodes[i], oldChildren[i], newChildren[i]); } }
			if(oldChildren.length < newChildren.length) {for(let i in oldChildren) {renderElement(node.childNodes[i], oldChildren[i], newChildren[i]); } for(let i=oldChildren.length;i<newChildren.length;i++) {node.appendChild(createElement(newChildren[i])); } }
		}
		function createElement(node) {
			if(isString(node)) {return document.createTextNode(i18n(node)) }
			let elem = document.createElement(node.type);
			for(let key in node.props) {elemSetAttr(elem, key, node.props[key]); }
			if(node.children) {for(let children of node.children) {elem.appendChild(createElement(children)); } }
			return elem;
		}
		function elemSetAttr(node, key, val, old) {
			switch(key) {
			case 'class': node.className = val; break;
			case 'style': node.style.cssText = val; break;
			case 'value': if(node.tagName.toUpperCase() === 'INPUT' || node.tagName.toUpperCase() === 'TEXTAREA') {node.value = val; } else {node.setAttribute(key, val); } break;
			case 'text': node.innerText = i18n(val); break;
			case 'html': node.innerHTML = val; break;
			case 'href': node.setAttribute(key, getHref(val));break;
			default: if (key.indexOf("on")!=-1){if(old){node.removeEventListener(key.slice(2), old); } node.addEventListener(key.slice(2), val); }else {node.setAttribute(key, val); }
			}
		}
		function elemSetType(type, dst){if(isString(dst)){return {type: type, props: {text: dst, } } } return {...{type: type}, ...dst} }
		function formatAtter(data) {
			if(isArray(data)) {
				let newdata = []
				for(let i in data) { newdata.push(formatAtter(data[i])) }
				return newdata 
			}
			if(isJson(data)){let newdata = {type: data.type || 'div', props: data.props || {}, children: [] }
				for(let key in data) {
					switch (key){
					case 'type': case 'props': break; 
					case 'id': case 'class': case 'text': case 'html': case 'value':
					case 'style': newdata.props[key] = data[key];break;
					case 'children': for(let i in data.children) {newdata.children.push(formatAtter(data.children[i])) } break;
					case 'bind': let bind = data[key];
						if(newdata.props.type=='checkbox'){if(bind[0][bind[1]]){newdata.props.checked='1'} }else{newdata.props['value'] = bind[0][bind[1]]; }
						newdata.props['onchange']=(e)=>{if(e.target.type=='checkbox'){bind[0][bind[1]]=e.target.checked}else {bind[0][bind[1]]=e.target.value}
							if(bind[2]){bind[2](e)} }; break;
					default:
						if (key.indexOf("on")!=-1){newdata.props[key] = data[key];
						}else if(isArray(data[key])) {for(let val of data[key]) {newdata.children.push(formatAtter(elemSetType(key,val))) }
						}else {newdata.children.push(formatAtter(elemSetType(key, data[key]))) }
					}
				}
				return newdata
			}
			return data
		}
		let dom = document.querySelector(name)
		let data = []
		function render(newdata) {
			if(!isArray(newdata)) {newdata = [newdata] } 
			newdata = formatAtter(newdata);
			renderChildren(dom, data, newdata); data = newdata }
		return render
	}
	function NewWatcher() { // 双向绑定
		function observer(data,notify, paths){
			if (!data || typeof data !== "object") {return; }
			paths=paths||[];
			if(Array.isArray(data)){
				data['__proto__'] = arrayMethods(notify, paths)
				for(let i in data) {observer(data[i],notify,paths.concat(i)) }
			}else{
				Object.keys(data).forEach(key=>{defineReactive(data,key,data[key],notify, paths.concat(key))})
			}
		}
		function defineReactive(data,key,val,notify, paths){
			if (!data || typeof data !== "object") {return; }
			observer(val, notify, paths);
			Object.defineProperty(data, key, {
				enumerable: true,
				configurable: true,
				get(){return val },
				set(newVal){
					if (val == newVal) return;
					val = newVal;
					data[key] = newVal;
					defineReactive(data, key, newVal, notify, paths);
					notify(paths.join('.'), newVal);
				}
			});
		}
		function arrayMethods(notify, paths){
			const arrayProto = Array.prototype
			const arrayMethods = Object.create(arrayProto)
			const methods = ['push', 'pop', 'shift', 'unshift', 'splice', 'sort', 'reverse']
			methods.forEach(function (method) {
				const original = arrayProto[method]
				Object.defineProperty(arrayMethods, method, {
					value: function v(...args) {
						if(method=='push') {for(let i in args) {observer(args[i], notify, paths.concat(this.length)) } }
						let val = original.apply(this, args);
						notify(paths.join('.'), args);
						return val
					}
				})
			})
			return arrayMethods
		}
		return observer
	}
	function NewLogger() { // 日志输出
		return {
			Debug: console.debug,
			Info: console.info,
			Error: console.error,
		}
	}
	function NewConfig(config) { // 配置管理
		return {
			Save() {localStorage.setItem(this.App.ConfigStore, JSON.stringify(this)) },
			Load() {
				if(window.localStorage) {
					let local = localStorage.getItem(this.App.ConfigStore) 
					if(local) {local = JSON.parse(local);for(let key in local){this[key] = local[key]; }}
				}
			},
			View() {
				return [{type: 'textarea', id: 'setting-input', value: JSON.stringify(this, null, 4), oninput: function(){
					this.style.height = '${0}px'.format(this.scrollHeight);
				}},
					{type: 'div', children: [
						{type: 'button', text: 'reset', onclick:()=> {localStorage.setItem(this.App.ConfigStore, ''); this.Load() } },
						{type: 'button', text: 'submit', onclick:()=>{localStorage.setItem(this.App.ConfigStore, document.getElementById('setting-input').value); this.Load() } }
				] } ]
			},
			...config
		}
	}
	function NewFetch(config) { // xhr
		return {
			Request(data) {
				data.method = data.method||'GET';
				if(data.url.indexOf('/')!=0){data.url = config.FetchGroup + data.url}
				data.headers = {...config.FetchHeaders, ...data.headers}
				if(data.data) {
					if(data.method=='GET'||data.method=='HEAD'){
						data.url+= '?'+(new URLSearchParams(data.data)).toString()
					}else {
						data.body = data.data
					}
				}
				return fetch(data.url,data).then((response)=>{
					data.response = response; if(!this.History){this.History=[]}; this.History.push(data);
					if(response.status>399) {throw new Error('request ${0} status ${1}'.format(data.url,response.status)); }
					let contexttype = response.headers.get('Content-Type')
					if(contexttype&&contexttype.indexOf('json')!= -1) {return response.json() }
					return response.text()
				}).then(data.success).catch((err)=>{this.Error(err) })
			}
		}
	}
	config = NewConfig({ // app配置初始化
		...config, App: {
			MaxEvent: 100,
			RouterPrefix: '#',
			ConfigStore: 'eudore-app',
			RenderStore: 'eudore-i18n',
			RenderLanguage: navigator.language,
			FetchGroup: '',
			FetchHeaders: {Accept: 'application/json', 'Cache-Control': 'no-cache'},
			...config.App
		},
	})
	config.Load()
	let app = {
		Events: [], Current: {Params: {}, Handler: {Close: ()=>{} } },
		Router: NewRouter(), Render: NewRender(name, config.App), Watcher: NewWatcher(), Logger: NewLogger(), Config: config, Fetch: NewFetch(config.App),
		Add(path, handler) {this.Router.Add(path, handler); },
		Goto(e) {
			if (this.Events.length == 0) {
				this.Watcher(this.Config, (path, val)=>{
					app.Config.Save();
					app.Current.Goto({Event: 'config', Location: location, Key: path, Value: val })
				});
			}

			let prefix = config.App.RouterPrefix
			if (!e) {e = {Event: 'reload', Location: location } }
			if (!e.Event) {
				e = {Event: 'goto', Path: e }
				if (prefix.indexOf('#') == 0) {
					history.pushState({}, '', location.pathname + location.search + prefix + e.Path);
				} else {
					history.pushState({}, '', prefix + e.Path + location.search + location.hash);
				}
			}
			if (e.Location) {
				if (prefix.indexOf('#') == 0) {
					e.Path = e.Location.hash.trimPrefix(prefix) || '/'
				} else {
					e.Path = e.Location.pathname.trimPrefix(prefix)+e.Location.search || '/'
				}
			}

			this.Current.Handler.Close(this.Current);
			let ctx = {...this.Router.Match(e.Path), ...app.Logger, Config: this.Config, Goto: (e) => this.Goto(e), Fetch: this.Fetch.Request }
			this.Current = ctx; ctx.Handler.Init(ctx); this.Reload(e);
			this.Watcher(ctx.Handler, (path, val)=>{this.Events.push(); this.Reload({Event: 'data', Key: path, Value: val }) });
		},
		Reload(e) {
			e.Time = new Date();	
			this.Events.push(e); let full = this.Events.length - config.App.MaxEvent; if (full > 0) {this.Events = this.Events.slice(full) };
			this.Render(this.Current.Handler.View(this.Current));
		},
	}
	window.onpopstate = (e)=>{app.Goto({Event: 'popstate', Location: e.target.location }) }
	window.addEventListener('click', (e)=>{if (e.target.tagName === 'A' && e.target.origin == location.origin) {e.preventDefault(); history.pushState({}, '', e.target.href); app.Goto({Event: 'click', Location: new URL(e.target.href), Value: e.target }) } })
	return app
}

function NewHandlerIndex() {
	return {
		View(ctx){return ['eudore index page'] }
	}
}

function NewHandlerDump() {
	function b64DecodeUnicode(str) {try {return decodeURIComponent(atob(str).split('').map(function(c) {return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2); }).join('')); }catch{return str } }
	return {
		ws: null,
		Data: [],
		Init(ctx){
			this.ws = new WebSocket("ws://"+location.host+ctx.Config.App.FetchGroup+ "dump/connect")
			try {
				this.ws.onopen = ()=>{fetch("/hello", {method: 'PUT',body:"request hello body",cache: 'no-cache',}) }
				this.ws.onmessage = (e)=> {let data = JSON.parse(e.data)||{}; data.display=false; data.info="basic"; this.Data.push(data); }
				this.ws.onclose = ()=>{this.Close() }
				this.ws.onerror = (e)=>{ctx.Error("dump server error:",e); this.Close(); }
			} catch (e) {ctx.Error(e.message); }
			return true;
		},
		Close(){if (this.ws) {this.ws.close(); this.ws = null; };console.log("close dump") },
		View(){
			let doms = [{type: 'div', id: 'dump-search', children: [{type: 'input',props: {type:'text', name: 'match'} }, {type: 'button', text:"Dump"}]}]
			for(let i in this.Data) {doms.push(this.DumpNewMessage(this.Data[i])) }
			return doms
		},
		DumpNewMessage(data) {
			let id = this.Data.length-1
			return {
				type: 'div', class: 'dump-line', children:[
					{
						type: 'div', class:"dump-message", onclick: ()=>{data.display=!data.display; },
						children: [{type: 'span', text: data["Method"]}, {type: 'span', text: data["Host"]+data["Path"]}, {type: 'span', text: data["Status"]}, ]
					},
					{
						type: 'div', class: 'dump-info', style: "display: "+(data.display?"block":"none"),
						children: [
							{
								type: 'ul', li: [
									{text: 'Basic Info', onclick: ()=>{this.Data[id].info="basic"; }},
									{text: 'Request Info', onclick: ()=>{this.Data[id].info="request"; }},
									{text: 'Response Info', onclick: ()=>{this.Data[id].info="response"; }},
								]
							},
							{
								type: 'div', class:"dump-info-basic", style: "display: "+(data.info=="basic"?"block":"none"),
								table: {tbody: {tr: [
									{td: [{text:"Method"},	{text:data["Method"]}]},
									{td: [{text:"URI"},	{text:data["RequestURI"]}]},
									{td: [{text:"Proto"},	{text:data["Proto"]}]},
									{td: [{text:"Host"},	{text:data["Host"]}]},
									{td: [{text:"Status"},	{text:data["Status"]}]},
									{td: [{text:"Time"},	{text:data["Time"]}]},
									{td: [{text:"Params"},	{p: this.getParamsDom(data["Params"])}]},
									{td: [{text:"Handlers"},{p: this.getHandlerDom(data["Handlers"])}]},
								]}}
							},
							{type: 'div', class: 'dump-info-request', style: "display: "+(data.info=="request"?"block":"none"), children: [
								{type: 'table', tr: this.getHeaderDom(data["RequestHeader"]) },
								{type: 'div', pre: {html: b64DecodeUnicode(data['RequestBody'])}},
							]},
							{type: 'div', class: 'dump-info-response', style: "display: "+(data.info=="response"?"block":"none"), children: [
									{type: 'table', tr: this.getHeaderDom(data["ResponseHeader"]) },
								{type: 'div', pre: {html: b64DecodeUnicode(data['ResponseBody'])}},
							]}
						]
					}
				
				]
			}
		},
		getHeaderDom(data) {let result=[]; for(let k in data) {result.push({td:[{text:k},{text:data[k].toString()}]}); } return  result },
		getParamsDom(data) {let result = []; for(let i in data) {result.push({text: i+"="+data[i]}); } return result },
		getHandlerDom(data) {let result = []; for(let i in data) {result.push({text: data[i]}); } return result }
	}
}

function NewHandlerBlack() {
	return {
		Numw: 0, Numb: 0, Data: {},
		Init(ctx) {ctx.Fetch({url: 'black/data', success: (data) =>{this.Data = data }})},
		View() {
			let doms = [{type: 'div', id: 'black-nav',children: [
				{type: 'span', text: 'eudore black list manager has ${0} while rule and ${1} black rule.'.format(this.Numw,this.Numb)},
				{type: 'button', id: 'black-insert', text: 'insert rule'}
			]}];
			for (let i of this.Data["white"]||[]) {doms.push(this.blackCreateList(i, "white")); }
			for (let i of this.Data["black"]||[]) {doms.push(this.blackCreateList(i, "black")); }
			this.Numw = (this.Data["white"]||[]).length;
			this.Numb = (this.Data["black"]||[]).length;
			return doms
		},
		blackCreateList(data, state) {
			let addr = data["addr"] + '/' + data["mask"]
			return {
				type: 'div', id: state + "-" + addr, class: "black-node black-" + state,
				children: [
					{type: 'span', text: addr}, {type: 'span', text: data["count"]},
					{type: 'button', text: "delete", onclick: (e)=> {
						fetch(apiGroup+'/black/'+state+'/'+data['addr']+'?mask='+data['mask'], {method: 'DELETE', cache: 'no-cache', }).then((response) =>{if (response.status==200) {this.Data[state] =this.Data[state].filter((i)=>{return i.addr != data["addr"] && i.mask != data["mask"] }) } })
						},
					},
				]
			}
		}
	}
}

function NewHandlerBreaker() {
	let states = ["closed", "half-open", "open"]
	return {
		Data: [],
		Init(ctx) {ctx.Fetch({url: 'breaker/data', success: (data)=>{let routes = []; for(let i in data) {data[i].display=false;routes.push(data[i]); } this.Data = routes }})},
		View(ctx) {
			let doms = [{}]; let state = {totalSuccesses:0, totalFailures:0, closed:0, open:0, 'half-open':0};
			for(let i in this.Data) {
				let route = this.Data[i]; state.totalSuccesses += route.totalsuccesses; state.totalFailures += route.totalfailures; state[route.state]++;
				doms.push( {
					type: 'div', class: "breaker", div: {
						class: "breaker-line " + route.state,
						span: [{text: route.name},
							{text:(route.totalsuccesses.toFixed(2)/(route.totalsuccesses+route.totalfailures).toFixed(2)*100).toFixed(2)+"%"}
						],
						svg: {
							html: '<svg viewBox="0 0 1024 1024" ns="http://www.w3.org/2000/svg"><path d="M513.3 101.9c-190.7 0-348.6 138.8-379.2 320.8H0l160.4 192.5 160.4-192.5H199C228.8 276.4 358.4 166 513.3 166c176.9 0 320.8 143.9 320.8 320.8S690.2 807.7 513.3 807.7c-78 0-148.7-29.1-204.4-75.6l-42.1 50.5c66.8 55.7 152.7 89.3 246.4 89.3 212.6 0 385-172.4 385-385s-172.2-385-384.9-385z" p-id="1482"></path></svg>',
							onclick: this.clickFlushFunc(ctx, i, route.id)
						},
						onclick: this.clickDisplayFunc(i)
					},
					table: {
						class: "breaker-info", style: "display: "+(route.display?"block":"none"), tbody: {tr: [
							{td: [{text: 'state'}, {select: {
								class: "breaker-select",
								children: [
									{type: 'option', text: "closed", props: (route.state=="closed"?{selected: 'selected'}:{})},
									{type: 'option', text: "half-open", props: (route.state=="half-open"?{selected: 'selected'}:{})},
									{type: 'option', text: "open", props: (route.state=="open"?{selected: 'selected'}:{})},
								],
								onchange: this.stateChangeFunc(ctx, i, route.id),
							}}]},
							{td: [{text: 'LastTime'}, {text: route.lasttime.slice(0, 19).replace("T", " ")}]},
							{td: [{text: 'totalsuccesses'}, {text: route.totalsuccesses}]},
							{td: [{text: 'totalfailures'}, {text: route.totalfailures}]},
							{td: [{text: 'consecutivesuccesses'}, {text: route.consecutivesuccesses}]},
							{td: [{text: 'consecutivefailures'}, {text: route.consecutivefailures}]},
						]}
					}
				});
			}
			doms[0] = {type: 'div', id:"state", children: [
				{type: 'p', text: 'totalsuccesses: ' + state.totalSuccesses + " totalfailures: " + state.totalFailures},
				{type: 'p', text: "closed: " + state.closed + " half-open: " + state['half-open'] + " open: " + state.open}
			] }
			return doms
		},
		clickFlushFunc(ctx, i, id){return (e)=>{this.stateFlush(ctx, i, id); e.stopPropagation(); } },
		clickDisplayFunc(i){return ()=> {this.Data[i].display = !this.Data[i].display;this.Data.push(); } },
		stateChangeFunc(ctx, i, id) {return (e)=>{
 			let state = e.target.selectedIndex;
			ctx.Fetch({method: 'PUT', url: "breaker/"+ id + "/state/" + state, success: ()=>{this.stateFlush(ctx, i, id)} })
		}},
		stateFlush(ctx, i, id){
			ctx.Fetch({url: "breaker/"+id, success: (route)=>{route.display = this.Data[i].display; this.Data[i] = route; this.Data.push()}})
		}
	}	
}

function NewHandlerRequest() {
	let AllMethods = ['GET', 'POST', 'PUT', 'DELETE', 'HEAD', 'PATCH']
	let AllHeaders = ['Accept', 'Accept-Language', 'Authorization', 'Cache-Control', 'Content-Encoding', 'Content-Language', 'Content-Location', 'Content-Type', 'From', 'If-Match', 'If-Modified-Since', 'If-None-Match', 'If-Unmodified-Since', 'Range', 'Pragma']
	let HeaderValues = {
		"Accept": ["application/json", "application/xml"],
		"Accept-Encoding": ["gzip", "compress", "deflate", "br", "identity", "*"],
		"Cache-Control": ["max-age", "max-stale", "min-fresh=", "no-cache", "no-store", "no-transform", "only-if-cached"],
		"Content-Type": ['application/json','application/xml','application/x-www-form-urlencoded','multipart/form-data','application/octet-stream'],
		"Connection": ["keep-alive", "close"]
	}
	let BodyType = ['*/*','application/json','application/xml','application/x-www-form-urlencoded','multipart/form-data','application/octet-stream']

	return {
		Names: {},
		Routes: {GET:[]},
		Input: {method:'GET',uri:'',args:[],headers:[],Output:{}},
		History: [],
		Init(ctx) {
			ctx.Fetch({
				url: 'router/data',
				success: (data)=>{
					let names = {};let routes = {}; for(let i of AllMethods){routes[i] = []; }
					for(let i in data.methods) {
						let path = data.paths[i].split(' ')[0]; if(AllMethods.includes(data.methods[i])) {routes[data.methods[i]].push(path); }
						if(data.methods[i]=="ANY") {for(let m of AllMethods){routes[m].push(path); } } names[path] = data.handlernames[i];
					}
					names[""] = ""; this.Names = names; this.Routes = routes;
				}
			})
		},
		View() {
			let output = this.Input.Output
			return [
				{type: 'div', id: 'request-prompt', style: 'display: none;', onclick: ()=>{}, ul: {id: 'request-prompt-list'} },
				{
					type: 'div', id: 'data-list', children: [
						{type: 'datalist', id: 'route_list', children: this.getOptions(this.Routes[this.Input.method]) },
						{type: 'datalist', id: 'headers', children: this.getOptions(AllHeaders)},
						...(()=>{let doms = []; for(let key in HeaderValues) {doms.push({type: 'datalist', id: 'headers-'+key, children: this.getOptions(HeaderValues[key])},) } return doms })()
					]
				},
				{
					type: 'div', id: 'request-ui',
					children: [
						{
							type: 'div', id: 'request-line',
							children: [
								{type: 'select' ,id: 'request-method-select', children: this.getOptions(AllMethods), onchange: (e)=>{this.Input.method=AllMethods[e.target.selectedIndex]} },
								{
									type: 'div', 
									input: {props: {type:'text', name: 'uri', list: 'route_list', title: this.Names[this.Input.uri]||''}, onchange: (e)=>{this.Input.uri =e.target.value}, }
								},
							 	{type:'button',text: 'send', onclick: ()=>{this.sendRequest()}}
							]
						},
						{
							type: 'fieldset', legend: "request agrs", children: (()=>{
								let data =this.Input.args; if(!data[data.length-1] || data[data.length-1].key) {data.push({key:''}) }
								let inputs = []; for(let i in data) {inputs.push(this.getInputLine('args',data[i])) } return inputs
							})()
						},
						{
							type: 'fieldset', legend: "request header", children: (()=>{
								let data =this.Input.headers; if(!data[data.length-1] || data[data.length-1].key) {data.push({key:''}) }
								let inputs = []; for(let i in data) {inputs.push(this.getInputLine('headers',data[i])) } return inputs
							})()
						},
						{
							type: 'fieldset', legend: "body", select: {
								onchange: (e)=>{this.Input.contentType=BodyType[e.target.selectedIndex]},
								children: this.getOptions(BodyType)
							},
							div: {children: [{type: 'div', textarea: {
								onchange: (e)=>{this.Input.body=e.target.value},
							}}]}
						},
						{
							type: 'fieldset', legend: "response status line", span: {
								text: output.status+' '+output.text
							}
						},
						{
							type: 'fieldset', legend: "response hader", children: (()=>{
								let headers =[];for(let i in output.headers){
									headers.push({type:'tr', children:[
										{type:'td',text:i},{type:'td',text:output.headers[i]}
									]})
								}
								return headers
							})()
						},
						{
							type: 'fieldset', legend: "response-body", pre: {
								text: output.body
							}
						},
					]
				}
			]
	
		},
		getInputLine(key, data) {
			return {type: 'div', children: [
				{type: 'input', props: {type: 'text', list: key}, onchange: (e)=>{data.key = e.target.value; }},
				{type: 'input', props: {type: 'text', list: key+'-'+data.key}, onchange: (e)=>{data.val = e.target.value; }}
			] }
		},
		getOptions(data) {
			let options = []; for(let i of data) {options.push({type: 'option', value: i, text: i}) };return options;
		},
		methodChangeFunc(e) {
			console.log(this)
			
		},
		uriChangeFunc(e) {
			
		},
		sendRequest(){
			let output = {headers:{}}
			let args = {}; for(let i of this.Input.args) {if(i.key){args[i.key]=i.val } }
			let header = {}; for(let i of this.Input.headers) {if(i.key){header[i.key]=i.val } }
			fetch(this.parseParams(this.Input.uri,args), {method: this.Input.method, headers: header })
			.then((response)=>{ 
				output.status =response.status;
				output.text = response.statusText;
				for (let key of response.headers.keys()) {
					output.headers[key]= response.headers.get(key)
				}
				return response.text()
			}).then((body)=>{
				output.body = body;
				this.Input.Output = output;
			})
		},
		parseParams(uri, params)  {
			const paramsArray = []
			Object.keys(params).forEach(key => params[key] && paramsArray.push('${0}=${1}'.format(key, params[key])))
			if (uri.search(/\?/) === -1) {
				uri += '?${0}'.format(paramsArray.join('&'))
			} else {
				uri += '&${0}'.format(paramsArray.join('&'))
			}
			return uri
		} 
	}
}

function NewHandlerPprof() {
	return {
		Data: [],
		Init(ctx) {ctx.Fetch({url: 'pprof/?format=json',success: (data)=>{this.Data = data}}) },
		View(ctx) {
			let tr = [{type:'tr',children:[{type:'td',text:'Count'}, {type:'td',text:'Profile'}, {type:'td',text:'Descriptions'}]}]
			for(let i of this.Data){
				tr.push({type: 'tr', children: [
					{type:'td', text: i.count},
					{type:'td', a: {props: {href: '/pprof/'+i.href, text: i.name}} },
					{type:'td', text: i.desc},
				]})
			}
			return [
				"Types of profiles available:",
				{type: 'table', children: tr},
				{type: 'a', props:{href: '/pprof/goroutine?debug=2', text: 'full goroutine stack dump'}}
			]
		}
	}
}

function NewHandlerPprofPage() {
	return {
		Data: '',
		Init(ctx){ctx.Fetch({url: 'pprof/'+ctx.Params.path, data: {...ctx.Querys, format: 'text'}, success: (data)=>{this.Data = data}, }) },
		View(ctx){return [{type: 'pre', text: this.Data}]}
	}
}


function NewHandlerPprofGoroutine() {
	let godoc = "https://golang.org"
	return {
		Data: [],
		Init(ctx){
			if(ctx.Config.Look.Godoc){godoc=ctx.Config.Look.Godoc.trimSuffix('/')||godoc }
			ctx.Fetch({url: 'pprof/goroutine', data: {...ctx.Querys, format: 'json'}, success: (data)=>{this.Data = data}, })
		},
		View(ctx){
			var doms = []
			if(ctx.Querys.debug==1){
				doms.push('goroutine profile: total '+this.Data.length+'\n')
				for(let i of this.Data){
					doms.push('['+i.args.join(' ')+']\n')
					for(let j of i.lines){
						doms.push('#	0x${0}	'.format(j.pointer), ...this.getPackage(j.func), '+0x${0}${1}'.format(j.pos, j.space),...this.getSource(j.file, j.line),'\n')
					}
					doms.push('\n')
				}
			}else {
				for(let i of this.Data){
					doms.push('goroutine ${0} [${1}]:\n'.format(i.number, i.state))
					for(let j of i.lines){
						if(j.created){doms.push('created by ',...this.getPackage(j.func),'\n\t')
						}else {doms.push(...this.getPackage(j.func), '(${0})\n\t'.format(j.args)) }
						doms.push(...this.getSource(j.file, j.line))
						if(j.pos){doms.push(' +0x${0}'.format(j.pos)) }
						doms.push('\n')
					}
					doms.push('\n')
				}
			}
			return {type: 'pre', children: doms}
		},
		getPackage(pkg){
			if(pkg == "main.main"){return [pkg] }
			let pos = pkg.indexOf('/')
			if(pos==-1){pos=0 }
			pos =pkg.slice(pos).indexOf('.')+pos
			let fn = pkg.slice(pos+1)
			pkg = pkg.slice(0, pos)

			let strs = fn.split('.')
			let obj = strs[0].trimPrefix('(').trimPrefix('*').trimSuffix(')')
			if(obj==''||obj[0] < 'A' || 'Z' < obj[0]) {
				return [{type:'a',props: {text: pkg, href: "${0}/pkg/${1}".format(godoc,pkg), target: '_Blank'}},'.'+fn]
			}
			if(strs.length==2 && 'A' <= strs[1][0] && strs[1][0] <= 'Z' ) {
				return [{type:'a',props: {text: '${0}.${1}'.format(pkg,fn), href: "${0}/pkg/${1}#${2}.${3}".format(godoc,pkg,obj,strs[1]), target: '_Blank'}}]
			}
			pos = fn.indexOf('.')
			if(pos==-1){
				return [{type:'a',props: {text: '${0}.${1}'.format(pkg,fn), href: "${0}/pkg/${1}#${2}".format(godoc,pkg,obj), target: '_Blank'}}]
			}
			return [{type:'a',props: {text: '${0}.${1}'.format(pkg,fn.slice(0,pos)), href: "${0}/pkg/${1}#${2}".format(godoc,pkg,obj), target: '_Blank'}},fn.slice(pos)]
		},
		getSource(file, line){
			let pos = file.indexOf('/src/')
			if(pos!=-1){
				return [{type: 'a', props: {text: file, href: "${0}${1}#L${2}".format(godoc, file.slice(pos), line), target: '_Blank'}}, ':'+line]
			}
			return [file + ":" + line]
		},
	}
}

function NewHandlerLook() {
	let paths = []
	let querys = ""
	let doc = "https://golang.org"
	return {
		Data: {},
		Review: 0,
		Init(ctx){
			if(ctx.Config.Look.Godoc){doc=ctx.Config.Look.Godoc }
			paths=[('/look/'+ctx.Params.path).trimSuffix('/')]
			ctx.Fetch({url: 'look/${0}?format=json&all=${1}&d=${2}'.format(ctx.Params.path, ctx.Config.Look.Showall, ctx.Config.Look.Depth),success: (data)=> {this.Data = data }})
		},
		View(ctx) {
			return [
				{type: 'div', children: [
					{type: 'input',bind: [ctx.Config.Look, 'Depth'], props: {type: 'text'}},
					{type: 'input',bind: [ctx.Config.Look, 'Showall'], props: {type: 'checkbox'}},
					{type: 'label', props:{for: 'look-showall'}, text: 'show all'}
				]},
				{type: 'pre', id:"look-value", children: this.getTemplate(this.Data)}
			]
		},
		addpath(path) {paths.push(path)},
		subpath() {paths.pop()},
		getpath() {return paths.join('/')},
		getTemplate(data) {
			let doms = []
			if(data.package&&data.name){
				doms.push({type: 'a', props: {text: data.package+"."+data.name, href: '${0}/pkg/${1}#${2}'.format(doc, data.package, data.name), target: '_Blank'} })
			}else {
				let name = ""
				if(data.package){name=data.package+"."}
				if(data.name){name=name+data.name }
				if(name){doms.push(name) }
			}
			switch(data.kind){
				case "bool":case "int": case "string": case "float": case "uint": case "complex":
					if(data.string){doms.push('("${0}")'.format(data.string)) }else if(data.kind=='string'){doms.push('("${0}")'.format(data.value))}else{doms.push('(${0})'.format(data.value)) } break;
				case "struct": case "map":
					doms.push("{"); 
					if(data.keys){ 
						doms.push({type: 'span', text: data.fold?"+":"-", onclick:()=>{data.fold=!data.fold; this.Review++; }});
						let fields = [];
						for(let i in data.keys){
							this.addpath(data.keys[i]);
							fields.push({type: 'a', props: {text: data.keys[i], href: this.getpath()} });
							fields.push(": "); fields=fields.concat(this.getTemplate(data.vals[i])); this.subpath();fields.push("\n");
						}
						doms.push({type:'pre', style: "display: "+(data.fold?"none":"block"),children:fields})
					}
					doms.push("}");
					break
				case "slice": case "array":
					doms.push("["); 
					if(data.vals){ 
						doms.push({type: 'span', text: data.fold?"+":"-", onclick:()=>{data.fold=!data.fold; this.Review++; }});
						let fields = [];
						for(let i in data.vals){
							this.addpath('${0}'.format(i));
							fields.push({type: 'a', props: {text: i, href: this.getpath(), target: '_Blank'} });
							fields.push(": "); fields=fields.concat(this.getTemplate(data.vals[i])); this.subpath();fields.push("\n");
						}
						doms.push({type:'pre', style: "display: "+(data.fold?"none":"block"),children:fields})
					}
					doms.push("]");
					break
				case "interface":
					if(data.elem){doms.push(" "); doms=doms.concat(this.getTemplate(data.elem)); }else {doms.push("(nil)"); } break;
				case "func": case "chan":
					if(data.pointer) {doms.push('(0x${0})'.format(data.pointer.toString(16))) }else{doms.push('(nil)') } break;
				default:
					if (data.elem){
						if(data.kind=="ptr"){doms.push("&"); doms=doms.concat(this.getTemplate(data.elem)); }
					}else if(data.pointer) {
						doms.push('(CYCLIC REFERENCE 0x${0})'.format((data.pointer||0).toString(16)))
					}else {
						doms.push('(nil)')						
					}
			}
			return doms
		}
	}
}

function NewHandlerExpvar() {
	return {
		Data: {},
		Init(ctx){ctx.Fetch({url: 'pprof/expvar', success: (data)=> {this.Data = data }})},
		View() {return [{type: 'pre', text: JSON.stringify(this.Data, null, "\t") }]}
	}
}

let app = NewApp("#eudore-app", {
	App: {FetchGroup: window.location.href.split(/(.*\/)admin\//)[1].trimPrefix(location.origin)},
	Look: {
		Showall: false,
		Depth: 10,
		Godoc: 'https://golang.org'
	}
})
app.Add('/', NewHandlerIndex())
app.Add('/dump', NewHandlerDump())
app.Add('/black', NewHandlerBlack())
app.Add('/breaker', NewHandlerBreaker())
app.Add('/request', NewHandlerRequest())
app.Add('/pprof/', NewHandlerPprof())
app.Add('/pprof/goroutine', NewHandlerPprofGoroutine())
app.Add('/pprof/*path', NewHandlerPprofPage())
app.Add('/look/*path', NewHandlerLook())
app.Add('/expvar', NewHandlerExpvar())
app.Add('/setting', {View: ()=>{return app.Config.View()}})
app.Goto()
</script>
</html>
`
