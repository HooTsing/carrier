package src

import "sync"

//TMsgPool message pool
type TMsgPool struct {
	*sync.Pool
	sz int
}

//Get impl sync.Pool Get() method
func (mp *TMsgPool) Get() []byte {
	return mp.Pool.Get().([]byte)
}

//Put impl sync.Pool Put() method
func (mp *TMsgPool) Put(data []byte) {
	if cap(data) == mp.sz {
		mp.Pool.Put(data[0:mp.sz])
	}
}

//NewMsgPool new msg_pool
func NewMsgPool(size int) *TMsgPool {
	p := &TMsgPool{
		sz: size,
	}
	p.Pool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, p.sz)
		},
	}
	return p
}
