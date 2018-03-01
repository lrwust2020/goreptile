package middleware

//停止信号
type StopSign interface {
	//发出停止信号
	//若已发出过停止信号，则返回false
	Sign() bool
	//判断停止信号是否发出
	Signed() bool

	Reset()
	//处理停止信号
	//参数code代表停止信号处理方的代号，该代号会出现在处理记录中
	Deal(code string)
	//停止信号处理方的处理计数
	DealCount() uint32
	//被处理总计数
	DealTotal() uint32
	Summary() string
}
