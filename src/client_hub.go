package src

import "time"

//TClientHub manages client links
type TClientHub struct {
	*Hub
	sent uint16
	rcvd uint16
}

func (chub *TClientHub) heartbeat() {
	heartbeat := getHeartbeat()
	timeout := getTimeout()
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	maxSpan := int(timeout / heartbeat)
	if maxSpan <= tunnelMinSpan {
		maxSpan = tunnelMinSpan
	}
	Debug("max span is: %d.", maxSpan)

	for range ticker.C {
		// id overflow
		span := int(^uint16(0) + chub.sent - chub.rcvd + 1)
		if span >= maxSpan {
			Error("tunnel[%v] timeout, sent: %d, rcvd: %d", chub.Hub.tunnel, chub.sent, chub.rcvd)
			chub.Hub.Close()
			break
		}

		chub.sent++
		if !chub.sendCmd(chub.sent, HEARTBEAT) {
			break
		}
	}
}

func (chub *TClientHub) onCtrl(cmd TCmd) bool {
	id := cmd.ID
	switch cmd.Cmd {
	case HEARTBEAT:
		chub.rcvd = id
		return true
	}
	return false
}

func newClientHub(tunnel *Tunnel) *TClientHub {
	chub := &TClientHub{
		Hub: newHub(tunnel),
	}
	chub.Hub.onCtrlFilter = chub.onCtrl
	go chub.heartbeat()
	return chub
}
