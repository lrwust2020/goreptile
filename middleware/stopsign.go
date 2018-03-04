package middleware

import (
	"sync"
)

//停止信号
type StopSign interface {
	//发出停止信号
	//若已发出过停止信号，则返回false
	Sign() bool
	//判断停止信号是否发出
	Signed() bool

	Reset()
	//处理停止信号
	//参数code代表停止信号处理方的代号，该代号会出现在处理记录中
	Deal(code string)
	//停止信号处理方的处理计数
	DealCount() uint32
	//被处理总计数
	DealTotal() uint32
	Summary() string
}

type myStopSign struct {
	signed       bool
	dealCountMay map[string]uint32
	rwMutex      sync.RWMutex
}

func (this *myStopSign) Sign() bool {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	if this.signed {
		return false
	}
	this.signed = true
	return true
}

func (this *myStopSign) Signed() bool {
	return this.signed
}

func (this *myStopSign) Reset() {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	this.signed = false
	this.dealCountMay = make(map[string]uint32)
}

func (this *myStopSign) Deal(code string) {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	if !this.signed {
		return
	}
	if _, ok := this.dealCountMay[code]; !ok {
		this.dealCountMay[code] = 1
	} else {
		this.dealCountMay[code] += 1
	}
}

func (this *myStopSign) DealCount() uint32 {
	panic("implement me")
}

func (this *myStopSign) DealTotal() uint32 {
	panic("implement me")
}

func (this *myStopSign) Summary() string {
	panic("implement me")
}

func NewStopSign() StopSign {
	return &myStopSign{
		dealCountMay: make(map[string]uint32),
	}
}
