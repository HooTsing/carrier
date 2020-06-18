package src

//ClientHubItem client hub item
type ClientHubItem struct {
	*TClientHub
	priority int // current link count
	index    int // index in the heap
}

//Status cli hub item status
func (chi *ClientHubItem) Status() {
	chi.Hub.Status()
	Log("priority: %d, index: %d", chi.priority, chi.index)
}

//ClientHubQueue client_hub queue
type ClientHubQueue []*ClientHubItem

//Len chq len
func (chq ClientHubQueue) Len() int {
	return len(chq)
}

func (chq ClientHubQueue) Less(i, j int) bool {
	return chq[i].priority < chq[j].priority
}

func (chq ClientHubQueue) Swap(i, j int) {
	chq[i], chq[j] = chq[j], chq[i]
	chq[i].index = i
	chq[j].index = j
}

//Push chq push
func (chq *ClientHubQueue) Push(x interface{}) {
	n := len(*chq)
	hub := x.(*ClientHubItem)
	hub.index = n
	*chq = append(*chq, hub)
}

//Pop chq pop
func (chq *ClientHubQueue) Pop() interface{} {
	old := *chq
	n := len(old)
	hub := old[n-1]
	hub.index = -1
	*chq = old[:n-1]
	return hub
}
