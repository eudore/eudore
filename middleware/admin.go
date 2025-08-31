package middleware

// adminStatic defines the content of admin.html.
const adminStatic = `<!DOCTYPE html>
<html>
<head>
	<title>Eudore Admin</title>
	<meta charset="utf-8">
	<meta name="author" content="eudore">
	<meta name="referrer" content="always">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="Eudore admin manage all eudore web ui">
	<style type="text/css">
	*{margin:0px;padding:0px;} 
	html, body{width:100%;height:100%;}
	.input{height: 24px; padding: 6px 16px; margin: 4px; font-size: 14px; line-height: 1.42857143; color: #2e2e2e; background-color: #fff; background-image: none; border: 1px solid #e5e5e5;}
	.button{display:inline-block;margin:4px 10px;padding:0 16px;height:32px;border:2px solid #000;border-radius:8px;vertical-align:baseline;line-height:24px;font-size:16px;box-sizing:border-box;background:transparent;color:#000;text-align:center;text-decoration:none;cursor:pointer;outline:none;transition:border-color .2s ease;}
	#eudore-app{max-width:960px;margin:auto;padding:4px;}
	#eudore-message {display: flex; flex-direction: column; position: fixed; margin: 20px;margin-top: 10px;top:0;right:0; width: 25%;max-width: 40px; min-width:300px; max-width: 600px; height: auto; z-index: 99; }
	#eudore-message >span {-webkit-transition: margin .5s ease-in-out; -moz-transition: margin .5s ease-in-out; display: flex; word-break: break-all; white-space: break-spaces; margin: 4px; padding: 4px; min-height: 32px; align-items:center; border-radius: 6px; }
	#eudore-message .debug {background-color: rgb(172,137,191);}
	#eudore-message .info {background-color: #090;}
	#eudore-message .warring {background-color: #bb0;}
	#eudore-message .error {background-color: #b00;}
	#nav{width:100%;background-color:#000;}
	#nav>ul{max-width:960px;margin:auto;display:flex;flex-wrap:wrap;}
	#nav li{height:40px;line-height:40px;float:left;list-style:none;padding:0 10px;}
	#nav li a:link{color: #ccc;text-decoration:none;}
	.state{display:flex;flex-direction:row;flex-wrap:nowrap;width:calc(100% - 32px);height:40px;padding:0 16px;margin:4px 0;line-height:40px;border:2px solid;border-radius:8px;}
	.state span {white-space:nowrap;overflow:hidden;text-overflow:ellipsis;}
	.state span:first-child{width:70%}
	.state span:nth-of-type(2){display:inline-block;width:20%}
	.state svg{width:20px;height:20px;vertical-align:middle;fill:currentColor;overflow:hidden;}
	.state-info{background-color:#ecfaf4;border-color:#49cc90}
	.state-warning{background-color:#fff5ea;border-color:#fca130}
	.state-error{background-color:#feebeb;border-color:#f93e3e}
	.dialog-container{position:fixed;inset:0;overflow:auto;background-color:rgba(0,0,0,0.5);}
	.dialog{margin:15vh auto 50px;width:50vw;background:rgba(0,0,0,0.1);background-color:canvas;box-sizing:border-box;border-radius:2px;box-shadow:0 1px 3px rgba(0,0,0,.3);border-radius:8px;overflow-y:hidden;}
	@media screen and (max-width: 600px){.dialog{left:10vw;width:80vw;}}
	.dump-node .state>span:first-child{width:10%}
	.dump-node .state>span:nth-of-type(2){width:70%}
	.dump-info{padding:8px;background-color:#bfe1ac;border-radius:4px;}
	.dump-info>ul{display:block;height:30px;}
	.dump-info>ul>li{float:left;list-style:none;width:100px;height:100%;}
	.dump-info tr td:nth-of-type(2){word-break:break-all;}
	.dump-info code{text-wrap:auto;}
	.dump-info, dump-info-basic,.dump-info-request,.dump-info-response{display:flex;flex-direction:column;}
	.dump-info-request>div,.dump-info-response>div {word-break:break-all;overflow:hidden;}
	.black-nav{display:flex;flex-direction:row;flex-wrap:wrap;align-items:center;}
	.black-nav textarea{width:calc(100% - 40px);min-height:300px;margin:20px 20px 0;overflow:hidden;padding:4px 16px;line-height:24px;box-sizing:border-box;border:1px solid #999;font-family:inherit;font-size:1rem;resize:none;outline:none;}
	.black-nav textarea:focus{border-color: #007bff;box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.25);}
	.black-buton{display:flex;flex-direction:row;flex-wrap:wrap;align-items:center;padding:20px;}
	.black-node > button {display:inline-block;width:10%;}
	.policy-nav ul{display:flex; flex-wrap: wrap; flex-direction: row;}
	.policy-nav li{float:left;list-style:none;min-width:80px;height:32px;line-height:32px;text-align:center;}
	.policy-nav input{padding:0 16px;min-width: 120px;height:24px;}
	.policy-node svg{margin: 0 4px;}
	.policy-node pre{padding:10px;background:#f9f9f9;border:1px solid #ccc;white-space:pre-wrap;word-wrap:break-word;tab-size:4;}
	.policy-node pre[contenteditable="true"]{background-color:#eef;outline:2px solid #4CAF50;}
	.member-node .state>span:first-child{width:20%;}
	.member-node .state>span:nth-of-type(2){width:70%;}
	.member-node ul li {padding:0 10px;line-height:24px;list-style:none;}
	.member-node ul li span:first-child{display:inline-block;width:100px;}
	fieldset.policy-form{border:1px solid #ccc;padding:20px;border-radius:8px;background-color:#f9f9f9;}
	.policy-form legend{font-weight:bold;font-size:1.2em;padding:0 8px;}
	.policy-form{display:grid;grid-template-columns:120px 1fr;grid-gap:16px;align-items:center;}
	.policy-form label{text-align:right;padding-right:10px;}
	.policy-form input,.policy-form textarea{width:100%;padding:8px;border:1px solid #ddd;border-radius:4px;}
	.policy-form textarea {min-height:80px;resize:vertical;}
	.policy-form input[type="button"]{grid-column:1 / -1;background-color:#007bff;color:white;border:none;cursor:pointer;margin-top:8px;}
	.policy-form input.disable{background-color:rgba(0,0,0,0.5);cursor:default}
	@media screen and (max-width: 600px){.policy-form{grid-template-columns: 60px 1fr;}}
	#pprof .profile-name{display:inline-block;width:6rem;}
	.look-nav{display:flex;align-items:center;}
	.look-node{overflow-x:auto;}
	.look-node pre{margin-left:40px;white-space: pre-wrap;}
	.expvar-node{white-space: pre-wrap;tab-size:4;}
	#setting-input{width:100%;min-height:400px;margin-top:20px;}
	</style>
</head>
<body>
<div id="nav">
	<ul>
		<li><a href="#/metadata">metadata</a></li>
		<li><a href="#/dump">dump</a></li>
		<li><a href="#/black">black</a></li>
		<li><a href="#/breaker">breaker</a></li>
		<li><a href="#/policy">policy</a></li>
		<li><a href="#/look/">look</a></li>
		<li><a href="#/pprof/">pprof</a></li>
		<li><a href="#/expvar">expvar</a></li>
		<li><a href="#/setting">setting</a></li>
	</ul>
</div>
<div id="eudore-message"></div>
<div id="eudore-app"></div>
</body>
<script type="text/javascript">
"use strict";String.prototype.trimPrefix=function(str){if(this.startsWith(str))return this.slice(str.length);return this};String.prototype.trimSuffix=function(str){if(this.endsWith(str))return this.slice(0,-str.length);return this};String.prototype.format=function(...args){return this.replace(/\${(\d+)}/g,(match,number)=>{return args[number]!==undefined?args[number]:match})};Date.prototype.Format=function(fmt){const Month=["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec",];let o={"M+":this.getMonth()+1,"d+":this.getDate(),"h+":this.getHours(),"m+":this.getMinutes(),"s+":this.getSeconds(),"q+":Math.floor((this.getMonth()+3)/3),S:this.getMilliseconds(),};if(/(y+)/.test(fmt))fmt=fmt.replace(RegExp.$1,(this.getFullYear()+"").substr(4-RegExp.$1.length),);for(let k in o)if(new RegExp("("+k+")").test(fmt))fmt=fmt.replace(RegExp.$1,RegExp.$1.length==1?o[k]:("00"+o[k]).substr((""+o[k]).length),);return fmt};function isMobile(){const userAgent=navigator.userAgent.toLowerCase();const mobileKeywords=["iphone","ipod","android","windows phone","blackberry",];return mobileKeywords.some((keyword)=>userAgent.indexOf(keyword)>-1)}
function NewApp(name,config){
	function NewRouter(){function getSplitPath(key){if(key.length<2){return["/"]}let strs=[];let num=-1;let isall=0;let isconst=false;for(let i=0;i<key.length;i++){if(isall>0){if(key[i]=="{"){isall++}else if(key[i]=="}"){isall--}if(isall>0){strs[num]=strs[num]+key.slice(i,i+1)}continue}if(key[i]=="/"){if(!isconst){num++;strs.push("");isconst=true}}else if(key[i]==":"){isconst=false;num++;strs.push("")}else if(key[i]=="*"){strs.push(key.slice(i));return strs}else if(key[i]=="{"){isall++;continue}strs[num]=strs[num]+key.slice(i,i+1)}return strs}function getSubsetPrefix(str1,str2){let findSubset=false;for(let i=0;i<str1.length&&i<str2.length;i++){if(str1[i]!=str2[i]){return{path:str1.slice(0,i),find:findSubset}}findSubset=true}if(str1.length>str2.length){return{path:str2,find:findSubset}}else if(str1.length==str2.length){return{path:str1,find:str1==str2}}return{path:str1,find:findSubset}}const stdNodeKindConst=1;const stdNodeKindParam=2;const stdNodeKindWildcard=3;function newRouterNode(path){let node={Kind:stdNodeKindConst,Name:"",Params:{},Cchild:[],Pchild:[],Wchild:null,insertNode(path,node){if(path.length==0){return this}node.Path=path;if(node.Kind==stdNodeKindConst){return this.insertNodeConst(path,node)}else if(node.Kind==stdNodeKindParam){for(let i of this.Pchild){if(i.Path==path){return i}}this.Pchild.push(node)}else if(node.Kind==stdNodeKindWildcard){if(this.Wchild==null){this.Wchild=node}else{this.Wchild.Name=node.Name}return this.Wchild}return node},insertNodeConst(path,node){for(let i in this.Cchild){let result=getSubsetPrefix(path,this.Cchild[i].Path);if(result.find){let subStr=result.path;if(subStr!=this.Cchild[i].Path){this.Cchild[i].Path=this.Cchild[i].Path.trimPrefix(subStr);let newnode=newRouterNode();newnode.Kind=stdNodeKindConst;newnode.Path=subStr;newnode.Cchild=[this.Cchild[i]];this.Cchild[i]=newnode}return this.Cchild[i].insertNode(path.trimPrefix(subStr),node)}}this.Cchild.push(node);for(let i=this.Cchild.length-1;i>0;i--){if(this.Cchild[i].Path[0]<this.Cchild[i-1].Path[0]){this.Cchild[i],(this.Cchild[i-1]=this.Cchild[i-1]),this.Cchild[i]}}return node},lookNode(searchKey,params){if(searchKey.length==0&&this.Handler!=null)return this;if(searchKey.length>0){for(let child of this.Cchild){if(child.Path[0]>=searchKey[0]){if(searchKey.startsWith(child.Path)){let ctx=child.lookNode(searchKey.trimPrefix(child.Path),params,);if(ctx!=null)return ctx}}}if(this.Pchild.length!=0){let pos=searchKey.indexOf("/");if(pos==-1)pos=searchKey.length;let currentKey=searchKey.slice(0,pos);let nextSearchKey=searchKey.slice(pos);for(let child of this.Pchild){let ctx=child.lookNode(nextSearchKey,params);if(ctx!=null){params[child.Name]=currentKey;return ctx}}}}if(this.Wchild!=null){params[this.Wchild.Name]=searchKey;return this.Wchild}return null},};if(path==undefined){return node}if(path.startsWith("*")){node.Kind=stdNodeKindWildcard;if(path.length==1){node.Name="*"}else{node.Name=path.slice(1)}}else if(path.startsWith(":")){node.Kind=stdNodeKindParam;node.Name=path.slice(1)}return node}return{Root:newRouterNode(),Handler404:{Mount(){return true},View(ctx){return["404 Page not found "+ctx.Path]},Unmount(){}},Add(path,handler){let node=this.Root;for(let i of getSplitPath(path)){node=node.insertNode(i,newRouterNode(i))}if(typeof handler=="function"){handler={View:handler}}if(typeof handler.Mount!="function"){handler.Mount=()=>{}}if(typeof handler.Unmount!="function"){handler.Unmount=()=>{}}node.Params.Route=path;node.Handler=handler},Match(path){let ctx={Path:path||"/",Params:{},Querys:{}};let pos=path.indexOf("?");if(pos!=-1){ctx.Path=path.slice(0,pos);for(let pair of new URLSearchParams(path.slice(pos)).entries()){ctx.Querys[pair[0]]=pair[1]}}let node=this.Root.lookNode(ctx.Path,ctx.Params);if(node==null){ctx.Params.Route="404";ctx.Handler=this.Handler404;return ctx}ctx.Params={...node.Params,...ctx.Params};ctx.Handler=node.Handler;return ctx},}}
	function NewRender(name,config){const lang=JSON.parse(localStorage.getItem(config.RenderStore)||"{}")[config.RenderLanguage];const i18n=(str)=>(lang&&lang[str])||str;const matchMenu=(name)=>{if(config.RenderMenus){const menuPattens=config.RenderMenus.map((patten)=>new RegExp("^"+patten.replace("*",".*")+"$"));return menuPattens.every((patten)=>!patten.test(name))}return true};const getHref=(i)=>{if(/^https?:\/\//.test(i)){return i}let routerPrefix=config.RouterPrefix;if(routerPrefix.startsWith("#")){return location.pathname+location.search+routerPrefix+i}return routerPrefix+i+location.search+location.hash};function renderchildWithRef(node,oldNodes=[],newNodes=[],ref){let num=Math.min(oldNodes.length,newNodes.length);for(let i=0;i<num;i++){if(oldNodes[i].type!==undefined){if(node.childNodes[i]==ref)renderElement(node.childNodes[i],oldNodes[i],newNodes[i]);renderchildWithRef(node.childNodes[i],oldNodes[i].child,newNodes[i].child,ref,)}}}function renderchildWithKeys(node,oldNodes=[],newNodes=[],start,num,){let keys=newNodes.filter((v)=>v.type!==undefined&&v.props.key!==undefined).map((v)=>v.props.key);oldNodes=oldNodes.map((v,i)=>{return{data:v,key:v.props&&v.props.key,index:i,dom:node.childNodes[i],use:v.props&&keys.includes(v.props.key),}});for(let i=start;i<num;i++){let key=newNodes[i].props&&newNodes[i].props.key;if(oldNodes[start].key===key&&(oldNodes[start].data.type===undefined)==(newNodes[i].type===undefined)){renderElement(node.childNodes[i],oldNodes[start].data,newNodes[i]);oldNodes.splice(start,1)}else{let val=oldNodes.slice(start).find((v)=>v.key===key)||oldNodes.slice(start).find((v)=>!v.use);if(val===undefined){node.insertBefore(createNode(newNodes[i]),node.childNodes[i]);continue}node.insertBefore(val.dom,node.childNodes[i]);renderElement(node.childNodes[i],val.data,newNodes[i]);oldNodes.splice(oldNodes.indexOf(val),1)}}oldNodes.forEach((v)=>node.removeChild(v.dom));for(let i=num;i<newNodes.length;i++)node.appendChild(createNode(newNodes[i]))}function renderchild(node,oldNodes=[],newNodes=[]){let num=Math.min(oldNodes.length,newNodes.length);for(let i=0;i<num;i++){if(newNodes[i].props&&newNodes[i].props.key!==undefined){return renderchildWithKeys(node,oldNodes,newNodes,i,num)}renderElement(node.childNodes[i],oldNodes[i],newNodes[i])}for(let i=num;i<newNodes.length;i++)node.appendChild(createNode(newNodes[i]));for(let i=num;i<oldNodes.length;i++)node.removeChild(node.childNodes[num])}function renderElement(node,o,n){if(n===undefined){node.parentNode.removeChild(node)}else if(o.type===undefined&&n.type===undefined){if(o!==n)node.textContent=i18n(n)}else if(o.type===n.type){for(let key of Object.keys(o.props)){if(!n.props.hasOwnProperty(key))removeElementAttr(node,key)}for(let[key,value]of Object.entries(n.props)){if(!o.props.hasOwnProperty(key)||o.props[key]!==value){setElementAttr(node,key,value,o.props[key])}}if(n.props.html===undefined)renderchild(node,o.child,n.child)}else{node.parentNode.replaceChild(createNode(n),node)}}function createNode(node){if(node.type===undefined){return document.createTextNode(i18n(node))}let elem=node.type==="svg"?document.createElementNS("http://www.w3.org/2000/svg","svg"):document.createElement(node.type);for(let[key,value]of Object.entries(node.props)){setElementAttr(elem,key,value)}if(node.props.$data!==undefined){node.props.$data.$ref=elem}if(node.child){for(let child of node.child){elem.appendChild(createNode(child))}}return elem}function removeElementAttr(node,key){switch(key){case"class":node.className="";break;case"style":node.style.cssText="";break;case"html":node.innerHTML="";break;case"key":case"$data":break;default:node.removeAttribute(key)}}function setElementAttr(node,key,val,old){switch(key){case"class":node.className=val;break;case"style":node.style.cssText=val;break;case"value":let tag=node.tagName.toUpperCase();if(tag==="INPUT"||tag==="TEXTAREA"){node.value=val}else{node.setAttribute(key,val)}break;case"html":node.innerHTML=val;break;case"href":node.setAttribute(key,getHref(val));break;case"key":case"$data":break;default:if(key.startsWith("on")){if(old)node.removeEventListener(key.slice(2),old);node.addEventListener(key.slice(2),val)}else{node.setAttribute(key,val)}}}const isJson=(o)=>typeof o==="object"&&!Array.isArray(o)&&o!==null;const elemSetType=(type,dst)=>isJson(dst)?{...{type},...dst}:{type,props:{text:dst}};function formatAttribute(data){if(Array.isArray(data)){return data.map(formatAttribute)}if(!isJson(data)){return data}let{type="div",props={},...rest}=data;const child=[];for(const[key,val]of Object.entries(rest)){switch(key){case"id":case"class":case"html":case"value":case"style":case"href":case"key":props[key]=val;break;case"text":if(val!==undefined){child.push(val)}break;case"menu":if(matchMenu(val)){return[]}break;case"if":if(!val){return[]}break;case"$data":if(!val){return[]}props[key]=val;break;case"child":if(Array.isArray(val))child.push(...val.map(formatAttribute));else child.push(val);break;case"options":break;case"bind":bindProps(type,props,data.child,val);break;case"component":delete data["component"];return formatAttribute(val(data));break;case"onenter":props.onkeydown=(e)=>{if(e.key==="Enter"){e.preventDefault();val(e)}};break;default:if(key.startsWith("on")){props[key]=val}else if(Array.isArray(val)){child.push(...val.map((v)=>formatAttribute(elemSetType(key,v))),)}else{child.push(formatAttribute(elemSetType(key,val)))}}}return{type,props,child}}function bindProps(type,props,child,[obj,key]){if(props.type==="checkbox"){if(obj[key]){props.checked=""}props.onchange=(e)=>{obj[key]=e.target.checked}}else if(props.type==="number"){props.onchange=(e)=>{obj[key]=parseInt(e.target.value,10)}}else if(type==="select"){let op=child.find((v)=>v.text==obj[key]);if(op!=null){op.props?(op.props.selected=""):(op.props={selected:""})}props.onchange=(e)=>{let d=e.target.childNodes[e.target.selectedIndex];obj[key]=d.value||d.text}}else{props.onchange=(e)=>{obj[key]=e.target.value}}if(obj[key]!==undefined){props.value=obj[key]}}let data=[];let dom=document.querySelector(name);while(dom.hasChildNodes())dom.removeChild(dom.firstChild);return(newdata,ref)=>{newdata=formatAttribute(newdata);if(!Array.isArray(newdata)){newdata=[newdata]}if(ref!==undefined){renderchildWithRef(dom,data,newdata,ref)}else{renderchild(dom,data,newdata)}data=newdata}}
	function NewWatcher(){function observer(data,notify,paths=[],ref){if(!data||typeof data!=="object")return;if(Array.isArray(data)){Object.setPrototypeOf(data,arrayMethods(data,notify,paths,ref));for(const[i,value]of data.entries()){observer(value,notify,[...paths,i],ref)}return}ref=data.$ref||ref;for(const[key,value]of Object.entries(data)){defineReactive(data,key,value,notify,[...paths,key],ref)}}function defineReactive(data,key,val,notify,paths,ref){if(key==="$ref")return;observer(val,notify,paths,ref);Object.defineProperty(data,key,{enumerable:true,configurable:true,get(){return val},set(newVal){if(val===newVal){return}val=newVal;if(typeof newVal==="object"){observer(newVal,notify,paths,ref)}notify(paths.join("."),newVal,ref)},})}function arrayMethods(data,notify,paths,ref){const arrayProto=Array.prototype;const arrayMethods=Object.create(arrayProto);const methods=["push","pop","shift","unshift","splice","sort","reverse",];methods.forEach(function(method){const original=arrayProto[method];Object.defineProperty(arrayMethods,method,{value:function(...args){if(method==="push"){args.forEach(function(value,i){observer(value,notify,[...paths,data.length+i],ref)})}const result=original.apply(this,args);if(args.length>0){notify(paths.join("."),data)}return result},})});Object.defineProperty(arrayMethods,"set",{value:function(index,value){observer(value,notify,[...paths,index],ref);data[index]=value;notify(paths.concat(index).join("."),value,ref)},});return arrayMethods}return observer}
	function NewLogger(){let index=0;return{Message:[],add(level,msg){this.Message.push({Time:new Date(),Level:level,Index:index++,Message:msg.toString(),})},Debug(...args){this.add("debug",args);console.debug(...args)},Info(...args){this.add("info",args);console.info(...args)},Error(...args){this.add("error",args);console.error(...args)},}}
	function NewConfig(config){return{Save(){localStorage.setItem(this.App.ConfigStore,JSON.stringify(this))},Load(){if(window.localStorage){try{let local=JSON.parse(localStorage.getItem(this.App.ConfigStore));if(local["App"]["Name"]!=config["App"]["Name"]){throw 0;}for(let key in local){this[key]=local[key]}}catch{this.Save()}}},View(){return[{type:"textarea",id:"setting-input",value:JSON.stringify(this,null,4),oninput:function(){this.style.height="${0}px".format(this.scrollHeight)},},{type:"div",child:[{type:"button",class:"button",text:"reset",onclick:()=>{localStorage.setItem(this.App.ConfigStore,"");this.Load()}},{type:"button",class:"button",text:"submit",onclick:()=>{localStorage.setItem(this.App.ConfigStore,document.getElementById("setting-input").value,);this.Load()},},],},]},...config,}}
	function NewFetch(config){function fatil(data,req){if(data.message)return [req.method,req.url,"message:",data.message].join(" ");if(data.error)return [req.method,req.url,"error:",data.error].join(" ");return [req.method, req.url,req.response.status,req.response.statusText].join(' ')}return function(req){req.method=req.method||"GET";if(req.url.indexOf("/")!==0&&config.FetchGroup)req.url=config.FetchGroup+req.url;req.headers={...config.FetchHeaders,...req.headers};req.referrerPolicy="no-referrer";if(req.data){if(req.method==="GET"||req.method==="HEAD"){const params=new URLSearchParams(req.data).toString();if(params!=="")req.url+="?"+params}else{if(req.headers["Content-Type"]===undefined)req.headers["Content-Type"]="application/json";switch(req.headers["Content-Type"]){case"application/json":req.body=JSON.stringify(req.data);break;case"multipart/form-data":default:req.body=req.data}}}if(req.bind){req.success=(d)=>{if(d===null)d=req.bind[2];req.bind[0][req.bind[1]]=d}}if(config.FetchHook){config.FetchHook(req)}return fetch(req.url,req).then((resp)=>{req.response=resp;if(req.reader!==undefined)return req.reader(resp);const contentType=resp.headers.get("Content-Type")||"";if(contentType.startsWith("application/json"))return resp.json();if(contentType.startsWith("application/octet-stream"))return resp.blob();return resp.text()}).then((data)=>{if(req.response.ok){req.success(data,req.response)}else if(req.fatil){req.fatil(data,req)}else{this.Error(fatil(data,req))}}).catch((err)=>{this.Error(err)})}}
	config=NewConfig({...config,App:{Name:"eudore",MaxEvent:100,RouterPrefix:"#",ConfigStore:"eudore-admin",RenderStore:"eudore-i18n",RenderMenus:["*"],RenderLanguage:navigator.language,FetchGroup:"",FetchHeaders:{Accept:"application/json","Cache-Control":"no-cache"},...config.App,},});config.Load();
	if(isMobile()){document.querySelector(name).classList.add("mobile")}
	const app={Events:[],Context:{Handler:{Unmount:()=>{}}},Router:NewRouter(),Render:NewRender(name,config.App),Watcher:NewWatcher(),Logger:NewLogger(),Config:config,Fetch:NewFetch(config.App),Add(path,handler){if(Array.isArray(path)){for(let p of path){this.Router.Add(p,handler)}}else{this.Router.Add(path,handler)}},Goto(e){if(this.Events.length==0){if(this.Init){this.Init(this)}this.Watcher(this.Config,(path,val)=>{app.Config.Save();app.Reload({Event:"config",Location:location,Key:path,Value:val,})})}let prefix=config.App.RouterPrefix;if(!e){e={Event:"reload",Location:location}}if(!e.Event){e={Event:"goto",Path:e};if(prefix.startsWith("#")){history.pushState({},"",location.pathname+location.search+prefix+e.Path,)}else{history.pushState({},"",prefix+e.Path+location.search+location.hash,)}}if(e.Location){if(prefix.startsWith("#")){e.Path=e.Location.hash.trimPrefix(prefix)||"/"}else{e.Path=e.Location.pathname.trimPrefix(prefix)+e.Location.search||"/"}}this.Context.Handler.Unmount(this.Context);let ctx={...this.Router.Match(e.Path),Config:this.Config,Fetch:this.Fetch,Goto:(e)=>this.Goto(e),Reload:(e)=>this.Reload(e),Debug:(e)=>this.Logger.Debug(e),Info:(e)=>this.Logger.Info(e),Error:(e)=>this.Logger.Error(e),};this.Context=ctx;ctx.Handler.Mount(ctx);this.Reload(e);this.Watcher(ctx.Handler,(path,val,ref)=>{this.Reload({Event:"data",Key:path,Value:val,Ref:ref})})},Reload(e){if(!e){this.Render(this.Context.Handler.View(this.Context));return}e.Index=(app.Events[app.Events.length-1]||{Index:-1}).Index+1;e.Time=new Date();e.toString=function(){let str = "event: ${0} event_id: ${1}".format(this.Event, this.Index);if(this.Path){str+=", path is "+this.Path};if(this.Key){str += ", key is " + this.Key};return str;};this.Logger.Debug(e);this.Events.push(e);let full=this.Events.length-config.App.MaxEvent;if(full>0){this.Events=this.Events.slice(full)}this.Render(this.Context.Handler.View(this.Context),e.Ref)},};
	window.onpopstate=(e)=>{app.Goto({Event:"popstate",Location:e.target.location})};
	window.onerror=(msg,file,line,_,error)=>{app.Logger.Error("${0} in ${1}:${2}".format(error,file,line));console.log("error: "+error)};
	window.addEventListener("click",(e)=>{let a=e.target;if(a.tagName==="A"&&a.origin==location.origin&&a.protocol==location.protocol){e.preventDefault();history.pushState({},"",a.href);app.Goto({Event:"click",Location:new URL(a.href),Value:a})}});
	if(document.querySelector("#eudore-message")){const log={Logger:app.Logger,Render:NewRender("#eudore-message",{}),Watcher:NewWatcher(),View(){return app.Logger.Message.filter((m)=>m.Level!="debug").map((m)=>{return{type:"span",class:m.Level,text:m.Message,onclick:()=>{app.Logger.Message=app.Logger.Message.filter((i)=>i.Index!=m.Index,)},}})},};log.Watcher(app.Logger,()=>{log.Render(log.View())});}
	return app;
}
const svgFlush = "<svg viewBox='0 0 1024 1024' version='1.1' xmlns='http://www.w3.org/2000/svg' p-id='4731'><path d='M286.016 737.792A319.104 319.104 0 0 1 467.2 195.488L461.632 92.8a421.568 421.568 0 0 0-248 717.28L128 895.744l268.128 14.72-14.72-268.128zM810.304 213.888L896 128.256l-268.128-14.72 14.72 268.128 95.424-95.424A319.072 319.072 0 0 1 556.8 828.512l5.568 102.688a421.472 421.472 0 0 0 247.936-717.312z' fill='' p-id='4732'></path></svg>"
const svgCopy = "<svg viewBox='0 0 1024 1024' version='1.1' xmlns='http://www.w3.org/2000/svg' p-id='5352'><path d='M853.333333 981.333333h-384c-72.533333 0-128-55.466667-128-128v-384c0-72.533333 55.466667-128 128-128h384c72.533333 0 128 55.466667 128 128v384c0 72.533333-55.466667 128-128 128z m-384-554.666666c-25.6 0-42.666667 17.066667-42.666666 42.666666v384c0 25.6 17.066667 42.666667 42.666666 42.666667h384c25.6 0 42.666667-17.066667 42.666667-42.666667v-384c0-25.6-17.066667-42.666667-42.666667-42.666666h-384zM213.333333 682.666667H170.666667c-72.533333 0-128-55.466667-128-128V170.666667c0-72.533333 55.466667-128 128-128h384c72.533333 0 128 55.466667 128 128v42.666666c0 25.6-17.066667 42.666667-42.666667 42.666667s-42.666667-17.066667-42.666667-42.666667V170.666667c0-25.6-17.066667-42.666667-42.666666-42.666667H170.666667c-25.6 0-42.666667 17.066667-42.666667 42.666667v384c0 25.6 17.066667 42.666667 42.666667 42.666666h42.666666c25.6 0 42.666667 17.066667 42.666667 42.666667s-17.066667 42.666667-42.666667 42.666667z' p-id='5353'></path></svg>"
const svgUpdate = "<svg viewBox='0 0 1024 1024' version='1.1' xmlns='http://www.w3.org/2000/svg' p-id='4499'><path d='M960 458.666667c0 23.466667-19.2 42.666667-42.666667 42.666667l-128 0 0 213.333333c0 23.466667-19.2 42.666667-42.666667 42.666667L277.333333 757.333333c-23.466667 0-42.666667-19.2-42.666667-42.666667l0-213.333333L106.666667 501.333333c-23.466667 0-42.666667-19.2-42.666667-42.666667 0-12.8 4.266667-23.466667 12.8-29.866667l0 0 405.333333-405.333333 0 0c8.533333-8.533333 19.2-12.8 29.866667-12.8s23.466667 4.266667 29.866667 12.8l0 0 405.333333 405.333333 0 0C955.733333 435.2 960 445.866667 960 458.666667zM512 113.066667 209.066667 416 277.333333 416c23.466667 0 42.666667 19.2 42.666667 42.666667l0 213.333333 384 0 0-213.333333c0-23.466667 19.2-42.666667 42.666667-42.666667l68.266667 0L512 113.066667zM277.333333 800l469.333333 0c23.466667 0 42.666667 19.2 42.666667 42.666667 0 23.466667-19.2 42.666667-42.666667 42.666667L277.333333 885.333333c-23.466667 0-42.666667-19.2-42.666667-42.666667C234.666667 819.2 253.866667 800 277.333333 800zM277.333333 928l469.333333 0c23.466667 0 42.666667 19.2 42.666667 42.666667s-19.2 42.666667-42.666667 42.666667L277.333333 1013.333333c-23.466667 0-42.666667-19.2-42.666667-42.666667S253.866667 928 277.333333 928z' p-id='4500'></path></svg>"
const svgDelete = "<svg viewBox='0 0 1024 1024' version='1.1' xmlns='http://www.w3.org/2000/svg' p-id='5179'><path d='M137.216 194.048c0-26.624 20.992-48.64 47.104-48.64h234.496v-48.64c0-26.624 20.992-48.64 47.104-48.64h93.696c25.6 0 47.104 21.504 47.104 48.64v48.64h234.496c26.112 0 47.104 21.504 47.104 48.64v48.64H137.216v-48.64z m702.976 145.408V937.472c0 26.624-20.992 48.64-47.104 48.64H230.912c-26.112 0-47.104-21.504-47.104-48.64V290.816h656.384v48.64zM371.2 436.224c0-26.624-20.992-48.64-47.104-48.64s-47.104 21.504-47.104 48.64v403.968c0 26.624 20.992 48.64 47.104 48.64 25.6 0 47.104-21.504 47.104-48.64V436.224z m187.904 0c0-26.624-20.992-48.64-47.104-48.64s-47.104 21.504-47.104 48.64v403.968c0 26.624 20.992 48.64 47.104 48.64s47.104-21.504 47.104-48.64V436.224z m187.392 0c0-26.624-20.992-48.64-47.104-48.64s-47.104 21.504-47.104 48.64v403.968c0 26.624 20.992 48.64 47.104 48.64s47.104-21.504 47.104-48.64V436.224z' fill='' p-id='5180'></path></svg>"
const ipv4Pattern = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
const cidr4Pattern = /^(?:[0-9]|[1-2][0-9]|3[0-2])$/;
const ipv6Pattern = /^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$/i;
const ipv6CompressedPattern = /((^|:)([0-9A-Fa-f]{1,4})){1,8}(::|((:([0-9A-Fa-f]{1,4})){1,7}))$/i;
const cidr6Pattern = /^(?:[0-9]{1,2}|1[01][0-9]|12[0-8])$/;
function copyValue(v){const t=document.createElement('textarea');t.value=v;t.style.position='fixed';t.style.opacity='0';t.style.width='1px';t.style.height='1px';document.body.appendChild(t);t.focus();t.select();document.execCommand('copy');document.body.removeChild(t);app.Context.Info("copysuccess")}
function onTextareaAuth(e){e.target.style.height="100px";e.target.style.height=e.target.scrollHeight+'px';}
function getEventIndex(e){let target = e.target;while(target.getAttribute('index')===null){target = target.parentElement;}return parseInt(target.getAttribute('index'),10)}
function checkIP(ip){const parts=ip.split('/');const ipPart=parts[0];const cidrPart=parts[1];if(parts.length>2){return false}const isIPv4=ipv4Pattern.test(ipPart);if(isIPv4){if(cidrPart!==undefined&&!cidr4Pattern.test(cidrPart)){return false}return true}const isIPv6=ipv6Pattern.test(ipPart)||ipv6CompressedPattern.test(ipPart);if(isIPv6){if(cidrPart!==undefined&&!cidr6Pattern.test(cidrPart)){return false}return true}return false}
function NewHandlerIndex(){return{View(ctx){return["eudore index page"]}}}` + metedataScript + dumpScript + blackScript + breakerScript + policyScript +
	`function NewHandlerPprof(){return{Data:[],Mount(ctx){ctx.Fetch({url:"pprof/?format=json",success:(data)=>{this.Data=data},})},View(ctx){let tr=[{type:"tr",child:[{type:"td",text:"Count"},{type:"td",text:"Profile"},{type:"td",text:"Descriptions"},],},];for(let i of this.Data){tr.push({type:"tr",child:[{type:"td",text:i.count},{type:"td",a:{text:i.name,props:{href:"/pprof/"+i.href}},},{type:"td",text:i.desc},],})}return["Types of profiles available:",{type:"table",child:tr},{type:"a",props:{href:"/pprof/goroutine?debug=2",text:"full goroutine stack dump",},},]},}}
	function NewHandlerPprofPage(){return{Data:"",Mount(ctx){ctx.Fetch({url:"pprof/"+ctx.Params.path,data:{...ctx.Querys,format:"text"},success:(data)=>{this.Data=data},})},View(ctx){return[{type:"pre",text:this.Data}]},}}
	function NewHandlerPprofGoroutine(){let godoc="https://golang.org";return{Data:[],Mount(ctx){if(ctx.Config.Look.Godoc){godoc=ctx.Config.Look.Godoc.trimSuffix("/")||godoc}ctx.Fetch({url:"pprof/goroutine",data:{...ctx.Querys,format:"json"},success:(data)=>{this.Data=data},})},View(ctx){console.log(22);var doms=[];if(ctx.Querys.debug==1){doms.push("goroutine profile: total "+this.Data.length+"\n");for(let i of this.Data){doms.push("["+i.args.join(" ")+"]\n");for(let j of i.lines){doms.push("#	0x${0}	".format(j.pointer),...this.getPackage(j.func),"+0x${0}${1}".format(j.pos,j.space),...this.getSource(j.file,j.line),"\n",)}doms.push("\n")}}else{for(let i of this.Data){doms.push("goroutine ${0} [${1}]:\n".format(i.number,i.state));for(let j of i.lines){if(j.args!=""){doms.push(...this.getPackage(j.func),"(${0})\n\t".format(j.args),)}else{doms.push("created by ",...this.getPackage(j.func),j.created?" in goroutine ${}\n\t".format(j.created):"\n\t",)}doms.push(...this.getSource(j.file,j.line));if(j.pos){doms.push(" +0x${0}".format(j.pos))}doms.push("\n")}doms.push("\n")}}return{type:"pre",child:doms}},getPackage(pkg){if(pkg=="main.main"){return[pkg]}let pos=pkg.indexOf("/");if(pos==-1){pos=0}pos=pkg.slice(pos).indexOf(".")+pos;let fn=pkg.slice(pos+1);pkg=pkg.slice(0,pos);let strs=fn.split(".");let obj=strs[0].trimPrefix("(").trimPrefix("*").trimSuffix(")");if(obj==""||obj[0]<"A"||"Z"<obj[0]){return[{type:"a",text:pkg,props:{href:"${0}/pkg/${1}".format(godoc,pkg),target:"_Blank",},},"."+fn,]}if(strs.length==2&&"A"<=strs[1][0]&&strs[1][0]<="Z"){return[{type:"a",text:"${0}.${1}".format(pkg,fn),props:{href:"${0}/pkg/${1}#${2}.${3}".format(godoc,pkg,obj,strs[1]),target:"_Blank",},},]}pos=fn.indexOf(".");if(pos==-1){return[{type:"a",text:"${0}.${1}".format(pkg,fn),props:{href:"${0}/pkg/${1}#${2}".format(godoc,pkg,obj),target:"_Blank",},},]}return[{type:"a",text:"${0}.${1}".format(pkg,fn.slice(0,pos)),props:{href:"${0}/pkg/${1}#${2}".format(godoc,pkg,obj),target:"_Blank",},},fn.slice(pos),]},getSource(file,line){let pos=file.indexOf("/src/");if(pos!=-1){return[{type:"a",text:file,props:{href:"${0}${1}#L${2}".format(godoc,file.slice(pos),line),target:"_Blank",},},":"+line,]}return[file+":"+line]},}}` +
	lookScript +
	expvarScript +
	`let app=NewApp("#eudore-app",{
		App:{FetchGroup: window.location.href.split(/(.*\/)admin\//)[1].trimPrefix(location.origin)},
		Look:{Showall: false,Depth: 10,Godoc: "https://golang.org"},
	});
	app.Init = (app)=>{
		setInterval(()=>{
			let now = new Date();
			let log = app.Logger.Message.filter((i)=>!((i.Level=="debug"&&now-i.Time>1200)||(i.Level=="info"&&now-i.Time>4000)||(i.Level=="error"&& now-i.Time>7000)));
			if (app.Logger.Message.length > 0 && app.Logger.Message.length != log.length) {app.Logger.Message = log}
		}, 400);
		if (location.protocol.includes("https:")){app.Logger.Debug=()=>{}}
		app.Logger.Debug("app init")
	}
	app.Add("/",NewHandlerIndex());
	app.Add("/metadata",NewHandlerMetadata());
	app.Add("/dump",NewHandlerDump());
	app.Add("/black",NewHandlerBlack());
	app.Add("/breaker",NewHandlerBreaker());
	app.Add(["/policy","/member","/policy-new"],NewHandlerPolicy());
	app.Add("/pprof/",NewHandlerPprof());
	app.Add("/pprof/goroutine",NewHandlerPprofGoroutine());
	app.Add("/pprof/*path",NewHandlerPprofPage());
	app.Add("/look/*path",NewHandlerLook());
	app.Add("/expvar",NewHandlerExpvar());
	app.Add("/setting",{View:()=>{return app.Config.View()},});
	app.Goto();
</script>
</html>`

