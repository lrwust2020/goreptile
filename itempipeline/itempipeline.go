package itempipeline

import "github.com/fmyxyz/goreptile/base"

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
type ProcessItem func(item base.Item) (result base.Item,err error)

