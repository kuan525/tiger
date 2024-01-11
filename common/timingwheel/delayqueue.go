package timingwheel

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"
)

type DelayQueue struct {
	C  chan interface{}
	mu sync.Mutex
	pq priorityQueue

	sleeping int32
	wakeupC  chan struct{}
}

func NewDelayQueue(size int) *DelayQueue {
	return &DelayQueue{
		C:       make(chan interface{}),
		pq:      newPriorityQueue(size),
		wakeupC: make(chan struct{}),
	}
}

func (dq *DelayQueue) Push(elem interface{}, expiration int64) {
	item := &item{
		Value:    elem,
		Priority: expiration,
	}
	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	index := item.Index
	dq.mu.Unlock()

	if index == 0 {
		// 当前添加的bucket的expiration最早，需要唤醒
		if atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
			dq.wakeupC <- struct{}{}
		}
	}
}

func (dq *DelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
	for {
		now := nowF() // 用于获取当前状态的函数，比如时间，或者当前分数
		dq.mu.Lock()
		item, delta := dq.pq.PeekAndShift(now)
		if item == nil { // 队列为空，或者最早的timer时间未到
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()

		if item == nil { // 队列为空，或者最早的timer时间未到
			if delta == 0 { // 队列为空
				select {
				case <-dq.wakeupC: // 等待被唤醒
					continue
				case <-exitC: // 等待退出指令，这里是优雅关闭
					goto exit
				}
			} else if delta > 0 { // 最早的timer时间未到
				select {
				case <-dq.wakeupC: // 等待被唤醒
					continue
				case <-time.After(time.Duration(delta) * time.Millisecond): // 等待堆顶
					// 满足的情况只有一种：
					// 此时通过push来了一个更早expiration的bucket，sleeping已经是0了，调用者就被阻塞
					// 所以这里需要防止调用者被阻塞
					if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
						<-dq.wakeupC
					}
					continue
				case <-exitC: // 优雅退出
					goto exit
				}
			}
		}

		// 只有获取到item才会到这个地方
		select {
		case dq.C <- item.Value: // 到期的value放入dp.C中
		case <-exitC: // 优雅关闭
			goto exit
		}
	}

exit:
	atomic.StoreInt32(&dq.sleeping, 0)
}
