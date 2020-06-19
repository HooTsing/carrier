package src

//THeader tunnel packet header
//	a tunnel packet consists of a  header and a body
//	Len is the length of subsequent packet body
// https://blog.csdn.net/Xiayan_ucas/article/details/80367812 大写为公有属性 外界可访问
type THeader struct {
	LinkID  uint16
	BodyLen uint16 //连着的body长度
}
