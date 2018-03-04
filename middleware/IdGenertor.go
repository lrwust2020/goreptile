package middleware

import (
	"sync/atomic"
	"sync"
	"math"
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

type cyclicIdGenertor struct {
	sn    uint32
	ended bool
	mux   sync.Mutex
}

func NewCyclicIdGenertor() IdGenertor {
	return &cyclicIdGenertor{}
}
func (this *cyclicIdGenertor) GetUint32() uint32 {
	this.mux.Lock()
	defer this.mux.Unlock()

	if this.ended {
		defer func() {
			this.ended = false
		}()
		this.sn = 0
		return this.sn
	}

	var id = this.sn
	if id < math.MaxUint32 {
		this.sn++
	} else {
		this.ended = true
	}
	return id

}
