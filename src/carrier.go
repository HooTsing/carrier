package src

import (
	"crypto/aes"
	"crypto/md5"
	"time"
)

const (
	LINK_DATA uint8 = iota
	LINK_CREATE
	LINK_CLOSE_ALL
	LINK_CLOSE_RECV
	LINK_CLOSE_SEND
	HEARTBEAT
)

const (
	//TaaTokenSize auth token size
	TaaTokenSize int = aes.BlockSize
	//TaaSignatureSize sign size
	TaaSignatureSize int = md5.Size
	//TaaBlockSize auth block size
	TaaBlockSize int = TaaTokenSize + TaaSignatureSize
)

const (
	//TunnelMaxID IDAllocator max id
	TunnelMaxID = ^uint16(0)
	//TunnelPacketSize tunnel packet size
	TunnelPacketSize = 8192
	//TunnelKeepAlivePeriod keep alive time 180s
	TunnelKeepAlivePeriod = time.Second * 180
	//TunnelCount tunnel count
	TunnelCount = 10
	//StackBufferMaxCap 2^15
	StackBufferMaxCap = 32768

	//defaultHeartbeat 默认心跳间隔
	defaultHeartbeat = 1
	// 3次心跳无回应则断开
	tunnelMinSpan = 3
)

var (
	//Heartbeat interval for tunnel hearbeat, second.
	Heartbeat int = 1
	//Timeout interval for tunnel write/read, second.
	Timeout int = 0

	//LogLevel log level
	LogLevel uint = 1
	//MsgPool msg pool
	MsgPool = NewMsgPool(TunnelPacketSize)
)

func getHeartbeat() time.Duration {
	if Heartbeat <= 0 {
		Heartbeat = defaultHeartbeat
	}
	return time.Duration(Heartbeat) * time.Second
}

func getTimeout() time.Duration {
	return time.Duration(Timeout) * time.Second
}
