package itempipeline

import (
	"errors"
	"fmt"
	"github.com/fmyxyz/goreptile/base"
	"sync/atomic"
)

//条目处理管道
type ItemPipeline interface {
	//发送条目
	Send(item base.Item) []error
	//failfast 方法返回一个布尔值，该值表示当前条目处理管道是否是快速失败的
	//这里的快速失败是指：只有对某个条目的处理流程上在某一个步骤上出错，那么条目处理管道就会忽略后续的所有处理步骤并报告错误
	FailFast() bool
	//设置是否快速失败
	SetFailFast(failFast bool)
	//获取已发送、已接收、已处理的条目的计数值
	//切片长度为3
	Count() []uint64
	//正在被处理的条目数量
	ProcessingNumber() uint64
	//摘要信息
	Summary() string
}

//处理条目
type ProcessItem func(item base.Item) (result base.Item, err error)

type myItemPipeline struct {
	itemProcessors   []ProcessItem
	failFast         bool
	sent             uint64
	accepted         uint64
	processed        uint64
	processingNumber uint64
}

func (this *myItemPipeline) Send(item base.Item) []error {
	atomic.AddUint64(&this.processingNumber, 1)
	defer atomic.AddUint64(&this.processingNumber, ^uint64(0))
	atomic.AddUint64(&this.sent, 1)
	errs := make([]error, 0)
	if item == nil {
		errs = append(errs, errors.New("The item is invalid!"))
		return errs
	}
	atomic.AddUint64(&this.accepted, 1)
	var currentItem = item
	for _, itemProcesser := range this.itemProcessors {
		processedItem, err := itemProcesser(currentItem)
		if err != nil {
			errs = append(errs, err)
			if this.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	atomic.AddUint64(&this.processed, 1)
	return errs
}

func (this *myItemPipeline) FailFast() bool {
	return this.failFast
}

func (this *myItemPipeline) SetFailFast(failFast bool) {
	this.failFast = failFast
}

func (this *myItemPipeline) Count() []uint64 {
	var counts = make([]uint64, 3)
	counts[0] = atomic.LoadUint64(&this.sent)
	counts[1] = atomic.LoadUint64(&this.accepted)
	counts[2] = atomic.LoadUint64(&this.processed)
	return counts
}

var summaryTemplate = "failFast:%v,processornumber:%d,sent:%d,accepted:%d,processed:%d,processingNumber:%d"

func (this *myItemPipeline) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&this.processingNumber)
}

func (this *myItemPipeline) Summary() string {
	var counts = this.Count()
	var summary = fmt.Sprintf(summaryTemplate, this.FailFast(), len(this.itemProcessors), counts[0], counts[1], counts[2], this.ProcessingNumber())
	return summary
}

func NewItemPipeline(itemProcessors []ProcessItem) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New("Invalid item processor list!"))
	}
	var innerItemProcessors = make([]ProcessItem, 0)
	for i, ip := range itemProcessors {
		if ip == nil {
			panic(errors.New(fmt.Sprintf("Invalid item processor[%d]!", i)))
		}
		innerItemProcessors = append(innerItemProcessors, ip)

	}
	return &myItemPipeline{itemProcessors: innerItemProcessors}
}
