package scheduler

import (
	"errors"
	"fmt"
	"github.com/fmyxyz/goreptile/analyzer"
	"github.com/fmyxyz/goreptile/base"
	"github.com/fmyxyz/goreptile/downloader"
	"github.com/fmyxyz/goreptile/itempipeline"
	"github.com/fmyxyz/goreptile/middleware"
	"log"
	"net/http"
	"os"
	"strings"
	//"sync"
	"sync/atomic"
	"time"
)

type Scheduler interface {
	//启动调度器
	Start(channelArgs base.ChannelArgs,
		poolBaseArgs base.PoolBaseArgs,
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

type myScheduler struct {
	channelArgs  base.ChannelArgs
	poolBaseArgs base.PoolBaseArgs

	crawlDepth    uint32 //深度
	primaryDomain string //主域名

	chanman      middleware.ChannelManager     //通道管理器
	stopSign     middleware.StopSign           //停止信号
	dlpool       downloader.PageDownloaderPool //网页下载器池
	analyzerPool analyzer.AnalyzerPool         //分析器池
	itempipeline itempipeline.ItemPipeline     //条目处理管道

	running uint32 //运行标记 0未运行 1已运行 2已停止

	reqCache requestCache //请求缓存

	urlMap map[string]bool //已请求的URL

	//wg sync.WaitGroup
}

func NewScheduler() Scheduler {
	return &myScheduler{
		reqCache: newRequestCache(),
		urlMap:   make(map[string]bool),
	}
}

var logger = log.New(os.Stdout, "scheduler:", log.LstdFlags)

func (this *myScheduler) Start(channelArgs base.ChannelArgs,
	poolBaseArgs base.PoolBaseArgs,
	crawlDepth uint32,
	httpClientGenerator GenHttpClient,
	respParsers []analyzer.ParseResponse,
	itemProcessors []itempipeline.ProcessItem,
	firstHttpReq *http.Request) (err error) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal  Scheduler Error:%s\n", p)
			logger.Println(errMsg)
			err = errors.New(errMsg)
		}
	}()
	if atomic.LoadUint32(&this.running) == 1 {
		return errors.New("The Scheduler has bean started!\n")
	}
	atomic.StoreUint32(&this.running, 1)

	if err := channelArgs.Check(); err != nil {
		return errors.New("The channel max length(capacity) can not be 0!\n")
	}
	this.channelArgs = channelArgs
	if err := poolBaseArgs.Check(); err != nil {
		return errors.New("The pool size can not be 0!\n")
	}
	this.poolBaseArgs = poolBaseArgs
	this.crawlDepth = crawlDepth

	this.chanman = generateChannelManager(this.channelArgs)

	if httpClientGenerator == nil {
		return errors.New("The http client generator list is ivalid!")
	}
	dlpool, err := generatePageDownloadPool(this.poolBaseArgs.PageDownloaderPoolSize(), httpClientGenerator)
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get page downloader pool:%s\n", err)
		return errors.New(errMsg)
	}
	this.dlpool = dlpool
	analyzerPool, err := generateAnalyzerPool(this.poolBaseArgs.AnalyzerPoolSize())
	if err != nil {
		errMsg := fmt.Sprintf("Occur error when get analyzer pool:%s\n", err)
		return errors.New(errMsg)
	}
	this.analyzerPool = analyzerPool

	if itemProcessors == nil {
		return errors.New("The item processor list is invalid!")
	}
	for i, ip := range itemProcessors {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
		}
	}

	this.itempipeline = generateItemPipelLine(itemProcessors)

	if this.stopSign == nil {
		this.stopSign = middleware.NewStopSign()
	} else {
		this.stopSign.Reset()
	}

	this.urlMap = make(map[string]bool)
	//this.wg.Add(4)
	//defer this.wg.Wait()
	this.startDownlaoding()
	this.activateAnalyzers(respParsers)
	this.openItemPipeLine()
	this.schedule(10 * time.Millisecond)

	if firstHttpReq == nil {
		return errors.New("The frist http request is invalid!")
	}
	pd, err := getPrimaryDomain(firstHttpReq.Host)
	if err != nil {
		return err
	}
	this.primaryDomain = pd

	firstReq := base.NewRequest(firstHttpReq, 0)
	this.reqCache.put(firstReq)
	return nil
}

