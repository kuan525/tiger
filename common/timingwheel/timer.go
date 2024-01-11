package timingwheel

import (
	"container/list"
	"sync/atomic"
	"unsafe"
)

// 计时器
type Timer struct {
	expiration int64          // 过期时间
	task       func()         // 执行函数
	b          unsafe.Pointer // 所在的桶
	element    *list.Element  // 所在桶中的链表中的位置
}

// get
func (t *Timer) getBucket() *bucket {
	return (*bucket)(atomic.LoadPointer(&t.b))
}

// set
func (t *Timer) setBucket(b *bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// stop
func (t *Timer) Stop() bool {
	stopped := false
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		// timer中存储对应的t，当前timer可能在移动过程中，所以这里要反复获取，直到将timer中的bucket置为nil
		stopped = b.Remove(t)
	}
	return stopped
}
