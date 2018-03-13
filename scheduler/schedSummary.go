package scheduler

import (
	"bytes"
)

type SchedSummary interface {
	String() string               //摘要一般信息
	Detail() string               //摘要详细信息
	Same(other SchedSummary) bool //是否与另一个摘要信息相同
}

type mySchedSummary struct {
	prefix     string
	running    uint32
	poolSize   uint32
	channelLen uint
	crawlDepth uint32

	chanmanSummary      string
	reqCacheSummary     string
	itemPipelineSummary string
	stopSignSummary     string

	dlPoolLen       uint32
	dlPoolCap       uint32
	analyzerPoolLen uint32
	analyzerPoolCap uint32

	urlCount  int
	urlDetail string
}

func (this *mySchedSummary) String() string {
	panic("implement me")
}

func (this *mySchedSummary) Detail() string {
	panic("implement me")
}

func (this *mySchedSummary) Same(other SchedSummary) bool {
	panic("implement me")
}

func NewSchedSummary(sched *myScheduler, prefix string) SchedSummary {

	urlCount := len(sched.urlMap)

	var urlDetail string
	if urlCount > 0 {
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for k, _ := range sched.urlMap {
			buffer.WriteString(prefix)
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		urlDetail = buffer.String()
	} else {
		urlDetail = "\n"
	}
	return &mySchedSummary{
		prefix:              prefix,
		running:             sched.running,
		poolSize:            sched.poolSize,
		channelLen:          sched.channelLen,
		crawlDepth:          sched.crawlDepth,
		chanmanSummary:      sched.chanman.Summary(),
		reqCacheSummary:     sched.reqCache.summary(),
		dlPoolLen:           sched.dlpool.Used(),
		dlPoolCap:           sched.dlpool.Total(),
		analyzerPoolLen:     sched.analyzerPool.Used(),
		analyzerPoolCap:     sched.analyzerPool.Total(),
		itemPipelineSummary: sched.itempipeline.Summary(),
		urlCount:            urlCount,
		urlDetail:           urlDetail,
		stopSignSummary:     sched.stopSign.Summary(),
	}
}
