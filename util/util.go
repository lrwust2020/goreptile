package util

import (
	"sync/atomic"
)

type IdGenertor interface {
	GetUint32() uint32
}
type myDefaultIdGenertor struct {
	id uint32
}

func (this *myDefaultIdGenertor) GetUint32() uint32 {
	this.id = atomic.AddUint32(&this.id, 1)
	var offset uint32
	for {
		offset = this.id
		if atomic.CompareAndSwapUint32(&this.id, offset, (1 + this.id)) {
			break
		}
	}
	return this.id
}

func DefaultIdGenertor() IdGenertor {
	return &myDefaultIdGenertor{}
}
