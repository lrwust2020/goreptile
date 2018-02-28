package middleware

import "github.com/fmyxyz/goreptile/base"

//通道管器接口
type ChannelManager interface {
	// 初始化
	Init(channelLen uint, reset bool) bool
	//关闭
	Close() bool
	ReqChan() (chan base.Request, error)
	RespChan() (chan base.Response, error)
	ItemChan() (chan base.Item, error)
	ErrorChan() (chan error, error)
	ChannelLen() uint
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
