package src

import "encoding/binary"

//TAuthToken autj token
type TAuthToken struct {
	challange uint64
	timestamp uint64
}

func (at *TAuthToken) fromBytes(buf []byte) {
	at.challange = binary.LittleEndian.Uint64(buf)
	at.timestamp = binary.LittleEndian.Uint64(buf[8:])
}

func (at *TAuthToken) toBytes() []byte {
	buf := make([]byte, TaaTokenSize)
	binary.LittleEndian.PutUint64(buf, at.challange)
	binary.LittleEndian.PutUint64(buf[8:], at.timestamp)
	return buf
}

func (at TAuthToken) complement() TAuthToken {
	return TAuthToken{
		challange: ^at.challange,
		timestamp: ^at.timestamp,
	}
}

func (at *TAuthToken) isComplemenary(t TAuthToken) bool {
	if at.challange != ^t.challange || at.timestamp != ^t.timestamp {
		return false
	}
	return true
}
