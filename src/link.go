package src

import (
	"errors"
	"net"
	"sync"
	"time"
)

var errPeerClosed = errors.New("errPeerClosed")

//TLink link
type TLink struct {
	id   uint16
	conn *net.TCPConn
	wbuf *TBuffer // write buffer

	lock sync.Mutex // protects below fields
	rerr error      // if read close, error to give reads
}

func (link *TLink) allClose() {
	link.readClose()
	link.writeClose()
}

func (link *TLink) readClose() bool {
	return link.wbuf.Close()
}

func (link *TLink) writeClose() bool {
	return link.setRerr(errPeerClosed)
}

func (link *TLink) setRerr(err error) bool {
	defer link.lock.Unlock()
	link.lock.Lock()

	if link.rerr != nil {
		return false
	}

	link.rerr = err
	return true
}

func (link *TLink) read() ([]byte, error) {
	if link.rerr != nil {
		Error("link read err, err: %s", link.rerr.Error())
		return nil, link.rerr
	}
	buf := MsgPool.Get()
	n, err := link.conn.Read(buf)
	if err != nil {
		link.setRerr(err)
		return nil, link.rerr
	}
	if link.rerr != nil {
		return nil, link.rerr
	}
	return buf[:n], nil
}

func (link *TLink) write(data []byte) bool {
	return link.wbuf.Put(data)
}

func (link *TLink) setConn(conn *net.TCPConn) {
	if link.conn != nil {
		Panic("link[%d] repeated set conn", link.id)
	}
	link.conn = conn
}

func (hub *Hub) createLink(id uint16) *TLink {
	Info("link[%d] new link start, tunnel: %s", id, hub.tunnel)

	_, ok := hub.links[id]
	if ok {
		Info("link[%d] repeated over, tunnel: %s", id, hub.tunnel)
		return nil
	}

	link := &TLink{
		id:   id,
		wbuf: newBuffer(16),
	}
	hub.links[id] = link
	return link
}

func (hub *Hub) getLink(id uint16) *TLink {
	defer hub.linkMutex.RUnlock()
	hub.linkMutex.RLock()

	return hub.links[id]
}

func (hub *Hub) deleteLink(id uint16) {
	hub.linkMutex.Lock()
	defer hub.linkMutex.Unlock()

	//map delete
	delete(hub.links, id)
	Info("link[%d] deleted.", id)
}

func (hub *Hub) startLink(link *TLink, conn *net.TCPConn) {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(60 * time.Second)
	link.setConn(conn)

	Info("link[%d] start link to remote[%v]", link.id, conn.RemoteAddr())

	//参考https://studygolang.com/static/pkgdoc/ sync.WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer link.conn.CloseRead()

		for {
			data, err := link.read()
			if err != nil {
				if err != errPeerClosed {
					hub.sendCmd(link.id, LINK_CLOSE_SEND)
				}
				break
			}

			hub.Send(link.id, data)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer link.conn.CloseWrite()

		for {
			data, ok := link.wbuf.Pop()
			if !ok {
				Error("link[%d] write_buf pop failed, write_buf closed.", link.id)
				break
			}
			_, err := link.conn.Write(data)
			MsgPool.Put(data)
			if err != nil {
				Error("link[%d] write to remote failed, err: %s", link.id, err.Error())
				hub.sendCmd(link.id, LINK_CLOSE_RECV)
			}
		}
	}()
	wg.Wait()
	Info("link[%d] closed...", link.id)
}

func (hub *Hub) resetAllLink() {
	defer hub.linkMutex.RUnlock()
	hub.linkMutex.RLock()

	for _, link := range hub.links {
		link.allClose()
	}

	Error("reset all links, links num: %d", len(hub.links))
}
