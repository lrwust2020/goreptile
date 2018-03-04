package scheduler

import (
	"github.com/fmyxyz/goreptile/analyzer"
	"github.com/fmyxyz/goreptile/itempipeline"
	"net/http"
)

type Scheduler interface {
	//启动调度器
	Start(channelLen uint,
		poolSize uint32,
		crawlDepth uint32,
		httpClientGenerator GenHttpClient,
		respParsers []analyzer.ParseResponse,
		itemProcessors []itempipeline.ProcessItem,
		firstHttpReq *http.Request) (err error)
	//停止调度器，所有处理模块都会停止
	Stop() bool
	//调度器是否在运行
	Running() bool
	//错误通道，调度器及各个处理模块出现的错误
	//nil 表示通道不可用或调度器已停止
	ErrorChan() <-chan error
	//判断所有处理模块是否空闲
	Idle() bool
	//摘要
	Summary(prefix string) SchedSummary
}

//生成HTTP客户端
type GenHttpClient func() *http.Client

type SchedSummary interface {
	String() string               //摘要一般信息
	Detail() string               //摘要详细信息
	Same(other SchedSummary) bool //是否与另一个摘要信息相同
}

type myScheduler struct {
}

func (this *myScheduler) Start(channelLen uint,
	poolSize uint32,
	crawlDepth uint32,
	httpClientGenerator GenHttpClient,
	respParsers []analyzer.ParseResponse,
	itemProcessors []itempipeline.ProcessItem,
	firstHttpReq *http.Request) (err error) {
	panic("implement me")
}

func (this *myScheduler) Stop() bool {
	panic("implement me")
}

func (this *myScheduler) Running() bool {
	panic("implement me")
}

func (this *myScheduler) ErrorChan() <-chan error {
	panic("implement me")
}

func (this *myScheduler) Idle() bool {
	panic("implement me")
}

func (this *myScheduler) Summary(prefix string) SchedSummary {
	panic("implement me")
}
