package analyzer

import (
	"errors"
	"fmt"
	"github.com/fmyxyz/goreptile/base"
	"github.com/fmyxyz/goreptile/middleware"
	"log"
	"net/http"
	"os"
	"reflect"
)

type Analyzer interface {
	Id() uint32
	Analyze(respParsers []ParseResponse, resp base.Response) ([]base.Data, []error) //根据规则分析响应返回请求和条目
}

//解析响应的函数类型
// data 请求或者条目
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)

//分析器池
type AnalyzerPool interface {
	Take() (Analyzer, error)
	Return(pl Analyzer) error
	Total() uint32
	Used() uint32
}

type myAnalyzer struct {
	id uint32
}

func (this *myAnalyzer) Id() uint32 {
	return this.id
}

var logger = log.New(os.Stdout, "analyzer", log.LstdFlags)

func (this *myAnalyzer) Analyze(respParsers []ParseResponse, resp base.Response) ([]base.Data, []error) {
	if respParsers == nil {
		err := errors.New("The response parser list is invaild!")
		return nil, []error{err}
	}
	var httpResp = resp.HttpReq()
	if httpResp == nil {
		err := errors.New("The http response is invaild!")
		return nil, []error{err}
	}
	var reqUrl = httpResp.Request.URL
	logger.Printf("Parser the response (reqUrl=%s)...\n", reqUrl)
	var dataList = make([]base.Data, 0)
	var errorList = make([]error, 0)
	var respDepth = resp.Depth()
	for i, respParser := range respParsers {
		if respParser == nil {
			err := errors.New(fmt.Sprintf("The document parser [%d] is invaild!", i))
			errorList = append(errorList, err)
			continue
		}
		pDataList, pErrorList := respParser(httpResp, respDepth)
		if pDataList != nil {
			for _, pData := range pDataList {
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}
		if pErrorList != nil {
			for _, pErr := range pErrorList {
				errorList = appendErrorList(errorList, pErr)
			}
		}

	}
	return dataList, errorList
}

func appendDataList(dataList []base.Data, data base.Data, respDepth uint32) []base.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*base.Request)
	if !ok {
		return append(dataList, data)
	}
	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = base.NewRequest(req.HttpReq(), newDepth)
	}
	return append(dataList, req)
}

func appendErrorList(errorList []error, err error) []error {
	if err == nil {
		return errorList
	}
	return append(errorList, err)
}

func NewAnalyzer() Analyzer {
	return &myAnalyzer{id: genAnalyzerId()}
}

var analyzerGenIdGenertor = middleware.NewCyclicIdGenertor()

func genAnalyzerId() uint32 {
	return analyzerGenIdGenertor.GetUint32()
}

type myAnalyzerPool struct {
	pool  middleware.Pool
	etype reflect.Type
}

func (this *myAnalyzerPool) Take() (Analyzer, error) {
	entity, err := this.pool.Take()
	if err != nil {
		return nil, err
	}
	dl, ok := entity.(Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", this.etype)
		panic(errors.New(errMsg))
	}
	return dl, nil

}

func (this *myAnalyzerPool) Return(pl Analyzer) error {
	return this.pool.Return(pl)
}

func (this *myAnalyzerPool) Total() uint32 {
	return this.pool.Total()
}

func (this *myAnalyzerPool) Used() uint32 {
	return this.pool.Used()
}

func NewAnalyzerPool(total uint32, gen GenAnalyzerPool) (AnalyzerPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() middleware.Entity {
		return gen()
	}
	pool, err := middleware.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	thisPool := &myAnalyzerPool{
		pool:  pool,
		etype: etype,
	}
	return thisPool, nil
}

type GenAnalyzerPool func() Analyzer
