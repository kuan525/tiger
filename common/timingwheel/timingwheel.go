package timingwheel

import (
	"errors"
	"sync/atomic"
	"time"
	"unsafe"
)

type TimingWheel struct {
	tick          int64 // 基本时间单位（单位ms）可以理解为精度
	wheelSize     int64 // 定时轮中槽的数量
	interval      int64 // 定时轮每次转动的时间间隔，单位是tick（ms） tickMs * wheelSize
	currentTime   int64 // 当前轮创建的时间
	buckets       []*bucket
	queue         *DelayQueue
	overflowWheel unsafe.Pointer // 超出当前时间轮的时间跨度，放入溢出轮
	exitC         chan struct{}
	waitGroup     waitGroupWrapper
}

func newTimingWheel(tickMs, wheelSize, startMs int64, queue *DelayQueue) *TimingWheel {
	buckets := make([]*bucket, wheelSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	return &TimingWheel{
		tick:        tickMs,
		wheelSize:   wheelSize,
		currentTime: truncate(startMs, tickMs),
		interval:    tickMs * wheelSize,
		buckets:     buckets,
		queue:       queue,
		exitC:       make(chan struct{}),
	}
}

// 精度：tick ｜ wheelSize：槽数
func NewTimingWheel(tick time.Duration, wheelSize int64) *TimingWheel {
	tickMs := int64(tick / time.Millisecond)
	if tickMs <= 0 {
		panic(errors.New("tick must be greater than or equal to 1ms"))
	}
	// 取当前时刻的毫秒数，会损失ns精度
	startMs := timeToMs(time.Now().UTC())
	return newTimingWheel(tickMs, wheelSize, startMs, NewDelayQueue(int(wheelSize)))
}

func (tw *TimingWheel) add(t *Timer) bool {
	currentTime := atomic.LoadInt64(&tw.currentTime)

	if t.expiration < currentTime+tw.tick { // 即将运行
		return false
	} else if t.expiration < currentTime+tw.interval { // 在当前轮内
		virtualID := t.expiration / tw.tick // 找到当前轮中对应的槽
		b := tw.buckets[virtualID%tw.wheelSize]
		b.Add(t)

		// 如果到期时间不一致，则再次进去一下延迟队列
		if b.SetExpiration(virtualID * tw.tick) {
			tw.queue.Push(b, b.Expiration())
		}
		return true
	} else { // 不在当前轮，则初始化紧跟着的溢出轮
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				// tickMs : interval 一轮更比一轮大，刻度变大
				unsafe.Pointer(newTimingWheel(tw.interval, tw.wheelSize, currentTime, tw.queue)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		// 进行一个递归，放入溢出轮
		return (*TimingWheel)(overflowWheel).add(t)
	}
}

func (tw *TimingWheel) addOrRun(t *Timer) {
	if !tw.add(t) { // 只有当过期时才立即运行，省去了加入时间轮的开销
		go t.task()
	}
}

// 时间轮转动
func (tw *TimingWheel) advanceClock(expiration int64) {
	currentTime := atomic.LoadInt64(&tw.currentTime)

	if expiration >= currentTime+tw.tick {
		currentTime = truncate(expiration, tw.tick)
		// 这里是，跳过前面的空槽
		atomic.StoreInt64(&tw.currentTime, currentTime)

		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*TimingWheel)(overflowWheel).advanceClock(currentTime)
		}
	}
}

func (tw *TimingWheel) Start() {
	// 开启延迟队列
	tw.waitGroup.Wrap(func() {
		tw.queue.Poll(tw.exitC, func() int64 {
			return timeToMs(time.Now().UTC())
		})
	})
	// 保持时间轮在不断转动
	tw.waitGroup.Wrap(func() {
		for {
			select {
			case elem := <-tw.queue.C:
				b := elem.(*bucket)
				tw.advanceClock(b.Expiration()) // bucket到达时间，转动时间轮
				b.Flush(tw.addOrRun)            // 将bucket中的timer执行
			case <-tw.exitC:
				return
			}
		}
	})
}

func (tw *TimingWheel) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

// 通过这个函数来指定定时时间和执行函数
func (tw *TimingWheel) AfterFunc(d time.Duration, f func()) *Timer {
	// 声明一个定时器对象
	t := &Timer{
		expiration: timeToMs(time.Now().UTC().Add(d)), // 存储的是具体的时间ms
		task:       f,
	}

	tw.addOrRun(t)
	return t
}

type Scheduler interface {
	Next(time.Time) time.Time
}

func (tw *TimingWheel) ScheduleFunc(s Scheduler, f func()) (t *Timer) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		return
	}
	t = &Timer{
		expiration: timeToMs(expiration),
		task: func() {
			expiration := s.Next(msToTime(t.expiration))
			if !expiration.IsZero() {
				t.expiration = timeToMs(expiration)
				tw.addOrRun(t)
			}
			f()
		},
	}
	tw.addOrRun(t)
	return
}
