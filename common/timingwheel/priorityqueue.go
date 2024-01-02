package timingwheel

import "container/heap"

type item struct {
	Value    interface{}
	Priority int64
	Index    int
}

type priorityQueue []*item

func newPriorityQueue(capacity int) priorityQueue {
	return make(priorityQueue, 0, capacity)
}

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = j
	pq[j].Index = i
}

func (pq *priorityQueue) Push(x interface{}) {
	n, c := len(*pq), cap(*pq)
	if n+1 > c {
		npq := make(priorityQueue, n, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : n+1]
	newItem := x.(*item)
	newItem.Index = n
	(*pq)[n] = newItem
}

func (pq *priorityQueue) Pop() interface{} {
	n, c := len(*pq), cap(*pq)
	if n < (c/2) && c > 25 {
		npq := make(priorityQueue, n, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	oldItem := (*pq)[n-1]
	oldItem.Index = -1 // 安全
	*pq = (*pq)[0 : n-1]
	return oldItem
}

// 当且仅当有元素，且堆顶元素的Priority小于等于mx，才返回item
func (pq *priorityQueue) PeekAndShift(mx int64) (*item, int64) {
	if pq.Len() == 0 {
		return nil, 0
	}
	item := (*pq)[0]
	if item.Priority > mx {
		return nil, item.Priority - mx
	}
	heap.Remove(pq, 0)
	return item, 0
}
