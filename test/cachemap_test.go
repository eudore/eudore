package test

import (
	"github.com/eudore/eudore"
	"testing"
	"time"
)

func TestCacheMap(t *testing.T) {
	c := &eudore.CacheMap{}
	var (
		key1 = "key1"
		key2 = "key2"
	)
	// put
	c.Set(key1, "val", 1*time.Second)
	c.Set(key2, "val 1", 100*time.Second)
	// get
	t.Log("1 key1: ", c.Get(key1))
	t.Log("1 key2: ", c.Get(key2))
	// exprie
	time.Sleep(2 * time.Second)
	t.Log("2 key1: ", c.Get(key1))
	t.Log("2 key2: ", c.Get(key2))
	// test delete
	c.Delete(key2)
	t.Log("3 key2: ", c.Get(key2))
}
