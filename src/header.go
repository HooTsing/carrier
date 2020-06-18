package src

// tunnel packet header
//	a tunnel packet consists of a  header and a body
//	Len is the length of subsequent packet body
type THeader struct {
	linkID  uint16
	bodyLen uint16 //连着的body长度
}
