package src

import (
	"net"
	"time"
)

//NewTCPListener new_listen
func NewTCPListener(laddr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", laddr)
	if err != nil {
		return nil, err
	}
	tl := ln.(*net.TCPListener)
	return tl, nil
}

//dial dial
func dial(raddr string) (net.Conn, error) {
	//5s超时时间
	conn, err := net.DialTimeout("tcp", raddr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	tcpConn := conn.(*net.TCPConn)
	//保持长连接
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(TunnelKeepAlivePeriod)
	return tcpConn, nil
}
