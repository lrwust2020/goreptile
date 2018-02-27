package analyzer

import (
	"github.com/fmyxyz/goreptile/base"
	"net/http"
)

type Analyzer interface {
	Id() uint32
	Analyze(respParsers []ParseResponse, resp base.Response) ([]base.Data, error) //根据规则分析响应返回请求和条目
}

//解析响应的函数类型
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)


//分析器池
type AnalyzerPool interface {
	Tack() (Analyzer ,error)
	Return(pl Analyzer) error
	Total()uint32
	Used() uint32
}
