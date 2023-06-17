package gcache

import (
	"fmt"
	"sync"
	"time"
)

type GCache struct {
	cache    map[string]Object
	mtx      sync.Mutex
	watchDog time.Ticker
}

var neverExpired time.Time = time.Unix(0, 0)

var gcache *GCache

type IGCache interface {
	Get(k string) interface{}
	GetPtr(k string) *interface{}
	Add(k string, v interface{}, expire time.Duration) error
	AddDefault(k string, v interface{}) error
	Set(k string, v interface{}) error
	runWatchDog()
	check()
	onExpired(k string)
}

func init() {
	gcache = &GCache{
		cache:    map[string]Object{},
		mtx:      sync.Mutex{},
		watchDog: *time.NewTicker(500 * time.Millisecond),
	}
	gcache.runWatchDog()
	fmt.Println("gcache inited")
}

func (c *GCache) Get(k string) interface{} {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	obj, ok := c.cache[k]
	if !ok {
		return nil
	}
	if obj.CheckHasExpired() {
		c.onExpired(k)
		return nil
	}
	return obj.GetObject()
}

func (c *GCache) GetPtr(k string) *interface{} {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	obj, ok := c.cache[k]
	if !ok {
		return nil
	}
	if obj.CheckHasExpired() {
		c.onExpired(k)
		return nil
	}
	return obj.GetObjectPtr()
}

func (c *GCache) Add(k string, v interface{}, expire time.Duration) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.cache[k]; ok {
		return fmt.Errorf("key: %s has existed, if you want to update, please use set", k)
	}
	c.cache[k] = Object{
		obj:    v,
		expire: time.Now().Add(expire),
	}
	return nil
}

func (c *GCache) AddDefault(k string, v interface{}) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.cache[k]; ok {
		return fmt.Errorf("key: %s has existed, if you want to update, please use set", k)
	}
	c.cache[k] = Object{
		obj:    v,
		expire: neverExpired,
	}
	return nil
}

func (c *GCache) Set(k string, v interface{}) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.cache[k]; !ok {
		return fmt.Errorf("key: %s not exist, please add first", k)
	}
	oldObj := c.cache[k]
	if oldObj.CheckHasExpired() {
		c.onExpired(k)
		return fmt.Errorf("object of key: %s has expired", k)
	}
	newObj := Object{
		obj:    oldObj.GetObject(),
		expire: oldObj.GetExpire(),
	}
	c.cache[k] = newObj
	return nil
}

func (c *GCache) runWatchDog() {
	go func() {
		for {
			<-c.watchDog.C
			c.check()
		}
	}()
}

func (c *GCache) onExpired(k string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.cache, k)
}

func (c *GCache) check() {
	for k, o := range c.cache {
		if o.CheckHasExpired() {
			c.onExpired(k)
		}
	}
}
