/*
Cache

定义一个数据缓存对象。
*/
package eudore


import (
	"fmt"
	"time"
	"unsafe"
	"strings"
	"sync"
	"sync/atomic"
)

type ( 
	Cache interface {
		Component
		// get cached value by key.
		Get(string) interface{}
		// set cached value with key and expire time.
		Set(string, interface{}, time.Duration) error
		// delete cached value by key.
		Delete(string) error
		// check if cached value exists or not.
		IsExist(string) bool
		// get all keys
		GetAllKeys() []string
		// get keys size
		Count() int
		// clean all cache.
		CleanAll() error
		// start gc routine based on config string settings.
		// StartAndGC(config string) error
	}
	// Add an expiration time implementation based on sync.Map.
	//
	// 基于sync.Map添加过期时间实现。
	CacheMap struct {
		mu			sync.Mutex
		read 		atomic.Value
		// *interface{}
		dirty		map[string]*cacheEntry
		misses int
	}
	readOnly struct {
		m			map[string]*cacheEntry
		amended		bool
	}
	cacheEntry struct {
		expire		time.Time
		p			unsafe.Pointer // *interface{}
	}
	CacheGroupConfig struct {
		Keys	[]string
		Vals	[]interface{}
	}
	// Cache combination to match different keys to the specified cache processing.
	//
	// 缓存组合，将不同键匹配给指定缓存处理。
	CacheGroup struct {
		*CacheGroupConfig
		Caches	[]Cache
	}
)

func NewCache(name string, arg interface{}) (Cache, error) {
	name = ComponentPrefix(name, "cache")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	l, ok := c.(Cache)
	if ok {
		return l, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Cache type", name)
}



func NewCacheMap() (Cache, error) {
	return &CacheMap{}, nil
}

func NewCacheGroup(arg interface{}) (Cache, error) {
	cf, ok := arg.(*CacheGroupConfig)
	if !ok {
		return nil, fmt.Errorf("create cachegroup arg not is CacheGroupConfig Pointer.")
	}
	c := &CacheGroup{
		CacheGroupConfig:	cf,
		Caches:				make([]Cache, len(cf.Keys)),
	}
	var err error
	for i, v := range cf.Vals {
		c.Caches[i], err = NewCache(ComponentGetName(v), v)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

var expunged = unsafe.Pointer(new(interface{}))

func (c *CacheMap) Get(key string) interface{} {
	read, _ := c.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		c.mu.Lock()
		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read, _ = c.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = c.dirty[key]
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			c.missLocked()
		}
		c.mu.Unlock()
	}
	if !ok {
		return nil
	}
	return e.load()
}


func (e *cacheEntry) load() interface{} {
	// check time
	if time.Now().After(e.expire) {
		return nil
	}
	// get value
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expunged {
		return nil
	}
	return *(*interface{})(p)
}




// Store sets the value for a key.
func (c *CacheMap) Set(key string, val interface{}, timeout time.Duration) error {
	read, _ := c.read.Load().(readOnly)
	if e, ok := read.m[key]; ok && e.tryStore(&val, timeout) {
		return nil
	}

	c.mu.Lock()
	read, _ = c.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			c.dirty[key] = e
		}
		e.storeLocked(&val)
	} else if e, ok := c.dirty[key]; ok {
		e.storeLocked(&val)
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			c.dirtyLocked()
			c.read.Store(readOnly{m: read.m, amended: true})
		}
		c.dirty[key] = &cacheEntry{expire: time.Now().Add(timeout), p: unsafe.Pointer(&val)}
	}
	c.mu.Unlock()
	return nil
}


func (e *cacheEntry) tryStore(i *interface{}, timeout time.Duration) bool {
	p := atomic.LoadPointer(&e.p)
	if p == expunged {
		return false
	}
	for {
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			e.expire = time.Now().Add(timeout)
			return true
		}
		p = atomic.LoadPointer(&e.p)
		if p == expunged {
			return false
		}
	}
}



func (e *cacheEntry) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

func (e *cacheEntry) storeLocked(i *interface{}) {
	atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

// Delete deletes the value for a key.
func (c *CacheMap) Delete(key string) error {
	read, _ := c.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		c.mu.Lock()
		read, _ = c.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			delete(c.dirty, key)
		}
		c.mu.Unlock()
	}
	if ok {
		e.delete()
	}
	return nil
}

func (e *cacheEntry) delete() (hadValue bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == expunged {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return true
		}
	}
}



func (c *CacheMap) missLocked() {
	c.misses++
	if c.misses < len(c.dirty) {
		return
	}
	c.read.Store(readOnly{m: c.dirty})
	c.dirty = nil
	c.misses = 0
}


func (c *CacheMap) dirtyLocked() {
	if c.dirty != nil {
		return
	}

	read, _ := c.read.Load().(readOnly)
	c.dirty = make(map[string]*cacheEntry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			c.dirty[k] = e
		}
	}
}

func (e *cacheEntry) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
	}
	return p == expunged
}

func (c *CacheMap) IsExist(key string) bool {
	return c.Get(key) != nil
}

func (c *CacheMap) GetAllKeys() []string {
	read, _ := c.read.Load().(readOnly)
	strs := make([]string, 0, len(read.m))
	for k, _ := range read.m {
		strs = append(strs, k)
	}
	return strs
}

func (c *CacheMap) Count() int {
	read, _ := c.read.Load().(readOnly)
	return len(read.m)
}


func (c *CacheMap) CleanAll() error {
	return nil
}


func (c *CacheMap) GetName() string {
	return ComponentCacheMapName
}

func (c *CacheMap) Version() string {
	return ComponentCacheMapVersion
}



func (c *CacheGroup) getCache(key string) Cache {
	for i, k := range c.Keys {
		if strings.HasPrefix(k, key) {
			return c.Caches[i]
		}
	}
	return nil
}

func (c *CacheGroup) Get(key string) interface{} {
	return c.getCache(key).Get(key)
}

func (c *CacheGroup) Set(key string, val interface{}, timeout time.Duration) error {
	return c.getCache(key).Set(key, val, timeout)
}

func (c *CacheGroup) Delete(key string) error {
	return c.getCache(key).Delete(key)
}

func (c *CacheGroup) IsExist(key string) bool {
	return c.getCache(key).IsExist(key)
}

func (c *CacheGroup) GetAllKeys() []string {
	return nil
}

func (c *CacheGroup) Count() (n int) {
	for _, i := range c.Caches {
		n += i.Count()
	}
	return
}

func (c *CacheGroup) CleanAll() (err error) {
	for _, i := range c.Caches {
		err = i.CleanAll()
		if err != nil {
			return err
		}
	}
	return nil
} 

func (*CacheGroup) Version() string {
	return ComponentCacheGroupVersion
}


func (*CacheGroupConfig) GetName() string {
	return ComponentCacheGroupName
}
