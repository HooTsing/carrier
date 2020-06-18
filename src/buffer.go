package src

import "sync"

type TBuffer struct {
	start  int
	end    int
	buf    [][]byte
	cond   *sync.Cond //buffer notify
	closed bool
}

func (b *TBuffer) bufferLen() int {
	return (b.bufferCap() + b.end - b.start) % b.bufferCap()
}

func (b *TBuffer) bufferCap() int {
	return cap(b.buf)
}

func (b *TBuffer) Close() bool {
	defer b.cond.L.Unlock()
	b.cond.L.Lock()

	if b.closed {
		return false
	}

	b.closed = true
	b.cond.Broadcast()
	return true
}

func (b *TBuffer) Put(data []byte) bool {
	defer b.cond.L.Unlock()
	b.cond.L.Lock()

	if b.closed {
		return false
	}

	oldBufCap := b.bufferCap()
	if (b.end+1)%oldBufCap == b.start {
		buf := make([][]byte, cap(b.buf)*2) //扩大两倍
		if b.end > b.start {
			copy(buf, b.buf[b.start:b.end])
		} else if b.end < b.start {
			copy(buf, b.buf[b.start:oldBufCap])
			copy(buf[oldBufCap-b.start:], b.buf[:b.end])
		}
		b.buf = buf
		b.start = 0
		b.end = oldBufCap - 1
	}

	b.buf[b.end] = data
	b.end = (b.end + 1) % b.bufferCap()
	b.cond.Signal()
	return true
}

func (b *TBuffer) Pop() ([]byte, bool) {
	for {
		b.cond.L.Lock()
		if b.bufferLen() > 0 {
			data := b.buf[b.start]
			b.start = (b.start + 1) % b.bufferCap()
			b.cond.L.Unlock()
			return data, true
		}
		if b.closed {
			b.cond.L.Unlock()
			return nil, false
		}
		b.cond.Wait()
		b.cond.L.Unlock()
	}
}

func newBuffer(sz int) *TBuffer {
	var mutex sync.Mutex
	buf := &TBuffer{
		start: 0,
		end:   0,
		buf:   make([][]byte, sz),
		cond:  sync.NewCond(&mutex),
	}
	return buf
}
