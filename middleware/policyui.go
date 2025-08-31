package middleware

const policyScript = `
function NewHandlerPolicy() {
	let classs=["state state-info", "state state-warning", "state state-error"];
	function policyEffects(p) {
		if(!Array.isArray(p.statement)) return "";
		const effect=[];
		if (p.statement.some(stmt => stmt.effect)) {effect.push("Allow"); }
		if (p.statement.some(stmt => !stmt.effect)) {effect.push("Deny"); }
		if (p.statement.some(stmt => stmt.data)) {effect.push("Data"); }
		return effect;
	}
	function policyStringify(stmt) {
		return JSON.stringify(stmt,null,"\t").replace(/"\S+": \[[\s\S]*?\]/g, (s) => {
			if (s.split(",").length > 3 ) {return s; }
			return s.replace(/[\t\n]/g, '').replace(/,/g, ", ");
		});
	}

	const h={
	Policys: [], Members: [],Commit: {pn:"",ps:"",mn:"",bp:"",bd:"",cd:""},Search:"",
	Mount(ctx) {
		ctx.Fetch({url:"policys", success: (data) => {this.Policys=(data||[]).map(p=>{
			return {...p, state: 0, display: false, effect:policyEffects(p), value: policyStringify(p.statement)}
		})}})
		ctx.Fetch({url:"members", success: (data) => {this.Members=(data||[]).map(m=>{
			return { ...m, state: 0, display: false, policylist:(m.policy||[]).join(', '),datalist:(m.data||[]).join(', ')}
		})}})
	},		
	View(ctx){
		let child=[]
		if(ctx.Params.Route=="/policy"){
			child=this.Policys.filter(data => data.name.indexOf(this.Search)!=-1).map((data, index)=>{ return {
				class: "policy-node",
				props: {index:index},
				div: {class: classs[data.state],onclick:()=>{data.display=!data.display}, span: [
					{text: data.name},
					{text: data.effect},
					{html:svgFlush, onclick: this.onPolicyFlush},
					{html:svgCopy, onclick: this.onPolicyCopy},
					{html:svgUpdate, onclick: this.onPolicyUpdate},
					{html:svgDelete, onclick: this.onPolicyDelete}
				]},
				pre: {if: data.display, text: data.value,
					ondblclick: this.onPolicyEdit, onblur: this.onPolicyBlur
				},
			}})
		}else if(ctx.Params.Route=="/member"){
			child=this.Members.filter(data => data.user.indexOf(this.Search)!=-1).map((data, index)=>{
				let child=[{class: classs[data.state],onclick:()=>{data.display=!data.display}, span: [
					{text: data.user},
					{text: data.policylist},
					{html:svgFlush, onclick: this.onMemberFlush},
					{html:svgCopy, onclick: this.onMemberCopy},
					{html:svgUpdate, onclick: this.onMemberUpdate},
					{html:svgDelete, onclick: this.onMemberDelete}
				]}]
				if(data.display) {
					child.push({type: "ul", child: (data.policy||[]).map((p,i)=>{return {
						type:'li', props:{draggable:true},
						span: [{text: "Policy"}, {text: p}],
						ondragstart: (e)=>{data.movepolicy=i },
						ondragend: (e)=>{e.preventDefault();data.movepolicy=null },
						ondragenter: (e)=>{
							e.preventDefault();
							const moving=data.policy[data.movepolicy];
							if(moving===undefined)return
							data.policy.splice(data.movepolicy, 1);
							data.policy.splice(i, 0, moving);
							data.movepolicy=i;
							data.state=data.policy.join(', ')==data.policylist ? 0 : 1
						},
						ontouchstart: (e)=>{e.preventDefault();data.movepolicy=i;data.movepos=e.touches[0].pageY},
						ontouchend: (e)=>{e.preventDefault();data.movepolicy=null;data.movepos=null},
						ontouchmove(e) {e.preventDefault();const size=e.touches[0].pageY-data.movepos;
							if (size < -24 || size > 24) {
								const moving=data.policy[data.movepolicy];
								const index=data.movepolicy + (size > 0 ? Math.floor(size / 24) : Math.ceil(size / 24));
								if(moving===undefined)return
								data.policy.splice(data.movepolicy, 1);
								data.policy.splice(index, 0, moving);
								data.movepolicy=index;
								data.movepos=e.touches[0].pageY;
								data.state=data.policy.join(', ')==data.policylist ? 0 : 1
							}
						},
					}})})
					child.push({type: "ul", child: (data.data||[]).map((p,i)=>{return {
						type:'li', props:{draggable:true},
						span: [{text: "Data"}, {text: p}],
						ondragstart: (e)=>{data.movedata=i },
						ondragend: (e)=>{e.preventDefault();data.movedata=null },
						ondragenter: (e)=>{
							e.preventDefault();
							const moving=data.data[data.movedata];
							if(moving===undefined)return
							data.data.splice(data.movedata, 1);
							data.data.splice(i, 0, moving);
							data.movedata=i;
							data.state=data.data.join(', ')==data.datalist ? 0 : 1
						},
						ontouchstart: (e)=>{e.preventDefault();data.movedata=i;data.movepos=e.touches[0].pageY},
						ontouchend: (e)=>{e.preventDefault();data.movedata=null;data.movepos=null},
						ontouchmove(e) {e.preventDefault();const size=e.touches[0].pageY-data.movepos;
							if (size < -24 || size > 24) {
								const moving=data.data[data.movedata];
								const index=data.movedata + (size > 0 ? Math.floor(size / 24) : Math.ceil(size / 24));
								if(moving===undefined)return
								data.data.splice(data.movedata, 1);
								data.data.splice(index, 0, moving);
								data.movedata=index;
								data.movepos=e.touches[0].pageY;
								data.state=data.data.join(', ')==data.datalist ? 0 : 1
							}
						},
					}})})
				}
				return {class: "member-node", props: {index:index}, child: child}
			})		
		}else if(ctx.Params.Route=="/policy-new"){
			child=[
				{type:"fieldset",class:"policy-form",child:[
					{type:"legend",text:"New Policy"},
					{type:"label",props:{for:"policy-name"},text:"policy name"},
					{type:"input",id:"policy-name",props:{name:"policy-name", list:"policy-list"},bind:[this.Commit,"pn"]},
					{type:"label",props:{for:"policy-stmt"},text:"policy stmt"},
					{type:"textarea",id:"policy-stmt",oninput:onTextareaAuth,props:{name:"policy-stmt",placeholder:"policy statement"},bind:[this.Commit,"ps"]},
					{type:"input",class:this.checkPolicy()?"":"disable",props:{type:"button",value:"Commit"},onclick:this.commitPolicy}
				]},
				{type:"fieldset",class:"policy-form",child:[
					{type:"legend",text:"New Member"},
					{type:"label",props:{for:"member-name"},text:"member name"},
					{type:"input",id:"member-name",props:{name:"member-name",list:"member-list"},bind:[this.Commit,"mn"]},
					{type:"label",props:{for:"bind-policy"},text:"bind policy"},
					{type:"input",id:"bind-policy",props:{name:"bind-policy",placeholder:"select policy name",list:"policy-list"},bind:[this.Commit,"bp"]},
					{type:"label",props:{for:"bind-data"},text:"bind data"},
					{type:"input",id:"bind-data",props:{name:"bind-data",placeholder:"select data name",list:"policy-list"},bind:[this.Commit,"bd"]},
					{type:"input",class:this.checkMember()?"":"disable",props:{type:"button",value:"Commit"},onclick:this.commitMemeber}
				]},
				{type:"fieldset",class:"policy-form",child:[
					{type:"legend",text:"Custom Data"},
					{type:"label",props:{for:"custom-data"},text:"custom data"},
					{type:"textarea",id:"custom-data",oninput:onTextareaAuth,props:{name:"custom-data",placeholder:"custom input json or array data"},bind:[this.Commit,"cd"]},
					{type:"input",class:this.checkCustom()?"":"disable",props:{type:"button",value:"Commit"},onclick:this.commitCustom}
				]},
    			{type:"datalist",id:"policy-list",child:this.Policys.map((data)=>{return{type:"option",value:data.name}})},
    			{type:"datalist",id:"member-list",child:this.Members.map((data)=>{return{type:"option",value:data.user}})},
			]
		}
		child.unshift({class: "policy-nav", ul: {li: [
			{text: "Policys", onclick: ()=>{ctx.Goto("/policy")}},
			{text: "Memebr", onclick: ()=>{ctx.Goto("/member")}},
			{text: "New", onclick: ()=>{ctx.Goto("/policy-new")}},
			{if: ctx.Params.Route!="/policy-new",input:{id:"policy-search",class:"input",bind:[this,"Search"],props:{type:"text",placeholder:"Search"}}}
		]}})
		return child
	}}
	h.addPolicy=(p,s)=>{
		const index=h.Policys.findIndex(item=>item.name === p.name);
		const data={...p,state: 0, display: false, effect:policyEffects(p),value:policyStringify(p.statement, null, "\t")}
		if(index>=0){data.display=h.Policys[index].display;h.Policys.set(index,data)}else{h.Policys.push(data)}
		app.Context.Info((s===true?"flush":"update")+" policy ${0} success".format(p.name));
	}
	h.addMember=(m,s) => {
		const index=h.Members.findIndex(item=>item.user === m.user);
		const data={ ...m, state: 0, display: false, policylist:(m.policy||[]).join(', '),datalist:(m.data||[]).join(', ')}
		if(index>=0){data.display=h.Members[index].display;h.Members.set(index,data)}else{h.Members.push(data)}
		app.Context.Info((s===true?"flush":"update")+" member ${0} success".format(m.user));
	}

	h.onPolicyEdit=(e)=>{let pre=e.target; pre.setAttribute('contenteditable', 'true'); pre.focus();}
	h.onPolicyBlur=(e)=>{
		const index=getEventIndex(e); let data=h.Policys[index]; let pre=e.target; let text=pre.innerText;
		pre.setAttribute('contenteditable', 'false'); pre.innerHTML=text;
		try {
			let parsed=JSON.parse(text); data.state=JSON.stringify(data.statement)===JSON.stringify(parsed)?0:1;
		} catch (err) {
			app.Context.Error(err.message); data.state=2;
		}
		data.value=text
	}
	
	h.onPolicyCopy=(e)=>{
		e.stopPropagation(); const index=getEventIndex(e); const policy=h.Policys[index];
		const body='{"name":"${0}", "statement":${1}}'.format(policy.name,policy.value)
		copyValue(policyStringify(JSON.parse(body), null, "\t"))
	}
	h.onPolicyFlush=(e)=>{
		e.stopPropagation(); const index=getEventIndex(e); const policy=h.Policys[index];
		app.Context.Fetch({url:"policys/"+policy.name, success: (p) => h.addPolicy(p, true)})
	}
	h.onPolicyUpdate=(e) =>{
		e.stopPropagation(); const index=getEventIndex(e); const policy=h.Policys[index];
		if(h.Policys[index].state==2){app.Context.Error("update policy ${0} error: data is invalid json".format(policy.name));return}
		const body='{"name":"${0}", "statement":${1}}'.format(policy.name,policy.value)
		app.Context.Fetch({url:"policys/"+policy.name, method: 'PUT', body: body,success:h.addPolicy})
	}
	h.onPolicyDelete=(e) =>{
		e.stopPropagation(); const index=getEventIndex(e); const policy=h.Policys[index];
		app.Context.Fetch({url:"policys/"+policy.name, method: 'DELETE', success: (p) => {
			h.Policys.splice(index, 1);
			app.Context.Info("delete policy ${0} success".format(policy.name));
		}})
	}
	h.onMemberCopy=(e)=>{
		e.stopPropagation(); const index=getEventIndex(e); const member=h.Members[index];
		const body={user:member.user,policy:member.policy,data:member.data}
		copyValue(policyStringify(body))
	}
	h.onMemberFlush=(e) =>{
		e.stopPropagation(); const index=getEventIndex(e); const member=h.Members[index];
		app.Context.Fetch({url:"members/"+member.user, success: (m)=>{h.addMember(m, true)}})
	}
	h.onMemberUpdate=(e) =>{
		e.stopPropagation(); const index=getEventIndex(e); const member=h.Members[index];
		const body={user:member.user,policy:member.policy,data:member.data}
		app.Context.Fetch({url:"members/"+member.user, method: 'PUT', data: body,success: h.addMember})
	}
	h.onMemberDelete=(e) =>{
		e.stopPropagation(); const index=getEventIndex(e); const user=h.Members[index].user;
		app.Context.Fetch({url:"members/"+user, method: 'DELETE', success: (p) => {
			h.Members.splice(index, 1);
			app.Context.Info("delete member ${0} success".format(user));
		}})
	}
 	h.checkPolicy=()=>{try{let parsed=JSON.parse(h.Commit.ps);h.Commit.ps=policyStringify(parsed);
 		return Array.isArray(parsed)&& h.Commit.pn.trim()!== "";}catch(err){return false}}
 	h.checkMember=()=>{return h.Commit.mn.trim()&&(h.Commit.bp.trim()||h.Commit.bd.trim())}
 	h.checkCustom=()=>{try{let parsed=JSON.parse(h.Commit.cd);h.Commit.cd=policyStringify(parsed);return true}catch(err){return false}}
	h.commitPolicy=(e) =>{
		if(!h.checkPolicy())return
		const body='{"name":"${0}", "statement":${1}}'.format(h.Commit.pn,h.Commit.ps)
		app.Context.Fetch({url:"policys/"+h.Commit.pn, method: 'PUT', body: body, success: h.addPolicy})
	}
	h.commitMemeber=(e) =>{
		if(!h.checkMember())return
		const body={user:h.Commit.mn, policy:h.Commit.bp?[h.Commit.bp]:null,data:h.Commit.bd?[h.Commit.bd]:null}
		app.Context.Fetch({url:"members/"+h.Commit.mn,method: 'POST',data: body, success: h.addMember})
	}
	h.commitCustom=(_,d) =>{
		let data=d||JSON.parse(h.Commit.cd);
		if(Array.isArray(data)){for(let v of data){h.commitCustom(_,v)}return}; d=data;
		if(d.name){app.Context.Fetch({url:"policys/"+d.name, method: 'PUT', data: d, success:h.addPolicy})}
		if(d.user){app.Context.Fetch({url:"members/"+d.user, method: 'PUT', data: d, success:h.addMember})}
	}
	return h
}
`