func (this *myScheduler) schedule(interval time.Duration) {
	go func() {
		//defer this.wg.Done()
		for {
			remainder := cap(this.getReqChan()) - len(this.getReqChan())
			var temp *base.Request
			for remainder > 0 {
				temp = this.reqCache.get()
				if temp == nil {
					return
				}
				this.getReqChan() <- *temp
				remainder--
			}
			time.Sleep(interval)
			if this.stopSign.Signed() {
				this.stopSign.Deal(SCHEDULER_CODE)
				return
			}
		}
	}()
}

func (this *myScheduler) openItemPipeLine() {
	go func() {
		//defer this.wg.Done()
		this.itempipeline.SetFailFast(true)
		code := ITEMPIPELINE_CODE
		for item := range this.getITemChan() {
			go func(item base.Item) {
				defer func() {
					if p := recover(); p != nil {
						errMsg := fmt.Sprintf("Fatal Item Processing Error:%s\n", p)
						logger.Println(errMsg)
					}

				}()
				errs := this.itempipeline.Send(item)
				if errs != nil {
					for _, err := range errs {
						this.SendError(err, code)
					}
				}
			}(item)
		}
	}()
}
func (this *myScheduler) Stop() bool {
	if atomic.LoadUint32(&this.running) != 1 {
		return false
	}
	this.stopSign.Sign()
	this.chanman.Close()
	this.reqCache.close()

	atomic.StoreUint32(&this.running, 2)
	return true
}

func (this *myScheduler) Running() bool {
	return atomic.LoadUint32(&this.running) == 1
}

func (this *myScheduler) ErrorChan() <-chan error {
	if this.chanman.Status() != middleware.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return this.getErrorChan()
}

func (this *myScheduler) Idle() bool {
	idleDlPool := this.dlpool.Used() == 0
	idleAnalyzerPool := this.analyzerPool.Used() == 0
	idleItemPipeline := this.itempipeline.ProcessingNumber() == 0
	return idleAnalyzerPool && idleDlPool && idleItemPipeline
}

func (this *myScheduler) Summary(prefix string) SchedSummary {
	return NewSchedSummary(this, prefix)
}

const (
	DOWNLOADER_CODE   = "DOWNLOADER"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "itempipeline"
	SCHEDULER_CODE    = "SCHEDULER"
)

func (this *myScheduler) startDownlaoding() {
	go func() {
		//defer this.wg.Done()
		for {
			req, ok := <-this.getReqChan()
			if !ok {
				break
			}
			go this.download(req)

		}

	}()
}

func (this *myScheduler) getReqChan() chan base.Request {
	reqchan, err := this.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqchan
}

func (this *myScheduler) getRespChan() chan base.Response {
	respchan, err := this.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respchan
}

func (this *myScheduler) getErrorChan() chan error {
	errchan, err := this.chanman.ErrorChan()
	if err != nil {
		panic(err)
	}
	return errchan
}
func (this *myScheduler) getITemChan() chan base.Item {
	itemchan, err := this.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemchan
}

func (this *myScheduler) download(req base.Request) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Download Error :%s\n", p)
			//logger.Println(errMsg)
			logger.Println(errMsg)
		}
	}()

	downloader, err := this.dlpool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Downloader pool error:%s\n", err)
		this.SendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := this.dlpool.Return(downloader)
		if err != nil {
			errMsg := fmt.Sprintf("Downloader pool error:%s\n", err)
			this.SendError(errors.New(errMsg), SCHEDULER_CODE)
			return
		}

	}()

	code := generateCode(DOWNLOADER_CODE, downloader.Id())
	resp, err := downloader.Download(req)
	if resp != nil {
		this.SendResp(*resp, code)
	}
	if err != nil {
		this.SendError(err, code)
	}
}

