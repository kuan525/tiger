package timingwheel

import (
	"fmt"
	"testing"
	"time"
)

func TestPoll(t *testing.T) {
	deq := NewDelayQueue(8)
	exitC := make(chan struct{})
	nowF := func() int64 {
		return timeToMs(time.Now().UTC())
	}

	go deq.Poll(exitC, nowF)
	go func() {
		for elem := range deq.C {
			arr := elem.(string)
			fmt.Println(arr)
		}
	}()

	deq.Push("kuan-ddd", nowF())

	exitC <- struct{}{}
	time.Sleep(time.Second)
}
