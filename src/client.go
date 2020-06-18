package src

import (
	"container/heap"
	"errors"
	"net"
	"sync"
	"time"
)

//TClient client struct
type TClient struct {
	laddr   string
	backend string
	secret  string
	tunnel  uint

	alloc *IDAllocator //id 分配器
	cq    ClientHubQueue
	lock  sync.Mutex
}

func (cli *TClient) newClientHubItem() (*ClientHubItem, error) {
	conn, err := dial(cli.backend)
	if err != nil {
		Error("newClientHubItem dail failed, err: %s", err)
		return nil, err
	}

	tunnel := newTunnel(conn)
	//ReadPacket
	_, challenge, err := tunnel.ReadPacket()
	if err != nil {
		Error("tunnel[%v] read challenge failed, err is: %s", tunnel, err)
		return nil, err
	}

	auth := newTaa(cli.secret)
	//交换秘钥
	token, ok := auth.ExchangeCipherBlock(challenge)
	if !ok {
		err = errors.New("exchange challenge failed")
		Error("tunnel[%v] write token failed, err is: %s", tunnel, err)
		return nil, err
	}
	//WritePacket
	err = tunnel.WritePacket(0, token)
	if err != nil {
		Error("tunnel[%v] write token failed, err is: %s", tunnel, err)
		return nil, err
	}

	tunnel.SetCipherKey(auth.getRc4key())

	chi := &ClientHubItem{
		TClientHub: newClientHub(tunnel),
	}
	return chi, nil
}

func (cli *TClient) addClientHubItem(item *ClientHubItem) {
	cli.lock.Lock()
	heap.Push(&cli.cq, item)
	cli.lock.Unlock()
}

func (cli *TClient) removeClientHubItem(item *ClientHubItem) {
	cli.lock.Lock()
	heap.Remove(&cli.cq, item.index)
	cli.lock.Unlock()
}

func (cli *TClient) fetchClientHubItem() *ClientHubItem {
	defer cli.lock.Unlock()
	cli.lock.Lock()

	if len(cli.cq) == 0 {
		Error("fetchClientHubItem failed, client_hub_queue is empty.")
		return nil
	}
	item := cli.cq[0]
	item.priority++
	//调整heap上位置0的item
	heap.Fix(&cli.cq, 0)
	return item
}

func (cli *TClient) dropClientHubItem(item *ClientHubItem) {
	cli.lock.Lock()
	item.priority--
	heap.Fix(&cli.cq, item.index)
	cli.lock.Unlock()
}

func (cli *TClient) handleConn(hub *ClientHubItem, conn *net.TCPConn) {
	defer Recover()
	defer cli.dropClientHubItem(hub)
	defer conn.Close()

	// allocate id
	id := cli.alloc.Acquire()
	defer cli.alloc.Release(id)

	//cli.alloc.Acquire()
	chub := hub.Hub
	link := hub.createLink(id)
	defer hub.deleteLink(id)

	chub.sendCmd(id, LINK_CREATE)
	chub.startLink(link, conn)
}

func (cli *TClient) listen() error {
	ln, err := net.Listen("tcp", cli.laddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	tcpListener := ln.(*net.TCPListener)
	for {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Temporary() {
				Log("accept failed temporary, err: %s", netErr.Error())
				continue
			} else {
				return err
			}
		}
		Info("new connection from: %v", conn.RemoteAddr())

		chub := cli.fetchClientHubItem()
		if chub == nil {
			Error("no active client hub")
			conn.Close()
			continue
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(60 * time.Second)
		go cli.handleConn(chub, conn)
	}
}

//StartOne start all
func (cli *TClient) StartOne(index int) error {
	for {
		hub, err := cli.newClientHubItem()
		if err != nil {
			Error("tunnel[%d] reconnect failed, err: %s", index, err)
			time.Sleep(3 * time.Second)
			continue
		}

		Error("tunnel[%d] connnect succeed", index)
		cli.addClientHubItem(hub)
		hub.Start()
		cli.removeClientHubItem(hub)
		Info("tunnel[%d] disconnected.", index)
	}
}

//Start impl Start() interface
func (cli *TClient) Start() error {
	size := cap(cli.cq)
	for i := 0; i < size; i++ {
		//并发处理
		go func(index int) {
			cli.StartOne(index)
		}(i)
	}
	return cli.listen()
}

//Status impl Status() interface
func (cli *TClient) Status() {
	for _, hub := range cli.cq {
		hub.Status()
	}
}

//NewClient create a carrier client
func NewClient(laddr, backend, secret string, tunnels uint) (*TClient, error) {
	cli := &TClient{
		laddr:   laddr,
		backend: backend,
		secret:  secret,
		tunnel:  tunnels,

		alloc: newAllocator(),
		cq:    make(ClientHubQueue, tunnels)[0:0],
	}
	return cli, nil
}
