package base

import (
	"bytes"
	"fmt"
	"net/http"
)

//请求
type Request struct {
	httpReq *http.Request //http 请求
	depth   uint32        //请求深度
}

//创建新请求
func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

//获取http请求
func (this *Request) HttpReq() *http.Request {
	return this.httpReq
}

//获取深度
func (this *Request) Depth() uint32 {
	return this.depth
}
func (this *Request) Valid() bool {
	return this.httpReq != nil && this.httpReq.URL != nil
}


//响应
type Response struct {
	httpResp *http.Response
	depth    uint32
}

//创建新响应
func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

//获取http响应
func (this *Response) HttpReq() *http.Response {
	return this.httpResp
}

//获取深度
func (this *Response) Depth() uint32 {
	return this.depth
}
func (this *Response) Valid() bool {
	return this.httpResp != nil && this.httpResp.Body != nil
}


//条目
type Item map[string]interface{}

func (this Item) Valid() bool {
	return this != nil
}

//数据接口
type Data interface {
	Valid() bool // 数据是否有效
}
type ErrorType string

type CrawlerError interface {
	Type() ErrorType
	Error() string
}

type myCrawlerError struct {
	errType    ErrorType //错误类型
	errMsg     string    //错误信息
	fullErrMsg string    //完整错误信息
}

func (this *myCrawlerError) Type() ErrorType {
	return this.errType
}
func (this *myCrawlerError) Error() string {
	if this.fullErrMsg == "" {
		this.genFullErrMsg()
	}
	return this.fullErrMsg
}

func (this *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error :")
	if this.errType != "" {
		buffer.WriteString(string(this.errType))
		buffer.WriteString(":")
	}
	buffer.WriteString(this.errMsg)
	this.fullErrMsg = fmt.Sprintf("%s\n", buffer.String())
}

const (
	DOWNLOADER_ERROR     ErrorType = "Downloader Error"
	ANALYZER_ERROR       ErrorType = "Analyzer Error"
	ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

//创建错误
func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{errMsg: errMsg, errType: errType}
}
