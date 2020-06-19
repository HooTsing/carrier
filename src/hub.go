package src

import (
	"bytes"
	"encoding/binary"
	"sync"
)

//Hub hub struct
type Hub struct {
	tunnel *Tunnel

	linkMutex sync.RWMutex //protect links
	links     map[uint16]*TLink

	onCtrlFilter func(cmd TCmd) bool
}

//Send send
func (hub *Hub) Send(id uint16, data []byte) bool {
	err := hub.tunnel.WritePacket(id, data)
	if err != nil {
		Error("link[%d] write to %v failed, err: %s", id, hub.tunnel, err.Error())
		return false
	}
	return true
}

func (hub *Hub) onCtrl(cmd TCmd) {
	if cmd.Cmd == HEARTBEAT {
		Debug("tunnel[%v] recv heartbeat: %d", hub.tunnel, cmd.ID)
	} else {
		Info("link[%d] recv cmd: %d", cmd.ID, cmd.Cmd)
	}

	//处理回调
	if hub.onCtrlFilter != nil && hub.onCtrlFilter(cmd) {
		return
	}

	id := cmd.ID
	link := hub.getLink(id)
	if link != nil {
		Error("link[%d] recv cmd: %d, no link.", id, cmd.Cmd)
		return
	}

	switch cmd.Cmd {
	case LINK_CLOSE_ALL:
		link.allClose()
	case LINK_CLOSE_RECV:
		link.readClose()
	case LINK_CLOSE_SEND:
		link.writeClose()
	default:
		Error("link[%d] RECV unknown cmd: %v", id, cmd)
	}
}

func (hub *Hub) onData(linkid uint16, data []byte) {
	Info("link[%d] recv data bytes is: %d", linkid, len(data))

	link := hub.getLink(linkid)
	if link == nil {
		Error("link[%d] no link...", linkid)
		MsgPool.Put(data)
		return
	}

	if !link.write(data) {
		Error("link[%d] put data failed, write_buff closed...", linkid)
		MsgPool.Put(data)
		return
	}
	return
}

//Start hub start
func (hub *Hub) Start() {
	defer hub.tunnel.Close()

	for {
		id, data, err := hub.tunnel.ReadPacket()
		if err != nil {
			Error("%s read failed, err: %s", hub.tunnel, err)
			break
		}

		if id == 0 {
			var cmd TCmd
			buf := bytes.NewBuffer(data)
			err := binary.Read(buf, binary.LittleEndian, &cmd)
			MsgPool.Put(data)
			if err != nil {
				Error("parse message failed, break dispatch, err: %s", err.Error())
				break
			}
			hub.onCtrl(cmd)
		} else {
			hub.onData(id, data)
		}
	}

	//tunnel disconnect, so reset all link
	hub.resetAllLink()
	Log("hub[%v] quit...", hub.tunnel)
}

//Close hub close
func (hub *Hub) Close() {
	hub.tunnel.Close()
}

//Status hub status
func (hub *Hub) Status() {
	Log("[status] %s", hub.tunnel)
}

func newHub(tunnel *Tunnel) *Hub {
	hub := &Hub{
		tunnel: tunnel,
		links:  make(map[uint16]*TLink),
	}
	return hub
}
