package domain

type stateWindow struct {
	stateQueue []*Stat
	statChan   chan *Stat
	sumStat    *Stat
	idx        int64
}

const windowSize = 5

func newStateWindow() *stateWindow {
	return &stateWindow{
		stateQueue: make([]*Stat, windowSize),
		statChan:   make(chan *Stat),
		sumStat:    &Stat{},
	}
}

func (sw *stateWindow) getStat() *Stat {
	res := sw.sumStat.Clone()
	res.Avg(windowSize)
	return res
}

// 更新stateWindow
func (sw *stateWindow) appendStat(s *Stat) {
	sw.sumStat.Sub(sw.stateQueue[sw.idx%windowSize]) // 没有则不减
	sw.stateQueue[sw.idx%windowSize] = s             // 环形数组，更新
	sw.sumStat.Add(s)                                // 加入
	sw.idx++                                         // 环形指针
}