func (this *myScheduler) activateAnalyzers(respParsers []analyzer.ParseResponse) {
	go func() {
		//defer this.wg.Done()
		for {
			resp, ok := <-this.getRespChan()
			if !ok {
				break
			}
			go this.analyze(respParsers, resp)

		}

	}()
}

func (this *myScheduler) analyze(respParsers []analyzer.ParseResponse, resp base.Response) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Analysis Error :%s\n", p)
			//logger.Println(errMsg)
			logger.Println(errMsg)
		}
	}()

	analyzer, err := this.analyzerPool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Analyzer pool error:%s\n", err)
		this.SendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := this.analyzerPool.Return(analyzer)
		if err != nil {
			errMsg := fmt.Sprintf("Analyzer pool error:%s\n", err)
			this.SendError(errors.New(errMsg), SCHEDULER_CODE)
			return
		}

	}()

	code := generateCode(ANALYZER_CODE, analyzer.Id())
	datalist, errs := analyzer.Analyze(respParsers, resp)
	if datalist != nil {
		for _, data := range datalist {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *base.Request:
				this.savaReqToCache(*d, code)
			case *base.Item:
				this.sendItem(*d, code)
			default:
				errMsg := fmt.Sprintf("Unsupported data type '%T'! (value=%v)\n", d, d)
				this.SendError(errors.New(errMsg), code)
			}

		}
	}
	if errs != nil {
		for _, err := range errs {
			if err == nil {
				continue
			}
			this.SendError(err, code)
		}
	}
}

func (this *myScheduler) savaReqToCache(req base.Request, code string) bool {
	httpReq := req.HttpReq()
	if httpReq == nil {
		logger.Println("Ignore the requst ! It is HTTP request is invalid!")
		return false
	}
	reqUrl := httpReq.URL
	if reqUrl == nil {
		logger.Println("Ignore the requst ! It is url is invalid!")

		return false
	}
	if strings.ToLower(reqUrl.Scheme) != "htttp" {
		logger.Printf("Ignore the requst ! It is url scheme '%s' ,but should be 'http' !\n", reqUrl.Scheme)
		return false
	}

	if _, ok := this.urlMap[reqUrl.String()]; ok {
		logger.Printf("Ignore the requst ! It is url is repeated. (requestUrl='%s')\n", reqUrl)
		return false
	}

	if pd, _ := getPrimaryDomain(httpReq.Host); pd != this.primaryDomain {
		logger.Printf("Ignore the requst ! It is host '%s' is not in primary domain '%s' (requestUrl='%s')\n", httpReq.Host, this.primaryDomain, reqUrl)
		return false
	}

	if this.stopSign.Signed() {
		this.stopSign.Deal(code)
		return false
	}
	this.reqCache.put(&req)

	this.urlMap[reqUrl.String()] = true
	return true
}

func (this *myScheduler) SendResp(resp base.Response, code string) bool {
	if this.stopSign.Signed() {
		this.stopSign.Deal(code)
		return false
	}
	this.getRespChan() <- resp
	return true
}

func (this *myScheduler) sendItem(item base.Item, code string) bool {
	if this.stopSign.Signed() {
		this.stopSign.Deal(code)
		return false
	}
	this.getITemChan() <- item
	return true
}

func (this *myScheduler) SendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType base.ErrorType

	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = base.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = base.ANALYZER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = base.ITEM_PROCESSOR_ERROR
	}
	cError := base.NewCrawlerError(errorType, err.Error())

	if this.stopSign.Signed() {
		this.stopSign.Deal(code)
		return false
	}
	go func() {
		this.getErrorChan() <- cError
	}()
	return true
}
