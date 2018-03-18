package base

import (
	"errors"
	"fmt"
)

type Args interface {
	Check() error
	String() string
}

type ChannelArgs struct {
	reqChanLen   uint
	respChanLen  uint
	itemChanLen  uint
	errorChanLen uint
	description  string
}

func NewChannelArgs(reqChanLen, respChanLen, itemChanLen, errorChanLen uint) ChannelArgs {
	return ChannelArgs{
		reqChanLen:   reqChanLen,
		respChanLen:  respChanLen,
		itemChanLen:  itemChanLen,
		errorChanLen: errorChanLen,
	}
}
func (this *ChannelArgs) Check() error {
	if (this.reqChanLen == 0 || this.respChanLen == 0 || this.errorChanLen == 0 || this.itemChanLen == 0){
		return errors.New("ChannelArgs Check error!")
	}else{
		return nil
	}
}

func (this *ChannelArgs) String() string {
	return fmt.Sprintf(`reqChanLen:   %d,
		respChanLen:   %d,
		itemChanLen:   %d,
		errorChanLen:  %d
`, this.reqChanLen, this.respChanLen, this.itemChanLen, this.errorChanLen)
}

func (this ChannelArgs) ReqChanLen() uint {
	return this.reqChanLen
}
func (this ChannelArgs) RespChanLen() uint {
	return this.respChanLen
}
func (this ChannelArgs) ItemChanLen() uint {
	return this.itemChanLen
}
func (this ChannelArgs) ErrorChanLen() uint {
	return this.errorChanLen
}

type PoolBaseArgs struct {
	pageDownloaderPoolSize uint32
	analyzerPoolSize       uint32
	description            string
}

func NewPoolBaseArgs(pageDownloaderPoolSize, analyzerPoolSize uint32) PoolBaseArgs {
	return PoolBaseArgs{
		pageDownloaderPoolSize: pageDownloaderPoolSize,
		analyzerPoolSize:       analyzerPoolSize,
	}
}
func (this *PoolBaseArgs) Check() error {
	if this.pageDownloaderPoolSize == 0 || this.analyzerPoolSize == 0 {
		return errors.New("PoolBaseArgs Check error!")
	}else{
		return nil
	}
}

func (this *PoolBaseArgs) String() string {
	return fmt.Sprintf(`pageDownloaderPoolSize:   %d,
		analyzerPoolSize:   %d
`, this.pageDownloaderPoolSize, this.analyzerPoolSize)
}
func (this *PoolBaseArgs) PageDownloaderPoolSize() uint32 {
	return this.pageDownloaderPoolSize
}

func (this *PoolBaseArgs) AnalyzerPoolSize() uint32 {
	return this.analyzerPoolSize
}
