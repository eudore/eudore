package debug

var (
	UIpath   = ``
	UIString = `<!DOCTYPE html>
<html>
<head>
	<title></title>
</head>
<body>
	<div id="input-prompt" onclick="clickprompt(event)" style="display: none;">
		<ul id="input-prompt-list"></ul>
	</div>
	<div id=ui>
		<div id="request-line">
			<select id="request-method-select">
				<option value="GET">GET</option>
				<option value="POST">POST</option>
				<option value="PUT">PUT</option>
				<option value="PATCH">PATCH</option>
				<option value="DELETE">DELETE</option>
				<option value="COPY">COPY</option>
				<option value="HEAD">HEAD</option>
				<option value="OPTIONS">OPTIONS</option>
				<option value="LINK">LINK</option>
				<option value="UNLINK">UNLINK</option>
				<option value="PURGE">PURGE</option>
			</select>
			<div>
				<input id="request-uri" type="text" name="uri">
			</div>
			<button onclick="send()">send</button>
		</div>
		<fieldset id="request-args">
			<legend>request agrs</legend>
		</fieldset>
		<fieldset id="request-header">
			<legend>request header</legend>
		</fieldset>
		<fieldset>
			<legend>body</legend>
			<select id="request-body-select" onchange="chbody()">
				<option>*/*</option>
				<option>application/json</option>
				<option>application/xml</option>
				<option>application/x-www-form-urlencoded</option>
				<option>multipart/form-data</option>
				<option>application/octet-stream</option>
			</select>
			<div id="request-body">
				<div id="request-body-input-text">
					<textarea id="request-body-input-text-value"></textarea>
				</div>
				<div id="request-body-input-value" style="display: none;">
				</div>
				<div id="request-body-input-file"></div>
			</div>
		</fieldset>
		<fieldset>
			<legend>response status line</legend>
			<span id="response-status">200</span>
		</fieldset>
		<fieldset id="response-header">
			<legend>response hader</legend>
		</fieldset>
		<fieldset>
			<legend>body</legend>
			<pre id="response-body"></pre>
		</fieldset>
	</div>
</body>
<script type="text/javascript">
	"use strict";
	function display(id, dis) {
		document.getElementById(id).style.display = dis
	}

	function longestCommonPrefix(strs) {
		let one = strs.length > 0 ? String(strs[0]).split("") : false;
		let a = "";
		if (!one) {
			return a;
		};
		for (let i = 0; i < one.length; i++) {
			let num = 0;
			strs.map(da => {
				da.charAt(i) == one[i] ? num++ : null
			})
			if (num === strs.length) {
				a = a + one[i]
			} else {
				break
			}
		}
		return a
	};

	function Radix(path, child) {
		this.path = path || ""
		this.child = child || []
		this.has = (typeof path == "string" && path != "")
	}
	Radix.prototype = {
		insert: function(path) {
			for(var i of this.child) {
				var prefix = longestCommonPrefix([path, i.path])
				if(prefix == i.path) {
					i.insert(path.slice(prefix.length))
					return 
				}else if(prefix != "") {
					var newnode = new Radix(i.path.slice(prefix.length), i.child)
					newnode.has = i.has

					i.path = prefix
					i.child = [newnode, new Radix(path.slice(prefix.length))]
					i.has = false
					return
				}
			}
			this.child.push(new Radix(path))
		},
		insertArray: function(paths) {
			for(var i of paths || []) {
				this.insert(i)
			}
		},
		lookup: function(path) {
			var data = []
			if(this.has && (this.path.startsWith(path) || path == "")) {
				data.push("")
			}
			for(var i of this.child) {
				var prefix = longestCommonPrefix([path, i.path])
				if(prefix != "" || path == "") {
					data = data.concat(i.lookup(path.slice(prefix.length)))
				}
			}
			for(var i in data) {
				data[i] = this.path + data[i]
			}
			return data
		},
		point: function(path) {
			var data = []
			if(this.has) {
				// data.push(this.path)
			}
			for(var i of this.child) {
				var prefix = longestCommonPrefix([path, i.path])
				if(prefix == i.path) {
					for(var i of i.point(path.slice(prefix.length))) {
						data.push(this.path + i)
					}
					return data
				}else if(prefix != "") {
					return [this.path + i.path]
				}
			}
			for(var i of this.child) {
				data.push(this.path + i.path)
			}
			return data
		}
	}
</script>
<script type="text/javascript">	
	"use strict";
	var host = window.location.protocol + "//" + window.location.host
	function send() {
		var methodSelect = document.getElementById("request-method-select")
		var method = methodSelect.options[methodSelect.selectedIndex].value
		var uri = document.getElementById("request-uri").value 

		// args
		var argsdata = getJson("request-args")
		// args route
		var reg = /\/([\:\*]\w*)/g
		var uri2 = uri
		var tmp
		while(tmp = reg.exec(uri2)) {
			var key = tmp[1]
			var val = argsdata[key]
			if(val != "" && typeof val == 'string') {
				uri = uri.replace(key, val)
				delete argsdata[key]
			}
		}
		// args uri
		var args = new URLSearchParams()
		for(var i in argsdata) {
			args.append(i, argsdata[i])
		}

		// send
		var url = host + uri + "?" + args.toString()
		var request = {
			method: method,
			cache: 'no-cache',
			headers: getJson("request-header"),
		}
		if(method != "HEAD" && method != "GET") {
			request.body = getBody()
		}
		fetch(url, request).then(function(response){
			// console.log(response)

			// console.log(response.status)
			document.getElementById("response-status").innerText = response.status
			var header = document.getElementById("response-header")
			header.innerHTML = '<legend>response hader</legend>'
			for(var key of response.headers.keys()) {
				// console.log(key, response.headers.get(key)); 
				var line = document.createElement("div")
				var name = document.createElement("span")
				var val = document.createElement("span")
				name.innerText = key
				val.id = "response-header-" + key
				val.innerText = response.headers.get(key)
				line.appendChild(name)
				line.appendChild(val)
				header.appendChild(line)
			}

			return response.text()
		}).then(function(text) {
			var headercontenttype = document.getElementById("response-header-content-type")
			if(headercontenttype != null && headercontenttype.innerText.indexOf("application/json") != -1 ) {
				document.getElementById("response-body").innerText = JSON.stringify(JSON.parse(text), null, 4) 
				return
			}
			document.getElementById("response-body").innerText = text
		})
		// console.log(getJson("request-args"))
		// console.log(getJson("request-header"))
	}
	function getJson(id) {
		var data = {}
		for(var i of document.getElementById(id).getElementsByTagName("div")) {
			var name = i.childNodes[0].value || ""
			var val = i.childNodes[1].value || ""
			 // console.log(name,val, i.childNodes)
			if(name != "" && val != "") {
				data[name] = val
			}
		}
		return data
	}
	function getBody() {
		if(bodytype == "application/x-www-form-urlencoded") {
			var data = getJson("request-body-input-value")
			var body = new URLSearchParams()
			for(var i in data) {
				body.append(i, data[i])
			}
			return body.toString()
		}else if(bodytype == "multipart/form-data") {
			var data = getJson("request-body-input-value")
			var body = new FormData()
			for(var i in data) {
				body.append(i, data[i])
			}
			return data
		}else {
			return document.getElementById("request-body-input-text-value").value
		}
	}

	var bodytype = "*/*"
	var bodyinput = {
		"*/*": "request-body-input-text",
		"application/json": "request-body-input-text",
		"application/xml": "request-body-input-text",
		"application/x-www-form-urlencoded": "request-body-input-value",
		"multipart/form-data": "request-body-input-value",
		"application/octet-stream": "request-body-input-text",
	}
	function chbody() {
		console.log("chbody")
		var select = document.getElementById("request-body-select")
		bodytype = select.options[select.selectedIndex].text
		var id = bodyinput[bodytype]
		console.log(id)
		for(var i of ["request-body-input-text", "request-body-input-value"]) {
			document.getElementById(i).style.display = (i==id ? "block": "none")
			console.log(i)
		}
	}

</script>
<script type="text/javascript">
	"use strict";
	function addinput(id) {
		console.log(id)
		var line = document.getElementById(id).lastChild.parentNode
		for(var i of line.getElementsByTagName("input")) {
			if(i.value == "") {
				return
			}
		}
		var line = document.createElement("div")
		var name = document.createElement("input")
		var val = document.createElement("input")
		name.type = "text"
		val.type = "text"
		addprompt(name, id)
		addprompt(val, "request-value")
		name.onchange = function() {
			addinput(id)
		}
		val.onchange = function() {
			addinput(id)
		}
		line.appendChild(name)
		line.appendChild(val)
		document.getElementById(id).appendChild(line)
	}

	function addprompt(dom, id) {
		console.log(dom, id)
		var fn = function(e) {
			point(id, e.target, "input-prompt-list")

		}
		dom.oninput = fn
		dom.onfocus = fn
		dom.onblur = function(e) {
			console.log(e)
			setTimeout('document.getElementById("input-prompt").style.display = "none"', 150)
			updateselect(pointindex, "")
			pointindex = -1
		}
		var updateselect = function(index, style) {
			var lines = document.getElementById("input-prompt-list").getElementsByTagName("li")
			if(index < 0) {
				pointindex = 0
				return
			}
			if(index >= lines.length) {
				pointindex = lines.length -1 
				return
			}
			lines[index].className = style
			console.log(index,style)
		}
		dom.onkeydown = function(e) {
			if(e.which == 40) {
				updateselect(pointindex, "")
				pointindex++
				var node = document.getElementById("input-prompt")
				node.scrollTo(node.scrollLeft, pointindex * 24)
				updateselect(pointindex, "prompt-select")
			}else if(e.which == 38) {
				updateselect(pointindex, "")
				pointindex--
				updateselect(pointindex, "prompt-select")
				var node = document.getElementById("input-prompt")
				node.scrollTo(node.scrollLeft, pointindex * 24)
			}else if(e.which == 13) {
				var lines = document.getElementById("input-prompt-list").getElementsByTagName("li")
				if(-1 < pointindex && pointindex < lines.length) {
					dom.value = lines[pointindex].innerText
				}else {
					return
				}
			}else if(e.which == 9) {
				var data = getpointdata(id, dom).point(dom.value)
				if(data.length == 1 && dom.value != data[0]) {
					dom.value = data[0]
				}else if(data.length > 0) {
				}else {
					return
				}
			}else {
				return
			}
			e.preventDefault()
		}
	}
	function point(id, dom, p) {
		pointdom = dom
		// console.log(id, document.getElementById(id).value)
		console.log(id, dom.value, p)
		var data = getpointdata(id, dom).lookup(dom.value)
		if(data.length == 0 ) {
			document.getElementById("input-prompt").style.display = "none"
			return
		}
		data.sort()

		var lines = document.getElementById(p).getElementsByTagName("li")
		if(lines.length < data.length) {
			var len = data.length-lines.length
			var ul = document.getElementById(p)
			for(var i=0;i< len; i++) {
				ul.appendChild(document.createElement("li"))
			}
		}

		if(data.length == 1 && data[0] == dom.value) {
			document.getElementById("input-prompt").style.display = "none"
			return
		}

		for(var i in data) {
			lines[i].innerText = data[i]
			lines[i].style.display = "block"
		}
		if(lines.length > data.length) {
			for(var i=data.length;i<lines.length;i++) {
				lines[i].innerText = ""
				lines[i].style.display = "none"
			}
		}

		var d = document.getElementById("input-prompt")
		d.parentNode.removeChild(d)
		dom.parentNode.appendChild(d)
		console.log("show")
		document.getElementById("input-prompt").style.display = "block"
	}
	function clickprompt(event) {
		console.log(event.target.innerText)
		pointdom.value = event.target.innerText
		document.getElementById("input-prompt").style.display = "none"
	}
	addinput("request-args")
	addinput("request-header")
	addinput("request-body-input-value")
	addprompt(document.getElementById("request-uri"), "request-uri")
	document.getElementById("request-uri").value = "/eudore/debug/router/ui"
	document.getElementById("request-uri").value = "/eudore/debug/:router/ui/:name/*"


	var pointindex = -1
	var pointdata = {
		"request-uri": new Radix(),
		"request-args": new Radix(),
		"request-header": new Radix(),
	}
	var pointdom
	var pointdatavalue = {
		"Accept": ["application/json", "application/xml"],
		"Accept-Encoding": ["gzip", "compress", "deflate", "br", "identity", "*"],
		"Cache-Control": ["max-age", "max-stale", "min-fresh=", "no-cache", "no-store", "no-transform", "only-if-cached"],
		"Connection": ["keep-alive", "close"]
	}
	pointdata["request-header"].insertArray(["Accept", "Accept-Charset", "Accept-Encoding", "Accept-Language", "Accept-Ranges", "Access-Control-Allow-Credentials", "Access-Control-Allow-Headers", "Access-Control-Allow-Methods", "Access-Control-Allow-Origin", "Access-Control-Expose-Headers", "Access-Control-Max-Age", "Access-Control-Request-Headers", "Access-Control-Request-Method", "Allow", "Authorization", "Cache-Control", "Connection", "Cookie", "DNT", "Date", "Early-Data", "Expect", "From", "Host", "If-Match", "If-Modified-Since", "If-None-Match", "If-Range", "If-Unmodified-Since", "Index", "Keep-Alive", "Origin", "Pragma", "Proxy-Authorization", "Range", "Referer", "TE", "Upgrade-Insecure-Requests", "User-Agent", "Via", "Warning"])

	function getpointdata(id, dom) {
		var data = pointdata[id]
		if(id == "request-uri") {
			var methodSelect = document.getElementById("request-method-select")
			var method = methodSelect.options[methodSelect.selectedIndex].value
			data = data[method]
		}else if(id == "request-value") {
			var tree = new Radix()
			tree.insertArray(pointdatavalue[dom.previousSibling.value])
			return tree
		}
		return data || new Radix()
	}

	fetch("/eudore/debug/router/data", {
		headers: {
			"Accept": "application/json, text/*"
		}
	}).then(function(response) {
		return response.json();
	}).then(function(data) {
		var routes = {}
		for(var i of ["HEAD", "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"]) {
			routes[i] = new Radix()
		}
		for(var i in data.paths) {
			if(data.methods[i] == "ANY") {
				for(var method in routes) {
					routes[method].insert(data.paths[i].split(" ")[0])
				}
			}else {
				routes[data.methods[i]].insert(data.paths[i].split(" ")[0])
			}
		}
		pointdata["request-uri"] = routes
	})
</script>
<style type="text/css">
*{
	margin: 0;
	padding: 0;
}
fieldset {
	border: 0;
}
legend {
	display: block;
	width: 100%;
	padding: 0;
	margin-bottom: 20px;
	font-size: 21px;
	line-height: inherit;
	color: #2e2e2e;
	border: 0;
	border-bottom: 1px solid #e5e5e5;
}
input {
	display: inline-block;
	width: 40%;
	height: 24px;
	padding: 6px 16px;
	margin: 4px;
	font-size: 14px;
	line-height: 1.42857143;
	color: #2e2e2e;
	background-color: #fff;
	background-image: none;
	border: 1px solid #e5e5e5;
}
li {
	list-style: none; 
	display: block;
	height: 24px;
	/*padding: 8px 10px;*/
	line-height: 24px;
	border-bottom: 4px solid transparent;
	box-sizing: border-box;
	font-size: 12px;
	letter-spacing: .05em;
	white-space: nowrap;
}
#ui {
	width: 960px;
	margin-left: auto;
	margin-right: auto;
}
#ui > * {
	margin-top: 10px
}
#request-line {
	display: flex;
}
#request-line > div {
	width: 40%;
}
#request-uri {
	width: 100%;
}
#request-line > button {
	display: inline-block;
}
#request-body-input-text > textarea {
	width: 100%;
}
#response-header > div > span:first-child {
	display: inline-block;
	line-height: 30px;
	width: 200px;
}

#input-prompt {
	position: absolute;
	z-index: 1000;
	height: 200px;
	width: 40%;
	border: 1px;
	border-style: solid;

	overflow-y: auto;
	overflow-x: all;
	background-color: #444;
}
.prompt-select {
	border: 2px;
	border-style: solid;
	background-color: #aaa;
}

#input-prompt::-webkit-scrollbar-track {
	-webkit-box-shadow: inset 0 0 6px rgba(0,0,0,0.3);
	border-radius: 10px;
	background-color: #F5F5F5;
}
#input-prompt::-webkit-scrollbar {
	width: 12px;
	background-color: #F5F5F5;
}
#input-prompt::-webkit-scrollbar-thumb {
	border-radius: 10px;
	-webkit-box-shadow: inset 0 0 6px rgba(0,0,0,.3);
	background-color: #555;
}
</style>
</html>`
)
