package middleware

import (
	"sort"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// Define the breaker state.
const (
	breakerStatueClosed = iota
	breakerStatueHalfOpen
	breakerStatueOpen
)

// breakerStatues defines the breaker status string.
var breakerStatues = []string{"closed", "half-open", "open"}

type breaker struct {
	Entrys             sync.Map
	GetKeyFunc         func(eudore.Context) string
	GetBreakrEntryFunc func(string) breakerEntry
}

type breakerEntry interface {
	OnAccess() bool
	OnSucceed() bool
	OnFailed() bool
	SetState(state int)
}

// The NewCircuitBreakerFunc function creates middleware to implement
// handle request breaking.
//
// This middleware does not support cluster mode.
//
// options: [NewOptionKeyFunc]
// [NewOptionCircuitBreakerConfig] [NewOptionRouter].
func NewCircuitBreakerFunc(options ...Option) Middleware {
	b := &breaker{
		GetKeyFunc: func(ctx eudore.Context) string {
			return ctx.GetParam(eudore.ParamRoute)
		},
		GetBreakrEntryFunc: newBreakerEntryfunc(10, 10,
			400*time.Microsecond, 10*time.Second,
		),
	}
	applyOption(b, options)

	release := func(ctx eudore.Context, entry breakerEntry, name string) {
		if ctx.Response().Status() < eudore.StatusInternalServerError {
			if entry.OnSucceed() {
				ctx.Infof("Breaker route %s change state to %s",
					name, breakerStatues[breakerStatueClosed],
				)
			}
		} else if entry.OnFailed() {
			ctx.Infof("Breaker route %s change state to %s", name, breakerStatues[breakerStatueOpen])
		}
	}
	return func(ctx eudore.Context) {
		key := b.GetKeyFunc(ctx)
		if key == "" {
			return
		}

		val, ok := b.Entrys.Load(key)
		if !ok {
			val = b.GetBreakrEntryFunc(key)
			b.Entrys.Store(key, val)
		}
		entry := val.(breakerEntry)
		if !entry.OnAccess() {
			writePage(ctx, eudore.StatusServiceUnavailable, DefaultPageCircuitBreaker, key)
			ctx.End()
			return
		}

		defer release(ctx, entry, key)
		ctx.Next()
	}
}

func (b *breaker) Inject(_ eudore.Controller, router eudore.Router) error {
	router.GetFunc("/breaker Action=middleware:breaker:GetBreaker", b.GetBreaker)
	router.GetFunc("/breaker/:key Action=middleware:breaker:GetBreakerByKey", b.GetBreakerByKey)
	router.PutFunc("/breaker/:key/state/:state Action=middleware:breaker:PutBreakerByKeyStateByState", b.PutBreakerByKeyStateByState)
	return nil
}

func (b *breaker) GetBreaker(ctx eudore.Context) {
	type pair struct {
		K string
		V any
	}
	var data []pair
	b.Entrys.Range(func(key, value any) bool {
		data = append(data, pair{key.(string), value})
		return true
	})
	sort.Slice(data, func(i, j int) bool {
		return data[i].K < data[j].K
	})

	entrys := make([]any, len(data))
	for i := range data {
		entrys[i] = data[i].V
	}
	_ = ctx.Render(entrys)
}

func (b *breaker) GetBreakerByKey(ctx eudore.Context) {
	key := ctx.GetParam("key")
	key64, err := base64Encoding.DecodeString(key)
	if err == nil {
		key = string(key64)
	}

	entry, ok := b.Entrys.Load(key)
	if !ok {
		ctx.Fatalf("key is invalid %s", key)
		return
	}
	_ = ctx.Render(entry)
}

func (b *breaker) PutBreakerByKeyStateByState(ctx eudore.Context) {
	key := ctx.GetParam("key")
	key64, err := base64Encoding.DecodeString(key)
	if err == nil {
		key = string(key64)
	}

	entry, ok := b.Entrys.Load(key)
	if !ok {
		ctx.Fatalf("key is inval %s", key)
		return
	}
	state := eudore.GetAnyByString[int](ctx.GetParam("state"))
	if state < -1 || state > 2 {
		ctx.Fatal("state is invalid")
		return
	}
	entry.(breakerEntry).SetState(state)
	ctx.Infof("Breaker route %s set state to %s", key, breakerStatues[state])
	_ = ctx.Render(entry)
}

// breakerEntryDefault defines the breaker data for a single entry.
type breakerEntryDefault struct {
	sync.Mutex `json:"-"`
	State      int    `json:"state"`
	Name       string `json:"name"`
	// config
	MaxConsecutiveSuccesses int           `json:"-"`
	MaxConsecutiveFailures  int           `json:"-"`
	HalfOpenWait            time.Duration `json:"-"`
	HalfOpenInterval        time.Duration `json:"-"`
	HalfOpenLast            time.Time     `json:"-"`
	// state
	LastTime             time.Time `json:"lastTime"`
	ConsecutiveSuccesses int       `json:"consecutiveSuccesses"`
	ConsecutiveFailures  int       `json:"consecutiveFailures"`
	TotalSuccesses       uint64    `json:"totalSuccesses"`
	TotalFailures        uint64    `json:"totalFailures"`
}

func newBreakerEntryfunc(maxSuccesses, maxFailures int, t, wait time.Duration,
) func(name string) breakerEntry {
	return func(name string) breakerEntry {
		return &breakerEntryDefault{
			Name:                    name,
			MaxConsecutiveSuccesses: maxSuccesses,
			MaxConsecutiveFailures:  maxFailures,
			HalfOpenInterval:        t,
			HalfOpenWait:            wait,
		}
	}
}

func (c *breakerEntryDefault) OnAccess() bool {
	now := time.Now()
	c.Lock()
	c.LastTime = now
	allow := c.State == breakerStatueClosed ||
		(c.State == breakerStatueHalfOpen && c.OnHalfOpen(now))
	c.Unlock()
	return allow
}

func (c *breakerEntryDefault) OnHalfOpen(now time.Time) bool {
	if now.After(c.HalfOpenLast.Add(c.HalfOpenInterval)) {
		c.HalfOpenLast = now
		return true
	}
	return false
}

func (c *breakerEntryDefault) OnSucceed() bool {
	c.Lock()
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
	change := c.State != breakerStatueClosed &&
		c.ConsecutiveSuccesses >= c.MaxConsecutiveSuccesses
	if change {
		c.ConsecutiveSuccesses = 0
		c.State = breakerStatueClosed
	}
	c.Unlock()
	return change
}

func (c *breakerEntryDefault) OnFailed() bool {
	c.Lock()
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
	change := c.State != breakerStatueOpen &&
		c.ConsecutiveFailures >= c.MaxConsecutiveFailures
	if change {
		c.ConsecutiveFailures = 0
		c.State = breakerStatueOpen
		c.RetryClose()
	}
	c.Unlock()
	return change
}

func (c *breakerEntryDefault) SetState(state int) {
	c.Lock()
	c.State = state
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
	c.RetryClose()
	c.Unlock()
}

func (c *breakerEntryDefault) RetryClose() {
	if c.State == breakerStatueOpen {
		go func() {
			time.Sleep(c.HalfOpenWait)
			c.Lock()
			if c.State == breakerStatueOpen {
				c.State--
			}
			c.Unlock()
		}()
	}
}

const breakerScript = `
function NewHandlerBreaker() {
	const states = ["closed", "half-open", "open"];
	const classs = ["state-info", "state-warning", "state-error"];
	const reservedChars = ['/', '?', '&', '#', ':', ' '];
	function encodeKey(key) {
		if(reservedChars.some(char => key.includes(char))){
			return btoa(key).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
		}
		return key
	}
	const h = {
	Entrys: [],
	Mount(ctx) {
		ctx.Fetch({url: "breaker", success: (data) => {
			this.Entrys = data.map((b) => {return {...b, display: false}})
		}})
	},
	View(ctx) {
		let state = {totalSuccesses: 0, totalFailures: 0, closed: 0, open: 0, "half-open": 0};
		let child = this.Entrys.map((data, index)=>{
			state[states[data.state]]++;
			state.totalSuccesses += data.totalSuccesses;
			state.totalFailures += data.totalFailures;
			return {type: "div", class: "breaker-node", props: {index:index},
				div: {class: "state " + classs[data.state], onclick:()=>{data.display = !data.display}, span: [
					{text: data.name},
					{text:(data.totalSuccesses.toFixed(2)/(data.totalSuccesses+data.totalFailures).toFixed(2)*100).toFixed(2)+"%"},
					{html: svgFlush, onclick: this.onFlush}
				]},
				table: {
					style: "display: "+(data.display?"block":"none"), tbody: {tr: [
						{td: [{text: 'state'}, {select: {
							class: "breaker-select",
							props: {name:'breaker-select'},
							child: [
								{type: 'option', text: "closed", props: (data.state==0?{selected: 'selected'}:{})},
								{type: 'option', text: "half-open", props: (data.state==1?{selected: 'selected'}:{})},
								{type: 'option', text: "open", props: (data.state==2?{selected: 'selected'}:{})},
							],
							onchange: this.onChange,
						}}]},
						{td: [{text: 'lastTime'}, {text: data.lastTime.slice(0, 19).replace("T", " ")}]},
						{td: [{text: 'totalSuccesses'}, {text: data.totalSuccesses}]},
						{td: [{text: 'totalFailures'}, {text: data.totalFailures}]},
						{td: [{text: 'consecutiveSuccesses'}, {text: data.consecutiveSuccesses}]},
						{td: [{text: 'consecutiveFailures'}, {text: data.consecutiveFailures}]},
					]}
				}
			}
		});
		child.unshift({type: 'div', id:"breaker-state", child: [
			{type: 'p', text: 'totalSuccesses: ' + state.totalSuccesses + " totalFailures: " + state.totalFailures},
			{type: 'p', text: "closed: " + state.closed + " half-open: " + state['half-open'] + " open: " + state.open}
		]})
		return child
	}}
	h.onFlush = (e) => {
		e.stopPropagation();
		const index = getEventIndex(e); 
		const entry = h.Entrys[index];
		app.Context.Fetch({url: "breaker/"+encodeKey(entry.name), success: (data)=>{
			h.Entrys.set(index, {...data, display: entry.display});
			app.Context.Info("flush state ${0}".format(entry.name))
		}})
	}
	h.onChange = (e) => {
		const index = getEventIndex(e); 
		const entry = h.Entrys[index];
		let state = e.target.selectedIndex;
		app.Context.Fetch({method: 'PUT', url: "breaker/"+ encodeKey(entry.name) + "/state/" + state, success: (data)=>{
			h.Entrys.set(index, {...data, display: entry.display});
			app.Context.Info("change state ${0} to ${1}".format(entry.name, states[state]))
		}})
	}
	return h
}
`