const metedataScript = `
function NewHandlerMetadata() {
	const order = ["app", "logger", "config", "router", "client", "server"]; 
	function formatList(data){let doms=[];for(let i in data){if(i!=0){doms.push({type:'p'})}doms.push({type:'span',class:'item',text:data[i]})}return doms}
	function formatMap(data){let doms=[];for(let i in data){if(i!=0){doms.push({type:'p'})}doms.push({type:'span',class:'item',text:i+": "+data[i]})}return doms}
	function formatParams(p){return p.reduce((t,c,i)=>{if(i%2===0){t.push("${0}=${1}".format(c,p[i+1])||'');}return t;},[]).join("\n")}
	function sortOder(a,b){let indexA=order.indexOf(a);let indexB=order.indexOf(b);return (indexA===-1?Infinity:indexA)-(indexB===-1?Infinity:indexB);}
	return {
	Meta: {},
	Mount(ctx){let init=(data)=>{for(let name in data){data[name].display=false};this.Meta = data};ctx.Fetch({url:'metadata/',success:init,fatil:init})},
	View() {return Object.keys(this.Meta).sort(sortOder).map(key=>{let data = this.Meta[key];return {class:"meta-node",child: [
		{class: "state state-"+(this.Meta[key].health?"info":"error"),onclick: ()=>{this.Meta[key].display=!data.display},child:[
			{type:'span',text:key},{type:'span',text:data.name}
		]},
		{if: data.display, id: "meta-"+data.name,...this.getMetaView(key,data)},
	]}})},
	getMetaView(key, data){
		if(key=="router"){return {child: [
			{type: 'span', text: "core "+data.core},
			{type: 'table', child: data.methods.map((_,i)=>{
				return {type: 'tr', child:[
					{type: 'td', text: data.methods[i]},
					{type: 'td', text: data.paths[i], props:{title:formatParams(data.params[i])}},
					{type: 'td', text: data.handlerNames[i].at(-1),props: {title:data.handlerNames[i].join("\n")}},
				]}
			})},
		]}}
		return {type: 'pre',  html: JSON.stringify(data, null, "\t") }
	},
}}
`

const expvarScript = `
function NewHandlerExpvar() {
	const regPause = /(\d+)(?:(,\n\t\t\t\d+){15})/g
	const regSize = /{\s+"Size":[^}]+}/g
	return {
	Data: {},
	Mount(ctx){ctx.Fetch({url: "pprof/expvar", success: (data)=> {this.Data = data }})},
	View() {return [{type: "pre", class: "expvar-node", text: JSON.stringify(this.Data, null, "\t")
		.replace(regPause, (m)=>{for(let i=0;i<4;i++){m=m.replace(/(\d+),[\n\t]+(\d+)/g,"$1, $2")} return m})
		.replace(regSize, (m)=>{ return m.replace(/[\n\t]+/g, "").replace(",", ", ") })
	}]}
}}
`
