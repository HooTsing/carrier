package src

import "net"

//TServerHub server hub struct
type TServerHub struct {
	*Hub
	addr *net.TCPAddr
}

func (shub *TServerHub) handleLink(link *TLink) {
	defer Recover()
	defer shub.deleteLink(link.id)

	// 监听地址自动分配
	conn, err := net.DialTCP("tcp", nil, shub.addr)
	if err != nil {
		Error("link[%d] connect to backend[%v] failed, err: %v", link.id, shub.addr, err)
		shub.sendCmd(link.id, LINK_CLOSE_ALL)
		shub.deleteLink(link.id)
		return
	}

	shub.startLink(link, conn)
}

func (shub *TServerHub) onCtrl(cmd TCmd) bool {
	id := cmd.ID
	switch cmd.Cmd {
	case LINK_CREATE:
		link := shub.createLink(id)
		if link != nil {
			go shub.handleLink(link)
		} else {
			shub.sendCmd(id, LINK_CLOSE_ALL)
		}
		return true
	case HEARTBEAT:
		shub.sendCmd(id, HEARTBEAT)
		return true
	}
	return false
}

func newServerHub(tunnel *Tunnel, addr *net.TCPAddr) *TServerHub {
	shub := &TServerHub{
		Hub:  newHub(tunnel),
		addr: addr,
	}
	shub.Hub.onCtrlFilter = shub.onCtrl
	return shub
}
