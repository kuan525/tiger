package timingwheel

import (
	"sync"
	"time"
)

// 将x向下（模m）取整
func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}

// 标准时间：毫秒时间戳
func timeToMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// 毫秒时间戳：标准时间，这里会有精度损失，但是不重要
func msToTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond)).UTC()
}

// 这是一个并发流封装器，保证父协程在子协程执行完后才退出
type waitGroupWrapper struct {
	sync.WaitGroup
}

func (w *waitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
