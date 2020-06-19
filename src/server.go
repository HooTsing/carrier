package src

import "net"

//TServer server struct
type TServer struct {
	listener net.Listener
	addr     *net.TCPAddr
	secret   string
}

func (svr *TServer) handleConn(conn net.Conn) {
	defer conn.Close()

	tunnel := newTunnel(conn)
	//authenticate connnection
	auth := newTaa(svr.secret)
	auth.GenToken()

	//tunnel.WritePacket()
	challenge := auth.GenCipherBlock(nil)
	err := tunnel.WritePacket(0, challenge)
	if err != nil {
		Error("tunnel[%v] write challenge failed, err is: %s", tunnel, err)
		return
	}

	_, token, err := tunnel.ReadPacket()
	if err != nil {
		Error("Tunnel[%v] read packet failed, err: %s", tunnel, err)
		return
	}

	//check token VerifyCipherBlock
	if !auth.VerifyCipherBlock(token) {
		Error("tunnel[%v] verify token failed, token is: %s", tunnel, token)
		return
	}

	tunnel.SetCipherKey(auth.getRc4key())
	shub := newServerHub(tunnel, svr.addr)
	shub.Start()
}

//Start impl Start() interface
func (svr *TServer) Start() error {
	defer svr.listener.Close()
	for {
		Info("server accept connection, waiting...")
		conn, err := svr.listener.Accept()
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Temporary() {
				Log("server accept failed, err is temporary, err: %s", netErr.Error())
				continue
			} else {
				Error("server accept failed, err: %s", netErr.Error())
				return err
			}
		}
		Log("new connection from %v", conn.RemoteAddr())
		go svr.handleConn(conn)
	}
}

//Status impl Status() interface
func (svr *TServer) Status() {
}

//NewServer create a carrier server
func NewServer(laddr, addr, secret string) (*TServer, error) {
	ln, err := NewTCPListener(laddr)
	if err != nil {
		return nil, err
	}

	//ResolveTCPAddr将addr作为TCP地址解析并返回
	baddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &TServer{
		listener: ln,
		addr:     baddr,
		secret:   secret,
	}
	return s, nil
}
