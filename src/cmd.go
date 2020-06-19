package src

import (
	"bytes"
	"encoding/binary"
)

//TCmd control command struct
type TCmd struct {
	Cmd uint8  //control command
	ID  uint16 // id
}

//sendCmd send cmd
func (hub *Hub) sendCmd(id uint16, Cmd uint8) bool {
	buf := bytes.NewBuffer(MsgPool.Get()[0:0])
	cmd := TCmd{
		Cmd: Cmd,
		ID:  id,
	}
	binary.Write(buf, binary.LittleEndian, &cmd)

	if Cmd == HEARTBEAT {
		Debug("tunnel[%v] SEND heartbeat: %d", hub.tunnel, id)
	} else {
		Info("link[%d] SEND cmd: %d", id, Cmd)
	}

	return hub.Send(0, buf.Bytes())
}
