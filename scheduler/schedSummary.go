package scheduler

import (
	"bytes"
	"fmt"
	"github.com/fmyxyz/goreptile/base"
)

type SchedSummary interface {
	String() string               //摘要一般信息
	Detail() string               //摘要详细信息
	Same(other SchedSummary) bool //是否与另一个摘要信息相同
}

type mySchedSummary struct {
	prefix  string
	running uint32

	channelArgs  base.ChannelArgs
	poolBaseArgs base.PoolBaseArgs

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
	return this.getSummary(false)
}

func (this *mySchedSummary) Detail() string {
	return this.getSummary(true)
}

func (this *mySchedSummary) Same(other SchedSummary) bool {
	if other == nil {

		return false
	}
	otherSs, ok := interface{}(other).(*mySchedSummary)
	if !ok {
		return false
	}

	if this.running != otherSs.running ||
		this.poolBaseArgs.AnalyzerPoolSize() != otherSs.poolBaseArgs.AnalyzerPoolSize() ||
		this.poolBaseArgs.PageDownloaderPoolSize() != otherSs.poolBaseArgs.PageDownloaderPoolSize() ||
		this.channelArgs.ErrorChanLen() != otherSs.channelArgs.ErrorChanLen() ||
		this.channelArgs.ItemChanLen() != otherSs.channelArgs.ItemChanLen() ||
		this.channelArgs.RespChanLen() != otherSs.channelArgs.RespChanLen() ||
		this.channelArgs.ReqChanLen() != otherSs.channelArgs.ReqChanLen() ||
		this.crawlDepth != otherSs.crawlDepth ||
		this.chanmanSummary != otherSs.chanmanSummary ||
		this.reqCacheSummary != otherSs.reqCacheSummary ||
		this.itemPipelineSummary != otherSs.itemPipelineSummary ||
		this.stopSignSummary != otherSs.stopSignSummary ||
		this.dlPoolLen != otherSs.dlPoolLen ||
		this.dlPoolCap != otherSs.dlPoolCap ||
		this.analyzerPoolLen != otherSs.analyzerPoolLen ||
		this.analyzerPoolCap != otherSs.analyzerPoolCap ||
		this.urlCount != otherSs.urlCount {
		return false
	} else {
		return true
	}

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
		poolBaseArgs:        sched.poolBaseArgs,
		channelArgs:         sched.channelArgs,
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

func (this *mySchedSummary) getSummary(detail bool) string {
	var template = this.prefix + "Running :%v \n" +
		this.prefix + "Pool base size args :%s \n" +
		this.prefix + "Channel args :%s \n" +
		this.prefix + "Crawl depth :%d \n" +
		this.prefix + "Channels manager :%s \n" +
		this.prefix + "Request cache :%s \n" +
		this.prefix + "Downloader pool :%d/%d \n" +
		this.prefix + "Analyzer pool :%d/%d \n" +
		this.prefix + "Item pipeline :%s \n" +
		this.prefix + "Url(%d) :%s \n" +
		this.prefix + "Stop sign :%s \n"
	return fmt.Sprintf(template,
		func() bool { return this.running == 1 }(),
		this.poolBaseArgs.String(),
		this.channelArgs.String(),
		this.crawlDepth,
		this.chanmanSummary,
		this.reqCacheSummary,
		this.dlPoolLen, this.dlPoolCap,
		this.analyzerPoolLen, this.analyzerPoolCap,
		this.itemPipelineSummary,
		this.urlCount,
		func() string {
			if detail {
				return this.urlDetail
			} else {
				return "<concealed>\n"
			}
		}(),
		this.stopSignSummary,
	)

}
