package gcache

import "time"

type Object struct {
	obj    interface{}
	expire time.Time
}

type IObject interface {
	GetObject() interface{}
	GetObjectPtr() *interface{}
	CheckExpired() bool
	GetLeft() (time.Duration, bool)
	GetExpire() time.Time
}

func (o *Object) GetObject() interface{} {
	return o.obj
}

func (o *Object) GetObjectPtr() *interface{} {
	return &o.obj
}

func (o *Object) CheckHasExpired() bool {
	if o.expire == neverExpired {
		return false
	}
	now := time.Now()
	return !now.After(o.expire)
}

func (o *Object) GetLeft() (time.Duration, bool) {
	if o.CheckHasExpired() {
		return 0, false
	}
	now := time.Now()
	return o.expire.Sub(now), true
}

func (o *Object) GetExpire() time.Time {
	return o.expire
}
