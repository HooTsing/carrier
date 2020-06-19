package src

//IDAllocator id allocator
type IDAllocator struct {
	freeList chan uint16
}

//Acquire acquire id
func (alloc *IDAllocator) Acquire() uint16 {
	return <-alloc.freeList
}

//Release release id
func (alloc *IDAllocator) Release(id uint16) {
	alloc.freeList <- id
}

func newAllocator() *IDAllocator {
	freeList := make(chan uint16, TunnelMaxID)

	var id uint16
	for id = 1; id < TunnelMaxID; id++ {
		freeList <- id
	}

	allocator := &IDAllocator{
		freeList: freeList,
	}
	return allocator
}
