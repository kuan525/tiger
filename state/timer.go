package state

import (
	"time"

	"github.com/kuan525/tiger/common/timingwheel"
)

var wheel *timingwheel.TimingWheel

func InitTimer() {
	wheel = timingwheel.NewTimingWheel(time.Millisecond, 20)
	wheel.Start()
}

func CloseTimer() {
	wheel.Stop()
}

func AfterFunc(d time.Duration, f func()) *timingwheel.Timer {
	t := wheel.AfterFunc(d, f)
	return t
}
