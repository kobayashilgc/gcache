package gcache

import (
	"context"
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
	GetLeft(k string) (time.Duration, error)
	GetExpire(k string) (time.Time, error)
	Add(k string, v interface{}, expire time.Duration) error
	AddDefault(k string, v interface{}) error
	Set(k string, v interface{}) error
	checkObjectExpired(k string) error
	runWatchDog(ctx context.Context)
	checkAll()
	onExpired(k string)
	checkKeyExist(k string) bool
}

func GCacheFactory(syncPeriod time.Duration, ctx context.Context) *GCache {
	c := &GCache{
		cache:    map[string]Object{},
		mtx:      sync.Mutex{},
		watchDog: *time.NewTicker(syncPeriod),
	}
	c.runWatchDog(ctx)
	fmt.Println("gcache inited")
	return c
}

func init() {
	gcache = GCacheFactory(500*time.Millisecond, context.Background())
}

func (c *GCache) Get(k string) interface{} {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if !c.checkKeyExist(k) {
		return nil
	}
	if c.checkObjectExpired(k) {
		return nil
	}
	obj := c.cache[k]
	return obj.getObject()
}

func (c *GCache) GetPtr(k string) *interface{} {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if !c.checkKeyExist(k) {
		return nil
	}
	if c.checkObjectExpired(k) {
		return nil
	}
	obj := c.cache[k]
	return obj.getObjectPtr()
}

func (c *GCache) GetLeft(k string) (time.Duration, error) {
	if !c.checkKeyExist(k) {
		return 0, fmt.Errorf("key: %s has existed", k)
	}
	obj := c.cache[k]
	return obj.getLeft(), nil
}

func (c *GCache) GetExpire(k string) (time.Time, error) {
	if !c.checkKeyExist(k) {
		return time.Now(), fmt.Errorf("key: %s has existed", k)
	}
	obj := c.cache[k]
	return obj.getExpire(), nil
}

func (c *GCache) Add(k string, v interface{}, strategy Strategy) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.checkKeyExist(k) {
		return fmt.Errorf("key: %s has existed, if you want to update, please use set", k)
	}
	if err := strategy.checkValid(); err != nil {
		return err
	}
	c.cache[k] = Object{
		obj:      v,
		strategy: strategy,
		expire:   time.Now().Add(strategy.Life),
	}
	return nil
}

func (c *GCache) AddDefault(k string, v interface{}) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if c.checkKeyExist(k) {
		return fmt.Errorf("key: %s has existed, if you want to update, please use set", k)
	}
	c.cache[k] = Object{
		obj:    v,
		expire: neverExpired,
		strategy: Strategy{
			Mode: Never,
		},
	}
	return nil
}

func (c *GCache) Set(k string, v interface{}) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if !c.checkKeyExist(k) {
		return fmt.Errorf("key: %s not exist, please add first", k)
	}
	oldObj := c.cache[k]
	if c.checkObjectExpired(k) {
		return fmt.Errorf("object of key: %s has expired", k)
	}
	newObj := Object{
		obj:    oldObj.getObject(),
		expire: oldObj.getExpire(),
	}
	c.cache[k] = newObj
	return nil
}

func (c *GCache) checkObjectExpired(k string) bool {
	oldObj := c.cache[k]
	if !oldObj.checkHasExpired() {
		return false
	}
	if oldObj.strategy.Mode == Lazy {
		c.onExpired(k)
	}
	return true
}

func (c *GCache) runWatchDog(ctx context.Context) {
	go func(ctx context.Context) {
		var spin bool = true
		for spin {
			select {
			case <-c.watchDog.C:
				c.checkAll()
			case <-ctx.Done():
				spin = false
			}
		}
	}(ctx)
}

func (c *GCache) onExpired(k string) {
	delete(c.cache, k)
}

func (c *GCache) checkAll() {
	for k, o := range c.cache {
		if o.checkHasExpired() && o.strategy.Mode == Auto {
			c.onExpired(k)
		}
	}
}

func (c *GCache) checkKeyExist(k string) bool {
	_, ok := c.cache[k]
	return ok
}
