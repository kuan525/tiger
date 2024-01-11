package timingwheel

import (
	"container/list"
	"sync"
	"sync/atomic"
)

type bucket struct {
	expiration int64      //当前桶的过期时间
	mu         sync.Mutex // 锁，保护临界区
	timers     *list.List // 存放timer的链表
}

func newBucket() *bucket {
	return &bucket{
		timers:     list.New(),
		expiration: -1,
	}
}

// 获取过期时间
func (b *bucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

// 更新过期时间，并查看原来和当前是否相等
func (b *bucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

// 当前bucket中加入一个timer
func (b *bucket) Add(t *Timer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e := b.timers.PushBack(t)

	// 以下两个是反向指针
	t.setBucket(b)
	t.element = e
}

// 类比一个底层操作，这里会被多次调用，Flush会调用，又不能多次加锁
func (b *bucket) remove(t *Timer) bool {
	if t.getBucket() != b {
		return false
	}
	b.timers.Remove(t.element)
	t.setBucket(nil)
	t.element = nil
	return true
}

func (b *bucket) Remove(t *Timer) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.remove(t)
}

// 时间轮的转动 reinsert: tw.addOrRun
func (b *bucket) Flush(reinsert func(*Timer)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for e := b.timers.Front(); e != nil; {
		next := e.Next()
		t := e.Value.(*Timer) // 得到timer
		b.remove(t)           // 移除timer
		reinsert(t)           // 重新处理：tw.addOrRun，将每个timer情况监控一遍，包括timer的移动
		e = next
	}

	b.SetExpiration(-1) // 当前轮废弃
}
