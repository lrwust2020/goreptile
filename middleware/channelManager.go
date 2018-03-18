package middleware

import (
	"errors"
	"fmt"
	"github.com/fmyxyz/goreptile/base"
	"sync"
)

//通道管器接口
type ChannelManager interface {
	// 初始化
	Init(channelArgs base.ChannelArgs, reset bool) bool
	//关闭
	Close() bool
	ReqChan() (chan base.Request, error)
	RespChan() (chan base.Response, error)
	ItemChan() (chan base.Item, error)
	ErrorChan() (chan error, error)
	//管道管理器状态
	Status() ChannelManagerStatus
	Summary() string
}

type ChannelManagerStatus uint8

const (
	CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManagerStatus = iota
	CHANNEL_MANAGER_STATUS_INITIALIZED
	CHANNEL_MANAGER_STATUS_CLOSED
)

type myChannelManager struct {
	reqCh   chan base.Request
	respCh  chan base.Response
	itemCh  chan base.Item
	errorCh chan error

	channelArgs base.ChannelArgs

	status ChannelManagerStatus

	rwMutex sync.RWMutex
}

var defaultChanLen uint = 64

//创建通道管理器
func NewChannelManager(channelArgs base.ChannelArgs) ChannelManager {
	var chanman = &myChannelManager{}
	chanman.Init(channelArgs, true)
	return chanman
}

func (this *myChannelManager) Init(channelArgs base.ChannelArgs, reset bool) bool {
	if err := channelArgs.Check(); err != nil {
		panic(errors.New("The channel length is invalid!"))
	}

	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()

	if this.status == CHANNEL_MANAGER_STATUS_INITIALIZED && !reset {
		return false
	}
	this.channelArgs = channelArgs
	this.reqCh = make(chan base.Request, channelArgs.ReqChanLen())
	this.respCh = make(chan base.Response, channelArgs.RespChanLen())
	this.itemCh = make(chan base.Item, channelArgs.ItemChanLen())
	this.errorCh = make(chan error, channelArgs.ErrorChanLen())
	this.status = CHANNEL_MANAGER_STATUS_INITIALIZED
	return true
}

func (this *myChannelManager) Close() bool {
	this.rwMutex.Lock()
	defer this.rwMutex.Unlock()
	if this.status != CHANNEL_MANAGER_STATUS_INITIALIZED {
		return false
	}
	close(this.reqCh)
	close(this.respCh)
	close(this.itemCh)
	close(this.errorCh)
	this.status = CHANNEL_MANAGER_STATUS_CLOSED
	return true
}

var statusNameMap = map[ChannelManagerStatus]string{
	CHANNEL_MANAGER_STATUS_UNINITIALIZED: "uninitialized",
	CHANNEL_MANAGER_STATUS_INITIALIZED:   "initialized",
	CHANNEL_MANAGER_STATUS_CLOSED:        "closed",
}

func (this *myChannelManager) checkStatus() error {
	if this.status == CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	statusname, ok := statusNameMap[this.status]
	if !ok {
		statusname = fmt.Sprintf("%d", this.status)
	}
	errMsg := fmt.Sprintf("The undesrirable status of channel manager : %s!\n", statusname)
	return errors.New(errMsg)
}

func (this *myChannelManager) ReqChan() (chan base.Request, error) {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	if err := this.checkStatus(); err != nil {
		return nil, err
	}
	return this.reqCh, nil
}

func (this *myChannelManager) RespChan() (chan base.Response, error) {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	if err := this.checkStatus(); err != nil {
		return nil, err
	}
	return this.respCh, nil
}

func (this *myChannelManager) ItemChan() (chan base.Item, error) {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	if err := this.checkStatus(); err != nil {
		return nil, err
	}
	return this.itemCh, nil
}

func (this *myChannelManager) ErrorChan() (chan error, error) {
	this.rwMutex.RLock()
	defer this.rwMutex.RUnlock()
	if err := this.checkStatus(); err != nil {
		return nil, err
	}
	return this.errorCh, nil
}

func (this *myChannelManager) Status() ChannelManagerStatus {
	return this.status
}

var chanmanSummaryTemplate = "status: %s," +
	"requestChannel: %d/%d," +
	"responseChannel: %d/%d," +
	"itemChannel: %d/%d," +
	"errorChannel: %d/%d"

func (this *myChannelManager) Summary() string {
	summaty := fmt.Sprintf(chanmanSummaryTemplate, statusNameMap[this.status],
		len(this.reqCh), cap(this.reqCh),
		len(this.respCh), cap(this.respCh),
		len(this.itemCh), cap(this.itemCh),
		len(this.errorCh), cap(this.errorCh),
	)
	return summaty
}
