package scheduler

import (
	"fmt"
	"github.com/fmyxyz/goreptile/base"
	"sync"
)

type requestCache interface {
	put(req *base.Request) bool
	get() *base.Request
	capacity() int
	length() int
	close()
	summary() string
}

type reqCacheBySlice struct {
	cache  []*base.Request
	mutex  sync.Mutex
	status byte //0 运行中 1已关闭
}

func (this *reqCacheBySlice) put(req *base.Request) bool {
	if req == nil {
		return false
	}
	if this.status == 1 {
		return false
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.cache = append(this.cache, req)
	return true
}

func (this *reqCacheBySlice) get() *base.Request {
	if this.length() == 0 {
		return nil
	}
	if this.status == 1 {
		return nil
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	req := this.cache[0]
	this.cache = this.cache[1:]
	return req

}

func (this *reqCacheBySlice) capacity() int {
	return cap(this.cache)
}

func (this *reqCacheBySlice) length() int {
	return len(this.cache)

}

func (this *reqCacheBySlice) close() {
	if this.status == 1 {
		return
	}
	this.status = 1
}

func (this *reqCacheBySlice) summary() string {
	return fmt.Sprintf("status:%s ,length:%d,capacity:%d", statusMap[this.status], this.length(), this.capacity())
}

var statusMap = map[byte]string{
	0: "running",
	1: "closed",
}

func newRequestCache() requestCache {
	rc := &reqCacheBySlice{
		cache: make([]*base.Request, 0),
	}
	return rc
}
