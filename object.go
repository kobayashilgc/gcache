package gcache

import "time"

type Object struct {
	obj      interface{}
	expire   time.Time
	strategy Strategy
}

type IObject interface {
	getObject() interface{}
	getObjectPtr() *interface{}
	checkHasExpired() bool
	getLeft() time.Duration
	getExpire() time.Time
}

func (o *Object) getObject() interface{} {
	return o.obj
}

func (o *Object) getObjectPtr() *interface{} {
	return &o.obj
}

func (o *Object) checkHasExpired() bool {
	if o.strategy.Mode == Never {
		return false
	}
	now := time.Now()
	return !now.After(o.expire)
}

func (o *Object) getLeft() time.Duration {
	if o.checkHasExpired() {
		return 0
	}
	now := time.Now()
	return o.expire.Sub(now)
}

func (o *Object) getExpire() time.Time {
	return o.expire
}
