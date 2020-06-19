package src

import (
	"bufio"
	"crypto/rc4"
	"encoding/binary"
	"io"
	"net"
	"sync"
)

//Tunnel tunnel struct
type Tunnel struct {
	net.Conn
	reader *bufio.Reader
	write  *bufio.Writer
	enc    *rc4.Cipher
	dec    *rc4.Cipher

	wlock sync.Mutex // protect concurrent write
	werr  error      // write error
}

//SetCipherKey set cipher key
func (tun *Tunnel) SetCipherKey(key []byte) {
	tun.enc, _ = rc4.NewCipher(key)
	tun.dec, _ = rc4.NewCipher(key)
}

//Flush flush write buf
func (tun *Tunnel) Flush() error {
	return tun.write.Flush()
}

// WritePacket write data
func (tun *Tunnel) WritePacket(linkid uint16, data []byte) error {
	defer MsgPool.Put(data)

	tun.wlock.Lock()
	defer tun.wlock.Unlock()

	if tun.werr != nil {
		return tun.werr
	}

	header := THeader{
		LinkID:  linkid,
		BodyLen: uint16(len(data)),
	}
	err := binary.Write(tun, binary.LittleEndian, header)
	if err != nil {
		tun.werr = err
		tun.Close()
		return err
	}

	_, err = tun.Write(data)
	if err != nil {
		tun.werr = err
		tun.Close()
		return err
	}

	err = tun.Flush()
	if err != nil {
		tun.werr = err
		tun.Close()
		return err
	}
	return nil
}

//ReadPacket read packet  can't read concurrently
func (tun *Tunnel) ReadPacket() (linkid uint16, data []byte, err error) {
	var h THeader
	binary.Read(tun, binary.LittleEndian, &h)
	if err != nil {
		Error("ReadPacket Read failed, err: %s", err)
		return
	}

	if h.BodyLen > TunnelPacketSize {
		Error("tunnel.Read: packet too large")
		return
	}

	data = MsgPool.Get()[:h.BodyLen]
	_, err = io.ReadFull(tun, data)
	if err != nil {
		Error("ReadPacket io.ReadFull() falied, err: %s", err)
		return
	}
	linkid = h.LinkID
	return
}

func newTunnel(conn net.Conn) *Tunnel {
	var mutex sync.Mutex
	tunnel := Tunnel{
		conn,
		bufio.NewReaderSize(conn, TunnelPacketSize*2),
		bufio.NewWriterSize(conn, TunnelPacketSize*2),
		nil,
		nil,
		mutex,
		nil,
	}
	return &tunnel
}
